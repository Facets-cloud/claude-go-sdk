# Claude Agent SDK for Go

![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go) ![Parity](https://img.shields.io/badge/TypeScript_SDK_parity-v0.2.81-blue?style=flat-square) ![Tests](https://img.shields.io/badge/tests-318_passing-brightgreen?style=flat-square)

An **unofficial** Go port of Anthropic's [Claude Agent SDK](https://github.com/anthropics/claude-agent-sdk-typescript) (TypeScript). Build AI agents with Claude Code's capabilities — autonomous agents that can understand codebases, edit files, run commands, and execute complex workflows.

> **Note:** This is a community-maintained Go port, not an official Anthropic product. For the official SDK, see [@anthropic-ai/claude-agent-sdk](https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk).

## Get started

```bash
go get github.com/Facets-cloud/claude-go-sdk
```

### Prerequisites

- Go 1.22+
- [Claude Code CLI](https://docs.claude.com/en/docs/claude-code) installed and authenticated

### Basic usage

```go
package main

import (
    "fmt"
    claudeagent "github.com/Facets-cloud/claude-go-sdk"
)

func main() {
    q := claudeagent.NewQuery(claudeagent.QueryParams{
        Prompt: "What is 2 + 2?",
        Options: &claudeagent.Options{
            MaxTurns: claudeagent.Int(1),
        },
    })
    defer q.Close()

    for msg := range q.Messages() {
        switch m := msg.(type) {
        case *claudeagent.SDKResultSuccess:
            fmt.Println("Result:", m.Result)
        case *claudeagent.SDKResultError:
            fmt.Println("Error:", m.Errors)
        }
    }
}
```

### Streaming multi-turn conversation

```go
input := make(chan claudeagent.SDKUserMessage)

q := claudeagent.NewQuery(claudeagent.QueryParams{
    Prompt: input,
    Options: &claudeagent.Options{
        SystemPrompt: "You are a helpful coding assistant.",
    },
})
defer q.Close()

// Send messages on the input channel
go func() {
    input <- claudeagent.SDKUserMessage{
        Type:      "user",
        Message:   json.RawMessage(`{"role":"user","content":"Explain goroutines"}`),
        SessionID: "session-1",
    }
}()

for msg := range q.Messages() {
    // Handle messages...
}
```

### Custom permissions

```go
q := claudeagent.NewQuery(claudeagent.QueryParams{
    Prompt: "Run the test suite",
    Options: &claudeagent.Options{
        CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts claudeagent.CanUseToolOptions) (claudeagent.PermissionResult, error) {
            if toolName == "Bash" {
                return claudeagent.PermissionResultAllow{Behavior: claudeagent.PermissionBehaviorAllow}, nil
            }
            return claudeagent.PermissionResultDeny{Behavior: claudeagent.PermissionBehaviorDeny, Message: "denied"}, nil
        },
    },
})
```

### Session management

```go
// List past sessions
sessions, _ := claudeagent.ListSessions(&claudeagent.ListSessionsOptions{
    Dir:   claudeagent.String("/path/to/project"),
    Limit: claudeagent.Int(10),
})

// Fork a session
result, _ := claudeagent.ForkSession("session-uuid", &claudeagent.ForkSessionOptions{
    Title: claudeagent.String("experiment branch"),
})
```

## Architecture

```
+----------------------------------------------------------+
|                   claude-go-sdk                           |
|                                                          |
|  claudeagent/   Core SDK (query, messages, options)      |
|  tools/         Tool input/output type definitions       |
|  bridge/        Alpha bridge session API                 |
|  browser/       WebSocket transport for browsers         |
|  examples/      6 runnable usage examples                |
+----------------------------------------------------------+
         |
         v
+----------------------------------------------------------+
|        Claude Code CLI (subprocess via os/exec)          |
|        JSON-over-stdin/stdout protocol                   |
+----------------------------------------------------------+
```

The SDK spawns the Claude Code CLI as a subprocess and communicates via JSON lines over stdin/stdout. Go channels replace TypeScript's AsyncGenerator for streaming.

## Packages

| Package | Description | Coverage |
|---|---|---|
| `claudeagent` | Core SDK — query, messages, options, hooks, permissions, settings | 85.9% |
| `tools` | Tool input/output schemas (Bash, Read, Edit, Glob, Grep, etc.) | 100% |
| `bridge` | Alpha bridge session API for claude.ai integration | 100% |
| `browser` | WebSocket-based browser transport | Types only |

## Feature parity with TypeScript SDK

This Go SDK tracks the official TypeScript SDK v0.2.81 with full type and function parity:

- 150+ exported types with JSON serialization
- All 24 SDKMessage types with discriminated union parsing
- All 23 hook event types with callback support
- Complete Options struct (~50 fields)
- Full Settings struct (~80 fields)
- All control request/response types (28 subtypes)
- All tool input/output schemas
- Session management (list, get, fork, rename, tag)
- V2 Session API (alpha)
- Bridge API (alpha)
- Browser/WebSocket transport

### Keeping in sync

A `/sync-upstream` slash command is included for Claude Code users to check for and apply updates from the TypeScript SDK. See `docs/SYNC_PROMPT.md` for the manual process.

## Examples

See the [examples/](examples/) directory:

| Example | Description |
|---|---|
| [basic](examples/basic/) | Simple single-turn query |
| [streaming](examples/streaming/) | Multi-turn streaming conversation |
| [custom_tools](examples/custom_tools/) | MCP server with custom tools |
| [permissions](examples/permissions/) | Custom permission handler |
| [hooks](examples/hooks/) | Hook callbacks for lifecycle events |
| [session_management](examples/session_management/) | List, fork, and resume sessions |

## Testing

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests (requires Claude Code CLI)
go test -tags integration -timeout 120s ./...
```

## Reporting bugs

File a [GitHub issue](https://github.com/Facets-cloud/claude-go-sdk/issues) to report bugs or request features.

## Upstream references

- [Official TypeScript SDK](https://github.com/anthropics/claude-agent-sdk-typescript)
- [Claude Agent SDK documentation](https://docs.claude.com/en/api/agent-sdk/overview)
- [Claude Code](https://docs.claude.com/en/docs/claude-code)

## License

This project is an independent community port. Use of the underlying Claude Code CLI is governed by Anthropic's [Commercial Terms of Service](https://www.anthropic.com/legal/commercial-terms).
