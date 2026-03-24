# Sync Go SDK with upstream TypeScript SDK

You are syncing the Go SDK (`github.com/anthropics/claude-agent-sdk-go`) with the latest release of the upstream TypeScript SDK (`@anthropic-ai/claude-agent-sdk`).

Source repo: https://github.com/anthropics/claude-agent-sdk-typescript
NPM package: @anthropic-ai/claude-agent-sdk

## Step 1: Check current versions

Read the current Go SDK version from `claudeagent.go` (look for `const Version`).

Then fetch the latest TypeScript SDK version:
```bash
npm view @anthropic-ai/claude-agent-sdk version
```

Compare the two. If they match, report "Go SDK is up to date with TypeScript SDK vX.Y.Z" and stop.

## Step 2: Download the new TypeScript SDK

```bash
cd /tmp && npm pack @anthropic-ai/claude-agent-sdk && tar xzf anthropic-ai-claude-agent-sdk-*.tgz
```

## Step 3: Read the CHANGELOG

Read `/tmp/package/CHANGELOG.md` and identify all entries NEWER than the current Go SDK version. Summarize what changed:
- New types added
- New fields on existing types
- New functions added
- New enum values
- Behavior changes
- Bug fixes

## Step 4: Diff the type definitions

For each `.d.ts` file, compare against the Go SDK types:

| TypeScript file | Go SDK location |
|---|---|
| `sdk.d.ts` | Root package: `messages.go`, `options.go`, `models.go`, `mcp.go`, `permissions.go`, `hooks.go`, `settings.go`, `sandbox.go`, `control.go`, `enums.go`, `session.go`, `session_v2.go`, `elicitation.go`, `query.go` |
| `sdk-tools.d.ts` | `tools/` package |
| `bridge.d.ts` | `bridge/` package |
| `browser-sdk.d.ts` | `browser/` package |

## Step 5: Apply changes (TDD)

For each change found:

1. **Write a failing test first** that exercises the new type/field/function
2. **Implement the change** in the correct Go file
3. **Verify the test passes**
4. **Commit** with message: `feat: sync <specific change> from TypeScript SDK vX.Y.Z`

### Type mapping reference

| TypeScript | Go |
|---|---|
| `string` | `string` |
| `number` | `int` or `float64` |
| `boolean` | `bool` |
| `string \| undefined` | `*string` with `omitempty` |
| `number \| null` | `*int` or `*float64` |
| `A \| B \| C` (union) | Interface with marker method + concrete structs |
| `Record<string, T>` | `map[string]T` |
| `T[]` | `[]T` |
| `Promise<T>` | `(T, error)` return |
| `AsyncGenerator<T>` | `<-chan T` |
| `AbortController` | `context.Context` |
| `unknown` / `any` | `interface{}` |
| `UUID` | `string` |

### Change type checklist

**New message subtype:**
- Add struct in `messages.go` with `sdkMessage()` and `MessageType()` methods
- Add case in `ParseSDKMessage()` or `parseSystemMessage()` in `json.go`
- Add to interface satisfaction test in `messages_test.go`
- Write JSON round-trip test

**New hook event:**
- Add `HookEvent` const in `enums.go`
- Add to `AllHookEvents()` return
- Add input struct in `hooks.go`
- Add output struct if applicable
- Update `TestHookEvent_Values` count

**New Option field:**
- Add field to `Options` struct in `options.go`
- Add CLI flag mapping in `process.go` `buildProcessArgs()`
- If it maps to an initialize control request field, add to `control.go`

**New Query method:**
- Add method to `Query` struct in `query.go`
- Add control request/response types in `control.go`

**New tool type:**
- Add input/output structs in appropriate `tools/*.go` file

**Settings field added:**
- Add to `Settings` struct in `settings.go`

## Step 6: Update version constants

In `claudeagent.go`, update:
```go
const Version = "NEW_VERSION"
const ClaudeCodeVersion = "NEW_CLI_VERSION"  // from package.json claudeCodeVersion field
```

## Step 7: Final verification

```bash
go test -timeout 60s -cover ./...
go build ./...
```

All tests must pass. Report coverage numbers.

## Step 8: Summary commit

```bash
git add -A
git commit -m "feat: sync with TypeScript SDK vX.Y.Z

Changes:
- [list each change]
"
```

Report what was synced and the new version numbers.
