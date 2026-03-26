package claudeagent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// defaultSpawnedProcess wraps an os/exec.Cmd as a SpawnedProcess.
type defaultSpawnedProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
}

func (p *defaultSpawnedProcess) Stdin() io.WriteCloser  { return p.stdin }
func (p *defaultSpawnedProcess) Stdout() io.ReadCloser  { return p.stdout }
func (p *defaultSpawnedProcess) Wait() error            { return p.cmd.Wait() }
func (p *defaultSpawnedProcess) Kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}

// defaultSpawn spawns the Claude Code CLI as an os/exec subprocess.
func defaultSpawn(opts SpawnOptions) SpawnedProcess {
	cmd := exec.Command(opts.Command, opts.Args...)
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}
	if opts.Env != nil {
		env := os.Environ()
		for k, v := range opts.Env {
			env = append(env, k+"="+v)
		}
		cmd.Env = env
	}

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = os.Stderr

	return &defaultSpawnedProcess{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
	}
}

// buildProcessArgs constructs the CLI argument list from Options.
// Flags are cross-referenced with the TypeScript SDK source (sdk.mjs) for accuracy.
func buildProcessArgs(opts *Options, prompt string) []string {
	args := []string{
		"--output-format", "stream-json",
		"--verbose",
	}

	if opts == nil {
		if prompt != "" {
			args = append(args, "--print", "--", prompt)
		}
		return args
	}

	// Thinking config (must be before other flags per TS SDK ordering)
	if opts.Thinking != nil {
		switch opts.Thinking.Type {
		case "adaptive":
			args = append(args, "--thinking", "adaptive")
		case "disabled":
			args = append(args, "--thinking", "disabled")
		case "enabled":
			if opts.Thinking.BudgetTokens != nil {
				args = append(args, "--max-thinking-tokens", fmt.Sprintf("%d", *opts.Thinking.BudgetTokens))
			} else {
				args = append(args, "--thinking", "adaptive")
			}
		}
	} else if opts.MaxThinkingTokens != nil {
		if *opts.MaxThinkingTokens == 0 {
			args = append(args, "--thinking", "disabled")
		} else {
			args = append(args, "--max-thinking-tokens", fmt.Sprintf("%d", *opts.MaxThinkingTokens))
		}
	}

	if opts.Effort != nil {
		args = append(args, "--effort", *opts.Effort)
	}

	if opts.MaxTurns != nil {
		args = append(args, "--max-turns", fmt.Sprintf("%d", *opts.MaxTurns))
	}

	if opts.MaxBudgetUsd != nil {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", *opts.MaxBudgetUsd))
	}

	if opts.Model != nil {
		args = append(args, "--model", *opts.Model)
	}

	if opts.Agent != nil {
		args = append(args, "--agent", *opts.Agent)
	}

	if len(opts.Betas) > 0 {
		betaStrs := make([]string, len(opts.Betas))
		for i, b := range opts.Betas {
			betaStrs[i] = string(b)
		}
		args = append(args, "--betas", strings.Join(betaStrs, ","))
	}

	// JSON schema for structured output
	if opts.OutputFormat != nil {
		ofJSON, err := json.Marshal(opts.OutputFormat.Schema)
		if err == nil {
			args = append(args, "--json-schema", string(ofJSON))
		}
	}

	if opts.DebugFile != nil {
		args = append(args, "--debug-file", *opts.DebugFile)
	} else if opts.Debug != nil && *opts.Debug {
		args = append(args, "--debug")
	}

	// Permission prompt tool (TS SDK: --permission-prompt-tool)
	if opts.CanUseTool != nil {
		if opts.PermissionPromptToolName != nil {
			// Can't use both
		} else {
			args = append(args, "--permission-prompt-tool", "stdio")
		}
	} else if opts.PermissionPromptToolName != nil {
		args = append(args, "--permission-prompt-tool", *opts.PermissionPromptToolName)
	}

	if opts.Continue != nil && *opts.Continue {
		args = append(args, "--continue")
	}

	if opts.Resume != nil {
		args = append(args, "--resume", *opts.Resume)
	}

	// AllowedTools — comma-separated (matches TS SDK: --allowedTools X,Y)
	if len(opts.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(opts.AllowedTools, ","))
	}

	// DisallowedTools — comma-separated
	if len(opts.DisallowedTools) > 0 {
		args = append(args, "--disallowedTools", strings.Join(opts.DisallowedTools, ","))
	}

	// Tools
	if opts.Tools != nil {
		switch t := opts.Tools.(type) {
		case []string:
			if len(t) == 0 {
				args = append(args, "--tools", "")
			} else {
				args = append(args, "--tools", strings.Join(t, ","))
			}
		case ToolPreset:
			args = append(args, "--tools", "default")
		}
	}

	// MCP servers — pass as JSON string (matches TS SDK)
	if len(opts.McpServers) > 0 {
		mcpConfig := map[string]interface{}{
			"mcpServers": opts.McpServers,
		}
		mcpJSON, err := json.Marshal(mcpConfig)
		if err == nil {
			args = append(args, "--mcp-config", string(mcpJSON))
		}
	}

	// Setting sources — comma-separated
	if len(opts.SettingSources) > 0 {
		sourceStrs := make([]string, len(opts.SettingSources))
		for i, s := range opts.SettingSources {
			sourceStrs[i] = string(s)
		}
		args = append(args, "--setting-sources", strings.Join(sourceStrs, ","))
	}

	if opts.StrictMcpConfig != nil && *opts.StrictMcpConfig {
		args = append(args, "--strict-mcp-config")
	}

	if opts.PermissionMode != nil {
		args = append(args, "--permission-mode", string(*opts.PermissionMode))
	}

	if opts.AllowDangerouslySkipPermissions != nil && *opts.AllowDangerouslySkipPermissions {
		args = append(args, "--allow-dangerously-skip-permissions")
	}

	if opts.FallbackModel != nil {
		args = append(args, "--fallback-model", *opts.FallbackModel)
	}

	if opts.IncludePartialMessages != nil && *opts.IncludePartialMessages {
		args = append(args, "--include-partial-messages")
	}

	for _, dir := range opts.AdditionalDirectories {
		args = append(args, "--add-dir", dir)
	}

	for _, plugin := range opts.Plugins {
		if plugin.Type == "local" {
			args = append(args, "--plugin-dir", plugin.Path)
		}
	}

	if opts.ForkSession != nil && *opts.ForkSession {
		args = append(args, "--fork-session")
	}

	if opts.ResumeSessionAt != nil {
		args = append(args, "--resume-session-at", *opts.ResumeSessionAt)
	}

	if opts.SessionID != nil {
		args = append(args, "--session-id", *opts.SessionID)
	}

	if opts.PersistSession != nil && !*opts.PersistSession {
		args = append(args, "--no-session-persistence")
	}

	// System prompt
	if opts.SystemPrompt != nil {
		switch sp := opts.SystemPrompt.(type) {
		case string:
			args = append(args, "--system-prompt", sp)
		case SystemPromptPreset:
			if sp.Append != nil {
				args = append(args, "--append-system-prompt", *sp.Append)
			}
		}
	}

	if len(opts.Agents) > 0 {
		agentsJSON, err := json.Marshal(opts.Agents)
		if err == nil {
			args = append(args, "--agents", string(agentsJSON))
		}
	}

	// Settings — passed via extra args mechanism in TS SDK.
	// Sandbox is embedded inside settings.
	if opts.Settings != nil || opts.Sandbox != nil {
		var settingsObj map[string]interface{}
		switch s := opts.Settings.(type) {
		case string:
			args = append(args, "--settings", s)
		case nil:
			settingsObj = make(map[string]interface{})
		default:
			settingsJSON, err := json.Marshal(s)
			if err == nil {
				json.Unmarshal(settingsJSON, &settingsObj)
			}
		}
		if opts.Sandbox != nil && settingsObj != nil {
			settingsObj["sandbox"] = opts.Sandbox
			combined, err := json.Marshal(settingsObj)
			if err == nil {
				args = append(args, "--settings", string(combined))
			}
		} else if settingsObj != nil && len(settingsObj) > 0 {
			combined, err := json.Marshal(settingsObj)
			if err == nil {
				args = append(args, "--settings", string(combined))
			}
		}
	}

	// Options handled via environment or init config (not CLI flags):
	// - Cwd: handled via cmd.Dir
	// - EnableFileCheckpointing: set via env CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING
	// - ToolConfig: set via env CLAUDE_CODE_QUESTION_PREVIEW_FORMAT
	// - PromptSuggestions: passed in initialize control request
	// - AgentProgressSummaries: passed in initialize control request

	// Extra args — escape hatch for any flags not covered above
	for k, v := range opts.ExtraArgs {
		if v == nil {
			args = append(args, "--"+k)
		} else {
			args = append(args, "--"+k, *v)
		}
	}

	// Prompt — in print mode, use --print with positional arg.
	// Use "--" separator to prevent variadic flags from consuming the prompt.
	if prompt != "" {
		args = append(args, "--print", "--", prompt)
	}

	return args
}

// processManager orchestrates the stdin/stdout/stderr goroutines for a
// SpawnedProcess and exposes a channel of parsed SDKMessages.
type processManager struct {
	proc     SpawnedProcess
	ctx      context.Context
	messages chan SDKMessage
	stderrCb func(string)

	stdinMu sync.Mutex
	done    chan struct{}
}

// newProcessManager creates a processManager and starts the stdout reader goroutine.
func newProcessManager(proc SpawnedProcess, ctx context.Context, stderrCb func(string)) *processManager {
	pm := &processManager{
		proc:     proc,
		ctx:      ctx,
		messages: make(chan SDKMessage, 64),
		stderrCb: stderrCb,
		done:     make(chan struct{}),
	}
	go pm.readStdout()
	return pm
}

// Messages returns the channel of parsed messages from stdout.
func (pm *processManager) Messages() <-chan SDKMessage {
	return pm.messages
}

// WriteJSON marshals v as JSON and writes it as a single line to the process stdin.
func (pm *processManager) WriteJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal stdin message: %w", err)
	}
	data = append(data, '\n')

	pm.stdinMu.Lock()
	defer pm.stdinMu.Unlock()

	_, err = pm.proc.Stdin().Write(data)
	if err != nil {
		return fmt.Errorf("write to stdin: %w", err)
	}
	return nil
}

// Kill terminates the underlying process.
func (pm *processManager) Kill() error {
	return pm.proc.Kill()
}

// Wait waits for the underlying process to exit.
func (pm *processManager) Wait() error {
	return pm.proc.Wait()
}

// readStdout reads JSON lines from the process stdout, parses each into an
// SDKMessage, and sends it on the messages channel. Malformed lines are
// silently skipped. The channel is closed when stdout is exhausted or the
// context is cancelled.
func (pm *processManager) readStdout() {
	defer close(pm.messages)
	defer close(pm.done)

	scanner := bufio.NewScanner(pm.proc.Stdout())
	// Increase buffer for large messages (16MB).
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		select {
		case <-pm.ctx.Done():
			return
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msg, err := ParseSDKMessage(line)
		if err != nil {
			// Malformed line — skip silently. In debug mode the stderr
			// callback can be used for diagnostics.
			continue
		}

		select {
		case pm.messages <- msg:
		case <-pm.ctx.Done():
			return
		}
	}
}
