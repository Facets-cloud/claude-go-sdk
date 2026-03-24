# Upstream Sync Prompt — Claude Agent SDK Go ↔ TypeScript Parity

> **Use this prompt whenever a new version of `@anthropic-ai/claude-agent-sdk` is released.**
> Copy this entire prompt into a new Claude Code session to sync the Go SDK.

---

## Prompt

```
You are syncing the Go SDK (`github.com/anthropics/claude-agent-sdk-go`) with a new
release of the TypeScript SDK (`@anthropic-ai/claude-agent-sdk`).

## Steps

1. **Get the new TypeScript SDK version:**
   Run: `npm pack @anthropic-ai/claude-agent-sdk && tar xzf anthropic-ai-claude-agent-sdk-*.tgz`
   This gives you the latest `package/sdk.d.ts`, `package/sdk-tools.d.ts`,
   `package/bridge.d.ts`, `package/browser-sdk.d.ts`, and `package/CHANGELOG.md`.

2. **Read the CHANGELOG entries** since the last synced version (check `Version` const in
   `claudeagent.go` for the current Go SDK version).

3. **Diff the type definitions:**
   For each `.d.ts` file, compare against the Go SDK types:
   - `sdk.d.ts` → root package types (`messages.go`, `options.go`, `models.go`, `mcp.go`,
     `permissions.go`, `hooks.go`, `settings.go`, `sandbox.go`, `control.go`, `enums.go`,
     `session.go`, `session_v2.go`, `elicitation.go`)
   - `sdk-tools.d.ts` → `tools/` package
   - `bridge.d.ts` → `bridge/` package
   - `browser-sdk.d.ts` → `browser/` package

4. **For each change, follow this pattern:**
   a. If a new TYPE was added → create the Go struct/interface in the correct file
   b. If a new FIELD was added to an existing type → add the field to the Go struct
   c. If a FIELD was removed → remove it from the Go struct
   d. If a new FUNCTION was added → implement in the correct file
   e. If a new ENUM VALUE was added → add the const
   f. If behavior changed → update implementation + tests

5. **For each change, write a test FIRST (TDD):**
   - Write a failing test that exercises the new type/field/function
   - Implement the change
   - Verify the test passes

6. **Update version constants:**
   In `claudeagent.go`, update:
   ```go
   const Version = "NEW_VERSION"
   const ClaudeCodeVersion = "NEW_CLI_VERSION"
   ```

7. **Run the full test suite:**
   ```bash
   go test ./...
   ```

8. **Commit with a message like:**
   ```
   feat: sync with TypeScript SDK vX.Y.Z

   Changes:
   - Added FooType
   - Added Bar field to Options
   - Updated BazMessage with new subtype
   ```

## Type Mapping Reference

| TypeScript | Go |
|---|---|
| `string` | `string` |
| `number` | `int` or `float64` |
| `boolean` | `bool` |
| `null` | nil (pointer type) |
| `string \| undefined` | `*string` with `omitempty` |
| `number \| null` | `*int` or `*float64` |
| `A \| B \| C` (union) | Interface with marker method + concrete structs |
| `Record<string, T>` | `map[string]T` |
| `T[]` | `[]T` |
| `Partial<T>` | Same struct, all fields have `omitempty` |
| `Promise<T>` | `(T, error)` return |
| `AsyncGenerator<T>` | `<-chan T` |
| `AbortController` | `context.Context` |
| `unknown` / `any` | `interface{}` |
| `readonly` | No equivalent needed (Go has no const fields) |
| `UUID` | `string` |
| `Readable` / `Writable` | `io.Reader` / `io.Writer` |

## Checklist for common sync operations

### New message subtype added
- [ ] Add struct in `messages.go`
- [ ] Add `sdkMessage()` and `MessageType()` methods
- [ ] Add case in `parseSystemMessage()` or `ParseSDKMessage()` in `json.go`
- [ ] Add to `SDKMessage` type list in test
- [ ] Write JSON round-trip test

### New hook event added
- [ ] Add `HookEvent` const in `enums.go`
- [ ] Add to `AllHookEvents()` return
- [ ] Add input struct in `hooks.go`
- [ ] Add output struct if applicable
- [ ] Update test count in `TestHookEvent_Values`

### New Option field added
- [ ] Add field to `Options` struct in `options.go`
- [ ] Add CLI flag mapping in `process.go` `buildProcessArgs()`
- [ ] If it maps to an initialize control request field, add to `control.go`
- [ ] Write test

### New Query method added
- [ ] Add method to `Query` struct in `query.go`
- [ ] Add control request type in `control.go`
- [ ] Add control response handling
- [ ] Write test

### New tool type added
- [ ] Add input struct in appropriate `tools/*.go` file
- [ ] Add output struct
- [ ] Write test

### Settings field added
- [ ] Add to `Settings` struct in `settings.go`
- [ ] Write JSON round-trip test

## Files to ALWAYS check during sync
1. `package/sdk.d.ts` — main API surface
2. `package/sdk-tools.d.ts` — tool schemas
3. `package/bridge.d.ts` — bridge API
4. `package/browser-sdk.d.ts` — browser API
5. `package/CHANGELOG.md` — what changed
6. `package/package.json` — version number, claudeCodeVersion
```

---

## Automated Sync Check (CI)

To detect drift, add this to CI:

```bash
#!/bin/bash
# scripts/check-parity.sh
LATEST=$(npm view @anthropic-ai/claude-agent-sdk version 2>/dev/null)
CURRENT=$(grep 'const Version' claudeagent.go | grep -oP '"[^"]+"' | tr -d '"')
if [ "$LATEST" != "$CURRENT" ]; then
  echo "⚠️  Go SDK ($CURRENT) is behind TypeScript SDK ($LATEST)"
  echo "Run the sync prompt in docs/SYNC_PROMPT.md"
  exit 1
fi
```
