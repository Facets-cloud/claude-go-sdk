# Claude Agent SDK for Go

## Project Structure
- Root package `claudeagent` contains core SDK (query, options, messages, etc.)
- `tools/` package contains tool input/output type definitions
- `bridge/` package contains the alpha bridge session API
- `browser/` package contains WebSocket-based browser transport
- `internal/` contains unexported implementation details
- `references/` contains the TypeScript SDK source for reference (read-only)
- `examples/` contains runnable usage examples
- `demos/` contains ported demos from anthropics/claude-agent-sdk-demos

## Conventions
- Follow standard Go conventions: exported types are PascalCase, unexported are camelCase
- Use `encoding/json` struct tags for all serializable types
- Union types use an interface with unexported marker method + concrete struct implementations
- Optional fields use pointer types (*string, *int, *bool) with `omitempty`
- Error handling follows Go idioms: return (value, error) pairs
- Context is threaded through for cancellation support
- Channels replace async generators for streaming

## Testing Requirements (MANDATORY)
Every change MUST include:

1. **Unit tests** — run `go test -timeout 60s ./...` before committing
2. **Live tests** — run `go test -tags live -timeout 120s -run TestLive ./...` for any change to query.go, process.go, control.go, or mcptools.go
3. **Server-context tests** — if changing pipe handling or subprocess management, test from `demos/ws-server/` to verify it works inside HTTP/WebSocket servers

Test commands:
```bash
# Unit tests (always run)
go test -timeout 60s ./...

# Live tests (requires Claude CLI + auth)
go test -tags live -v -timeout 120s -run TestLive ./...

# MCP live tests
go test -tags live -v -timeout 120s -run TestLive_Mcp ./...

# Server context test
go run ./demos/ws-server/ &
curl http://localhost:8765/test
```

## Key Architecture: Single-Turn vs Multi-Turn

**Single-Turn** (string prompt): Creates a new CLI process per query. stdin is closed after result.
```go
q := claudeagent.NewQuery(claudeagent.QueryParams{
    Prompt: "Hello",  // string = single-turn
})
```

**Multi-Turn** (channel prompt): CLI process stays alive. Multiple messages sent via channel.
```go
input := make(chan claudeagent.SDKUserMessage, 1)
q := claudeagent.NewQuery(claudeagent.QueryParams{
    Prompt: input,  // channel = multi-turn, CLI stays alive
})
// Send messages on input channel...
```

**IMPORTANT for server developers**: If you need multi-turn conversations (WebSocket chat), use the channel pattern. Creating a new `NewQuery` with string prompt per message will spawn a new CLI process each time and lose conversation context.

## CLI Protocol
The SDK uses `--print --output-format stream-json --verbose --input-format stream-json` for bidirectional communication with the Claude Code CLI. The `--print` flag is REQUIRED for `--input-format` to work.

## Dependencies
- `github.com/mark3labs/mcp-go` for in-process MCP server support
- `github.com/google/uuid` for UUID generation
- `golang.org/x/net/websocket` for WebSocket server demo
