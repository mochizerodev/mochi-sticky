package mcp

import (
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"testing"
	"time"
)

// TestPersistentConnection tests that the server can handle multiple sequential requests
// over the same connection, mimicking VSCode/Copilot behavior
func TestPersistentConnection(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	server, err := NewServer(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	// Create pipe for bidirectional communication
	serverIn, clientOut := io.Pipe()
	clientIn, serverOut := io.Pipe()

	// Start server in background
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.ServeContext(ctx, serverIn, serverOut)
	}()

	// Send multiple requests and verify responses
	encoder := json.NewEncoder(clientOut)
	decoder := json.NewDecoder(clientIn)

	// Request 1: initialize
	req1 := map[string]any{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"id":      1,
	}
	if err := encoder.Encode(req1); err != nil {
		t.Fatalf("encode req1: %v", err)
	}

	var resp1 rpcResponse
	if err := decoder.Decode(&resp1); err != nil {
		t.Fatalf("decode resp1: %v", err)
	}
	if resp1.Error != nil {
		t.Fatalf("initialize error: %+v", resp1.Error)
	}

	// Request 2: list_tools
	req2 := map[string]any{
		"jsonrpc": "2.0",
		"method":  "list_tools",
		"id":      2,
	}
	if err := encoder.Encode(req2); err != nil {
		t.Fatalf("encode req2: %v", err)
	}

	var resp2 rpcResponse
	if err := decoder.Decode(&resp2); err != nil {
		t.Fatalf("decode resp2: %v", err)
	}
	if resp2.Error != nil {
		t.Fatalf("list_tools error: %+v", resp2.Error)
	}

	// Request 3: list_resources
	req3 := map[string]any{
		"jsonrpc": "2.0",
		"method":  "list_resources",
		"id":      3,
	}
	if err := encoder.Encode(req3); err != nil {
		t.Fatalf("encode req3: %v", err)
	}

	var resp3 rpcResponse
	if err := decoder.Decode(&resp3); err != nil {
		t.Fatalf("decode resp3: %v", err)
	}
	if resp3.Error != nil {
		t.Fatalf("list_resources error: %+v", resp3.Error)
	}

	// Close client side to signal EOF
	_ = clientOut.Close()

	// Wait for server to finish
	select {
	case err := <-serverDone:
		if err != nil && err != io.EOF {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server didn't exit after EOF")
	}
}

// TestShutdownProtocol tests the shutdown and exit methods
func TestShutdownProtocol(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	server, err := NewServer(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	serverIn, clientOut := io.Pipe()
	clientIn, serverOut := io.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.ServeContext(ctx, serverIn, serverOut)
	}()

	encoder := json.NewEncoder(clientOut)
	decoder := json.NewDecoder(clientIn)

	// Send shutdown request
	shutdownReq := map[string]any{
		"jsonrpc": "2.0",
		"method":  "shutdown",
		"id":      1,
	}
	if err := encoder.Encode(shutdownReq); err != nil {
		t.Fatalf("encode shutdown: %v", err)
	}

	var resp rpcResponse
	if err := decoder.Decode(&resp); err != nil {
		t.Fatalf("decode shutdown response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("shutdown error: %+v", resp.Error)
	}

	// Server should exit after shutdown
	select {
	case err := <-serverDone:
		if err != nil && err != io.EOF && err != context.DeadlineExceeded {
			t.Fatalf("server error after shutdown: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server didn't exit after shutdown")
	}
}

// TestExitNotification tests that exit notification triggers server shutdown
func TestExitNotification(t *testing.T) {
	baseDir := t.TempDir()
	storageRoot := filepath.Join(baseDir, "storage")
	server, err := NewServer(baseDir, storageRoot)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	serverIn, clientOut := io.Pipe()
	_, serverOut := io.Pipe()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.ServeContext(ctx, serverIn, serverOut)
	}()

	encoder := json.NewEncoder(clientOut)

	// Send exit notification (no id = notification, no response expected)
	exitNotification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "exit",
	}
	if err := encoder.Encode(exitNotification); err != nil {
		t.Fatalf("encode exit: %v", err)
	}

	// Server should exit quickly
	select {
	case err := <-serverDone:
		if err != nil && err != io.EOF && err != context.DeadlineExceeded {
			t.Fatalf("server error after exit: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server didn't exit after exit notification")
	}
}
