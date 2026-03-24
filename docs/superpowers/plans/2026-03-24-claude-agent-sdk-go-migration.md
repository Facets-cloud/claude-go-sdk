# Claude Agent SDK — Go Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete, idiomatic Go SDK (`github.com/anthropics/claude-agent-sdk-go`) that provides 1:1 feature parity with the TypeScript `@anthropic-ai/claude-agent-sdk` v0.2.81, enabling Go developers to build AI agents with Claude Code's capabilities.

**Architecture:** The SDK wraps the Claude Code CLI as a subprocess, communicating via JSON-over-stdin/stdout. Go channels replace TypeScript's AsyncGenerator for streaming. The SDK is organized into packages: `claudeagent` (core), `claudeagent/types` (all message/config types), `claudeagent/tools` (tool input/output schemas), `claudeagent/bridge` (alpha bridge API), and `claudeagent/browser` (WebSocket transport). Process lifecycle, JSON serialization, and control request/response correlation are handled internally.

**Tech Stack:** Go 1.22+, `encoding/json`, `os/exec`, `nhooyr.io/websocket` (browser transport), `github.com/google/uuid`, standard library for everything else. No CGo. No heavy frameworks.

**Reference TypeScript source:** `references/` directory + `package/sdk.d.ts`, `package/sdk-tools.d.ts`, `package/bridge.d.ts`, `package/browser-sdk.d.ts`

---

## Architecture Overview

```
+------------------------------------------------------------------+
|                     claude-agent-sdk-go                            |
|                                                                    |
|  +-----------+  +--------+  +---------+  +--------+  +---------+ |
|  | claudeagent|  | types  |  | tools   |  | bridge |  | browser | |
|  | (core)     |  |        |  |         |  | (alpha)|  |         | |
|  +-----------+  +--------+  +---------+  +--------+  +---------+ |
|       |              |            |            |            |      |
|       v              v            v            v            v      |
|  +----------------------------------------------------------+    |
|  |              Internal: process management                  |    |
|  |   stdin/stdout JSON  |  control request/response           |    |
|  |   message routing    |  correlation via request_id         |    |
|  +----------------------------------------------------------+    |
|       |                                                           |
|       v                                                           |
|  +----------------------------------------------------------+    |
|  |          Claude Code CLI subprocess                        |    |
|  |          (spawned via os/exec or custom SpawnFunc)         |    |
|  +----------------------------------------------------------+    |
+------------------------------------------------------------------+
```

```
TypeScript → Go Mapping:
  AsyncGenerator<SDKMessage>  →  <-chan SDKMessage (+ Query struct with methods)
  Promise<T>                  →  (T, error)
  AbortController             →  context.Context with cancel
  AsyncIterable<SDKUserMessage> → <-chan SDKUserMessage
  interface Query extends ...  →  type Query struct with methods
  union types (A | B | C)     →  interface with unexported marker + concrete structs
  optional fields              →  pointer types (*string, *int) or omitempty
  Record<string, T>           →  map[string]T
  Partial<Record<K, V>>       →  map[K]V (Go maps are inherently partial)
  zod schemas                 →  struct tags + validation functions
```

---

## File Structure

```
claude-agent-sdk-go/
├── references/                          # Cloned TypeScript SDK (read-only reference)
├── docs/
│   └── superpowers/plans/               # This plan
├── go.mod                               # Module: github.com/anthropics/claude-agent-sdk-go
├── go.sum
│
├── claudeagent.go                       # Package doc, version const, top-level query() + session functions
├── options.go                           # Options struct, PermissionMode, SystemPrompt, etc.
├── query.go                             # Query struct (channel-based streaming + control methods)
├── process.go                           # Subprocess spawn, stdin/stdout management, lifecycle
├── control.go                           # Control request/response types and correlation engine
├── messages.go                          # SDKMessage union type + all concrete message types
├── hooks.go                             # Hook event types, HookCallback, HookCallbackMatcher
├── permissions.go                       # CanUseTool, PermissionResult, PermissionUpdate types
├── mcp.go                               # MCP server config types, McpServerStatus, McpSetServersResult
├── settings.go                          # Settings struct (large — all settings fields)
├── session.go                           # Session management: list, get, fork, rename, tag, messages
├── session_v2.go                        # V2 Session API (alpha): SDKSession, create, resume
├── models.go                            # ModelInfo, ModelUsage, AccountInfo, AgentInfo, AgentDefinition
├── sandbox.go                           # SandboxSettings, SandboxNetworkConfig, SandboxFilesystemConfig
├── errors.go                            # AbortError, custom error types
├── embed.go                             # CLI path resolution (equivalent to embed.d.ts)
├── json.go                              # JSON marshal/unmarshal helpers for union types
│
├── tools/                               # Tool input/output schemas (sdk-tools.d.ts equivalent)
│   ├── tools.go                         # Package doc
│   ├── agent.go                         # AgentInput, AgentOutput
│   ├── bash.go                          # BashInput, BashOutput
│   ├── file.go                          # FileReadInput/Output, FileEditInput/Output, FileWriteInput/Output
│   ├── glob.go                          # GlobInput, GlobOutput
│   ├── grep.go                          # GrepInput, GrepOutput
│   ├── notebook.go                      # NotebookEditInput, NotebookEditOutput
│   ├── mcp_tools.go                     # McpInput, McpOutput, ListMcpResourcesInput/Output, etc.
│   ├── web.go                           # WebFetchInput/Output, WebSearchInput/Output
│   ├── task.go                          # TaskOutputInput, TaskStopInput/Output
│   ├── todo.go                          # TodoWriteInput, TodoWriteOutput
│   ├── ask.go                           # AskUserQuestionInput, AskUserQuestionOutput
│   ├── config.go                        # ConfigInput, ConfigOutput
│   ├── worktree.go                      # EnterWorktreeInput/Output, ExitWorktreeInput/Output
│   └── plan.go                          # ExitPlanModeInput, ExitPlanModeOutput
│
├── bridge/                              # Bridge API (alpha — package/bridge.d.ts)
│   ├── bridge.go                        # AttachBridgeSession, BridgeSessionHandle
│   └── types.go                         # SessionState, AttachBridgeSessionOptions
│
├── browser/                             # Browser/WebSocket transport (package/browser-sdk.d.ts)
│   ├── browser.go                       # Query function (WebSocket-based)
│   └── types.go                         # BrowserQueryOptions, WebSocketOptions, AuthMessage
│
├── internal/                            # Unexported implementation details
│   ├── jsonrpc/                         # Minimal JSON-RPC message types (for MCP)
│   │   └── jsonrpc.go
│   └── process/                         # Low-level process spawn and I/O
│       └── spawn.go
│
├── examples/                            # Usage examples
│   ├── basic/main.go                    # Simple single-turn query
│   ├── streaming/main.go                # Multi-turn streaming conversation
│   ├── custom_tools/main.go             # MCP server with custom tools
│   ├── permissions/main.go              # Custom permission handler
│   ├── hooks/main.go                    # Hook callbacks
│   └── session_management/main.go       # List, fork, resume sessions
│
├── claudeagent_test.go                  # Integration tests for query()
├── options_test.go                      # Options validation tests
├── query_test.go                        # Query method tests
├── process_test.go                      # Process management tests
├── control_test.go                      # Control request/response tests
├── messages_test.go                     # Message JSON round-trip tests
├── hooks_test.go                        # Hook type tests
├── permissions_test.go                  # Permission type tests
├── mcp_test.go                          # MCP config tests
├── settings_test.go                     # Settings serialization tests
├── session_test.go                      # Session management tests
├── session_v2_test.go                   # V2 session tests
├── models_test.go                       # Model/account info tests
├── sandbox_test.go                      # Sandbox config tests
├── json_test.go                         # JSON helper tests
├── tools/
│   ├── agent_test.go
│   ├── bash_test.go
│   ├── file_test.go
│   ├── glob_test.go
│   ├── grep_test.go
│   ├── web_test.go
│   └── ...                              # One test file per tool file
├── bridge/
│   └── bridge_test.go
└── browser/
    └── browser_test.go
```

---

## Task Breakdown

### Task 1: Project Scaffolding & Module Init

**Files:**
- Create: `go.mod`
- Create: `claudeagent.go`
- Create: `CLAUDE.md`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Users/anshulsao/Facets/ai/claude-go-sdk
go mod init github.com/anthropics/claude-agent-sdk-go
```

- [ ] **Step 2: Create package doc file with version constant**

Create `claudeagent.go`:
```go
// Package claudeagent provides a Go SDK for building AI agents with Claude Code's
// capabilities. It enables programmatic interaction with Claude to build autonomous
// agents that can understand codebases, edit files, run commands, and execute
// complex workflows.
//
// The SDK works by spawning the Claude Code CLI as a subprocess and communicating
// via JSON over stdin/stdout. Messages are streamed to the caller via Go channels.
//
// Basic usage:
//
//	q := claudeagent.Query(claudeagent.QueryParams{
//	    Prompt: "Explain the main function in this project",
//	})
//	for msg := range q.Messages() {
//	    fmt.Println(msg)
//	}
package claudeagent

// Version is the SDK version, tracking parity with the TypeScript SDK.
const Version = "0.2.81"

// ClaudeCodeVersion is the minimum compatible Claude Code CLI version.
const ClaudeCodeVersion = "2.1.81"
```

- [ ] **Step 3: Create CLAUDE.md for the project**

Create `CLAUDE.md`:
```markdown
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
```

- [ ] **Step 4: Commit**

```bash
git add go.mod claudeagent.go CLAUDE.md
git commit -m "feat: initialize Go SDK module with package doc and version constants"
```

---

### Task 2: Core Error Types

**Files:**
- Create: `errors.go`
- Create: `errors_test.go`

- [ ] **Step 1: Write failing test for AbortError**

Create `errors_test.go`:
```go
package claudeagent

import (
	"errors"
	"testing"
)

func TestAbortError(t *testing.T) {
	err := &AbortError{Message: "operation cancelled"}
	if err.Error() != "operation cancelled" {
		t.Errorf("got %q, want %q", err.Error(), "operation cancelled")
	}

	var target *AbortError
	if !errors.As(err, &target) {
		t.Error("AbortError should be matchable with errors.As")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestAbortError -v ./...`
Expected: FAIL — AbortError not defined

- [ ] **Step 3: Implement AbortError**

Create `errors.go`:
```go
package claudeagent

// AbortError is returned when an operation is cancelled via context cancellation.
type AbortError struct {
	Message string
}

func (e *AbortError) Error() string {
	return e.Message
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -run TestAbortError -v ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add errors.go errors_test.go
git commit -m "feat: add AbortError type"
```

---

### Task 3: Enums and Constants

**Files:**
- Create: `enums.go`
- Create: `enums_test.go`

- [ ] **Step 1: Write failing tests for enum types**

Create `enums_test.go`:
```go
package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestPermissionMode_JSON(t *testing.T) {
	tests := []struct {
		mode PermissionMode
		json string
	}{
		{PermissionModeDefault, `"default"`},
		{PermissionModeAcceptEdits, `"acceptEdits"`},
		{PermissionModeBypassPermissions, `"bypassPermissions"`},
		{PermissionModePlan, `"plan"`},
		{PermissionModeDontAsk, `"dontAsk"`},
	}
	for _, tt := range tests {
		b, err := json.Marshal(tt.mode)
		if err != nil {
			t.Fatalf("Marshal(%v): %v", tt.mode, err)
		}
		if string(b) != tt.json {
			t.Errorf("Marshal(%v) = %s, want %s", tt.mode, b, tt.json)
		}
		var got PermissionMode
		if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", tt.json, err)
		}
		if got != tt.mode {
			t.Errorf("Unmarshal(%s) = %v, want %v", tt.json, got, tt.mode)
		}
	}
}

func TestExitReason_Values(t *testing.T) {
	expected := []ExitReason{
		ExitReasonClear, ExitReasonResume, ExitReasonLogout,
		ExitReasonPromptInputExit, ExitReasonOther, ExitReasonBypassPermissionsDisabled,
	}
	if len(expected) != 6 {
		t.Errorf("expected 6 exit reasons, got %d", len(expected))
	}
}

func TestHookEvent_Values(t *testing.T) {
	events := AllHookEvents()
	if len(events) != 23 {
		t.Errorf("expected 23 hook events, got %d", len(events))
	}
}

func TestPermissionBehavior_JSON(t *testing.T) {
	b, err := json.Marshal(PermissionBehaviorAllow)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"allow"` {
		t.Errorf("got %s, want %q", b, "allow")
	}
}

func TestFastModeState_JSON(t *testing.T) {
	b, err := json.Marshal(FastModeStateOff)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"off"` {
		t.Errorf("got %s, want %q", b, "off")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run "TestPermissionMode|TestExitReason|TestHookEvent|TestPermissionBehavior|TestFastModeState" -v ./...`
Expected: FAIL

- [ ] **Step 3: Implement all enum types**

Create `enums.go`:
```go
package claudeagent

// PermissionMode controls how tool executions are handled.
type PermissionMode string

const (
	PermissionModeDefault           PermissionMode = "default"
	PermissionModeAcceptEdits       PermissionMode = "acceptEdits"
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
	PermissionModePlan              PermissionMode = "plan"
	PermissionModeDontAsk           PermissionMode = "dontAsk"
)

// ExitReason describes why a session ended.
type ExitReason string

const (
	ExitReasonClear                     ExitReason = "clear"
	ExitReasonResume                    ExitReason = "resume"
	ExitReasonLogout                    ExitReason = "logout"
	ExitReasonPromptInputExit           ExitReason = "prompt_input_exit"
	ExitReasonOther                     ExitReason = "other"
	ExitReasonBypassPermissionsDisabled ExitReason = "bypass_permissions_disabled"
)

// HookEvent identifies a lifecycle event that hooks can intercept.
type HookEvent string

const (
	HookEventPreToolUse          HookEvent = "PreToolUse"
	HookEventPostToolUse         HookEvent = "PostToolUse"
	HookEventPostToolUseFailure  HookEvent = "PostToolUseFailure"
	HookEventNotification        HookEvent = "Notification"
	HookEventUserPromptSubmit    HookEvent = "UserPromptSubmit"
	HookEventSessionStart        HookEvent = "SessionStart"
	HookEventSessionEnd          HookEvent = "SessionEnd"
	HookEventStop                HookEvent = "Stop"
	HookEventStopFailure         HookEvent = "StopFailure"
	HookEventSubagentStart       HookEvent = "SubagentStart"
	HookEventSubagentStop        HookEvent = "SubagentStop"
	HookEventPreCompact          HookEvent = "PreCompact"
	HookEventPostCompact         HookEvent = "PostCompact"
	HookEventPermissionRequest   HookEvent = "PermissionRequest"
	HookEventSetup               HookEvent = "Setup"
	HookEventTeammateIdle        HookEvent = "TeammateIdle"
	HookEventTaskCompleted       HookEvent = "TaskCompleted"
	HookEventElicitation         HookEvent = "Elicitation"
	HookEventElicitationResult   HookEvent = "ElicitationResult"
	HookEventConfigChange        HookEvent = "ConfigChange"
	HookEventWorktreeCreate      HookEvent = "WorktreeCreate"
	HookEventWorktreeRemove      HookEvent = "WorktreeRemove"
	HookEventInstructionsLoaded  HookEvent = "InstructionsLoaded"
)

// AllHookEvents returns all valid hook event values.
func AllHookEvents() []HookEvent {
	return []HookEvent{
		HookEventPreToolUse, HookEventPostToolUse, HookEventPostToolUseFailure,
		HookEventNotification, HookEventUserPromptSubmit, HookEventSessionStart,
		HookEventSessionEnd, HookEventStop, HookEventStopFailure,
		HookEventSubagentStart, HookEventSubagentStop, HookEventPreCompact,
		HookEventPostCompact, HookEventPermissionRequest, HookEventSetup,
		HookEventTeammateIdle, HookEventTaskCompleted, HookEventElicitation,
		HookEventElicitationResult, HookEventConfigChange, HookEventWorktreeCreate,
		HookEventWorktreeRemove, HookEventInstructionsLoaded,
	}
}

// PermissionBehavior describes how a permission rule acts.
type PermissionBehavior string

const (
	PermissionBehaviorAllow PermissionBehavior = "allow"
	PermissionBehaviorDeny  PermissionBehavior = "deny"
	PermissionBehaviorAsk   PermissionBehavior = "ask"
)

// FastModeState indicates whether fast mode is active.
type FastModeState string

const (
	FastModeStateOff      FastModeState = "off"
	FastModeStateCooldown FastModeState = "cooldown"
	FastModeStateOn       FastModeState = "on"
)

// SDKStatus represents system status states.
type SDKStatus *string

// SDKAssistantMessageError enumerates assistant message error types.
type SDKAssistantMessageError string

const (
	AssistantErrorAuthFailed       SDKAssistantMessageError = "authentication_failed"
	AssistantErrorBilling          SDKAssistantMessageError = "billing_error"
	AssistantErrorRateLimit        SDKAssistantMessageError = "rate_limit"
	AssistantErrorInvalidRequest   SDKAssistantMessageError = "invalid_request"
	AssistantErrorServer           SDKAssistantMessageError = "server_error"
	AssistantErrorUnknown          SDKAssistantMessageError = "unknown"
	AssistantErrorMaxOutputTokens  SDKAssistantMessageError = "max_output_tokens"
)

// ApiKeySource identifies where the API key came from.
type ApiKeySource string

const (
	ApiKeySourceUser      ApiKeySource = "user"
	ApiKeySourceProject   ApiKeySource = "project"
	ApiKeySourceOrg       ApiKeySource = "org"
	ApiKeySourceTemporary ApiKeySource = "temporary"
	ApiKeySourceOAuth     ApiKeySource = "oauth"
)

// SettingSource identifies a settings file location.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// ConfigScope identifies a configuration scope.
type ConfigScope string

const (
	ConfigScopeLocal   ConfigScope = "local"
	ConfigScopeUser    ConfigScope = "user"
	ConfigScopeProject ConfigScope = "project"
)

// OutputFormatType identifies output format types.
type OutputFormatType string

const (
	OutputFormatTypeJSONSchema OutputFormatType = "json_schema"
)

// PermissionUpdateDestination identifies where permission updates are stored.
type PermissionUpdateDestination string

const (
	PermissionDestUserSettings    PermissionUpdateDestination = "userSettings"
	PermissionDestProjectSettings PermissionUpdateDestination = "projectSettings"
	PermissionDestLocalSettings   PermissionUpdateDestination = "localSettings"
	PermissionDestSession         PermissionUpdateDestination = "session"
	PermissionDestCLIArg          PermissionUpdateDestination = "cliArg"
)

// SdkBeta identifies available beta features.
type SdkBeta string

const (
	SdkBetaContext1M SdkBeta = "context-1m-2025-08-07"
)
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test -run "TestPermissionMode|TestExitReason|TestHookEvent|TestPermissionBehavior|TestFastModeState" -v ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add enums.go enums_test.go
git commit -m "feat: add all enum/constant types (PermissionMode, HookEvent, ExitReason, etc.)"
```

---

### Task 4: JSON Union Type Helpers

**Files:**
- Create: `json.go`
- Create: `json_test.go`

- [ ] **Step 1: Write failing test for typed JSON union marshaling**

Create `json_test.go`:
```go
package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestRawJSONMessage_Unmarshal(t *testing.T) {
	raw := `{"type":"assistant","message":{},"parent_tool_use_id":null,"uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	if _, ok := msg.(*SDKAssistantMessage); !ok {
		t.Errorf("expected *SDKAssistantMessage, got %T", msg)
	}
}

func TestRawJSONMessage_UnmarshalResult(t *testing.T) {
	raw := `{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":50,"is_error":false,"num_turns":1,"result":"done","stop_reason":null,"total_cost_usd":0.01,"usage":{"input_tokens":10,"output_tokens":20,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"server_tool_use":null,"service_tier":null,"cache_creation":null},"modelUsage":{},"permission_denials":[],"uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	if result, ok := msg.(*SDKResultSuccess); !ok {
		t.Errorf("expected *SDKResultSuccess, got %T", msg)
	} else if result.Result != "done" {
		t.Errorf("result = %q, want %q", result.Result, "done")
	}
}

func TestRawJSONMessage_UnmarshalSystem(t *testing.T) {
	raw := `{"type":"system","subtype":"init","agents":[],"apiKeySource":"user","claude_code_version":"2.1.81","cwd":"/tmp","tools":["Bash"],"mcp_servers":[],"model":"claude-sonnet-4-6","permissionMode":"default","slash_commands":[],"output_style":"normal","skills":[],"plugins":[],"uuid":"abc","session_id":"s1"}`
	msg, err := ParseSDKMessage([]byte(raw))
	if err != nil {
		t.Fatalf("ParseSDKMessage: %v", err)
	}
	if sys, ok := msg.(*SDKSystemMessage); !ok {
		t.Errorf("expected *SDKSystemMessage, got %T", msg)
	} else if sys.ClaudeCodeVersion != "2.1.81" {
		t.Errorf("version = %q, want %q", sys.ClaudeCodeVersion, "2.1.81")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestRawJSONMessage -v ./...`
Expected: FAIL

- [ ] **Step 3: Implement ParseSDKMessage dispatcher**

Create `json.go`:
```go
package claudeagent

import (
	"encoding/json"
	"fmt"
)

// messageEnvelope is used to peek at type/subtype before full deserialization.
type messageEnvelope struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype,omitempty"`
}

// ParseSDKMessage deserializes raw JSON into the appropriate SDKMessage concrete type.
func ParseSDKMessage(data []byte) (SDKMessage, error) {
	var env messageEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("parse message envelope: %w", err)
	}

	switch env.Type {
	case "assistant":
		var m SDKAssistantMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse assistant message: %w", err)
		}
		return &m, nil

	case "user":
		// Check for replay marker
		var peek struct {
			IsReplay bool `json:"isReplay"`
		}
		_ = json.Unmarshal(data, &peek)
		if peek.IsReplay {
			var m SDKUserMessageReplay
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse user replay message: %w", err)
			}
			return &m, nil
		}
		var m SDKUserMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse user message: %w", err)
		}
		return &m, nil

	case "result":
		switch env.Subtype {
		case "success":
			var m SDKResultSuccess
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse result success: %w", err)
			}
			return &m, nil
		default:
			var m SDKResultError
			if err := json.Unmarshal(data, &m); err != nil {
				return nil, fmt.Errorf("parse result error: %w", err)
			}
			return &m, nil
		}

	case "system":
		return parseSystemMessage(env.Subtype, data)

	case "stream_event":
		var m SDKPartialAssistantMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse stream event: %w", err)
		}
		return &m, nil

	case "tool_progress":
		var m SDKToolProgressMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse tool progress: %w", err)
		}
		return &m, nil

	case "tool_use_summary":
		var m SDKToolUseSummaryMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse tool use summary: %w", err)
		}
		return &m, nil

	case "auth_status":
		var m SDKAuthStatusMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse auth status: %w", err)
		}
		return &m, nil

	case "rate_limit_event":
		var m SDKRateLimitEvent
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse rate limit event: %w", err)
		}
		return &m, nil

	case "prompt_suggestion":
		var m SDKPromptSuggestionMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse prompt suggestion: %w", err)
		}
		return &m, nil

	default:
		return nil, fmt.Errorf("unknown message type: %q", env.Type)
	}
}

func parseSystemMessage(subtype string, data []byte) (SDKMessage, error) {
	switch subtype {
	case "init":
		var m SDKSystemMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse system init: %w", err)
		}
		return &m, nil
	case "status":
		var m SDKStatusMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse status: %w", err)
		}
		return &m, nil
	case "api_retry":
		var m SDKAPIRetryMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse api retry: %w", err)
		}
		return &m, nil
	case "compact_boundary":
		var m SDKCompactBoundaryMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse compact boundary: %w", err)
		}
		return &m, nil
	case "local_command_output":
		var m SDKLocalCommandOutputMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse local command output: %w", err)
		}
		return &m, nil
	case "hook_started":
		var m SDKHookStartedMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse hook started: %w", err)
		}
		return &m, nil
	case "hook_progress":
		var m SDKHookProgressMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse hook progress: %w", err)
		}
		return &m, nil
	case "hook_response":
		var m SDKHookResponseMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse hook response: %w", err)
		}
		return &m, nil
	case "task_notification":
		var m SDKTaskNotificationMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse task notification: %w", err)
		}
		return &m, nil
	case "task_started":
		var m SDKTaskStartedMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse task started: %w", err)
		}
		return &m, nil
	case "task_progress":
		var m SDKTaskProgressMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse task progress: %w", err)
		}
		return &m, nil
	case "files_persisted":
		var m SDKFilesPersistedEvent
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse files persisted: %w", err)
		}
		return &m, nil
	case "elicitation_complete":
		var m SDKElicitationCompleteMessage
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse elicitation complete: %w", err)
		}
		return &m, nil
	default:
		// Unknown system subtype — return as raw
		return nil, fmt.Errorf("unknown system subtype: %q", subtype)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass** (will fail — messages.go not yet created)

This task depends on Task 5 (messages) to compile. Run after Task 5.

- [ ] **Step 5: Commit** (combined with Task 5)

---

### Task 5: SDK Message Types (Core)

**Files:**
- Create: `messages.go`
- Create: `messages_test.go`

This is the largest single task — defines ALL SDKMessage concrete types.

- [ ] **Step 1: Write failing test for SDKMessage interface and key types**

Create `messages_test.go`:
```go
package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestSDKAssistantMessage_JSON(t *testing.T) {
	raw := `{"type":"assistant","message":{"id":"msg_01","type":"message","role":"assistant","content":[{"type":"text","text":"hello"}],"model":"claude-sonnet-4-6","stop_reason":"end_turn","usage":{"input_tokens":10,"output_tokens":5}},"parent_tool_use_id":null,"uuid":"550e8400-e29b-41d4-a716-446655440000","session_id":"sess1"}`
	var msg SDKAssistantMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("UUID = %q", msg.UUID)
	}
	if msg.SessionID != "sess1" {
		t.Errorf("SessionID = %q", msg.SessionID)
	}
}

func TestSDKSystemMessage_JSON(t *testing.T) {
	raw := `{"type":"system","subtype":"init","agents":["general"],"apiKeySource":"user","claude_code_version":"2.1.81","cwd":"/tmp","tools":["Bash","Read"],"mcp_servers":[],"model":"claude-sonnet-4-6","permissionMode":"default","slash_commands":[],"output_style":"normal","skills":[],"plugins":[],"uuid":"abc","session_id":"s1"}`
	var msg SDKSystemMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.ClaudeCodeVersion != "2.1.81" {
		t.Errorf("version = %q", msg.ClaudeCodeVersion)
	}
	if len(msg.Tools) != 2 {
		t.Errorf("tools = %v", msg.Tools)
	}
}

func TestSDKResultSuccess_JSON(t *testing.T) {
	raw := `{"type":"result","subtype":"success","duration_ms":1000,"duration_api_ms":800,"is_error":false,"num_turns":3,"result":"All done","stop_reason":"end_turn","total_cost_usd":0.05,"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"server_tool_use":null,"service_tier":null,"cache_creation":null},"modelUsage":{},"permission_denials":[],"uuid":"abc","session_id":"s1"}`
	var msg SDKResultSuccess
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.Result != "All done" {
		t.Errorf("result = %q", msg.Result)
	}
	if msg.NumTurns != 3 {
		t.Errorf("num_turns = %d", msg.NumTurns)
	}
}

func TestSDKUserMessage_JSON(t *testing.T) {
	raw := `{"type":"user","message":{"role":"user","content":"hello"},"parent_tool_use_id":null,"session_id":"s1"}`
	var msg SDKUserMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if msg.SessionID != "s1" {
		t.Errorf("session_id = %q", msg.SessionID)
	}
}

func TestSDKMessage_Interface(t *testing.T) {
	// Verify all message types satisfy the SDKMessage interface
	var _ SDKMessage = &SDKAssistantMessage{}
	var _ SDKMessage = &SDKUserMessage{}
	var _ SDKMessage = &SDKUserMessageReplay{}
	var _ SDKMessage = &SDKResultSuccess{}
	var _ SDKMessage = &SDKResultError{}
	var _ SDKMessage = &SDKSystemMessage{}
	var _ SDKMessage = &SDKPartialAssistantMessage{}
	var _ SDKMessage = &SDKCompactBoundaryMessage{}
	var _ SDKMessage = &SDKStatusMessage{}
	var _ SDKMessage = &SDKAPIRetryMessage{}
	var _ SDKMessage = &SDKLocalCommandOutputMessage{}
	var _ SDKMessage = &SDKHookStartedMessage{}
	var _ SDKMessage = &SDKHookProgressMessage{}
	var _ SDKMessage = &SDKHookResponseMessage{}
	var _ SDKMessage = &SDKToolProgressMessage{}
	var _ SDKMessage = &SDKAuthStatusMessage{}
	var _ SDKMessage = &SDKTaskNotificationMessage{}
	var _ SDKMessage = &SDKTaskStartedMessage{}
	var _ SDKMessage = &SDKTaskProgressMessage{}
	var _ SDKMessage = &SDKFilesPersistedEvent{}
	var _ SDKMessage = &SDKToolUseSummaryMessage{}
	var _ SDKMessage = &SDKRateLimitEvent{}
	var _ SDKMessage = &SDKElicitationCompleteMessage{}
	var _ SDKMessage = &SDKPromptSuggestionMessage{}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run "TestSDK.*Message|TestSDKResult|TestSDKMessage_Interface" -v ./...`
Expected: FAIL

- [ ] **Step 3: Implement all SDK message types**

Create `messages.go`:
```go
package claudeagent

import "encoding/json"

// SDKMessage is the interface implemented by all message types streamed from the SDK.
// Use a type switch to handle specific message types.
type SDKMessage interface {
	sdkMessage() // unexported marker method
	// MessageType returns the wire "type" field value.
	MessageType() string
}

// --- Usage types ---

// NonNullableUsage contains token usage information with all fields non-nullable.
type NonNullableUsage struct {
	InputTokens                int              `json:"input_tokens"`
	OutputTokens               int              `json:"output_tokens"`
	CacheCreationInputTokens   int              `json:"cache_creation_input_tokens"`
	CacheReadInputTokens       int              `json:"cache_read_input_tokens"`
	ServerToolUse              *ServerToolUse    `json:"server_tool_use"`
	ServiceTier                *string           `json:"service_tier"`
	CacheCreation              *CacheCreation    `json:"cache_creation"`
}

type ServerToolUse struct {
	WebSearchRequests int `json:"web_search_requests"`
	WebFetchRequests  int `json:"web_fetch_requests"`
}

type CacheCreation struct {
	Ephemeral1hInputTokens int `json:"ephemeral_1h_input_tokens"`
	Ephemeral5mInputTokens int `json:"ephemeral_5m_input_tokens"`
}

// ModelUsage contains per-model usage statistics.
type ModelUsage struct {
	InputTokens              int     `json:"inputTokens"`
	OutputTokens             int     `json:"outputTokens"`
	CacheReadInputTokens     int     `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int     `json:"cacheCreationInputTokens"`
	WebSearchRequests        int     `json:"webSearchRequests"`
	CostUSD                  float64 `json:"costUSD"`
	ContextWindow            int     `json:"contextWindow"`
	MaxOutputTokens          int     `json:"maxOutputTokens"`
}

// SDKPermissionDenial records a denied tool use.
type SDKPermissionDenial struct {
	ToolName  string                 `json:"tool_name"`
	ToolUseID string                 `json:"tool_use_id"`
	ToolInput map[string]interface{} `json:"tool_input"`
}

// --- Assistant Message ---

type SDKAssistantMessage struct {
	Type             string                    `json:"type"` // "assistant"
	Message          json.RawMessage           `json:"message"`
	ParentToolUseID  *string                   `json:"parent_tool_use_id"`
	Error            *SDKAssistantMessageError `json:"error,omitempty"`
	UUID             string                    `json:"uuid"`
	SessionID        string                    `json:"session_id"`
}
func (m *SDKAssistantMessage) sdkMessage()        {}
func (m *SDKAssistantMessage) MessageType() string { return "assistant" }

// --- User Messages ---

type SDKUserMessage struct {
	Type            string          `json:"type"` // "user"
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	IsSynthetic     *bool           `json:"isSynthetic,omitempty"`
	ToolUseResult   interface{}     `json:"tool_use_result,omitempty"`
	Priority        *string         `json:"priority,omitempty"` // "now" | "next" | "later"
	Timestamp       *string         `json:"timestamp,omitempty"`
	UUID            *string         `json:"uuid,omitempty"`
	SessionID       string          `json:"session_id"`
}
func (m *SDKUserMessage) sdkMessage()        {}
func (m *SDKUserMessage) MessageType() string { return "user" }

type SDKUserMessageReplay struct {
	Type            string          `json:"type"` // "user"
	Message         json.RawMessage `json:"message"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	IsSynthetic     *bool           `json:"isSynthetic,omitempty"`
	ToolUseResult   interface{}     `json:"tool_use_result,omitempty"`
	Priority        *string         `json:"priority,omitempty"`
	Timestamp       *string         `json:"timestamp,omitempty"`
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
	IsReplay        bool            `json:"isReplay"` // always true
}
func (m *SDKUserMessageReplay) sdkMessage()        {}
func (m *SDKUserMessageReplay) MessageType() string { return "user" }

// --- Result Messages ---

type SDKResultSuccess struct {
	Type              string                     `json:"type"`    // "result"
	Subtype           string                     `json:"subtype"` // "success"
	DurationMs        int                        `json:"duration_ms"`
	DurationAPIMs     int                        `json:"duration_api_ms"`
	IsError           bool                       `json:"is_error"`
	NumTurns          int                        `json:"num_turns"`
	Result            string                     `json:"result"`
	StopReason        *string                    `json:"stop_reason"`
	TotalCostUSD      float64                    `json:"total_cost_usd"`
	Usage             NonNullableUsage           `json:"usage"`
	ModelUsageMap     map[string]ModelUsage      `json:"modelUsage"`
	PermissionDenials []SDKPermissionDenial      `json:"permission_denials"`
	StructuredOutput  interface{}                `json:"structured_output,omitempty"`
	FastModeState     *FastModeState             `json:"fast_mode_state,omitempty"`
	UUID              string                     `json:"uuid"`
	SessionID         string                     `json:"session_id"`
}
func (m *SDKResultSuccess) sdkMessage()        {}
func (m *SDKResultSuccess) MessageType() string { return "result" }

type SDKResultError struct {
	Type              string                     `json:"type"`    // "result"
	Subtype           string                     `json:"subtype"` // "error_during_execution" | "error_max_turns" | "error_max_budget_usd" | "error_max_structured_output_retries"
	DurationMs        int                        `json:"duration_ms"`
	DurationAPIMs     int                        `json:"duration_api_ms"`
	IsError           bool                       `json:"is_error"`
	NumTurns          int                        `json:"num_turns"`
	StopReason        *string                    `json:"stop_reason"`
	TotalCostUSD      float64                    `json:"total_cost_usd"`
	Usage             NonNullableUsage           `json:"usage"`
	ModelUsageMap     map[string]ModelUsage      `json:"modelUsage"`
	PermissionDenials []SDKPermissionDenial      `json:"permission_denials"`
	Errors            []string                   `json:"errors"`
	FastModeState     *FastModeState             `json:"fast_mode_state,omitempty"`
	UUID              string                     `json:"uuid"`
	SessionID         string                     `json:"session_id"`
}
func (m *SDKResultError) sdkMessage()        {}
func (m *SDKResultError) MessageType() string { return "result" }

// --- System Messages ---

type SDKSystemMessage struct {
	Type             string            `json:"type"`    // "system"
	Subtype          string            `json:"subtype"` // "init"
	Agents           []string          `json:"agents,omitempty"`
	ApiKeySource     ApiKeySource      `json:"apiKeySource"`
	Betas            []string          `json:"betas,omitempty"`
	ClaudeCodeVersion string           `json:"claude_code_version"`
	Cwd              string            `json:"cwd"`
	Tools            []string          `json:"tools"`
	McpServers       []McpServerRef    `json:"mcp_servers"`
	Model            string            `json:"model"`
	PermissionMode   PermissionMode    `json:"permissionMode"`
	SlashCommands    []string          `json:"slash_commands"`
	OutputStyle      string            `json:"output_style"`
	Skills           []string          `json:"skills"`
	Plugins          []PluginRef       `json:"plugins"`
	FastModeState    *FastModeState    `json:"fast_mode_state,omitempty"`
	UUID             string            `json:"uuid"`
	SessionID        string            `json:"session_id"`
}
func (m *SDKSystemMessage) sdkMessage()        {}
func (m *SDKSystemMessage) MessageType() string { return "system" }

type McpServerRef struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PluginRef struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// --- Status Message ---

type SDKStatusMessage struct {
	Type           string          `json:"type"`    // "system"
	Subtype        string          `json:"subtype"` // "status"
	Status         *string         `json:"status"`  // "compacting" | null
	PermissionMode *PermissionMode `json:"permissionMode,omitempty"`
	UUID           string          `json:"uuid"`
	SessionID      string          `json:"session_id"`
}
func (m *SDKStatusMessage) sdkMessage()        {}
func (m *SDKStatusMessage) MessageType() string { return "system" }

// --- API Retry ---

type SDKAPIRetryMessage struct {
	Type          string                    `json:"type"`    // "system"
	Subtype       string                    `json:"subtype"` // "api_retry"
	Attempt       int                       `json:"attempt"`
	MaxRetries    int                       `json:"max_retries"`
	RetryDelayMs  int                       `json:"retry_delay_ms"`
	ErrorStatus   *int                      `json:"error_status"`
	Error         *SDKAssistantMessageError `json:"error"`
	UUID          string                    `json:"uuid"`
	SessionID     string                    `json:"session_id"`
}
func (m *SDKAPIRetryMessage) sdkMessage()        {}
func (m *SDKAPIRetryMessage) MessageType() string { return "system" }

// --- Compact Boundary ---

type SDKCompactBoundaryMessage struct {
	Type            string          `json:"type"`    // "system"
	Subtype         string          `json:"subtype"` // "compact_boundary"
	CompactMetadata CompactMetadata `json:"compact_metadata"`
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
}
func (m *SDKCompactBoundaryMessage) sdkMessage()        {}
func (m *SDKCompactBoundaryMessage) MessageType() string { return "system" }

type CompactMetadata struct {
	Trigger          string            `json:"trigger"` // "manual" | "auto"
	PreTokens        int               `json:"pre_tokens"`
	PreservedSegment *PreservedSegment `json:"preserved_segment,omitempty"`
}

type PreservedSegment struct {
	HeadUUID   string `json:"head_uuid"`
	AnchorUUID string `json:"anchor_uuid"`
	TailUUID   string `json:"tail_uuid"`
}

// --- Local Command Output ---

type SDKLocalCommandOutputMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "local_command_output"
	Content   string `json:"content"`
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}
func (m *SDKLocalCommandOutputMessage) sdkMessage()        {}
func (m *SDKLocalCommandOutputMessage) MessageType() string { return "system" }

// --- Hook Messages ---

type SDKHookStartedMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "hook_started"
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}
func (m *SDKHookStartedMessage) sdkMessage()        {}
func (m *SDKHookStartedMessage) MessageType() string { return "system" }

type SDKHookProgressMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "hook_progress"
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	Output    string `json:"output"`
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}
func (m *SDKHookProgressMessage) sdkMessage()        {}
func (m *SDKHookProgressMessage) MessageType() string { return "system" }

type SDKHookResponseMessage struct {
	Type      string `json:"type"`    // "system"
	Subtype   string `json:"subtype"` // "hook_response"
	HookID    string `json:"hook_id"`
	HookName  string `json:"hook_name"`
	HookEvent string `json:"hook_event"`
	Output    string `json:"output"`
	Stdout    string `json:"stdout"`
	Stderr    string `json:"stderr"`
	ExitCode  *int   `json:"exit_code,omitempty"`
	Outcome   string `json:"outcome"` // "success" | "error" | "cancelled"
	UUID      string `json:"uuid"`
	SessionID string `json:"session_id"`
}
func (m *SDKHookResponseMessage) sdkMessage()        {}
func (m *SDKHookResponseMessage) MessageType() string { return "system" }

// --- Stream Event (Partial Assistant Message) ---

type SDKPartialAssistantMessage struct {
	Type            string          `json:"type"` // "stream_event"
	Event           json.RawMessage `json:"event"`
	ParentToolUseID *string         `json:"parent_tool_use_id"`
	UUID            string          `json:"uuid"`
	SessionID       string          `json:"session_id"`
}
func (m *SDKPartialAssistantMessage) sdkMessage()        {}
func (m *SDKPartialAssistantMessage) MessageType() string { return "stream_event" }

// --- Tool Progress ---

type SDKToolProgressMessage struct {
	Type               string  `json:"type"` // "tool_progress"
	ToolUseID          string  `json:"tool_use_id"`
	ToolName           string  `json:"tool_name"`
	ParentToolUseID    *string `json:"parent_tool_use_id"`
	ElapsedTimeSeconds float64 `json:"elapsed_time_seconds"`
	TaskID             *string `json:"task_id,omitempty"`
	UUID               string  `json:"uuid"`
	SessionID          string  `json:"session_id"`
}
func (m *SDKToolProgressMessage) sdkMessage()        {}
func (m *SDKToolProgressMessage) MessageType() string { return "tool_progress" }

// --- Tool Use Summary ---

type SDKToolUseSummaryMessage struct {
	Type                 string   `json:"type"` // "tool_use_summary"
	Summary              string   `json:"summary"`
	PrecedingToolUseIDs  []string `json:"preceding_tool_use_ids"`
	UUID                 string   `json:"uuid"`
	SessionID            string   `json:"session_id"`
}
func (m *SDKToolUseSummaryMessage) sdkMessage()        {}
func (m *SDKToolUseSummaryMessage) MessageType() string { return "tool_use_summary" }

// --- Auth Status ---

type SDKAuthStatusMessage struct {
	Type             string   `json:"type"` // "auth_status"
	IsAuthenticating bool     `json:"isAuthenticating"`
	Output           []string `json:"output"`
	Error            *string  `json:"error,omitempty"`
	UUID             string   `json:"uuid"`
	SessionID        string   `json:"session_id"`
}
func (m *SDKAuthStatusMessage) sdkMessage()        {}
func (m *SDKAuthStatusMessage) MessageType() string { return "auth_status" }

// --- Task Messages ---

type SDKTaskNotificationMessage struct {
	Type       string     `json:"type"`    // "system"
	Subtype    string     `json:"subtype"` // "task_notification"
	TaskID     string     `json:"task_id"`
	ToolUseID  *string    `json:"tool_use_id,omitempty"`
	Status     string     `json:"status"` // "completed" | "failed" | "stopped"
	OutputFile string     `json:"output_file"`
	Summary    string     `json:"summary"`
	Usage      *TaskUsage `json:"usage,omitempty"`
	UUID       string     `json:"uuid"`
	SessionID  string     `json:"session_id"`
}
func (m *SDKTaskNotificationMessage) sdkMessage()        {}
func (m *SDKTaskNotificationMessage) MessageType() string { return "system" }

type TaskUsage struct {
	TotalTokens int `json:"total_tokens"`
	ToolUses    int `json:"tool_uses"`
	DurationMs  int `json:"duration_ms"`
}

type SDKTaskStartedMessage struct {
	Type        string  `json:"type"`    // "system"
	Subtype     string  `json:"subtype"` // "task_started"
	TaskID      string  `json:"task_id"`
	ToolUseID   *string `json:"tool_use_id,omitempty"`
	Description string  `json:"description"`
	TaskType    *string `json:"task_type,omitempty"`
	Prompt      *string `json:"prompt,omitempty"`
	UUID        string  `json:"uuid"`
	SessionID   string  `json:"session_id"`
}
func (m *SDKTaskStartedMessage) sdkMessage()        {}
func (m *SDKTaskStartedMessage) MessageType() string { return "system" }

type SDKTaskProgressMessage struct {
	Type         string     `json:"type"`    // "system"
	Subtype      string     `json:"subtype"` // "task_progress"
	TaskID       string     `json:"task_id"`
	ToolUseID    *string    `json:"tool_use_id,omitempty"`
	Description  string     `json:"description"`
	Usage        TaskUsage  `json:"usage"`
	LastToolName *string    `json:"last_tool_name,omitempty"`
	Summary      *string    `json:"summary,omitempty"`
	UUID         string     `json:"uuid"`
	SessionID    string     `json:"session_id"`
}
func (m *SDKTaskProgressMessage) sdkMessage()        {}
func (m *SDKTaskProgressMessage) MessageType() string { return "system" }

// --- Files Persisted ---

type SDKFilesPersistedEvent struct {
	Type        string              `json:"type"`    // "system"
	Subtype     string              `json:"subtype"` // "files_persisted"
	Files       []PersistedFile     `json:"files"`
	Failed      []PersistedFileFail `json:"failed"`
	ProcessedAt string              `json:"processed_at"`
	UUID        string              `json:"uuid"`
	SessionID   string              `json:"session_id"`
}
func (m *SDKFilesPersistedEvent) sdkMessage()        {}
func (m *SDKFilesPersistedEvent) MessageType() string { return "system" }

type PersistedFile struct {
	Filename string `json:"filename"`
	FileID   string `json:"file_id"`
}

type PersistedFileFail struct {
	Filename string `json:"filename"`
	Error    string `json:"error"`
}

// --- Rate Limit ---

type SDKRateLimitEvent struct {
	Type          string           `json:"type"` // "rate_limit_event"
	RateLimitInfo SDKRateLimitInfo `json:"rate_limit_info"`
	UUID          string           `json:"uuid"`
	SessionID     string           `json:"session_id"`
}
func (m *SDKRateLimitEvent) sdkMessage()        {}
func (m *SDKRateLimitEvent) MessageType() string { return "rate_limit_event" }

type SDKRateLimitInfo struct {
	Status                string   `json:"status"` // "allowed" | "allowed_warning" | "rejected"
	ResetsAt              *int64   `json:"resetsAt,omitempty"`
	RateLimitType         *string  `json:"rateLimitType,omitempty"`
	Utilization           *float64 `json:"utilization,omitempty"`
	OverageStatus         *string  `json:"overageStatus,omitempty"`
	OverageResetsAt       *int64   `json:"overageResetsAt,omitempty"`
	OverageDisabledReason *string  `json:"overageDisabledReason,omitempty"`
	IsUsingOverage        *bool    `json:"isUsingOverage,omitempty"`
	SurpassedThreshold    *float64 `json:"surpassedThreshold,omitempty"`
}

// --- Elicitation Complete ---

type SDKElicitationCompleteMessage struct {
	Type          string `json:"type"`    // "system"
	Subtype       string `json:"subtype"` // "elicitation_complete"
	McpServerName string `json:"mcp_server_name"`
	ElicitationID string `json:"elicitation_id"`
	UUID          string `json:"uuid"`
	SessionID     string `json:"session_id"`
}
func (m *SDKElicitationCompleteMessage) sdkMessage()        {}
func (m *SDKElicitationCompleteMessage) MessageType() string { return "system" }

// --- Prompt Suggestion ---

type SDKPromptSuggestionMessage struct {
	Type       string `json:"type"` // "prompt_suggestion"
	Suggestion string `json:"suggestion"`
	UUID       string `json:"uuid"`
	SessionID  string `json:"session_id"`
}
func (m *SDKPromptSuggestionMessage) sdkMessage()        {}
func (m *SDKPromptSuggestionMessage) MessageType() string { return "prompt_suggestion" }
```

- [ ] **Step 4: Run all message + JSON tests**

Run: `go test -v ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add messages.go messages_test.go json.go json_test.go
git commit -m "feat: add all SDKMessage types and JSON parser dispatcher"
```

---

### Task 6: Model, Account, and Agent Info Types

**Files:**
- Create: `models.go`
- Create: `models_test.go`

- [ ] **Step 1: Write failing test**

Create `models_test.go`:
```go
package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestModelInfo_JSON(t *testing.T) {
	raw := `{"value":"claude-sonnet-4-6","displayName":"Claude Sonnet 4.6","description":"Fast model","supportsEffort":true,"supportedEffortLevels":["low","medium","high"],"supportsAdaptiveThinking":true,"supportsFastMode":true,"supportsAutoMode":false}`
	var m ModelInfo
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if m.Value != "claude-sonnet-4-6" {
		t.Errorf("Value = %q", m.Value)
	}
	if m.SupportsEffort == nil || !*m.SupportsEffort {
		t.Error("SupportsEffort should be true")
	}
}

func TestAccountInfo_JSON(t *testing.T) {
	raw := `{"email":"test@example.com","organization":"Acme","apiProvider":"firstParty"}`
	var a AccountInfo
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if *a.Email != "test@example.com" {
		t.Errorf("Email = %v", a.Email)
	}
}

func TestAgentDefinition_JSON(t *testing.T) {
	raw := `{"description":"test runner","prompt":"Run tests","tools":["Bash","Read"],"model":"haiku"}`
	var a AgentDefinition
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if a.Description != "test runner" {
		t.Errorf("Description = %q", a.Description)
	}
	if len(a.Tools) != 2 {
		t.Errorf("Tools = %v", a.Tools)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run "TestModelInfo|TestAccountInfo|TestAgentDefinition" -v ./...`
Expected: FAIL

- [ ] **Step 3: Implement types**

Create `models.go`:
```go
package claudeagent

// ModelInfo describes an available model.
type ModelInfo struct {
	Value                    string   `json:"value"`
	DisplayName              string   `json:"displayName"`
	Description              string   `json:"description"`
	SupportsEffort           *bool    `json:"supportsEffort,omitempty"`
	SupportedEffortLevels    []string `json:"supportedEffortLevels,omitempty"`
	SupportsAdaptiveThinking *bool    `json:"supportsAdaptiveThinking,omitempty"`
	SupportsFastMode         *bool    `json:"supportsFastMode,omitempty"`
	SupportsAutoMode         *bool    `json:"supportsAutoMode,omitempty"`
}

// AccountInfo describes the authenticated user's account.
type AccountInfo struct {
	Email            *string `json:"email,omitempty"`
	Organization     *string `json:"organization,omitempty"`
	SubscriptionType *string `json:"subscriptionType,omitempty"`
	TokenSource      *string `json:"tokenSource,omitempty"`
	ApiKeySource     *string `json:"apiKeySource,omitempty"`
	ApiProvider      *string `json:"apiProvider,omitempty"` // "firstParty" | "bedrock" | "vertex" | "foundry"
}

// AgentDefinition defines a custom subagent.
type AgentDefinition struct {
	Description                      string              `json:"description"`
	Tools                            []string            `json:"tools,omitempty"`
	DisallowedTools                  []string            `json:"disallowedTools,omitempty"`
	Prompt                           string              `json:"prompt"`
	Model                            *string             `json:"model,omitempty"`
	McpServers                       []AgentMcpServerSpec `json:"mcpServers,omitempty"`
	CriticalSystemReminder           *string             `json:"criticalSystemReminder_EXPERIMENTAL,omitempty"`
	Skills                           []string            `json:"skills,omitempty"`
	MaxTurns                         *int                `json:"maxTurns,omitempty"`
}

// AgentMcpServerSpec is either a server name (string) or a map of name to config.
// In Go, this is represented as json.RawMessage for flexibility.
type AgentMcpServerSpec = interface{}

// AgentInfo describes an available subagent.
type AgentInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Model       *string `json:"model,omitempty"`
}

// SlashCommand describes an available slash command/skill.
type SlashCommand struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}
```

- [ ] **Step 4: Run tests**

Run: `go test -run "TestModelInfo|TestAccountInfo|TestAgentDefinition" -v ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add models.go models_test.go
git commit -m "feat: add ModelInfo, AccountInfo, AgentDefinition, AgentInfo types"
```

---

### Task 7: MCP Server Configuration Types

**Files:**
- Create: `mcp.go`
- Create: `mcp_test.go`

- [ ] **Step 1: Write failing tests for all MCP config types**

Create `mcp_test.go`:
```go
package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestMcpStdioServerConfig_JSON(t *testing.T) {
	raw := `{"command":"node","args":["server.js"],"env":{"PORT":"3000"}}`
	var c McpStdioServerConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.Command != "node" {
		t.Errorf("Command = %q", c.Command)
	}
}

func TestMcpServerStatus_JSON(t *testing.T) {
	raw := `{"name":"my-server","status":"connected","serverInfo":{"name":"test","version":"1.0"},"tools":[{"name":"mytool","description":"a tool"}]}`
	var s McpServerStatus
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "my-server" {
		t.Errorf("Name = %q", s.Name)
	}
	if len(s.Tools) != 1 {
		t.Errorf("Tools len = %d", len(s.Tools))
	}
}

func TestMcpSetServersResult_JSON(t *testing.T) {
	raw := `{"added":["s1"],"removed":["s2"],"errors":{"s3":"failed to connect"}}`
	var r McpSetServersResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Added) != 1 {
		t.Errorf("Added = %v", r.Added)
	}
	if r.Errors["s3"] != "failed to connect" {
		t.Errorf("Errors = %v", r.Errors)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run "TestMcp" -v ./...`
Expected: FAIL

- [ ] **Step 3: Implement MCP types**

Create `mcp.go`:
```go
package claudeagent

// McpStdioServerConfig defines a stdio-based MCP server.
type McpStdioServerConfig struct {
	Type    *string           `json:"type,omitempty"` // "stdio" or omitted
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// McpSSEServerConfig defines an SSE-based MCP server.
type McpSSEServerConfig struct {
	Type    string            `json:"type"` // "sse"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// McpHttpServerConfig defines an HTTP-based MCP server.
type McpHttpServerConfig struct {
	Type    string            `json:"type"` // "http"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

// McpSdkServerConfig defines an in-process SDK MCP server.
type McpSdkServerConfig struct {
	Type string `json:"type"` // "sdk"
	Name string `json:"name"`
}

// McpClaudeAIProxyServerConfig defines a claude.ai proxy MCP server.
type McpClaudeAIProxyServerConfig struct {
	Type string `json:"type"` // "claudeai-proxy"
	URL  string `json:"url"`
	ID   string `json:"id"`
}

// McpServerConfig is a union of all MCP server configuration types.
// Use json.RawMessage and inspect the "type" field to determine the concrete type.
type McpServerConfig = interface{}

// McpServerStatus describes the current status of an MCP server connection.
type McpServerStatus struct {
	Name       string               `json:"name"`
	Status     string               `json:"status"` // "connected" | "failed" | "needs-auth" | "pending" | "disabled"
	ServerInfo *McpServerInfo       `json:"serverInfo,omitempty"`
	Error      *string              `json:"error,omitempty"`
	Config     interface{}          `json:"config,omitempty"`
	Scope      *string              `json:"scope,omitempty"`
	Tools      []McpServerToolInfo  `json:"tools,omitempty"`
}

type McpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type McpServerToolInfo struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	Annotations *McpToolAnnotations      `json:"annotations,omitempty"`
}

type McpToolAnnotations struct {
	ReadOnly    *bool `json:"readOnly,omitempty"`
	Destructive *bool `json:"destructive,omitempty"`
	OpenWorld   *bool `json:"openWorld,omitempty"`
}

// McpSetServersResult is the result of a setMcpServers operation.
type McpSetServersResult struct {
	Added   []string          `json:"added"`
	Removed []string          `json:"removed"`
	Errors  map[string]string `json:"errors"`
}
```

- [ ] **Step 4: Run tests**

Run: `go test -run "TestMcp" -v ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add mcp.go mcp_test.go
git commit -m "feat: add MCP server configuration and status types"
```

---

### Task 8: Permission Types

**Files:**
- Create: `permissions.go`
- Create: `permissions_test.go`

- [ ] **Step 1: Write failing tests**

Create `permissions_test.go`:
```go
package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestPermissionResult_Allow_JSON(t *testing.T) {
	r := PermissionResultAllow{
		Behavior: PermissionBehaviorAllow,
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) == "" {
		t.Error("empty marshal")
	}
}

func TestPermissionUpdate_AddRules_JSON(t *testing.T) {
	raw := `{"type":"addRules","rules":[{"toolName":"Bash","ruleContent":"npm *"}],"behavior":"allow","destination":"session"}`
	var u PermissionUpdateAddRules
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		t.Fatal(err)
	}
	if u.UpdateType != "addRules" {
		t.Errorf("Type = %q", u.UpdateType)
	}
	if len(u.Rules) != 1 {
		t.Errorf("Rules = %v", u.Rules)
	}
}

func TestPermissionRuleValue_JSON(t *testing.T) {
	raw := `{"toolName":"Bash","ruleContent":"npm test"}`
	var r PermissionRuleValue
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if r.ToolName != "Bash" {
		t.Errorf("ToolName = %q", r.ToolName)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -run TestPermission -v ./...`
Expected: FAIL

- [ ] **Step 3: Implement permission types**

Create `permissions.go`:
```go
package claudeagent

import "context"

// CanUseTool is a callback invoked before each tool execution.
// Return a PermissionResult to allow or deny the tool use.
type CanUseTool func(ctx context.Context, toolName string, input map[string]interface{}, opts CanUseToolOptions) (PermissionResult, error)

// CanUseToolOptions provides context for permission decisions.
type CanUseToolOptions struct {
	Suggestions     []PermissionUpdate `json:"suggestions,omitempty"`
	BlockedPath     *string            `json:"blockedPath,omitempty"`
	DecisionReason  *string            `json:"decisionReason,omitempty"`
	Title           *string            `json:"title,omitempty"`
	DisplayName     *string            `json:"displayName,omitempty"`
	Description     *string            `json:"description,omitempty"`
	ToolUseID       string             `json:"toolUseID"`
	AgentID         *string            `json:"agentID,omitempty"`
}

// PermissionResult is the interface for permission decisions.
type PermissionResult interface {
	permissionResult()
}

// PermissionResultAllow approves a tool use.
type PermissionResultAllow struct {
	Behavior           PermissionBehavior `json:"behavior"` // "allow"
	UpdatedInput       map[string]interface{} `json:"updatedInput,omitempty"`
	UpdatedPermissions []PermissionUpdate     `json:"updatedPermissions,omitempty"`
	ToolUseID          *string                `json:"toolUseID,omitempty"`
}
func (r PermissionResultAllow) permissionResult() {}

// PermissionResultDeny denies a tool use.
type PermissionResultDeny struct {
	Behavior  PermissionBehavior `json:"behavior"` // "deny"
	Message   string             `json:"message"`
	Interrupt *bool              `json:"interrupt,omitempty"`
	ToolUseID *string            `json:"toolUseID,omitempty"`
}
func (r PermissionResultDeny) permissionResult() {}

// PermissionRuleValue identifies a permission rule.
type PermissionRuleValue struct {
	ToolName    string  `json:"toolName"`
	RuleContent *string `json:"ruleContent,omitempty"`
}

// PermissionUpdate is the interface for permission update operations.
type PermissionUpdate interface {
	permissionUpdate()
}

// PermissionUpdateAddRules adds new permission rules.
type PermissionUpdateAddRules struct {
	UpdateType  string                      `json:"type"` // "addRules"
	Rules       []PermissionRuleValue       `json:"rules"`
	Behavior    PermissionBehavior          `json:"behavior"`
	Destination PermissionUpdateDestination `json:"destination"`
}
func (u PermissionUpdateAddRules) permissionUpdate() {}

// PermissionUpdateReplaceRules replaces existing permission rules.
type PermissionUpdateReplaceRules struct {
	UpdateType  string                      `json:"type"` // "replaceRules"
	Rules       []PermissionRuleValue       `json:"rules"`
	Behavior    PermissionBehavior          `json:"behavior"`
	Destination PermissionUpdateDestination `json:"destination"`
}
func (u PermissionUpdateReplaceRules) permissionUpdate() {}

// PermissionUpdateRemoveRules removes permission rules.
type PermissionUpdateRemoveRules struct {
	UpdateType  string                      `json:"type"` // "removeRules"
	Rules       []PermissionRuleValue       `json:"rules"`
	Behavior    PermissionBehavior          `json:"behavior"`
	Destination PermissionUpdateDestination `json:"destination"`
}
func (u PermissionUpdateRemoveRules) permissionUpdate() {}

// PermissionUpdateSetMode changes the permission mode.
type PermissionUpdateSetMode struct {
	UpdateType  string                      `json:"type"` // "setMode"
	Mode        PermissionMode              `json:"mode"`
	Destination PermissionUpdateDestination `json:"destination"`
}
func (u PermissionUpdateSetMode) permissionUpdate() {}

// PermissionUpdateAddDirectories adds directories to the permission scope.
type PermissionUpdateAddDirectories struct {
	UpdateType  string                      `json:"type"` // "addDirectories"
	Directories []string                    `json:"directories"`
	Destination PermissionUpdateDestination `json:"destination"`
}
func (u PermissionUpdateAddDirectories) permissionUpdate() {}

// PermissionUpdateRemoveDirectories removes directories from the permission scope.
type PermissionUpdateRemoveDirectories struct {
	UpdateType  string                      `json:"type"` // "removeDirectories"
	Directories []string                    `json:"directories"`
	Destination PermissionUpdateDestination `json:"destination"`
}
func (u PermissionUpdateRemoveDirectories) permissionUpdate() {}
```

- [ ] **Step 4: Run tests**

Run: `go test -run TestPermission -v ./...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add permissions.go permissions_test.go
git commit -m "feat: add permission types (CanUseTool, PermissionResult, PermissionUpdate)"
```

---

### Task 9: Hook Types

**Files:**
- Create: `hooks.go`
- Create: `hooks_test.go`

- [ ] **Step 1: Write failing tests**

- [ ] **Step 2: Implement all hook input/output types**

Create `hooks.go` with all 23 hook event input types:
- `BaseHookInput`, `PreToolUseHookInput`, `PostToolUseHookInput`, `PostToolUseFailureHookInput`
- `NotificationHookInput`, `UserPromptSubmitHookInput`, `SessionStartHookInput`, `SessionEndHookInput`
- `StopHookInput`, `StopFailureHookInput`, `SubagentStartHookInput`, `SubagentStopHookInput`
- `PreCompactHookInput`, `PostCompactHookInput`, `PermissionRequestHookInput`, `SetupHookInput`
- `TeammateIdleHookInput`, `TaskCompletedHookInput`, `ElicitationHookInput`, `ElicitationResultHookInput`
- `ConfigChangeHookInput`, `WorktreeCreateHookInput`, `WorktreeRemoveHookInput`, `InstructionsLoadedHookInput`

Plus all hook-specific output types and `HookCallback`, `HookCallbackMatcher`, `HookJSONOutput`.

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add hooks.go hooks_test.go
git commit -m "feat: add all hook event input/output types and callback interfaces"
```

---

### Task 10: Options and Query Parameters

**Files:**
- Create: `options.go`
- Create: `options_test.go`

- [ ] **Step 1: Write failing tests for Options struct serialization**
- [ ] **Step 2: Implement Options struct** (mirrors the TypeScript `Options` type with ~50 fields)

Key fields to implement:
```go
type Options struct {
    AbortContext           context.Context
    AdditionalDirectories  []string
    Agent                  *string
    Agents                 map[string]AgentDefinition
    AllowedTools           []string
    CanUseTool             CanUseTool
    Continue               *bool
    Cwd                    *string
    DisallowedTools        []string
    Tools                  interface{} // []string | ToolPreset
    Env                    map[string]string
    Executable             *string // "bun" | "deno" | "node"
    ExecutableArgs         []string
    ExtraArgs              map[string]*string
    FallbackModel          *string
    EnableFileCheckpointing *bool
    ToolConfig             *ToolConfig
    ForkSession            *bool
    Betas                  []SdkBeta
    Hooks                  map[HookEvent][]HookCallbackMatcher
    OnElicitation          OnElicitation
    PersistSession         *bool
    IncludePartialMessages *bool
    Thinking               *ThinkingConfig
    Effort                 *string
    MaxThinkingTokens      *int
    MaxTurns               *int
    MaxBudgetUsd           *float64
    McpServers             map[string]McpServerConfig
    Model                  *string
    OutputFormat           *OutputFormat
    PathToClaudeCodeExecutable *string
    PermissionMode         *PermissionMode
    AllowDangerouslySkipPermissions *bool
    PermissionPromptToolName *string
    Plugins                []SdkPluginConfig
    PromptSuggestions      *bool
    AgentProgressSummaries *bool
    Resume                 *string
    SessionID              *string
    ResumeSessionAt        *string
    Sandbox                *SandboxSettings
    Settings               interface{} // string | *Settings
    SettingSources         []SettingSource
    Debug                  *bool
    DebugFile              *string
    Stderr                 func(string)
    StrictMcpConfig        *bool
    SystemPrompt           interface{} // string | SystemPromptPreset
    SpawnClaudeCodeProcess SpawnFunc
}
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add options.go options_test.go
git commit -m "feat: add Options struct and supporting types (ThinkingConfig, SystemPrompt, etc.)"
```

---

### Task 11: Sandbox Settings Types

**Files:**
- Create: `sandbox.go`
- Create: `sandbox_test.go`

- [ ] **Step 1: Write tests for sandbox config serialization**
- [ ] **Step 2: Implement SandboxSettings, SandboxNetworkConfig, SandboxFilesystemConfig**
- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add sandbox.go sandbox_test.go
git commit -m "feat: add sandbox settings types"
```

---

### Task 12: Settings Type (Large)

**Files:**
- Create: `settings.go`
- Create: `settings_test.go`

- [ ] **Step 1: Write tests for Settings JSON round-trip**
- [ ] **Step 2: Implement the full Settings struct** (~80 fields covering all settings.json options)

This mirrors the massive `Settings` interface from the TypeScript SDK including:
- `apiKeyHelper`, `permissions`, `model`, `hooks`, `sandbox`, `env`
- `worktree`, `enabledPlugins`, `extraKnownMarketplaces`, `strictKnownMarketplaces`
- All enterprise/managed settings fields
- Plugin configs, marketplace sources, etc.

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add settings.go settings_test.go
git commit -m "feat: add complete Settings struct with all configuration fields"
```

---

### Task 13: Control Request/Response Types

**Files:**
- Create: `control.go`
- Create: `control_test.go`

- [ ] **Step 1: Write tests for control request serialization**
- [ ] **Step 2: Implement all control request/response types**

Types to implement:
- `SDKControlRequest`, `SDKControlResponse`
- `SDKControlInitializeRequest`, `SDKControlInitializeResponse`
- `SDKControlInterruptRequest`, `SDKControlPermissionRequest`
- `SDKControlSetPermissionModeRequest`, `SDKControlSetModelRequest`
- `SDKControlSetMaxThinkingTokensRequest`
- `SDKControlMcpStatusRequest`, `SDKControlMcpSetServersRequest`
- `SDKControlMcpReconnectRequest`, `SDKControlMcpToggleRequest`
- `SDKControlRewindFilesRequest`, `SDKControlStopTaskRequest`
- `SDKControlApplyFlagSettingsRequest`, `SDKControlGetSettingsRequest`
- `SDKControlElicitationRequest`, `SDKControlCancelAsyncMessageRequest`
- `SDKControlCancelRequest`, `SDKControlEndSessionRequest`
- `SDKHookCallbackMatcher`, `SDKHookCallbackRequest`
- Request/response correlation engine (map of request_id → response channel)

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add control.go control_test.go
git commit -m "feat: add control request/response types and correlation engine"
```

---

### Task 14: Tool Input/Output Types (tools/ package)

**Files:**
- Create: `tools/tools.go`, `tools/agent.go`, `tools/bash.go`, `tools/file.go`, `tools/glob.go`, `tools/grep.go`, `tools/notebook.go`, `tools/mcp_tools.go`, `tools/web.go`, `tools/task.go`, `tools/todo.go`, `tools/ask.go`, `tools/config.go`, `tools/worktree.go`, `tools/plan.go`
- Create: `tools/agent_test.go`, `tools/bash_test.go`, `tools/file_test.go`, etc.

- [ ] **Step 1: Write failing tests for key tool types (agent, bash, file, grep)**
- [ ] **Step 2: Implement all tool input types** (directly from `sdk-tools.d.ts`)

Key types per file:
- `agent.go`: `AgentInput`, `AgentOutput` (completed + async_launched variants)
- `bash.go`: `BashInput`, `BashOutput`
- `file.go`: `FileReadInput`, `FileReadOutput` (text/image/notebook/pdf/parts variants), `FileEditInput/Output`, `FileWriteInput/Output`
- `glob.go`: `GlobInput`, `GlobOutput`
- `grep.go`: `GrepInput`, `GrepOutput`
- `notebook.go`: `NotebookEditInput/Output`
- `mcp_tools.go`: `McpInput/Output`, `ListMcpResourcesInput/Output`, `ReadMcpResourceInput/Output`, `SubscribeMcpResourceInput/Output`, `UnsubscribeMcpResourceInput/Output`, `SubscribePollingInput/Output`, `UnsubscribePollingInput/Output`
- `web.go`: `WebFetchInput/Output`, `WebSearchInput/Output`
- `task.go`: `TaskOutputInput`, `TaskStopInput/Output`
- `todo.go`: `TodoWriteInput/Output`
- `ask.go`: `AskUserQuestionInput/Output`
- `config.go`: `ConfigInput/Output`
- `worktree.go`: `EnterWorktreeInput/Output`, `ExitWorktreeInput/Output`
- `plan.go`: `ExitPlanModeInput/Output`

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add tools/
git commit -m "feat: add all tool input/output types (tools/ package)"
```

---

### Task 15: Process Management (Internal)

**Files:**
- Create: `internal/process/spawn.go`
- Create: `process.go`
- Create: `process_test.go`

- [ ] **Step 1: Write failing tests for process spawn and lifecycle**

```go
func TestBuildProcessArgs(t *testing.T) {
    args := buildProcessArgs(Options{
        Model: strPtr("claude-sonnet-4-6"),
        Cwd:   strPtr("/tmp"),
    })
    // Verify expected CLI flags
}
```

- [ ] **Step 2: Implement process management**

```go
// SpawnFunc allows custom process spawning (VMs, containers, remote).
type SpawnFunc func(opts SpawnOptions) SpawnedProcess

type SpawnOptions struct {
    Command string
    Args    []string
    Cwd     string
    Env     map[string]string
    Cancel  context.CancelFunc
}

// SpawnedProcess wraps a running Claude Code subprocess.
type SpawnedProcess interface {
    Stdin() io.WriteCloser
    Stdout() io.ReadCloser
    Stderr() io.ReadCloser
    Wait() error
    Kill() error
}
```

Key implementation:
- Resolve CLI executable path (embedded or custom)
- Build CLI argument list from Options
- Spawn subprocess with os/exec
- Wire stdin (JSON lines in), stdout (JSON lines out), stderr (debug)
- Handle process lifecycle (start, wait, kill, cleanup)

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add internal/process/ process.go process_test.go
git commit -m "feat: add process management for Claude Code subprocess"
```

---

### Task 16: Query Implementation (Core)

**Files:**
- Create: `query.go`
- Create: `query_test.go`

- [ ] **Step 1: Write failing tests for Query construction and message channel**
- [ ] **Step 2: Implement Query struct and all control methods**

```go
// QueryParams configures a query to Claude Code.
type QueryParams struct {
    // Prompt is either a string or a channel of SDKUserMessage for streaming input.
    Prompt  interface{} // string | <-chan SDKUserMessage
    Options *Options
}

// Query manages a conversation with the Claude Code subprocess.
type Query struct {
    messages <-chan SDKMessage
    // ... internal state
}

// NewQuery creates and starts a new query.
func NewQuery(params QueryParams) *Query

// Messages returns the channel of messages from the agent.
func (q *Query) Messages() <-chan SDKMessage

// Interrupt stops the current query execution.
func (q *Query) Interrupt(ctx context.Context) error

// SetPermissionMode changes the permission mode mid-session.
func (q *Query) SetPermissionMode(ctx context.Context, mode PermissionMode) error

// SetModel changes the model mid-session.
func (q *Query) SetModel(ctx context.Context, model *string) error

// SetMaxThinkingTokens sets the max thinking token budget.
func (q *Query) SetMaxThinkingTokens(ctx context.Context, tokens *int) error

// ApplyFlagSettings merges settings into the flag settings layer.
func (q *Query) ApplyFlagSettings(ctx context.Context, settings Settings) error

// InitializationResult returns the full init response.
func (q *Query) InitializationResult(ctx context.Context) (*SDKControlInitializeResponse, error)

// SupportedCommands returns available slash commands.
func (q *Query) SupportedCommands(ctx context.Context) ([]SlashCommand, error)

// SupportedModels returns available models.
func (q *Query) SupportedModels(ctx context.Context) ([]ModelInfo, error)

// SupportedAgents returns available subagents.
func (q *Query) SupportedAgents(ctx context.Context) ([]AgentInfo, error)

// McpServerStatus returns MCP server connection statuses.
func (q *Query) McpServerStatus(ctx context.Context) ([]McpServerStatus, error)

// AccountInfo returns authenticated account info.
func (q *Query) AccountInfo(ctx context.Context) (*AccountInfo, error)

// RewindFiles rewinds file changes to a specific message.
func (q *Query) RewindFiles(ctx context.Context, userMessageID string, opts *RewindFilesOptions) (*RewindFilesResult, error)

// ReconnectMcpServer reconnects a disconnected MCP server.
func (q *Query) ReconnectMcpServer(ctx context.Context, serverName string) error

// ToggleMcpServer enables or disables an MCP server.
func (q *Query) ToggleMcpServer(ctx context.Context, serverName string, enabled bool) error

// SetMcpServers replaces the dynamic MCP server set.
func (q *Query) SetMcpServers(ctx context.Context, servers map[string]McpServerConfig) (*McpSetServersResult, error)

// StreamInput sends user messages to the query.
func (q *Query) StreamInput(messages <-chan SDKUserMessage) error

// StopTask stops a running background task.
func (q *Query) StopTask(ctx context.Context, taskID string) error

// Close terminates the query and cleans up resources.
func (q *Query) Close()
```

Internal implementation:
- Goroutine reading stdout JSON lines → dispatching to messages channel
- Goroutine reading stderr → calling Options.Stderr callback
- Control request/response correlation via request_id map
- Initialization handshake (send initialize control request, wait for response)
- Hook callback routing to user-supplied HookCallback functions
- Permission request routing to CanUseTool callback
- Graceful shutdown on context cancellation

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: implement Query with streaming, control methods, and lifecycle management"
```

---

### Task 17: Top-Level API Functions

**Files:**
- Modify: `claudeagent.go`

- [ ] **Step 1: Write failing tests for top-level query() function**
- [ ] **Step 2: Implement the public query() function**

```go
// Query creates a new query to Claude Code.
// Returns a *Query that streams SDKMessage values via Messages().
//
// Example:
//
//	q := claudeagent.NewQuery(claudeagent.QueryParams{
//	    Prompt: "What does this code do?",
//	    Options: &claudeagent.Options{
//	        Model: claudeagent.String("claude-sonnet-4-6"),
//	    },
//	})
//	defer q.Close()
//	for msg := range q.Messages() {
//	    switch m := msg.(type) {
//	    case *claudeagent.SDKResultSuccess:
//	        fmt.Println(m.Result)
//	    }
//	}
func NewQuery(params QueryParams) *Query
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add claudeagent.go claudeagent_test.go
git commit -m "feat: add top-level NewQuery API function"
```

---

### Task 18: Session Management Functions

**Files:**
- Create: `session.go`
- Create: `session_test.go`

- [ ] **Step 1: Write failing tests for session CRUD operations**
- [ ] **Step 2: Implement all session functions**

```go
// ListSessions returns session metadata.
func ListSessions(opts *ListSessionsOptions) ([]SDKSessionInfo, error)

// GetSessionInfo returns metadata for a single session.
func GetSessionInfo(sessionID string, opts *GetSessionInfoOptions) (*SDKSessionInfo, error)

// GetSessionMessages reads conversation messages from a session transcript.
func GetSessionMessages(sessionID string, opts *GetSessionMessagesOptions) ([]SessionMessage, error)

// ForkSession creates a new session branched from an existing one.
func ForkSession(sessionID string, opts *ForkSessionOptions) (*ForkSessionResult, error)

// RenameSession changes a session's title.
func RenameSession(sessionID string, title string, opts *SessionMutationOptions) error

// TagSession adds a tag to a session.
func TagSession(sessionID string, tag string, opts *SessionMutationOptions) error
```

Plus all supporting types: `SDKSessionInfo`, `SessionMessage`, `ListSessionsOptions`, etc.

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add session.go session_test.go
git commit -m "feat: add session management functions (list, get, fork, rename, tag)"
```

---

### Task 19: V2 Session API (Alpha)

**Files:**
- Create: `session_v2.go`
- Create: `session_v2_test.go`

- [ ] **Step 1: Write failing tests for V2 session lifecycle**
- [ ] **Step 2: Implement V2 Session API**

```go
// SDKSession provides a multi-turn conversation interface (V2 API, alpha).
type SDKSession struct {
    sessionID string
    // ...
}

// CreateSession creates a new V2 session.
func CreateSession(opts SDKSessionOptions) (*SDKSession, error)

// ResumeSession resumes an existing V2 session.
func ResumeSession(sessionID string, opts SDKSessionOptions) (*SDKSession, error)

// Send sends a message to the agent.
func (s *SDKSession) Send(message interface{}) error // string | SDKUserMessage

// Stream returns a channel of messages from the agent.
func (s *SDKSession) Stream() <-chan SDKMessage

// Close terminates the session.
func (s *SDKSession) Close()

// SessionID returns the session identifier.
func (s *SDKSession) SessionID() string
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add session_v2.go session_v2_test.go
git commit -m "feat: add V2 session API (alpha — CreateSession, ResumeSession)"
```

---

### Task 20: Bridge API (Alpha)

**Files:**
- Create: `bridge/bridge.go`
- Create: `bridge/types.go`
- Create: `bridge/bridge_test.go`

- [ ] **Step 1: Write failing tests for bridge session handle**
- [ ] **Step 2: Implement bridge types and AttachBridgeSession**

```go
package bridge

// SessionState represents bridge session states.
type SessionState string

const (
    SessionStateIdle           SessionState = "idle"
    SessionStateRunning        SessionState = "running"
    SessionStateRequiresAction SessionState = "requires_action"
)

// BridgeSessionHandle is a per-session bridge transport handle.
type BridgeSessionHandle struct {
    sessionID string
    // ...
}

// AttachBridgeSession attaches to an existing bridge session (alpha).
func AttachBridgeSession(opts AttachBridgeSessionOptions) (*BridgeSessionHandle, error)

// All methods: Write, SendResult, SendControlRequest, SendControlResponse,
// SendControlCancelRequest, ReconnectTransport, ReportState, ReportMetadata,
// ReportDelivery, Flush, Close, GetSequenceNum, IsConnected, SessionID
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add bridge/
git commit -m "feat: add bridge API (alpha — AttachBridgeSession, BridgeSessionHandle)"
```

---

### Task 21: Browser/WebSocket Transport

**Files:**
- Create: `browser/browser.go`
- Create: `browser/types.go`
- Create: `browser/browser_test.go`

- [ ] **Step 1: Write failing tests for browser query options**
- [ ] **Step 2: Implement WebSocket-based query**

```go
package browser

// BrowserQueryOptions configures a browser-based query.
type BrowserQueryOptions struct {
    Prompt         <-chan claudeagent.SDKUserMessage
    WebSocket      WebSocketOptions
    AbortContext    context.Context
    CanUseTool     claudeagent.CanUseTool
    Hooks          map[claudeagent.HookEvent][]claudeagent.HookCallbackMatcher
    McpServers     map[string]claudeagent.McpServerConfig
    JsonSchema     map[string]interface{}
}

type WebSocketOptions struct {
    URL     string
    Headers map[string]string
    Auth    *AuthMessage
}

type AuthMessage struct {
    Type       string          `json:"type"` // "auth"
    Credential OAuthCredential `json:"credential"`
}

type OAuthCredential struct {
    Type  string `json:"type"` // "oauth"
    Token string `json:"token"`
}

// Query creates a browser-based query via WebSocket.
func Query(opts BrowserQueryOptions) *claudeagent.Query
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add browser/
git commit -m "feat: add browser/WebSocket transport (browser/ package)"
```

---

### Task 22: CLI Path Resolution (embed)

**Files:**
- Create: `embed.go`
- Create: `embed_test.go`

- [ ] **Step 1: Write failing test for CLI path resolution**
- [ ] **Step 2: Implement CLI path finder**

```go
// CLIPath returns the path to the Claude Code CLI executable.
// It checks PathToClaudeCodeExecutable option first, then looks for
// the bundled CLI, then falls back to PATH lookup.
func CLIPath(customPath *string) (string, error)
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add embed.go embed_test.go
git commit -m "feat: add CLI path resolution"
```

---

### Task 23: Helper Functions and Utilities

**Files:**
- Create: `helpers.go`
- Create: `helpers_test.go`

- [ ] **Step 1: Write tests for helper functions**
- [ ] **Step 2: Implement common helpers**

```go
// String returns a pointer to the given string. Convenience for optional fields.
func String(s string) *string { return &s }

// Int returns a pointer to the given int.
func Int(i int) *int { return &i }

// Bool returns a pointer to the given bool.
func Bool(b bool) *bool { return &b }

// Float64 returns a pointer to the given float64.
func Float64(f float64) *float64 { return &f }
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add helpers.go helpers_test.go
git commit -m "feat: add pointer helper functions"
```

---

### Task 24: Elicitation Types

**Files:**
- Modify: `hooks.go` (add elicitation-specific types if not already)
- Create: `elicitation.go`
- Create: `elicitation_test.go`

- [ ] **Step 1: Write tests for ElicitationRequest/Result**
- [ ] **Step 2: Implement types**

```go
// OnElicitation handles MCP elicitation requests.
type OnElicitation func(ctx context.Context, request ElicitationRequest) (*ElicitationResult, error)

type ElicitationRequest struct {
    ServerName      string                 `json:"serverName"`
    Message         string                 `json:"message"`
    Mode            *string                `json:"mode,omitempty"` // "form" | "url"
    URL             *string                `json:"url,omitempty"`
    ElicitationID   *string                `json:"elicitationId,omitempty"`
    RequestedSchema map[string]interface{} `json:"requestedSchema,omitempty"`
}

type ElicitationResult struct {
    Action  string                 `json:"action"` // "accept" | "decline" | "cancel"
    Content map[string]interface{} `json:"content,omitempty"`
}
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

```bash
git add elicitation.go elicitation_test.go
git commit -m "feat: add MCP elicitation request/result types"
```

---

### Task 25: Examples

**Files:**
- Create: `examples/basic/main.go`
- Create: `examples/streaming/main.go`
- Create: `examples/custom_tools/main.go`
- Create: `examples/permissions/main.go`
- Create: `examples/hooks/main.go`
- Create: `examples/session_management/main.go`

- [ ] **Step 1: Write basic example**

```go
// examples/basic/main.go
package main

import (
    "fmt"
    "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
    q := claudeagent.NewQuery(claudeagent.QueryParams{
        Prompt: "What is 2 + 2?",
        Options: &claudeagent.Options{
            SystemPrompt: "You are a helpful math assistant.",
            MaxTurns:     claudeagent.Int(1),
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

- [ ] **Step 2: Write streaming multi-turn example**
- [ ] **Step 3: Write custom tools example (MCP server)**
- [ ] **Step 4: Write permissions example**
- [ ] **Step 5: Write hooks example**
- [ ] **Step 6: Write session management example**
- [ ] **Step 7: Commit**

```bash
git add examples/
git commit -m "feat: add usage examples (basic, streaming, tools, permissions, hooks, sessions)"
```

---

### Task 26: Integration Tests

**Files:**
- Create: `integration_test.go`

- [ ] **Step 1: Write integration test that spawns real Claude Code**

```go
//go:build integration

package claudeagent_test

import (
    "testing"
    "github.com/anthropics/claude-agent-sdk-go"
)

func TestIntegration_BasicQuery(t *testing.T) {
    q := claudeagent.NewQuery(claudeagent.QueryParams{
        Prompt: "Reply with exactly: HELLO",
        Options: &claudeagent.Options{
            MaxTurns:       claudeagent.Int(1),
            PermissionMode: ptrTo(claudeagent.PermissionModeDontAsk),
            SystemPrompt:   "You are a test assistant. Follow instructions exactly.",
        },
    })
    defer q.Close()

    var result string
    for msg := range q.Messages() {
        if r, ok := msg.(*claudeagent.SDKResultSuccess); ok {
            result = r.Result
        }
    }
    if result == "" {
        t.Error("expected non-empty result")
    }
}

func TestIntegration_ListSessions(t *testing.T) {
    sessions, err := claudeagent.ListSessions(nil)
    if err != nil {
        t.Fatalf("ListSessions: %v", err)
    }
    // Just verify it doesn't error — may be empty
    _ = sessions
}

func TestIntegration_SupportedModels(t *testing.T) {
    q := claudeagent.NewQuery(claudeagent.QueryParams{
        Prompt: "hi",
        Options: &claudeagent.Options{MaxTurns: claudeagent.Int(1)},
    })
    defer q.Close()
    // Consume first message to ensure init completes
    <-q.Messages()

    models, err := q.SupportedModels(context.Background())
    if err != nil {
        t.Fatalf("SupportedModels: %v", err)
    }
    if len(models) == 0 {
        t.Error("expected at least one model")
    }
}
```

- [ ] **Step 2: Run integration tests** (requires Claude Code CLI)

Run: `go test -tags integration -v -timeout 120s ./...`

- [ ] **Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add integration tests (require Claude Code CLI)"
```

---

### Task 27: Sync Prompt Document

**Files:**
- Create: `docs/SYNC_PROMPT.md`

- [ ] **Step 1: Create the sync prompt for future upstream parity**

Create `docs/SYNC_PROMPT.md` — see below (this is the prompt that keeps the Go SDK in sync with new TypeScript SDK releases).

- [ ] **Step 2: Commit**

```bash
git add docs/SYNC_PROMPT.md
git commit -m "docs: add upstream sync prompt for TypeScript SDK parity"
```

---

## Dependency Graph

```
Task 1 (scaffolding)
  ├─> Task 2 (errors)
  ├─> Task 3 (enums)
  │     └─> Task 5 (messages) ←── Task 4 (JSON helpers)
  │           ├─> Task 6 (models)
  │           ├─> Task 7 (MCP)
  │           ├─> Task 8 (permissions)
  │           ├─> Task 9 (hooks)
  │           ├─> Task 13 (control)
  │           └─> Task 24 (elicitation)
  ├─> Task 10 (options) ←── Tasks 3,6,7,8,9,11
  ├─> Task 11 (sandbox)
  ├─> Task 12 (settings) ←── Tasks 7,11
  ├─> Task 14 (tools/) — independent
  └─> Task 23 (helpers) — independent

Task 15 (process) ←── Tasks 10
  └─> Task 16 (query) ←── Tasks 5,10,13,15
        └─> Task 17 (top-level API) ←── Task 16
              ├─> Task 18 (sessions)
              ├─> Task 19 (V2 sessions)
              ├─> Task 25 (examples)
              └─> Task 26 (integration tests)

Task 20 (bridge) ←── Tasks 5,7
Task 21 (browser) ←── Tasks 5,7,16
Task 22 (embed) — independent

Task 27 (sync prompt) — independent, can be done anytime
```

### Parallelizable task groups:
- **Group A** (types, can all run after Task 5): Tasks 6, 7, 8, 9, 11, 14, 23, 24
- **Group B** (depends on Group A): Tasks 10, 12, 13
- **Group C** (core runtime): Tasks 15, 16, 17
- **Group D** (features on top of runtime): Tasks 18, 19, 20, 21, 22
- **Group E** (examples and tests): Tasks 25, 26, 27

---

## Type Parity Checklist

Every exported type from `sdk.d.ts` must have a Go equivalent:

| TypeScript Type | Go Type | Task |
|---|---|---|
| `AbortError` | `AbortError` | 2 |
| `AccountInfo` | `AccountInfo` | 6 |
| `AgentDefinition` | `AgentDefinition` | 6 |
| `AgentInfo` | `AgentInfo` | 6 |
| `AgentMcpServerSpec` | `AgentMcpServerSpec` (interface{}) | 6 |
| `ApiKeySource` | `ApiKeySource` | 3 |
| `AsyncHookJSONOutput` | `AsyncHookJSONOutput` | 9 |
| `BaseHookInput` | `BaseHookInput` | 9 |
| `CanUseTool` | `CanUseTool` | 8 |
| `ConfigScope` | `ConfigScope` | 3 |
| `ElicitationRequest` | `ElicitationRequest` | 24 |
| `ElicitationResult` | `ElicitationResult` | 24 |
| `ExitReason` | `ExitReason` | 3 |
| `FastModeState` | `FastModeState` | 3 |
| `HookCallback` | `HookCallback` | 9 |
| `HookCallbackMatcher` | `HookCallbackMatcher` | 9 |
| `HookEvent` | `HookEvent` | 3 |
| `HookInput` (union) | Concrete types per event | 9 |
| `HookJSONOutput` | `HookJSONOutput` | 9 |
| `McpHttpServerConfig` | `McpHttpServerConfig` | 7 |
| `McpSSEServerConfig` | `McpSSEServerConfig` | 7 |
| `McpSdkServerConfig` | `McpSdkServerConfig` | 7 |
| `McpServerConfig` (union) | `McpServerConfig` (interface{}) | 7 |
| `McpServerStatus` | `McpServerStatus` | 7 |
| `McpSetServersResult` | `McpSetServersResult` | 7 |
| `McpStdioServerConfig` | `McpStdioServerConfig` | 7 |
| `ModelInfo` | `ModelInfo` | 6 |
| `ModelUsage` | `ModelUsage` | 5 |
| `NonNullableUsage` | `NonNullableUsage` | 5 |
| `OnElicitation` | `OnElicitation` | 24 |
| `Options` | `Options` | 10 |
| `OutputFormat` | `OutputFormat` | 10 |
| `PermissionBehavior` | `PermissionBehavior` | 3 |
| `PermissionMode` | `PermissionMode` | 3 |
| `PermissionResult` (union) | `PermissionResult` interface | 8 |
| `PermissionRuleValue` | `PermissionRuleValue` | 8 |
| `PermissionUpdate` (union) | `PermissionUpdate` interface | 8 |
| `PermissionUpdateDestination` | `PermissionUpdateDestination` | 3 |
| `Query` | `Query` | 16 |
| `RewindFilesResult` | `RewindFilesResult` | 16 |
| `SandboxSettings` | `SandboxSettings` | 11 |
| `SdkBeta` | `SdkBeta` | 3 |
| `SdkPluginConfig` | `SdkPluginConfig` | 10 |
| `SDKAssistantMessage` | `SDKAssistantMessage` | 5 |
| `SDKMessage` (union) | `SDKMessage` interface | 5 |
| `SDKPartialAssistantMessage` | `SDKPartialAssistantMessage` | 5 |
| `SDKResultError` | `SDKResultError` | 5 |
| `SDKResultSuccess` | `SDKResultSuccess` | 5 |
| `SDKSession` | `SDKSession` | 19 |
| `SDKSessionInfo` | `SDKSessionInfo` | 18 |
| `SDKSessionOptions` | `SDKSessionOptions` | 19 |
| `SDKSystemMessage` | `SDKSystemMessage` | 5 |
| `SDKUserMessage` | `SDKUserMessage` | 5 |
| `Settings` | `Settings` | 12 |
| `SettingSource` | `SettingSource` | 3 |
| `SlashCommand` | `SlashCommand` | 6 |
| All 23 hook input types | Corresponding Go structs | 9 |
| All hook-specific output types | Corresponding Go structs | 9 |
| All SDK tool input types | `tools/` package | 14 |
| All SDK tool output types | `tools/` package | 14 |
| `BridgeSessionHandle` | `bridge.BridgeSessionHandle` | 20 |
| `AttachBridgeSessionOptions` | `bridge.AttachBridgeSessionOptions` | 20 |
| `BrowserQueryOptions` | `browser.BrowserQueryOptions` | 21 |

### Function Parity Checklist

| TypeScript Function | Go Function | Task |
|---|---|---|
| `query()` | `NewQuery()` | 17 |
| `listSessions()` | `ListSessions()` | 18 |
| `getSessionInfo()` | `GetSessionInfo()` | 18 |
| `getSessionMessages()` | `GetSessionMessages()` | 18 |
| `forkSession()` | `ForkSession()` | 18 |
| `renameSession()` | `RenameSession()` | 18 |
| `tagSession()` | `TagSession()` | 18 |
| `createSdkMcpServer()` | N/A (Go uses interface) | — |
| `tool()` | N/A (Go uses struct) | — |
| `unstable_v2_createSession()` | `CreateSession()` | 19 |
| `unstable_v2_resumeSession()` | `ResumeSession()` | 19 |
| `attachBridgeSession()` | `bridge.AttachBridgeSession()` | 20 |
| Browser `query()` | `browser.Query()` | 21 |
| Query.interrupt() | `Query.Interrupt()` | 16 |
| Query.setPermissionMode() | `Query.SetPermissionMode()` | 16 |
| Query.setModel() | `Query.SetModel()` | 16 |
| Query.setMaxThinkingTokens() | `Query.SetMaxThinkingTokens()` | 16 |
| Query.applyFlagSettings() | `Query.ApplyFlagSettings()` | 16 |
| Query.initializationResult() | `Query.InitializationResult()` | 16 |
| Query.supportedCommands() | `Query.SupportedCommands()` | 16 |
| Query.supportedModels() | `Query.SupportedModels()` | 16 |
| Query.supportedAgents() | `Query.SupportedAgents()` | 16 |
| Query.mcpServerStatus() | `Query.McpServerStatus()` | 16 |
| Query.accountInfo() | `Query.AccountInfo()` | 16 |
| Query.rewindFiles() | `Query.RewindFiles()` | 16 |
| Query.reconnectMcpServer() | `Query.ReconnectMcpServer()` | 16 |
| Query.toggleMcpServer() | `Query.ToggleMcpServer()` | 16 |
| Query.setMcpServers() | `Query.SetMcpServers()` | 16 |
| Query.streamInput() | `Query.StreamInput()` | 16 |
| Query.stopTask() | `Query.StopTask()` | 16 |
| Query.close() | `Query.Close()` | 16 |
| `unstable_v2_prompt()` | `Prompt()` | 19 |

---

## Addendum: Review Fixes (Addressing Gaps Found by Plan Reviewer)

The following additions address all gaps identified during plan review.

### Task 28: PromptRequest/Response Types (MISSING FROM ORIGINAL PLAN)

**Files:**
- Modify: `messages.go` (add types)
- Modify: `messages_test.go`

These are exported types used for CLI-level prompt requests/responses.

- [ ] **Step 1: Write failing test**

```go
func TestPromptRequest_JSON(t *testing.T) {
    raw := `{"prompt":"p1","message":"Choose option","options":[{"key":"a","label":"Option A","description":"desc"}]}`
    var r PromptRequest
    if err := json.Unmarshal([]byte(raw), &r); err != nil {
        t.Fatal(err)
    }
    if r.Prompt != "p1" { t.Errorf("Prompt = %q", r.Prompt) }
    if len(r.Options) != 1 { t.Errorf("Options = %v", r.Options) }
}
```

- [ ] **Step 2: Implement types**

```go
// PromptRequest is a prompt displayed to the user with options.
type PromptRequest struct {
    Prompt  string                `json:"prompt"`
    Message string                `json:"message"`
    Options []PromptRequestOption `json:"options"`
}

type PromptRequestOption struct {
    Key         string  `json:"key"`
    Label       string  `json:"label"`
    Description *string `json:"description,omitempty"`
}

type PromptResponse struct {
    PromptResponse string `json:"prompt_response"`
    Selected       string `json:"selected"`
}
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

---

### Task 29: Missing Control Request Types (10 members)

**Files:**
- Modify: `control.go`
- Modify: `control_test.go`

These control request subtypes were missing from the `SDKControlRequestInner` union:

- [ ] **Step 1: Add all missing control request types**

```go
// SDKControlMcpAuthenticateRequest initiates MCP server authentication.
type SDKControlMcpAuthenticateRequest struct {
    Subtype    string `json:"subtype"` // "mcp_authenticate"
    ServerName string `json:"serverName"`
}

// SDKControlMcpClearAuthRequest clears MCP server auth credentials.
type SDKControlMcpClearAuthRequest struct {
    Subtype    string `json:"subtype"` // "mcp_clear_auth"
    ServerName string `json:"serverName"`
}

// SDKControlMcpOAuthCallbackUrlRequest provides an OAuth callback URL.
type SDKControlMcpOAuthCallbackUrlRequest struct {
    Subtype     string `json:"subtype"` // "mcp_oauth_callback_url"
    ServerName  string `json:"serverName"`
    CallbackUrl string `json:"callbackUrl"`
}

// SDKControlClaudeAuthenticateRequest initiates Claude authentication.
type SDKControlClaudeAuthenticateRequest struct {
    Subtype string `json:"subtype"` // "claude_authenticate"
}

// SDKControlClaudeOAuthCallbackRequest provides a Claude OAuth callback.
type SDKControlClaudeOAuthCallbackRequest struct {
    Subtype     string `json:"subtype"` // "claude_oauth_callback"
    CallbackUrl string `json:"callbackUrl"`
}

// SDKControlClaudeOAuthWaitForCompletionRequest waits for OAuth to complete.
type SDKControlClaudeOAuthWaitForCompletionRequest struct {
    Subtype string `json:"subtype"` // "claude_oauth_wait_for_completion"
}

// SDKControlRemoteControlRequest sends a remote control command.
type SDKControlRemoteControlRequest struct {
    Subtype string                 `json:"subtype"` // "remote_control"
    Action  string                 `json:"action"`
    Data    map[string]interface{} `json:"data,omitempty"`
}

// SDKControlSetProactiveRequest toggles proactive behavior.
type SDKControlSetProactiveRequest struct {
    Subtype   string `json:"subtype"` // "set_proactive"
    Proactive bool   `json:"proactive"`
}

// SDKControlGenerateSessionTitleRequest requests AI-generated session title.
type SDKControlGenerateSessionTitleRequest struct {
    Subtype string `json:"subtype"` // "generate_session_title"
}

// SDKControlSideQuestionRequest sends a side question outside the main turn.
type SDKControlSideQuestionRequest struct {
    Subtype  string `json:"subtype"` // "side_question"
    Question string `json:"question"`
}
```

- [ ] **Step 2: Update control dispatcher to handle `control_cancel_request` as separate top-level type**

Note: `SDKControlCancelRequest` has `type: "control_cancel_request"` (NOT `type: "control_request"`).
It must be dispatched BEFORE checking for `control_request` in the message parser.

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

---

### Task 30: ThinkingConfig Types

**Files:**
- Modify: `options.go`
- Modify: `options_test.go`

- [ ] **Step 1: Write failing test**

```go
func TestThinkingConfig_Adaptive_JSON(t *testing.T) {
    raw := `{"type":"adaptive"}`
    var tc ThinkingConfig
    if err := json.Unmarshal([]byte(raw), &tc); err != nil {
        t.Fatal(err)
    }
}
```

- [ ] **Step 2: Implement ThinkingConfig union types**

```go
// ThinkingConfig controls Claude's thinking/reasoning behavior.
type ThinkingConfig struct {
    Type         string `json:"type"` // "adaptive" | "enabled" | "disabled"
    BudgetTokens *int   `json:"budgetTokens,omitempty"` // only for "enabled"
}

// Convenience constructors
func ThinkingAdaptive() ThinkingConfig { return ThinkingConfig{Type: "adaptive"} }
func ThinkingEnabled(budgetTokens int) ThinkingConfig { return ThinkingConfig{Type: "enabled", BudgetTokens: Int(budgetTokens)} }
func ThinkingDisabled() ThinkingConfig { return ThinkingConfig{Type: "disabled"} }
```

- [ ] **Step 3: Run tests, verify pass**
- [ ] **Step 4: Commit**

---

### Task 31: ToolConfig, SpawnedProcess, SpawnOptions Types

**Files:**
- Modify: `options.go`
- Modify: `process.go`

- [ ] **Step 1: Implement missing types**

```go
// ToolConfig provides per-tool configuration.
type ToolConfig struct {
    AskUserQuestion *AskUserQuestionConfig `json:"askUserQuestion,omitempty"`
}

type AskUserQuestionConfig struct {
    PreviewFormat *string `json:"previewFormat,omitempty"` // "markdown" | "html"
}

// SpawnOptions configures how the Claude Code process is spawned.
type SpawnOptions struct {
    Command string
    Args    []string
    Cwd     string
    Env     map[string]string
    Signal  <-chan struct{}
}

// SpawnedProcess wraps a running Claude Code subprocess.
type SpawnedProcess interface {
    Stdin() io.WriteCloser
    Stdout() io.ReadCloser
    Stderr() io.ReadCloser
    Wait() error
    Kill() error
}

// SpawnFunc allows custom process spawning (VMs, containers, remote).
type SpawnFunc func(opts SpawnOptions) SpawnedProcess
```

- [ ] **Step 2: Run tests, verify pass**
- [ ] **Step 3: Commit**

---

### Task 32: Missing MCP Types (McpSdkServerConfigWithInstance, McpServerConfigForProcessTransport, McpServerStatusConfig)

**Files:**
- Modify: `mcp.go`

- [ ] **Step 1: Add missing types**

```go
// McpSdkServerConfigWithInstance is the SDK server config with a live server instance.
// In Go, this is not directly usable since Go doesn't have MCP SDK.
// Provided for type completeness — users provide custom implementations.
type McpSdkServerConfigWithInstance struct {
    McpSdkServerConfig
    Instance interface{} // User-provided MCP server implementation
}

// McpServerConfigForProcessTransport is the union of serializable MCP configs.
// In Go, use McpStdioServerConfig, McpSSEServerConfig, McpHttpServerConfig,
// or McpSdkServerConfig directly.
type McpServerConfigForProcessTransport = interface{}

// McpServerStatusConfig is the union of McpServerConfigForProcessTransport
// and McpClaudeAIProxyServerConfig.
type McpServerStatusConfig = interface{}
```

- [ ] **Step 2: Commit**

---

### Task 33: SDKResultMessage Type Alias

**Files:**
- Modify: `messages.go`

- [ ] **Step 1: Add type alias comment and helper**

```go
// SDKResultMessage is either *SDKResultSuccess or *SDKResultError.
// Use a type switch on SDKMessage to distinguish.
// This is a documentation alias — Go doesn't have union type aliases.
// Provided here for parity reference with the TypeScript SDK.

// IsResultMessage returns true if the given SDKMessage is a result message.
func IsResultMessage(msg SDKMessage) bool {
    switch msg.(type) {
    case *SDKResultSuccess, *SDKResultError:
        return true
    }
    return false
}
```

- [ ] **Step 2: Commit**

---

### Task 34: V2 Prompt Function

**Files:**
- Modify: `session_v2.go`

- [ ] **Step 1: Implement unstable_v2_prompt equivalent**

```go
// Prompt is a convenience function for single-turn V2 queries.
// It creates a session, sends the prompt, collects the result, and closes.
// Returns the result message (success or error).
func Prompt(prompt string, opts SDKSessionOptions) (SDKMessage, error) {
    sess, err := CreateSession(opts)
    if err != nil {
        return nil, err
    }
    defer sess.Close()

    if err := sess.Send(prompt); err != nil {
        return nil, err
    }

    var result SDKMessage
    for msg := range sess.Stream() {
        if IsResultMessage(msg) {
            result = msg
        }
    }
    if result == nil {
        return nil, fmt.Errorf("no result received")
    }
    return result, nil
}
```

- [ ] **Step 2: Commit**

---

### Task 35: Bridge API — Missing Callbacks

**Files:**
- Modify: `bridge/types.go`

- [ ] **Step 1: Add missing callbacks to AttachBridgeSessionOptions**

```go
type AttachBridgeSessionOptions struct {
    SessionID            string
    IngressToken         string
    ApiBaseUrl           string
    Epoch                *int
    InitialSequenceNum   *int
    HeartbeatIntervalMs  *int
    OnInboundMessage     func(msg claudeagent.SDKMessage) error
    OnPermissionResponse func(res claudeagent.SDKControlResponse)
    OnInterrupt          func()
    OnSetModel           func(model *string)              // ← was missing
    OnSetMaxThinkingTokens func(tokens *int)              // ← was missing
    OnSetPermissionMode  func(mode claudeagent.PermissionMode) *BridgePermissionModeResult
    OnClose              func(code *int)
}

type BridgePermissionModeResult struct {
    OK    bool
    Error string // only if OK is false
}
```

- [ ] **Step 2: Commit**

---

### Task 36: JSON Parser Fallthrough Handler

**Files:**
- Modify: `json.go`

- [ ] **Step 1: Add unknown message fallthrough**

Replace the `default` error return in `ParseSDKMessage` and `parseSystemMessage` with a raw fallback:

```go
// SDKRawMessage holds an unrecognized message type as raw JSON.
type SDKRawMessage struct {
    RawType    string          `json:"type"`
    RawSubtype string          `json:"subtype,omitempty"`
    Raw        json.RawMessage `json:"-"`
}
func (m *SDKRawMessage) sdkMessage()        {}
func (m *SDKRawMessage) MessageType() string { return m.RawType }

// In ParseSDKMessage default case:
// return &SDKRawMessage{RawType: env.Type, Raw: json.RawMessage(data)}, nil
```

This prevents silent failures when the CLI emits new/internal message types.

- [ ] **Step 2: Commit**

---

### Updated Type Parity Table (additions)

| TypeScript Type | Go Type | Task |
|---|---|---|
| `PromptRequest` | `PromptRequest` | 28 |
| `PromptResponse` | `PromptResponse` | 28 |
| `PromptRequestOption` | `PromptRequestOption` | 28 |
| `SessionMessage` | `SessionMessage` | 18 |
| `ThinkingConfig` | `ThinkingConfig` | 30 |
| `ThinkingAdaptive` | `ThinkingAdaptive()` constructor | 30 |
| `ThinkingEnabled` | `ThinkingEnabled()` constructor | 30 |
| `ThinkingDisabled` | `ThinkingDisabled()` constructor | 30 |
| `ToolConfig` | `ToolConfig` | 31 |
| `SpawnedProcess` | `SpawnedProcess` interface | 31 |
| `SpawnOptions` | `SpawnOptions` | 31 |
| `McpSdkServerConfigWithInstance` | `McpSdkServerConfigWithInstance` | 32 |
| `McpServerConfigForProcessTransport` | `McpServerConfigForProcessTransport` | 32 |
| `McpServerStatusConfig` | `McpServerStatusConfig` | 32 |
| `SDKResultMessage` | `IsResultMessage()` helper | 33 |
| `ForkSessionResult` | `ForkSessionResult` | 18 |
| `ForkSessionOptions` | `ForkSessionOptions` | 18 |
| `GetSessionInfoOptions` | `GetSessionInfoOptions` | 18 |
| `GetSessionMessagesOptions` | `GetSessionMessagesOptions` | 18 |
| `ListSessionsOptions` | `ListSessionsOptions` | 18 |
| `SessionMutationOptions` | `SessionMutationOptions` | 18 |
| `SyncHookJSONOutput` | `SyncHookJSONOutput` | 9 |
| `SDKControlRequest` | `SDKControlRequest` | 13 |
| `SDKControlResponse` | `SDKControlResponse` | 13 |
| `SDKControlInitializeResponse` | `SDKControlInitializeResponse` | 13 |
| `SDKRawMessage` | `SDKRawMessage` (fallthrough) | 36 |
| 10 missing control request types | See Task 29 | 29 |

### Updated Function Parity Table (additions)

| TypeScript Function | Go Function | Task |
|---|---|---|
| `unstable_v2_prompt()` | `Prompt()` | 34 |

### Updated Dependency Graph (additional tasks)

Tasks 28–36 all depend on their parent files already existing:
- Tasks 28, 33, 36 depend on Task 5 (messages)
- Task 29 depends on Task 13 (control)
- Tasks 30, 31 depend on Task 10 (options)
- Task 32 depends on Task 7 (MCP)
- Task 34 depends on Task 19 (V2 sessions)
- Task 35 depends on Task 20 (bridge)

All can be parallelized within their dependency group.
