# Claude Agent SDK for Go

## Project Structure
- Root package `claudeagent` contains core SDK (query, options, messages, etc.)
- `tools/` package contains tool input/output type definitions
- `bridge/` package contains the alpha bridge session API
- `browser/` package contains WebSocket-based browser transport
- `internal/` contains unexported implementation details
- `references/` contains the TypeScript SDK source for reference (read-only)
- `examples/` contains runnable usage examples

## Conventions
- Follow standard Go conventions: exported types are PascalCase, unexported are camelCase
- Use `encoding/json` struct tags for all serializable types
- Union types use an interface with unexported marker method + concrete struct implementations
- Optional fields use pointer types (*string, *int, *bool) with `omitempty`
- Error handling follows Go idioms: return (value, error) pairs
- Context is threaded through for cancellation support
- Channels replace async generators for streaming

## Testing
- Run all tests: `go test ./...`
- Run specific package: `go test ./tools/...`
- Integration tests require Claude Code CLI installed (tagged with `//go:build integration`)

## Dependencies
- Minimal external deps. Use stdlib where possible.
- `nhooyr.io/websocket` for browser package WebSocket support
- `github.com/google/uuid` for UUID generation
