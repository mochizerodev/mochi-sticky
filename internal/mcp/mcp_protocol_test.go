package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"mochi-sticky/internal/board"
)

func TestMCPInitializeResultShape(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	input := `{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"opencode","version":"test"},"capabilities":{"roots":{"listChanged":true}},"roots":[{"uri":"file://` + baseDir + `","name":"workspace"}]},"id":1}`
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}

	result := responses[0].Result.(map[string]any)
	if _, ok := result["protocolVersion"].(string); !ok {
		t.Fatalf("expected protocolVersion string, got %T", result["protocolVersion"])
	}
	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatalf("expected serverInfo object, got %T", result["serverInfo"])
	}
	if _, ok := serverInfo["name"].(string); !ok {
		t.Fatalf("expected serverInfo.name string, got %T", serverInfo["name"])
	}
	if _, ok := serverInfo["version"].(string); !ok {
		t.Fatalf("expected serverInfo.version string, got %T", serverInfo["version"])
	}

	capabilities, ok := result["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("expected capabilities object, got %T", result["capabilities"])
	}
	if _, ok := capabilities["tools"].(map[string]any); !ok {
		t.Fatalf("expected capabilities.tools object, got %T", capabilities["tools"])
	}
	if _, ok := capabilities["resources"].(map[string]any); !ok {
		t.Fatalf("expected capabilities.resources object, got %T", capabilities["resources"])
	}
}

func TestMCPToolsListUsesInputSchema(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	input := `{"jsonrpc":"2.0","method":"tools/list","id":1}`
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	tools, ok := result["tools"].([]any)
	if !ok || len(tools) == 0 {
		t.Fatalf("expected tools list, got %T (%v)", result["tools"], result["tools"])
	}
	for i, tool := range tools {
		m, ok := tool.(map[string]any)
		if !ok {
			t.Fatalf("expected tools[%d] object, got %T", i, tool)
		}
		if _, ok := m["inputSchema"].(map[string]any); !ok {
			t.Fatalf("expected tools[%d].inputSchema object, got %T (%v)", i, m["inputSchema"], m["inputSchema"])
		}
	}

	var listTasks map[string]any
	for _, tool := range tools {
		m, ok := tool.(map[string]any)
		if !ok {
			continue
		}
		if m["name"] == "list_tasks" {
			listTasks = m
			break
		}
	}
	if listTasks == nil {
		t.Fatalf("expected list_tasks tool descriptor")
	}
	if _, ok := listTasks["inputSchema"].(map[string]any); !ok {
		t.Fatalf("expected list_tasks.inputSchema object, got %T", listTasks["inputSchema"])
	}
	if _, ok := listTasks["input_schema"]; ok {
		t.Fatalf("did not expect legacy input_schema key in tools/list")
	}
}

func TestMCPToolsCallWrapsResultInContent(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	input := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"list_boards","arguments":{}},"id":1}`
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("expected content array, got %T (%v)", result["content"], result["content"])
	}
	first, ok := content[0].(map[string]any)
	if !ok {
		t.Fatalf("expected content[0] object, got %T", content[0])
	}
	if first["type"] != "text" {
		t.Fatalf("expected content[0].type text, got %v", first["type"])
	}
	text, ok := first["text"].(string)
	if !ok || text == "" {
		t.Fatalf("expected content[0].text string, got %T (%v)", first["text"], first["text"])
	}
	var payload any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("expected tool output JSON text, unmarshal error: %v", err)
	}
	obj := payload.(map[string]any)
	if obj["active"] != "default" {
		t.Fatalf("expected active default, got %v", obj["active"])
	}
}

func TestMCPResourcesReadWrapsContents(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")

	repo, err := board.NewRepositoryWithStorage(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	if err := repo.InitStoreContext(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	input := `{"jsonrpc":"2.0","method":"resources/read","params":{"uri":"config://active"},"id":1}`
	output := runServerWithStorage(t, baseDir, storageRoot, input)
	responses := decodeResponses(t, output)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	if responses[0].Error != nil {
		t.Fatalf("unexpected error: %+v", responses[0].Error)
	}
	result := responses[0].Result.(map[string]any)
	contents, ok := result["contents"].([]any)
	if !ok || len(contents) != 1 {
		t.Fatalf("expected single contents entry, got %T (%v)", result["contents"], result["contents"])
	}
	entry := contents[0].(map[string]any)
	if entry["uri"] != "config://active" {
		t.Fatalf("expected uri config://active, got %v", entry["uri"])
	}
	text, ok := entry["text"].(string)
	if !ok || text == "" {
		t.Fatalf("expected contents[0].text string, got %T (%v)", entry["text"], entry["text"])
	}
	var payload any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		t.Fatalf("expected resource JSON text, unmarshal error: %v", err)
	}
	obj := payload.(map[string]any)
	if obj["board_id"] != "default" {
		t.Fatalf("expected board_id default, got %v", obj["board_id"])
	}
}
