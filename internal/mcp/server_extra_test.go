package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mochi-sticky/internal/board"
	"mochi-sticky/internal/wiki"
)

func TestServerListBoardsAfterInit(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	output := runServerWithStorage(t, baseDir, storageRoot, `{"jsonrpc":"2.0","method":"list_boards","id":1}`)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	if result["active"] != "default" {
		t.Fatalf("expected active default, got %v", result["active"])
	}
	boards := result["boards"].([]any)
	if len(boards) != 1 {
		t.Fatalf("expected 1 board, got %d", len(boards))
	}
}

func TestServerReadConfigDefault(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	output := runServerWithStorage(t, baseDir, storageRoot, `{"jsonrpc":"2.0","method":"read_config","params":{},"id":2}`)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	if result["board_id"] != "default" {
		t.Fatalf("expected board_id default, got %v", result["board_id"])
	}
	cfg := result["config"].(map[string]any)
	rawColumns, ok := cfg["Columns"]
	if !ok {
		rawColumns = cfg["columns"]
	}
	columns, ok := rawColumns.([]any)
	if !ok || len(columns) == 0 {
		t.Fatalf("expected default columns, got %v", rawColumns)
	}
}

func TestServerCreateAndListTasks(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"create_task","params":{"title":"One"},"id":1}`,
		`{"jsonrpc":"2.0","method":"list_tasks","params":{},"id":2}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	task := responses[0].Result.(map[string]any)
	if strings.TrimSpace(task["title"].(string)) != "One" {
		t.Fatalf("expected created task title, got %v", task["title"])
	}

	if responses[1].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[1].Error)
	}
	list := responses[1].Result.([]any)
	if len(list) != 1 {
		t.Fatalf("expected 1 task, got %d", len(list))
	}
}

func TestServerReadBoardEmptyDescription(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	output := runServerWithStorage(t, baseDir, storageRoot, `{"jsonrpc":"2.0","method":"read_board","params":{"id":"default"},"id":3}`)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	if result["id"] != "default" {
		t.Fatalf("expected default board id, got %v", result["id"])
	}
}

func TestServerWikiListSectionsAndExport(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	wikiRoot := filepath.Join(storageRoot, "wiki")

	if err := os.MkdirAll(wikiRoot, 0o755); err != nil {
		t.Fatalf("mkdir wiki: %v", err)
	}
	if err := wiki.SavePage(filepath.Join(wikiRoot, "alpha.md"), wiki.Page{
		Title:   "Alpha",
		Slug:    "alpha",
		Section: "General",
		Status:  "published",
		Content: "Alpha body",
	}); err != nil {
		t.Fatalf("write page: %v", err)
	}
	if err := wiki.SavePage(filepath.Join(wikiRoot, "beta.md"), wiki.Page{
		Title:   "Beta",
		Slug:    "beta",
		Section: "General",
		Status:  "published",
		Content: "Beta body",
	}); err != nil {
		t.Fatalf("write page: %v", err)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"list_wiki_sections","params":{},"id":10}`,
		`{"jsonrpc":"2.0","method":"list_wiki_pages","params":{},"id":11}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	if responses[1].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[1].Error)
	}
	sections := responses[0].Result.([]any)
	if len(sections) == 0 {
		t.Fatalf("expected sections")
	}
	pages := responses[1].Result.([]any)
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}

	exportPath := filepath.Join(wikiRoot, "export.md")
	exportInputBytes, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  "export_wiki",
		"params": map[string]any{
			"format": "md",
			"output": exportPath,
		},
		"id": 12,
	})
	if err != nil {
		t.Fatalf("marshal export request: %v", err)
	}
	exportInput := string(exportInputBytes)
	exportOutput := runServerWithStorage(t, baseDir, storageRoot, exportInput)
	exportResponses := decodeResponses(t, exportOutput)
	if len(exportResponses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(exportResponses))
	}
	if exportResponses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", exportResponses[0].Error)
	}
	if _, err := os.Stat(exportPath); err != nil {
		t.Fatalf("expected export file: %v", err)
	}
}

func TestServerCreateBoardAndSetActive(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"create_board","params":{"name":"Work"},"id":20}`,
		`{"jsonrpc":"2.0","method":"set_active_board","params":{"id":"work"},"id":21}`,
		`{"jsonrpc":"2.0","method":"read_boards","id":22}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 3 {
		t.Fatalf("expected 3 responses, got %d", len(responses))
	}
	if responses[2].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[2].Error)
	}
	result := responses[2].Result.(map[string]any)
	if result["active"] != "work" {
		t.Fatalf("expected active work, got %v", result["active"])
	}
}

func TestServerWikiTemplatesAndCreateFromTemplate(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	templatesDir := filepath.Join(storageRoot, "templates", "wiki")
	wikiRoot := filepath.Join(storageRoot, "wiki")

	if err := os.MkdirAll(templatesDir, 0o755); err != nil {
		t.Fatalf("mkdir templates: %v", err)
	}
	if err := os.MkdirAll(wikiRoot, 0o755); err != nil {
		t.Fatalf("mkdir wiki: %v", err)
	}
	if err := wiki.SavePage(filepath.Join(templatesDir, "starter.md"), wiki.Page{
		Title:   "Starter",
		Section: "General",
		Status:  "draft",
		Content: "# Starter\n",
	}); err != nil {
		t.Fatalf("save template: %v", err)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"list_wiki_templates","id":30}`,
		`{"jsonrpc":"2.0","method":"create_wiki_from_template","params":{"template":"starter","title":"My Page","slug":"my-page"},"id":31}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].Error != nil || responses[1].Error != nil {
		t.Fatalf("unexpected error: %+v %+v", responses[0].Error, responses[1].Error)
	}
	templates := responses[0].Result.([]any)
	if len(templates) != 1 || templates[0].(string) != "starter" {
		t.Fatalf("expected starter template, got %v", templates)
	}
	if _, err := os.Stat(filepath.Join(wikiRoot, "my-page.md")); err != nil {
		t.Fatalf("expected created page: %v", err)
	}
}

func TestServerManifestLintUpdateDeleteWiki(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	wikiRoot := filepath.Join(storageRoot, "wiki")

	if err := os.MkdirAll(wikiRoot, 0o755); err != nil {
		t.Fatalf("mkdir wiki: %v", err)
	}
	if err := wiki.SavePage(filepath.Join(wikiRoot, "alpha.md"), wiki.Page{
		Title:   "Alpha",
		Slug:    "alpha",
		Section: "General",
		Status:  "published",
		Content: "# Alpha\n",
	}); err != nil {
		t.Fatalf("save page: %v", err)
	}
	if err := wiki.SaveIndexContext(context.Background(), filepath.Join(wikiRoot, "_index.yaml"), wiki.Index{
		Sections: []wiki.IndexSection{{Title: "General", Slug: "", Pages: []string{"alpha"}}},
	}); err != nil {
		t.Fatalf("save index: %v", err)
	}

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"manifest_wiki","params":{},"id":40}`,
		`{"jsonrpc":"2.0","method":"lint_wiki","params":{},"id":41}`,
		`{"jsonrpc":"2.0","method":"update_wiki_section","params":{"slug":"general","title":"General Updated"},"id":42}`,
		`{"jsonrpc":"2.0","method":"delete_wiki_page","params":{"slug":"alpha","update_index":true},"id":43}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 4 {
		t.Fatalf("expected 4 responses, got %d", len(responses))
	}
	if responses[0].Error != nil || responses[1].Error != nil || responses[2].Error != nil || responses[3].Error != nil {
		t.Fatalf("unexpected errors: %+v %+v %+v %+v", responses[0].Error, responses[1].Error, responses[2].Error, responses[3].Error)
	}
	issues := responses[1].Result.([]any)
	if len(issues) != 0 {
		t.Fatalf("expected no lint issues, got %v", issues)
	}
}

func TestServerReadTask(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}
	task, err := board.NewTask("Read me")
	if err != nil {
		t.Fatalf("new task: %v", err)
	}
	created, err := repo.CreateTaskContext(context.Background(), task)
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	input := `{"jsonrpc":"2.0","method":"read_task","params":{"id":"` + created.ID + `"},"id":50}`
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	if result["id"] != created.ID {
		t.Fatalf("expected task id, got %v", result["id"])
	}
	if result["board_id"] != "default" {
		t.Fatalf("expected board_id default, got %v", result["board_id"])
	}
}

func TestServerWikiLintReportsIssue(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	wikiRoot := filepath.Join(storageRoot, "wiki")

	if err := os.MkdirAll(wikiRoot, 0o755); err != nil {
		t.Fatalf("mkdir wiki: %v", err)
	}
	if err := wiki.SavePage(filepath.Join(wikiRoot, "alpha.md"), wiki.Page{
		Title:   "Alpha",
		Slug:    "alpha",
		Status:  "published",
		Content: "no heading\n",
	}); err != nil {
		t.Fatalf("save page: %v", err)
	}

	output := runServerWithStorage(t, baseDir, storageRoot, `{"jsonrpc":"2.0","method":"lint_wiki","params":{},"id":60}`)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	issues := responses[0].Result.([]any)
	if len(issues) == 0 {
		t.Fatalf("expected lint issues")
	}
}

func TestServerDeleteWikiPageRemovesIndexEntry(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	wikiRoot := filepath.Join(storageRoot, "wiki")

	if err := os.MkdirAll(wikiRoot, 0o755); err != nil {
		t.Fatalf("mkdir wiki: %v", err)
	}
	if err := wiki.SavePage(filepath.Join(wikiRoot, "alpha.md"), wiki.Page{
		Title:   "Alpha",
		Slug:    "alpha",
		Section: "General",
		Status:  "published",
		Content: "# Alpha\n",
	}); err != nil {
		t.Fatalf("save page: %v", err)
	}
	if err := wiki.SaveIndexContext(context.Background(), filepath.Join(wikiRoot, "_index.yaml"), wiki.Index{
		Sections: []wiki.IndexSection{{Title: "General", Slug: "", Pages: []string{"alpha"}}},
	}); err != nil {
		t.Fatalf("save index: %v", err)
	}

	output := runServerWithStorage(t, baseDir, storageRoot, `{"jsonrpc":"2.0","method":"delete_wiki_page","params":{"slug":"alpha","update_index":true},"id":70}`)
	responses := decodeResponses(t, output)
	if len(responses) != 1 || responses[0].Error != nil {
		t.Fatalf("unexpected delete response: %+v", responses)
	}

	index, err := wiki.LoadIndexContext(context.Background(), filepath.Join(wikiRoot, "_index.yaml"))
	if err != nil {
		t.Fatalf("load index: %v", err)
	}
	if len(index.Sections) == 0 || len(index.Sections[0].Pages) != 0 {
		t.Fatalf("expected index pages removed, got %+v", index.Sections)
	}
}

func TestServerWriteAndReadWikiPage(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	wikiRoot := filepath.Join(storageRoot, "wiki")
	if err := os.MkdirAll(wikiRoot, 0o755); err != nil {
		t.Fatalf("mkdir wiki: %v", err)
	}

	title := "New Page"
	content := "# New Page\\n"
	section := "Docs"
	status := "published"
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"write_wiki_page","params":{"slug":"new-page","title":"` + title + `","section":"` + section + `","status":"` + status + `","content":"` + content + `"},"id":80}`,
		`{"jsonrpc":"2.0","method":"read_wiki_page","params":{"slug":"new-page"},"id":81}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(responses))
	}
	if responses[0].Error != nil || responses[1].Error != nil {
		t.Fatalf("unexpected errors: %+v %+v", responses[0].Error, responses[1].Error)
	}
	read := responses[1].Result.(map[string]any)
	if read["slug"] != "new-page" {
		t.Fatalf("expected slug new-page, got %v", read["slug"])
	}
	if read["title"] != title {
		t.Fatalf("expected title, got %v", read["title"])
	}
}

func TestServerUpdateTaskFields(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	createOut := runServerWithStorage(t, baseDir, storageRoot, `{"jsonrpc":"2.0","method":"create_task","params":{"title":"Orig"},"id":90}`)
	createResp := decodeResponses(t, createOut)
	if len(createResp) != 1 || createResp[0].Error != nil {
		t.Fatalf("unexpected create response: %+v", createResp)
	}
	created := createResp[0].Result.(map[string]any)
	taskID := created["id"].(string)

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","method":"update_task_title","params":{"id":"` + taskID + `","title":"New"},"id":91}`,
		`{"jsonrpc":"2.0","method":"update_task_tags","params":{"id":"` + taskID + `","tags":["a","b"]},"id":92}`,
		`{"jsonrpc":"2.0","method":"update_task_priority","params":{"id":"` + taskID + `","priority":2},"id":93}`,
		`{"jsonrpc":"2.0","method":"update_task_content","params":{"id":"` + taskID + `","content":"hello"},"id":94}`,
		`{"jsonrpc":"2.0","method":"read_task","params":{"id":"` + taskID + `"},"id":95}`,
	}, "\n")
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 5 {
		t.Fatalf("expected 5 responses, got %d", len(responses))
	}
	last := responses[len(responses)-1]
	if last.Error != nil {
		t.Fatalf("unexpected error: %+v", last.Error)
	}
	read := last.Result.(map[string]any)
	content := read["content"].(string)
	if !strings.Contains(content, "hello") {
		t.Fatalf("expected updated content, got %v", content)
	}
}

func runServerWithStorage(t *testing.T, baseDir, storageRoot, input string) string {
	t.Helper()
	server, err := NewServer(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	var out bytes.Buffer
	if err := server.Serve(strings.NewReader(input), &out); err != nil {
		t.Fatalf("serve error: %v", err)
	}
	return out.String()
}
