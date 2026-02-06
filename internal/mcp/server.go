package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/shared"
	"mochi-sticky/internal/storage"
	"mochi-sticky/internal/wiki"
)

const (
	serverName    = "mochi-sticky"
	serverVersion = "0.1.0"
)

const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
	codeDenied         = -32000
)

// Server orchestrates JSON-RPC 2.0 calls forwarded over stdin/stdout.
// It wires the MCP tools to the board and wiki domain packages.
type Server struct {
	baseDir     string
	storageRoot string
	shutdown    bool
	mu          sync.RWMutex
}

// NewServer creates an MCP server tied to a workspace root and storage root.
func NewServer(baseDir, storageRoot string) (*Server, error) {
	if strings.TrimSpace(baseDir) == "" {
		return nil, fmt.Errorf("mcp: base directory is required")
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("mcp: failed to resolve base directory: %w", err)
	}
	if strings.TrimSpace(storageRoot) == "" {
		storageRoot = filepath.Join(absBase, ".sticky")
	} else if !filepath.IsAbs(storageRoot) {
		storageRoot = filepath.Join(absBase, storageRoot)
	}
	storageRoot, err = filepath.Abs(storageRoot)
	if err != nil {
		return nil, fmt.Errorf("mcp: failed to resolve storage root: %w", err)
	}
	return &Server{baseDir: absBase, storageRoot: storageRoot}, nil
}

func (s *Server) templatePaths() (storage.TemplatePaths, error) {
	cfg, err := storage.LoadConfigFromRoot(s.storageRoot)
	if err != nil {
		return storage.TemplatePaths{}, err
	}
	return storage.ResolveTemplates(s.baseDir, s.storageRoot, cfg)
}

func slugForPage(root, templatesRoot string, page wiki.Page) string {
	slug := strings.TrimSpace(page.Slug)
	if slug != "" {
		return slug
	}
	if page.FilePath == "" {
		return ""
	}
	if templatesRoot != "" && shared.IsSubpath(templatesRoot, page.FilePath) {
		return wiki.SlugFromPath(templatesRoot, page.FilePath)
	}
	return wiki.SlugFromPath(root, page.FilePath)
}

// Serve processes incoming JSON-RPC requests from in and writes responses to out.
// The loop keeps running until the input stream closes (EOF).
func (s *Server) Serve(in io.Reader, out io.Writer) error {
	return s.ServeContext(context.Background(), in, out)
}

// ServeContext processes incoming JSON-RPC requests and stops when ctx is canceled or EOF is reached.
func (s *Server) ServeContext(ctx context.Context, in io.Reader, out io.Writer) error {
	return s.ServeContextWithTimeout(ctx, in, out, 0)
}

// ServeContextWithTimeout processes incoming JSON-RPC requests with an optional idle timeout.
// If timeout is 0, no timeout is applied.
func (s *Server) ServeContextWithTimeout(ctx context.Context, in io.Reader, out io.Writer, timeout time.Duration) error {
	bufOut := bufio.NewWriter(out)
	decoder := json.NewDecoder(bufio.NewReader(in))
	encoder := json.NewEncoder(bufOut)
	encoder.SetEscapeHTML(false)

	for {
		// Check for shutdown or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		s.mu.RLock()
		shuttingDown := s.shutdown
		s.mu.RUnlock()
		if shuttingDown {
			return nil
		}

		// Set read deadline if timeout is specified
		if timeout > 0 {
			if reader, ok := in.(interface{ SetReadDeadline(time.Time) error }); ok {
				_ = reader.SetReadDeadline(time.Now().Add(timeout))
			}
		}

		var req rpcRequest
		if err := decoder.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			// For parse errors, send error response and continue (don't exit)
			s.writeErrorBuf(encoder, bufOut, nil, codeParseError, "parse error")
			continue
		}
		if req.JSONRPC != "2.0" || strings.TrimSpace(req.Method) == "" {
			s.writeErrorBuf(encoder, bufOut, req.ID, codeInvalidRequest, "invalid request")
			continue
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		result, rpcErr := s.dispatch(ctx, req)
		if isNotification(req.ID) {
			continue
		}
		if rpcErr != nil {
			s.writeErrorBuf(encoder, bufOut, req.ID, rpcErr.Code, rpcErr.Message)
			continue
		}
		resp := rpcResponse{JSONRPC: "2.0", Result: result, ID: req.ID}
		_ = encoder.Encode(resp)
		_ = bufOut.Flush()
	}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type listTasksParams struct {
	BoardID string   `json:"board_id"`
	Status  string   `json:"status"`
	Title   string   `json:"title"`
	Tags    []string `json:"tags"`
	TagMode string   `json:"tag_mode"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	Sort    string   `json:"sort"`
	Desc    bool     `json:"desc"`
}

type getTaskParams struct {
	BoardID string `json:"board_id"`
	ID      string `json:"id"`
}

type createTaskParams struct {
	BoardID  string   `json:"board_id"`
	Title    string   `json:"title"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags"`
	Priority int      `json:"priority"`
}

type listWikiParams struct {
	Status           string   `json:"status"`
	IncludeTemplates bool     `json:"include_templates"`
	Title            string   `json:"title"`
	Tags             []string `json:"tags"`
	TagMode          string   `json:"tag_mode"`
	Section          string   `json:"section"`
	Query            string   `json:"query"`
	CaseInsensitive  *bool    `json:"case_insensitive"`
}

type initializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Roots struct {
			ListChanged bool `json:"listChanged"`
		} `json:"roots"`
	} `json:"capabilities"`
	ClientInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
	Roots []struct {
		URI  string `json:"uri"`
		Name string `json:"name"`
	} `json:"roots"`
}

type listWikiSectionsParams struct {
	Tags       []string `json:"tags"`
	TagMode    string   `json:"tag_mode"`
	LinkType   string   `json:"link_type"`
	LinkTarget string   `json:"link_target"`
}

type updateWikiSectionParams struct {
	Slug  string             `json:"slug"`
	Title *string            `json:"title"`
	Order *int               `json:"order"`
	Tags  *[]string          `json:"tags"`
	Links *wiki.SectionLinks `json:"links"`
}

type readWikiParams struct {
	Slug string `json:"slug"`
}

type writeWikiParams struct {
	Slug    string    `json:"slug"`
	Title   *string   `json:"title"`
	Section *string   `json:"section"`
	Order   *int      `json:"order"`
	Tags    *[]string `json:"tags"`
	Status  *string   `json:"status"`
	Content *string   `json:"content"`
}

type searchWikiParams struct {
	Query            string `json:"query"`
	Status           string `json:"status"`
	IncludeTemplates bool   `json:"include_templates"`
	CaseInsensitive  *bool  `json:"case_insensitive"`
}

type listWikiTemplatesParams struct{}

type createWikiFromTemplateParams struct {
	Slug     string    `json:"slug"`
	Title    string    `json:"title"`
	Template string    `json:"template"`
	Section  *string   `json:"section"`
	Order    *int      `json:"order"`
	Tags     *[]string `json:"tags"`
	Status   *string   `json:"status"`
}

type lintWikiParams struct{}

type manifestWikiParams struct{}

type exportWikiParams struct {
	Format                string   `json:"format"`
	Output                string   `json:"output"`
	Roots                 []string `json:"roots"`
	Prefix                string   `json:"prefix"`
	Title                 string   `json:"title"`
	Author                string   `json:"author"`
	Template              string   `json:"template"`
	Page                  string   `json:"page"`
	Section               string   `json:"section"`
	FilterTitle           string   `json:"filter_title"`
	FilterTags            []string `json:"filter_tags"`
	FilterTagMode         string   `json:"filter_tag_mode"`
	FilterSection         string   `json:"filter_section"`
	FilterQuery           string   `json:"filter_query"`
	FilterCaseInsensitive *bool    `json:"filter_case_insensitive"`
	IncludeLinked         bool     `json:"include_linked"`
	LinkTypes             []string `json:"link_types"`
}

type deleteWikiParams struct {
	Slug        string `json:"slug"`
	UpdateIndex bool   `json:"update_index"`
}

type generateWikiIndexParams struct {
	IncludeTemplates bool   `json:"include_templates"`
	Write            bool   `json:"write"`
	Output           string `json:"output"`
}

type updateStatusParams struct {
	BoardID string `json:"board_id"`
	ID      string `json:"id"`
	Status  string `json:"status"`
}

type updatePriorityParams struct {
	BoardID  string `json:"board_id"`
	ID       string `json:"id"`
	Priority int    `json:"priority"`
}

type updateTitleParams struct {
	BoardID string `json:"board_id"`
	ID      string `json:"id"`
	Title   string `json:"title"`
}

type updateTagsParams struct {
	BoardID string   `json:"board_id"`
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
}

type updateContentParams struct {
	BoardID string `json:"board_id"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

type updateDepsParams struct {
	BoardID string   `json:"board_id"`
	ID      string   `json:"id"`
	Depends []string `json:"depends_on"`
}

type taskIDParams struct {
	BoardID string `json:"board_id"`
	ID      string `json:"id"`
	Force   bool   `json:"force"`
}

type boardIDParams struct {
	ID    string `json:"id"`
	Force bool   `json:"force"`
}

type boardNameParams struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type boardCreateParams struct {
	Name string `json:"name"`
}

type boardReadParams struct {
	ID string `json:"id"`
}

type updateBoardDescriptionParams struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type boardContextParams struct {
	BoardID string             `json:"board_id"`
	Context board.BoardContext `json:"context"`
}

type taskSummary struct {
	BoardID   string   `json:"board_id,omitempty"`
	BoardName string   `json:"board_name,omitempty"`
	ID        string   `json:"id"`
	UID       string   `json:"uid,omitempty"`
	Title     string   `json:"title"`
	Status    string   `json:"status"`
	Priority  int      `json:"priority"`
	Tags      []string `json:"tags,omitempty"`
	Created   string   `json:"created,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`
}

type taskDetail struct {
	taskSummary
	Content string `json:"content,omitempty"`
}

type boardSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Archived bool   `json:"archived"`
	Created  string `json:"created,omitempty"`
}

type toolDescriptor struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema,omitempty"`
}

type resourceDescriptor struct {
	Name        string `json:"name"`
	URI         string `json:"uri"`
	Description string `json:"description"`
}

func (s *Server) dispatch(ctx context.Context, req rpcRequest) (any, *rpcError) {
	select {
	case <-ctx.Done():
		return nil, &rpcError{Code: codeInternalError, Message: ctx.Err().Error()}
	default:
	}
	switch req.Method {
	case "initialize", "handshake":
		var params initializeParams
		if err := decodeParams(req.Params, &params); err == nil && len(params.Roots) > 0 {
			// Use first root as workspace directory
			if rootURI := params.Roots[0].URI; strings.HasPrefix(rootURI, "file://") {
				rootPath := strings.TrimPrefix(rootURI, "file://")
				if absPath, err := filepath.Abs(rootPath); err == nil {
					s.mu.Lock()
					s.baseDir = absPath
					// Update storageRoot relative to new baseDir
					s.storageRoot = filepath.Join(absPath, ".sticky")
					s.mu.Unlock()
				}
			}
		}
		return s.serverInfo(), nil
	case "shutdown":
		s.mu.Lock()
		s.shutdown = true
		s.mu.Unlock()
		return map[string]any{"success": true}, nil
	case "exit":
		// Exit notification - set shutdown flag (server loop will exit)
		s.mu.Lock()
		s.shutdown = true
		s.mu.Unlock()
		return nil, nil
	case "tools/list", "list_tools":
		return map[string]any{"tools": toolList()}, nil
	case "resources/list", "list_resources":
		return map[string]any{"resources": resourceList()}, nil
	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.readResource(ctx, params.URI)
	case "tools/call":
		// Standard MCP tools/call - extract tool name and arguments
		var callParams struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := decodeParams(req.Params, &callParams); err != nil {
			return nil, invalidParams(err)
		}
		// Convert arguments to JSON and dispatch to the tool
		argsJSON, err := json.Marshal(callParams.Arguments)
		if err != nil {
			return nil, invalidParams(err)
		}
		toolReq := rpcRequest{
			JSONRPC: "2.0",
			Method:  callParams.Name,
			Params:  argsJSON,
			ID:      req.ID,
		}
		return s.dispatch(ctx, toolReq)
	case "list_tasks":
		var params listTasksParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.listTasks(ctx, params)
	case "get_task":
		var params getTaskParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.getTask(params)
	case "create_task":
		var params createTaskParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.createTask(ctx, params)
	case "update_task_status":
		var params updateStatusParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateTaskStatus(ctx, params)
	case "update_task_priority":
		var params updatePriorityParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateTaskPriority(ctx, params)
	case "update_task_title":
		var params updateTitleParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateTaskTitle(ctx, params)
	case "update_task_tags":
		var params updateTagsParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateTaskTags(ctx, params)
	case "update_task_content":
		var params updateContentParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateTaskContent(ctx, params)
	case "update_task_dependencies":
		var params updateDepsParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateTaskDependencies(ctx, params)
	case "get_task_dependencies":
		var params getTaskParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.getTaskDependencies(params)
	case "list_ready_tasks":
		var params listTasksParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.listReadyTasks(ctx, params)
	case "archive_task":
		var params taskIDParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.archiveTask(ctx, params)
	case "restore_task":
		var params taskIDParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.restoreTask(ctx, params)
	case "delete_task":
		var params taskIDParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.deleteTask(ctx, params)
	case "list_archived_tasks":
		var params listTasksParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.listArchivedTasks(ctx, params)
	case "list_boards":
		return s.listBoards(ctx)
	case "create_board":
		var params boardCreateParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.createBoard(ctx, params)
	case "rename_board":
		var params boardNameParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.renameBoard(ctx, params)
	case "set_active_board":
		var params boardIDParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.setActiveBoard(ctx, params)
	case "archive_board":
		var params boardIDParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.archiveBoard(ctx, params)
	case "delete_board":
		var params boardIDParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.deleteBoard(ctx, params)
	case "update_board_description":
		var params updateBoardDescriptionParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateBoardDescription(ctx, params)
	case "update_board_context":
		var params boardContextParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateBoardContext(ctx, params)
	case "get_board_context":
		var params boardReadParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.getBoardContext(ctx, params)
	case "read_task":
		var params getTaskParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.readTask(params)
	case "read_board":
		var params boardReadParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.readBoard(ctx, params)
	case "list_wiki_pages":
		var params listWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.listWikiPages(ctx, params)
	case "list_wiki_sections":
		var params listWikiSectionsParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.listWikiSections(ctx, params)
	case "read_wiki_page":
		var params readWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.readWikiPage(params)
	case "write_wiki_page":
		var params writeWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.writeWikiPage(params)
	case "update_wiki_section":
		var params updateWikiSectionParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.updateWikiSection(ctx, params)
	case "search_wiki":
		var params searchWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.searchWiki(ctx, params)
	case "list_wiki_templates":
		var params listWikiTemplatesParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.listWikiTemplates(params)
	case "create_wiki_from_template":
		var params createWikiFromTemplateParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.createWikiFromTemplate(params)
	case "lint_wiki":
		var params lintWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.lintWiki(ctx, params)
	case "manifest_wiki":
		var params manifestWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.manifestWiki(ctx, params)
	case "export_wiki":
		var params exportWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.exportWiki(ctx, params)
	case "delete_wiki_page":
		var params deleteWikiParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.deleteWikiPage(ctx, params)
	case "generate_wiki_index":
		var params generateWikiIndexParams
		if err := decodeParams(req.Params, &params); err != nil {
			return nil, invalidParams(err)
		}
		return s.generateWikiIndex(ctx, params)
	case "read_config":
		return s.readConfig(ctx)
	case "read_boards":
		return s.readBoards(ctx)
	default:
		return nil, &rpcError{Code: codeMethodNotFound, Message: "method not found"}
	}
}

func (s *Server) serverInfo() map[string]any {
	return map[string]any{
		"name":    serverName,
		"version": serverVersion,
		"capabilities": map[string]any{
			"tools":     toolNames(),
			"resources": resourceNames(),
		},
	}
}

func toolNames() []string {
	tools := toolList()
	result := make([]string, 0, len(tools))
	for _, tool := range tools {
		result = append(result, tool.Name)
	}
	return result
}

func resourceNames() []string {
	resources := resourceList()
	result := make([]string, 0, len(resources))
	for _, resource := range resources {
		result = append(result, resource.Name)
	}
	return result
}

func toolList() []toolDescriptor {
	return []toolDescriptor{
		{
			Name:        "list_tasks",
			Description: "List tasks with optional filters",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"board_id": map[string]any{"type": "string", "description": "Board ID (optional)"},
					"status":   map[string]any{"type": "string", "description": "Filter by status"},
					"sort":     map[string]any{"type": "string", "description": "Sort field"},
				},
			},
		},
		{Name: "get_task", Description: "Get task details by id", InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string", "description": "Task ID"},
			},
			"required": []string{"id"},
		}},
		{Name: "create_task", Description: "Create a new task", InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title":  map[string]any{"type": "string", "description": "Task title"},
				"status": map[string]any{"type": "string", "description": "Task status"},
			},
			"required": []string{"title"},
		}},
		{Name: "update_task_status", Description: "Update a task status"},
		{Name: "update_task_priority", Description: "Update a task priority"},
		{Name: "update_task_title", Description: "Update a task title"},
		{Name: "update_task_tags", Description: "Update task tags"},
		{Name: "update_task_content", Description: "Update task content"},
		{Name: "update_task_dependencies", Description: "Set task dependencies"},
		{Name: "get_task_dependencies", Description: "Get task dependency list"},
		{Name: "list_ready_tasks", Description: "List tasks whose dependencies are satisfied"},
		{Name: "archive_task", Description: "Archive a task (requires force)"},
		{Name: "restore_task", Description: "Restore an archived task"},
		{Name: "delete_task", Description: "Delete a task (requires force)"},
		{Name: "list_archived_tasks", Description: "List archived tasks"},
		{Name: "list_boards", Description: "List boards"},
		{Name: "create_board", Description: "Create a new board"},
		{Name: "rename_board", Description: "Rename a board"},
		{Name: "set_active_board", Description: "Set active board"},
		{Name: "archive_board", Description: "Archive a board (requires force)"},
		{Name: "delete_board", Description: "Delete a board (requires force)"},
		{Name: "update_board_description", Description: "Update board description markdown"},
		{Name: "update_board_context", Description: "Update board context metadata"},
		{Name: "get_board_context", Description: "Read the board context block"},
		{Name: "list_wiki_pages", Description: "List wiki pages"},
		{Name: "list_wiki_sections", Description: "List wiki sections"},
		{Name: "read_wiki_page", Description: "Read a wiki page"},
		{Name: "write_wiki_page", Description: "Create or update a wiki page"},
		{Name: "update_wiki_section", Description: "Update wiki section metadata"},
		{Name: "search_wiki", Description: "Search wiki pages"},
		{Name: "list_wiki_templates", Description: "List wiki templates"},
		{Name: "create_wiki_from_template", Description: "Create a wiki page from a template"},
		{Name: "lint_wiki", Description: "Lint wiki pages"},
		{Name: "manifest_wiki", Description: "Generate a wiki manifest"},
		{Name: "export_wiki", Description: "Export wiki content (md or pdf)"},
		{Name: "delete_wiki_page", Description: "Delete a wiki page"},
		{Name: "generate_wiki_index", Description: "Generate wiki index from pages"},
	}
}

func resourceList() []resourceDescriptor {
	return []resourceDescriptor{
		{Name: "read_task", URI: "task://<id>", Description: "Read full task markdown"},
		{Name: "read_board", URI: "board://<id>", Description: "Read board description markdown"},
		{Name: "read_config", URI: "config://active", Description: "Read active board config"},
		{Name: "read_boards", URI: "boards://registry", Description: "Read board registry"},
	}
}

func (s *Server) listTasks(ctx context.Context, params listTasksParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	tasks, err := repo.GetAllTasksContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	fromDate, err := parseDate(params.From)
	if err != nil {
		return nil, invalidParams(err)
	}
	toDate, err := parseDate(params.To)
	if err != nil {
		return nil, invalidParams(err)
	}
	filtered := board.FilterAndSortTasks(tasks, board.ListOptions{
		Status:  params.Status,
		Title:   params.Title,
		Tags:    board.NormalizeTags(params.Tags),
		TagMode: params.TagMode,
		From:    fromDate,
		To:      toDate,
		SortBy:  params.Sort,
		Desc:    params.Desc,
	})

	result := make([]taskSummary, 0, len(filtered))
	for _, task := range filtered {
		result = append(result, toTaskSummary(task, boardID))
	}
	return result, nil
}

func (s *Server) listArchivedTasks(ctx context.Context, params listTasksParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	tasks, err := repo.ListArchivedTasksContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	fromDate, err := parseDate(params.From)
	if err != nil {
		return nil, invalidParams(err)
	}
	toDate, err := parseDate(params.To)
	if err != nil {
		return nil, invalidParams(err)
	}
	filtered := board.FilterAndSortTasks(tasks, board.ListOptions{
		Status:  params.Status,
		Title:   params.Title,
		Tags:    board.NormalizeTags(params.Tags),
		TagMode: params.TagMode,
		From:    fromDate,
		To:      toDate,
		SortBy:  params.Sort,
		Desc:    params.Desc,
	})
	result := make([]taskSummary, 0, len(filtered))
	for _, task := range filtered {
		result = append(result, toTaskSummary(task, boardID))
	}
	return result, nil
}

func (s *Server) getTask(params getTaskParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardID(params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	task, err := repo.GetTaskByID(params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return taskDetail{taskSummary: toTaskSummary(task, boardID), Content: task.Content}, nil
}

func (s *Server) createTask(ctx context.Context, params createTaskParams) (any, *rpcError) {
	if strings.TrimSpace(params.Title) == "" {
		return nil, invalidParams(fmt.Errorf("title is required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	task, err := board.NewTask(params.Title)
	if err != nil {
		return nil, internalError(err)
	}
	if strings.TrimSpace(params.Status) != "" {
		task.Status = params.Status
	}
	if params.Priority != 0 {
		task.Priority = params.Priority
	}
	if len(params.Tags) > 0 {
		task.Tags = params.Tags
	}
	created, err := repo.CreateTaskContext(ctx, task)
	if err != nil {
		return nil, internalError(err)
	}
	return toTaskSummary(created, boardID), nil
}

func (s *Server) updateTaskStatus(ctx context.Context, params updateStatusParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" || strings.TrimSpace(params.Status) == "" {
		return nil, invalidParams(fmt.Errorf("id and status are required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateTaskStatusContext(ctx, params.ID, params.Status); err != nil {
		return nil, internalError(err)
	}
	return s.getTask(getTaskParams{BoardID: boardID, ID: params.ID})
}

func (s *Server) updateTaskPriority(ctx context.Context, params updatePriorityParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateTaskPriorityContext(ctx, params.ID, params.Priority); err != nil {
		return nil, internalError(err)
	}
	return s.getTask(getTaskParams{BoardID: boardID, ID: params.ID})
}

func (s *Server) updateTaskTitle(ctx context.Context, params updateTitleParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" || strings.TrimSpace(params.Title) == "" {
		return nil, invalidParams(fmt.Errorf("id and title are required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateTaskTitleContext(ctx, params.ID, params.Title); err != nil {
		return nil, internalError(err)
	}
	return s.getTask(getTaskParams{BoardID: boardID, ID: params.ID})
}

func (s *Server) updateTaskTags(ctx context.Context, params updateTagsParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateTaskTagsContext(ctx, params.ID, params.Tags); err != nil {
		return nil, internalError(err)
	}
	return s.getTask(getTaskParams{BoardID: boardID, ID: params.ID})
}

func (s *Server) updateTaskContent(ctx context.Context, params updateContentParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateTaskContentContext(ctx, params.ID, params.Content); err != nil {
		return nil, internalError(err)
	}
	return s.getTask(getTaskParams{BoardID: boardID, ID: params.ID})
}

func (s *Server) updateTaskDependencies(ctx context.Context, params updateDepsParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateTaskDependenciesContext(ctx, params.ID, params.Depends); err != nil {
		return nil, internalError(err)
	}
	return s.getTask(getTaskParams{BoardID: boardID, ID: params.ID})
}

func (s *Server) getTaskDependencies(params getTaskParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardID(params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	task, err := repo.GetTaskByID(params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return map[string]any{"id": task.ID, "depends_on": task.DependsOn, "board_id": boardID}, nil
}

func (s *Server) listReadyTasks(ctx context.Context, params listTasksParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	tasks, err := repo.ListReadyTasksContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	result := make([]taskSummary, 0, len(tasks))
	for _, task := range tasks {
		result = append(result, toTaskSummary(task, boardID))
	}
	return result, nil
}

func (s *Server) archiveTask(ctx context.Context, params taskIDParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	if err := requireForce(params.Force, "archive_task"); err != nil {
		return nil, err
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	task, err := repo.ArchiveTaskContext(ctx, params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return toTaskSummary(task, boardID), nil
}

func (s *Server) restoreTask(ctx context.Context, params taskIDParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	task, err := repo.RestoreTaskContext(ctx, params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return toTaskSummary(task, boardID), nil
}

func (s *Server) deleteTask(ctx context.Context, params taskIDParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	if err := requireForce(params.Force, "delete_task"); err != nil {
		return nil, err
	}
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.DeleteTaskContext(ctx, params.ID); err != nil {
		return nil, internalError(err)
	}
	return map[string]string{"deleted": params.ID, "board_id": boardID}, nil
}

func (s *Server) listBoards(ctx context.Context) (any, *rpcError) {
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	boards, active, err := repo.ListBoardsContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	result := make([]boardSummary, 0, len(boards))
	for _, board := range boards {
		result = append(result, toBoardSummary(board))
	}
	return map[string]any{"active": active, "boards": result}, nil
}

func (s *Server) createBoard(ctx context.Context, params boardCreateParams) (any, *rpcError) {
	if strings.TrimSpace(params.Name) == "" {
		return nil, invalidParams(fmt.Errorf("name is required"))
	}
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	board, err := repo.CreateBoardContext(ctx, params.Name)
	if err != nil {
		return nil, internalError(err)
	}
	return toBoardSummary(board), nil
}

func (s *Server) renameBoard(ctx context.Context, params boardNameParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" || strings.TrimSpace(params.Name) == "" {
		return nil, invalidParams(fmt.Errorf("id and name are required"))
	}
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	board, err := repo.RenameBoardContext(ctx, params.ID, params.Name)
	if err != nil {
		return nil, internalError(err)
	}
	return toBoardSummary(board), nil
}

func (s *Server) setActiveBoard(ctx context.Context, params boardIDParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.SetActiveBoardContext(ctx, params.ID); err != nil {
		return nil, internalError(err)
	}
	return map[string]string{"active": params.ID}, nil
}

func (s *Server) archiveBoard(ctx context.Context, params boardIDParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	if err := requireForce(params.Force, "archive_board"); err != nil {
		return nil, err
	}
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	board, err := repo.ArchiveBoardContext(ctx, params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	return toBoardSummary(board), nil
}

func (s *Server) deleteBoard(ctx context.Context, params boardIDParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	if err := requireForce(params.Force, "delete_board"); err != nil {
		return nil, err
	}
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.DeleteBoardContext(ctx, params.ID); err != nil {
		return nil, internalError(err)
	}
	return map[string]string{"deleted": params.ID}, nil
}

func (s *Server) updateBoardDescription(ctx context.Context, params updateBoardDescriptionParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateBoardDescriptionContext(ctx, params.Description); err != nil {
		return nil, internalError(err)
	}
	return s.readBoard(ctx, boardReadParams{ID: boardID})
}

func (s *Server) updateBoardContext(ctx context.Context, params boardContextParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	if err := repo.UpdateBoardContextContext(ctx, params.Context); err != nil {
		return nil, internalError(err)
	}
	cfg, err := repo.LoadConfigContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	return cfg.Context, nil
}

func (s *Server) getBoardContext(ctx context.Context, params boardReadParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	cfg, err := repo.LoadConfigContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	return cfg.Context, nil
}

func (s *Server) readBoard(ctx context.Context, params boardReadParams) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	path, err := repo.BoardDescriptionPath()
	if err != nil {
		return nil, internalError(err)
	}
	content, err := repo.LoadBoardDescriptionContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	return map[string]string{"id": boardID, "path": path, "content": content}, nil
}

func (s *Server) readResource(ctx context.Context, uri string) (any, *rpcError) {
	// Parse URI format: scheme://id
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) != 2 {
		return nil, invalidParams(fmt.Errorf("invalid resource URI format: %s", uri))
	}
	scheme := parts[0]
	id := parts[1]

	switch scheme {
	case "task":
		return s.readTask(getTaskParams{ID: id})
	case "board":
		return s.readBoard(ctx, boardReadParams{ID: id})
	case "config":
		if id == "active" {
			return s.readConfig(ctx)
		}
		return nil, invalidParams(fmt.Errorf("unknown config resource: %s", id))
	case "boards":
		if id == "registry" {
			return s.readBoards(ctx)
		}
		return nil, invalidParams(fmt.Errorf("unknown boards resource: %s", id))
	default:
		return nil, invalidParams(fmt.Errorf("unknown resource scheme: %s", scheme))
	}
}

func (s *Server) readTask(params getTaskParams) (any, *rpcError) {
	if strings.TrimSpace(params.ID) == "" {
		return nil, invalidParams(fmt.Errorf("id is required"))
	}
	boardID, err := s.resolveBoardID(params.BoardID)
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	task, err := repo.GetTaskByID(params.ID)
	if err != nil {
		return nil, internalError(err)
	}
	data, err := os.ReadFile(task.FilePath)
	if err != nil {
		return nil, internalError(err)
	}
	return map[string]string{"id": task.ID, "path": task.FilePath, "content": string(data), "board_id": boardID}, nil
}

type wikiPageSummary struct {
	Title   string   `json:"title"`
	Slug    string   `json:"slug"`
	Section string   `json:"section"`
	Order   int      `json:"order"`
	Tags    []string `json:"tags"`
	Status  string   `json:"status"`
}

type wikiPageDetail struct {
	wikiPageSummary
	Content string `json:"content"`
}

type wikiSectionSummary struct {
	Title string            `json:"title"`
	Slug  string            `json:"slug"`
	Order int               `json:"order"`
	Tags  []string          `json:"tags"`
	Links wiki.SectionLinks `json:"links"`
}

func (s *Server) listWikiPages(ctx context.Context, params listWikiParams) (any, *rpcError) {
	root := s.wikiRoot()
	templatePaths, err := s.templatePaths()
	if err != nil {
		return nil, internalError(err)
	}
	indexPath := filepath.Join(root, "_index.yaml")
	index, err := wiki.LoadIndexContext(ctx, indexPath)
	if err != nil && !errors.Is(err, wiki.ErrIndexNotFound) {
		return nil, internalError(err)
	}

	var pages []wiki.Page
	if err == nil {
		pages, err = wiki.ListPagesFromIndexContext(ctx, root, index)
		if err != nil {
			return nil, internalError(err)
		}
		if params.IncludeTemplates {
			allPages, err := wiki.ListPagesWithTemplatesRootContext(ctx, root, true, templatePaths.Wiki)
			if err != nil {
				return nil, internalError(err)
			}
			pagesBySlug := make(map[string]wiki.Page, len(allPages))
			for _, page := range allPages {
				slug := slugForPage(root, templatePaths.Wiki, page)
				if slug == "" {
					continue
				}
				page.Slug = slug
				pagesBySlug[slug] = page
			}
			ordered := make([]wiki.Page, 0, len(pagesBySlug))
			seen := make(map[string]struct{}, len(pagesBySlug))
			for _, page := range pages {
				slug := slugForPage(root, templatePaths.Wiki, page)
				if slug == "" {
					continue
				}
				if full, ok := pagesBySlug[slug]; ok {
					ordered = append(ordered, full)
				} else {
					ordered = append(ordered, page)
				}
				seen[slug] = struct{}{}
			}
			remaining := make([]wiki.Page, 0)
			for slug, page := range pagesBySlug {
				if _, ok := seen[slug]; ok {
					continue
				}
				remaining = append(remaining, page)
			}
			sort.Slice(remaining, func(i, j int) bool {
				return strings.ToLower(remaining[i].Title) < strings.ToLower(remaining[j].Title)
			})
			pages = append(ordered, remaining...)
		}
	} else {
		pages, err = wiki.ListPagesWithTemplatesRootContext(ctx, root, params.IncludeTemplates, templatePaths.Wiki)
		if err != nil {
			return nil, internalError(err)
		}
		sort.Slice(pages, func(i, j int) bool {
			return strings.ToLower(pages[i].Title) < strings.ToLower(pages[j].Title)
		})
	}

	pages = wiki.FilterPagesByStatus(pages, params.Status)
	caseInsensitive := true
	if params.CaseInsensitive != nil {
		caseInsensitive = *params.CaseInsensitive
	}
	pages = wiki.FilterPages(pages, wiki.FilterOptions{
		Title:           params.Title,
		Tags:            params.Tags,
		TagMode:         params.TagMode,
		Section:         params.Section,
		Query:           params.Query,
		CaseInsensitive: caseInsensitive,
	})
	summaries := make([]wikiPageSummary, 0, len(pages))
	for _, page := range pages {
		slug := slugForPage(root, templatePaths.Wiki, page)
		summaries = append(summaries, wikiPageSummary{
			Title:   page.Title,
			Slug:    slug,
			Section: page.Section,
			Order:   page.Order,
			Tags:    page.Tags,
			Status:  page.Status,
		})
	}
	return summaries, nil
}

func (s *Server) listWikiSections(ctx context.Context, params listWikiSectionsParams) (any, *rpcError) {
	root := s.wikiRoot()
	indexPath := filepath.Join(root, "_index.yaml")
	index, err := wiki.LoadIndexContext(ctx, indexPath)
	if err != nil {
		if errors.Is(err, wiki.ErrIndexNotFound) {
			pages, err := wiki.ListPagesContext(ctx, root)
			if err != nil {
				return nil, internalError(err)
			}
			index, err = wiki.GenerateIndexContext(ctx, pages)
			if err != nil {
				return nil, internalError(err)
			}
		} else {
			return nil, internalError(err)
		}
	}

	sections := wiki.FilterSections(index.Sections, wiki.SectionFilterOptions{
		Tags:       params.Tags,
		TagMode:    params.TagMode,
		LinkType:   params.LinkType,
		LinkTarget: params.LinkTarget,
	})
	summaries := make([]wikiSectionSummary, 0, len(sections))
	for _, section := range sections {
		summaries = append(summaries, wikiSectionSummary{
			Title: section.Title,
			Slug:  section.Slug,
			Order: section.Order,
			Tags:  section.Tags,
			Links: section.Links,
		})
	}
	return summaries, nil
}

func (s *Server) readWikiPage(params readWikiParams) (any, *rpcError) {
	slug, err := wiki.NormalizeSlug(params.Slug)
	if err != nil {
		return nil, invalidParams(err)
	}
	path := filepath.Join(s.wikiRoot(), filepath.FromSlash(slug)+".md")
	page, err := wiki.LoadPage(path)
	if err != nil {
		return nil, internalError(err)
	}
	if strings.TrimSpace(page.Slug) == "" {
		page.Slug = slug
	}
	return wikiPageDetail{
		wikiPageSummary: wikiPageSummary{
			Title:   page.Title,
			Slug:    page.Slug,
			Section: page.Section,
			Order:   page.Order,
			Tags:    page.Tags,
			Status:  page.Status,
		},
		Content: page.Content,
	}, nil
}

func (s *Server) writeWikiPage(params writeWikiParams) (any, *rpcError) {
	slug, err := wiki.NormalizeSlug(params.Slug)
	if err != nil {
		return nil, invalidParams(err)
	}
	path := filepath.Join(s.wikiRoot(), filepath.FromSlash(slug)+".md")

	var page wiki.Page
	_, statErr := os.Stat(path)
	if statErr == nil {
		page, err = wiki.LoadPage(path)
		if err != nil {
			return nil, internalError(err)
		}
	} else if !os.IsNotExist(statErr) {
		return nil, internalError(statErr)
	}

	if params.Title != nil {
		title := strings.TrimSpace(*params.Title)
		if title == "" {
			return nil, invalidParams(fmt.Errorf("title is required"))
		}
		page.Title = title
	} else if page.Title == "" {
		return nil, invalidParams(fmt.Errorf("title is required"))
	}

	if params.Section != nil {
		page.Section = strings.TrimSpace(*params.Section)
	}
	if params.Order != nil {
		page.Order = *params.Order
	}
	if params.Tags != nil {
		page.Tags = *params.Tags
	}
	if params.Status != nil {
		status := strings.TrimSpace(*params.Status)
		if status == "" {
			return nil, invalidParams(fmt.Errorf("status is required"))
		}
		page.Status = status
	} else if strings.TrimSpace(page.Status) == "" {
		page.Status = "published"
	}
	if params.Content != nil {
		page.Content = *params.Content
	}

	page.Slug = slug
	if err := wiki.SavePage(path, page); err != nil {
		return nil, internalError(err)
	}

	return wikiPageSummary{
		Title:   page.Title,
		Slug:    page.Slug,
		Section: page.Section,
		Order:   page.Order,
		Tags:    page.Tags,
		Status:  page.Status,
	}, nil
}

func (s *Server) updateWikiSection(ctx context.Context, params updateWikiSectionParams) (any, *rpcError) {
	if strings.TrimSpace(params.Slug) == "" {
		return nil, invalidParams(fmt.Errorf("slug is required"))
	}
	indexPath := filepath.Join(s.wikiRoot(), "_index.yaml")
	index, err := wiki.LoadIndexContext(ctx, indexPath)
	if err != nil {
		if errors.Is(err, wiki.ErrIndexNotFound) {
			return nil, invalidParams(fmt.Errorf("wiki index not found"))
		}
		return nil, internalError(err)
	}

	section, idx, err := wiki.FindSection(index, params.Slug)
	if err != nil {
		return nil, invalidParams(err)
	}
	if params.Title != nil {
		section.Title = strings.TrimSpace(*params.Title)
	}
	if params.Order != nil {
		section.Order = *params.Order
	}
	if params.Tags != nil {
		section.Tags = normalizeSectionTags(*params.Tags)
	}
	if params.Links != nil {
		section.Links = normalizeSectionLinks(*params.Links)
	}
	index.Sections[idx] = section
	if err := wiki.SaveIndexContext(ctx, indexPath, index); err != nil {
		return nil, internalError(err)
	}
	return wikiSectionSummary{
		Title: section.Title,
		Slug:  section.Slug,
		Order: section.Order,
		Tags:  section.Tags,
		Links: section.Links,
	}, nil
}

func (s *Server) searchWiki(ctx context.Context, params searchWikiParams) (any, *rpcError) {
	if strings.TrimSpace(params.Query) == "" {
		return nil, invalidParams(fmt.Errorf("query is required"))
	}
	caseInsensitive := true
	if params.CaseInsensitive != nil {
		caseInsensitive = *params.CaseInsensitive
	}
	templatePaths, err := s.templatePaths()
	if err != nil {
		return nil, internalError(err)
	}
	results, err := wiki.SearchPagesContext(ctx, s.wikiRoot(), wiki.SearchOptions{
		Query:            params.Query,
		Status:           params.Status,
		IncludeTemplates: params.IncludeTemplates,
		CaseInsensitive:  caseInsensitive,
		TemplatesRoot:    templatePaths.Wiki,
	})
	if err != nil {
		return nil, internalError(err)
	}
	return results, nil
}

func (s *Server) listWikiTemplates(params listWikiTemplatesParams) (any, *rpcError) {
	templatePaths, err := s.templatePaths()
	if err != nil {
		return nil, internalError(err)
	}
	templatesDir := templatePaths.Wiki
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, internalError(err)
	}
	names := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".md" {
			continue
		}
		names = append(names, strings.TrimSuffix(name, ".md"))
	}
	sort.Strings(names)
	return names, nil
}

func (s *Server) createWikiFromTemplate(params createWikiFromTemplateParams) (any, *rpcError) {
	if strings.TrimSpace(params.Template) == "" {
		return nil, invalidParams(fmt.Errorf("template is required"))
	}
	if strings.TrimSpace(params.Title) == "" {
		return nil, invalidParams(fmt.Errorf("title is required"))
	}

	slug := strings.TrimSpace(params.Slug)
	if slug == "" {
		slug = wiki.Slugify(params.Title)
	}
	var err error
	slug, err = wiki.NormalizeSlug(slug)
	if err != nil {
		return nil, invalidParams(err)
	}

	templatePaths, err := s.templatePaths()
	if err != nil {
		return nil, internalError(err)
	}
	templatePath := filepath.Join(templatePaths.Wiki, params.Template+".md")
	templatePage, err := wiki.LoadPage(templatePath)
	if err != nil {
		return nil, internalError(fmt.Errorf("template not found: %s", params.Template))
	}

	page := wiki.Page{
		Title:   strings.TrimSpace(params.Title),
		Slug:    slug,
		Section: strings.TrimSpace(templatePage.Section),
		Order:   templatePage.Order,
		Tags:    templatePage.Tags,
		Status:  templatePage.Status,
		Content: templatePage.Content,
	}

	if params.Section != nil {
		page.Section = strings.TrimSpace(*params.Section)
	}
	if params.Order != nil {
		page.Order = *params.Order
	}
	if params.Tags != nil {
		page.Tags = *params.Tags
	}
	if params.Status != nil {
		page.Status = strings.TrimSpace(*params.Status)
	}
	if strings.TrimSpace(page.Status) == "" {
		page.Status = "published"
	}

	path := filepath.Join(s.wikiRoot(), filepath.FromSlash(slug)+".md")
	if _, err := os.Stat(path); err == nil {
		return nil, invalidParams(fmt.Errorf("page already exists: %s", slug))
	} else if !os.IsNotExist(err) {
		return nil, internalError(err)
	}
	if err := wiki.SavePage(path, page); err != nil {
		return nil, internalError(err)
	}

	return wikiPageSummary{
		Title:   page.Title,
		Slug:    page.Slug,
		Section: page.Section,
		Order:   page.Order,
		Tags:    page.Tags,
		Status:  page.Status,
	}, nil
}

func (s *Server) lintWiki(ctx context.Context, params lintWikiParams) (any, *rpcError) {
	pages, err := wiki.ListPagesContext(ctx, s.wikiRoot())
	if err != nil {
		return nil, internalError(err)
	}
	return wiki.LintPages(pages), nil
}

func (s *Server) manifestWiki(ctx context.Context, params manifestWikiParams) (any, *rpcError) {
	indexPath := filepath.Join(s.wikiRoot(), "_index.yaml")
	index, indexErr := wiki.LoadIndexContext(ctx, indexPath)
	if indexErr != nil && !errors.Is(indexErr, wiki.ErrIndexNotFound) {
		return nil, internalError(indexErr)
	}
	var pages []wiki.Page
	if indexErr == nil {
		var err error
		pages, err = wiki.ListPagesFromIndexContext(ctx, s.wikiRoot(), index)
		if err != nil {
			return nil, internalError(err)
		}
		manifest, err := wiki.BuildManifest(index, pages)
		if err != nil {
			return nil, internalError(err)
		}
		return manifest, nil
	}
	pages, err := wiki.ListPagesContext(ctx, s.wikiRoot())
	if err != nil {
		return nil, internalError(err)
	}
	return wiki.BuildManifestFromPages(pages), nil
}

func (s *Server) exportWiki(ctx context.Context, params exportWikiParams) (any, *rpcError) {
	select {
	case <-ctx.Done():
		return nil, &rpcError{Code: codeInternalError, Message: ctx.Err().Error()}
	default:
	}
	format := strings.ToLower(strings.TrimSpace(params.Format))
	if format == "" {
		format = "md"
	}
	if format != "md" && format != "pdf" {
		return nil, invalidParams(fmt.Errorf("unsupported format: %s", format))
	}
	pageSlug := strings.TrimSpace(params.Page)
	section := strings.TrimSpace(params.Section)
	if pageSlug != "" && section != "" {
		return nil, invalidParams(fmt.Errorf("choose either page or section"))
	}
	if (pageSlug != "" || section != "") && len(params.Roots) > 0 {
		return nil, invalidParams(fmt.Errorf("roots cannot be used with page or section export"))
	}
	filtersActive := strings.TrimSpace(params.FilterTitle) != "" ||
		strings.TrimSpace(params.FilterSection) != "" ||
		strings.TrimSpace(params.FilterQuery) != "" ||
		len(params.FilterTags) > 0
	if filtersActive && len(params.Roots) > 0 {
		return nil, invalidParams(fmt.Errorf("filters cannot be used with roots"))
	}

	indexPath := filepath.Join(s.wikiRoot(), "_index.yaml")
	index, indexErr := wiki.LoadIndexContext(ctx, indexPath)
	if indexErr != nil && !errors.Is(indexErr, wiki.ErrIndexNotFound) {
		return nil, internalError(indexErr)
	}

	select {
	case <-ctx.Done():
		return nil, &rpcError{Code: codeInternalError, Message: ctx.Err().Error()}
	default:
	}

	var pages []wiki.Page
	var err error
	if indexErr == nil {
		pages, err = wiki.ListPagesFromIndexContext(ctx, s.wikiRoot(), index)
	} else {
		pages, err = wiki.ListPagesContext(ctx, s.wikiRoot())
	}
	if err != nil {
		return nil, internalError(err)
	}

	select {
	case <-ctx.Done():
		return nil, &rpcError{Code: codeInternalError, Message: ctx.Err().Error()}
	default:
	}

	var manifest []wiki.ManifestEntry
	if indexErr == nil {
		manifest, err = wiki.BuildManifest(index, pages)
		if err != nil {
			return nil, internalError(err)
		}
	} else {
		manifest = wiki.BuildManifestFromPages(pages)
	}

	mainPrefix := strings.TrimSpace(params.Prefix)
	mainManifest := wiki.RootManifest{Root: s.wikiRoot(), Prefix: mainPrefix, Pages: manifest}
	selection := wiki.ExportSelection{Page: pageSlug, Section: section}
	caseInsensitive := true
	if params.FilterCaseInsensitive != nil {
		caseInsensitive = *params.FilterCaseInsensitive
	}
	filterOptions := wiki.FilterOptions{
		Title:           params.FilterTitle,
		Tags:            params.FilterTags,
		TagMode:         params.FilterTagMode,
		Section:         params.FilterSection,
		Query:           params.FilterQuery,
		CaseInsensitive: caseInsensitive,
	}

	select {
	case <-ctx.Done():
		return nil, &rpcError{Code: codeInternalError, Message: ctx.Err().Error()}
	default:
	}

	if pageSlug != "" {
		selected, err := wiki.BuildPageManifest(s.wikiRoot(), pages, pageSlug)
		if err != nil {
			return nil, internalError(err)
		}
		mainManifest.Pages = selected
	}
	if section != "" {
		sectionIndex := index
		if indexErr != nil {
			sectionIndex, err = wiki.GenerateIndex(pages)
			if err != nil {
				return nil, internalError(err)
			}
		}
		if params.IncludeLinked {
			baseSection, _, err := wiki.FindSection(sectionIndex, section)
			if err != nil {
				return nil, internalError(err)
			}
			linked, err := wiki.ResolveLinkedSections(sectionIndex, baseSection, params.LinkTypes)
			if err != nil {
				return nil, internalError(err)
			}
			sections := append([]wiki.IndexSection{baseSection}, linked...)
			selected, err := wiki.BuildManifestForSections(sectionIndex, pages, sections)
			if err != nil {
				return nil, internalError(err)
			}
			mainManifest.Pages = selected
		} else {
			selected, err := wiki.BuildSectionManifest(sectionIndex, pages, section)
			if err != nil {
				return nil, internalError(err)
			}
			mainManifest.Pages = selected
		}
	}

	if filtersActive {
		pageMap := make(map[string]wiki.Page, len(pages))
		for _, page := range pages {
			slug := strings.TrimSpace(page.Slug)
			if slug == "" && page.FilePath != "" {
				slug = wiki.SlugFromPath(s.wikiRoot(), page.FilePath)
			}
			if slug == "" {
				continue
			}
			page.Slug = slug
			pageMap[slug] = page
		}
		mainManifest.Pages = wiki.FilterManifest(mainManifest.Pages, pageMap, filterOptions)
	}

	rootManifests := []wiki.RootManifest{mainManifest}
	if pageSlug == "" && section == "" {
		for _, external := range params.Roots {
			parts := strings.SplitN(external, ":", 2)
			pathPart := strings.TrimSpace(parts[0])
			if pathPart == "" {
				continue
			}
			if !filepath.IsAbs(pathPart) {
				pathPart = filepath.Join(s.baseDir, pathPart)
			}
			prefix := ""
			if len(parts) == 2 {
				prefix = strings.TrimSpace(parts[1])
			}
			externalManifest, err := wiki.BuildRootManifestContext(ctx, pathPart, prefix)
			if err != nil {
				return nil, internalError(err)
			}
			rootManifests = append(rootManifests, externalManifest)
		}

		if err := wiki.ValidateManifests(rootManifests); err != nil {
			return nil, internalError(err)
		}
	}

	select {
	case <-ctx.Done():
		return nil, &rpcError{Code: codeInternalError, Message: ctx.Err().Error()}
	default:
	}

	output := strings.TrimSpace(params.Output)
	if output == "" {
		output = wiki.DefaultExportPath(s.wikiRoot(), format, selection)
	}

	var data []byte
	if len(rootManifests) > 1 {
		flat := wiki.FlattenManifests(rootManifests)
		if len(flat) == 0 {
			return map[string]string{"status": "empty"}, nil
		}
		data, err = wiki.ExportMarkdownMultiContext(ctx, rootManifests)
	} else {
		if len(mainManifest.Pages) == 0 {
			return map[string]string{"status": "empty"}, nil
		}
		data, err = wiki.ExportMarkdownContext(ctx, s.wikiRoot(), mainManifest.Pages)
	}
	if err != nil {
		return nil, internalError(err)
	}

	if format == "md" {
		if err := wiki.WriteExportContext(ctx, output, data); err != nil {
			return nil, internalError(err)
		}
		return map[string]string{"path": output}, nil
	}

	if _, err := wiki.WritePDFContext(ctx, data, wiki.PDFOptions{
		Output:   output,
		Title:    params.Title,
		Author:   params.Author,
		Template: params.Template,
		BaseDir:  s.baseDir,
		TempDir:  s.wikiRoot(),
	}); err != nil {
		return nil, internalError(err)
	}
	return map[string]string{"path": output}, nil
}

func (s *Server) deleteWikiPage(ctx context.Context, params deleteWikiParams) (any, *rpcError) {
	if strings.TrimSpace(params.Slug) == "" {
		return nil, invalidParams(fmt.Errorf("slug is required"))
	}
	slug, err := wiki.NormalizeSlug(params.Slug)
	if err != nil {
		return nil, invalidParams(err)
	}
	select {
	case <-ctx.Done():
		return nil, internalError(ctx.Err())
	default:
	}
	path := filepath.Join(s.wikiRoot(), filepath.FromSlash(slug)+".md")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil, invalidParams(fmt.Errorf("page not found: %s", slug))
		}
		return nil, internalError(err)
	}

	if params.UpdateIndex {
		indexPath := filepath.Join(s.wikiRoot(), "_index.yaml")
		index, err := wiki.LoadIndexContext(ctx, indexPath)
		if err == nil {
			if index.RemoveSlug(slug) {
				if err := wiki.SaveIndexContext(ctx, indexPath, index); err != nil {
					return nil, internalError(err)
				}
			}
		} else if !errors.Is(err, wiki.ErrIndexNotFound) {
			return nil, internalError(err)
		}
	}

	return map[string]string{"deleted": slug}, nil
}

func (s *Server) generateWikiIndex(ctx context.Context, params generateWikiIndexParams) (any, *rpcError) {
	templatePaths, err := s.templatePaths()
	if err != nil {
		return nil, internalError(err)
	}
	pages, err := wiki.ListPagesWithTemplatesRootContext(ctx, s.wikiRoot(), params.IncludeTemplates, templatePaths.Wiki)
	if err != nil {
		return nil, internalError(err)
	}
	index, err := wiki.GenerateIndexContext(ctx, pages)
	if err != nil {
		return nil, internalError(err)
	}

	write := params.Write
	if params.Output == "" && !write {
		return index, nil
	}
	if params.Output == "" {
		params.Output = filepath.Join(s.wikiRoot(), "_index.yaml")
	}
	if write {
		if err := wiki.SaveIndexContext(ctx, params.Output, index); err != nil {
			return nil, internalError(err)
		}
	}
	return map[string]any{"index": index, "path": params.Output}, nil
}

func (s *Server) readConfig(ctx context.Context) (any, *rpcError) {
	boardID, err := s.resolveBoardIDContext(ctx, "")
	if err != nil {
		return nil, internalError(err)
	}
	repo, err := s.repoForBoard(boardID)
	if err != nil {
		return nil, internalError(err)
	}
	cfg, err := repo.LoadConfigContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	return map[string]any{"board_id": boardID, "config": cfg}, nil
}

func (s *Server) readBoards(ctx context.Context) (any, *rpcError) {
	repo, err := s.boardRepo()
	if err != nil {
		return nil, internalError(err)
	}
	registry, err := repo.LoadRegistryContext(ctx)
	if err != nil {
		return nil, internalError(err)
	}
	result := make([]boardSummary, 0, len(registry.Boards))
	for _, board := range registry.Boards {
		result = append(result, toBoardSummary(board))
	}
	return map[string]any{"active": registry.Active, "boards": result}, nil
}

func (s *Server) repoForBoard(boardID string) (*board.Repository, error) {
	return board.NewRepositoryForBoardWithStorage(s.baseDir, boardID, s.storageRoot)
}

func (s *Server) wikiRoot() string {
	return filepath.Join(s.storageRoot, "wiki")
}

func normalizeSectionTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{})
	for _, tag := range tags {
		clean := strings.TrimSpace(tag)
		if clean == "" {
			continue
		}
		key := strings.ToLower(clean)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, clean)
	}
	return result
}

func normalizeSectionLinks(links wiki.SectionLinks) wiki.SectionLinks {
	normalize := func(values []string) []string {
		result := make([]string, 0, len(values))
		seen := make(map[string]struct{})
		for _, value := range values {
			clean := strings.TrimSpace(value)
			if clean == "" {
				continue
			}
			key := strings.ToLower(clean)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, clean)
		}
		return result
	}
	return wiki.SectionLinks{
		DependsOn: normalize(links.DependsOn),
		RelatedTo: normalize(links.RelatedTo),
	}
}

func (s *Server) boardRepo() (*board.BoardRepository, error) {
	return board.NewBoardRepositoryWithStorage(s.baseDir, s.storageRoot)
}

func (s *Server) resolveBoardID(boardID string) (string, error) {
	return s.resolveBoardIDContext(context.Background(), boardID)
}

func (s *Server) resolveBoardIDContext(ctx context.Context, boardID string) (string, error) {
	trimmed := strings.TrimSpace(boardID)
	if trimmed != "" {
		return trimmed, nil
	}
	repo, err := s.boardRepo()
	if err != nil {
		return "", err
	}
	_, active, err := repo.ListBoardsContext(ctx)
	if err != nil {
		return "", err
	}
	return active, nil
}

func toTaskSummary(task board.Task, boardID string) taskSummary {
	created := ""
	if !task.Created.IsZero() {
		created = task.Created.Format("2006-01-02")
	}
	priority := task.Priority
	if priority == 0 {
		priority = board.DefaultPriority
	}
	return taskSummary{
		BoardID:   boardID,
		BoardName: board.TaskBoardLabel(task),
		ID:        task.ID,
		UID:       task.UID,
		Title:     task.Title,
		Status:    task.Status,
		Priority:  priority,
		Tags:      task.Tags,
		Created:   created,
		DependsOn: task.DependsOn,
	}
}

func toBoardSummary(board board.Board) boardSummary {
	created := ""
	if !board.Created.IsZero() {
		created = board.Created.Format("2006-01-02")
	}
	return boardSummary{
		ID:       board.ID,
		Name:     board.Name,
		Path:     board.Path,
		Archived: board.Archived,
		Created:  created,
	}
}

func decodeParams(raw json.RawMessage, dest any) error {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	return nil
}

func parseDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q", value)
	}
	return parsed, nil
}

func invalidParams(err error) *rpcError {
	return &rpcError{Code: codeInvalidParams, Message: err.Error()}
}

func internalError(err error) *rpcError {
	return &rpcError{Code: codeInternalError, Message: err.Error()}
}

func requireForce(force bool, action string) *rpcError {
	if force {
		return nil
	}
	return &rpcError{Code: codeDenied, Message: fmt.Sprintf("%s requires force: true", action)}
}

func (s *Server) writeErrorBuf(encoder *json.Encoder, buf *bufio.Writer, id json.RawMessage, code int, message string) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		Error:   &rpcError{Code: code, Message: message},
		ID:      id,
	}
	_ = encoder.Encode(resp)
	_ = buf.Flush()
}
func isNotification(id json.RawMessage) bool {
	return len(id) == 0 || string(id) == "null"
}
