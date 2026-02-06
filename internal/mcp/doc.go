// Package mcp implements the Model Context Protocol helper server that exposes
// board/task and wiki operations over JSON-RPC 2.0 on stdin/stdout.
// Tools such as `mochi-sticky mcp` instantiate this server so agents can list
// and query boards, create or mutate tasks, read and write wiki pages, export
// content, and keep the workspace in sync with external tooling.
//
// The exported Server simply wraps a workspace root and storage directory, then
// dispatches a fixed set of JSON-RPC methods (e.g., initialize/handshake,
// list_tasks, read_wiki, create_task) over stdin/stdout. Each method maps cleanly
// to the existing board and wiki domain packages so agents never write to stdout
// or stderr directly and can rely on the same filtering, sorting, and validation
// logic the CLI and TUI already use.
//
// This package also exposes descriptors for the available tools and resources
// so agents can self-document their capabilities and support optional schema
// validation before they submit payloads.
package mcp
