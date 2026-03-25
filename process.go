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
func (p *defaultSpawnedProcess) Kill() error            { return p.cmd.Process.Kill() }

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
func buildProcessArgs(opts *Options, prompt string) []string {
	args := []string{
		"--print",
		"--output-format", "stream-json",
		"--verbose",
	}

	if opts == nil {
		if prompt != "" {
			args = append(args, prompt)
		}
		return args
	}

	if opts.Model != nil {
		args = append(args, "--model", *opts.Model)
	}

	if opts.PermissionMode != nil {
		args = append(args, "--permission-mode", string(*opts.PermissionMode))
	}

	if opts.AllowDangerouslySkipPermissions != nil && *opts.AllowDangerouslySkipPermissions {
		args = append(args, "--dangerously-skip-permissions")
	}

	if opts.MaxTurns != nil {
		args = append(args, "--max-turns", fmt.Sprintf("%d", *opts.MaxTurns))
	}

	if opts.MaxBudgetUsd != nil {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", *opts.MaxBudgetUsd))
	}

	if opts.Continue != nil && *opts.Continue {
		args = append(args, "--continue")
	}

	if opts.Resume != nil {
		args = append(args, "--resume", *opts.Resume)
	}

	if opts.SessionID != nil {
		args = append(args, "--session-id", *opts.SessionID)
	}

	// Cwd is handled by setting cmd.Dir on the subprocess, not via CLI flag.

	if opts.Agent != nil {
		args = append(args, "--agent", *opts.Agent)
	}

	if opts.Debug != nil && *opts.Debug {
		args = append(args, "--debug")
	}

	if opts.DebugFile != nil {
		args = append(args, "--debug-file", *opts.DebugFile)
	}

	if opts.ForkSession != nil && *opts.ForkSession {
		args = append(args, "--fork-session")
	}

	if opts.PersistSession != nil && !*opts.PersistSession {
		args = append(args, "--no-persist-session")
	}

	if opts.IncludePartialMessages != nil && *opts.IncludePartialMessages {
		args = append(args, "--include-partial-messages")
	}

	if opts.PromptSuggestions != nil && *opts.PromptSuggestions {
		args = append(args, "--prompt-suggestions")
	}

	if opts.AgentProgressSummaries != nil && *opts.AgentProgressSummaries {
		args = append(args, "--agent-progress-summaries")
	}

	if opts.StrictMcpConfig != nil && *opts.StrictMcpConfig {
		args = append(args, "--strict-mcp-config")
	}

	if opts.FallbackModel != nil {
		args = append(args, "--fallback-model", *opts.FallbackModel)
	}

	if opts.EnableFileCheckpointing != nil && *opts.EnableFileCheckpointing {
		args = append(args, "--enable-file-checkpointing")
	}

	if opts.MaxThinkingTokens != nil {
		args = append(args, "--max-thinking-tokens", fmt.Sprintf("%d", *opts.MaxThinkingTokens))
	}

	if opts.Effort != nil {
		args = append(args, "--effort", *opts.Effort)
	}

	if opts.ResumeSessionAt != nil {
		args = append(args, "--resume-session-at", *opts.ResumeSessionAt)
	}

	if opts.PermissionPromptToolName != nil {
		args = append(args, "--permission-prompt-tool-name", *opts.PermissionPromptToolName)
	}

	for _, dir := range opts.AdditionalDirectories {
		args = append(args, "--additional-directory", dir)
	}

	if len(opts.AllowedTools) > 0 {
		args = append(args, "--allowed-tools")
		args = append(args, opts.AllowedTools...)
	}

	if len(opts.DisallowedTools) > 0 {
		args = append(args, "--disallowed-tools")
		args = append(args, opts.DisallowedTools...)
	}

	for _, beta := range opts.Betas {
		args = append(args, "--beta", string(beta))
	}

	for _, source := range opts.SettingSources {
		args = append(args, "--setting-source", string(source))
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

	// Tools
	if opts.Tools != nil {
		switch t := opts.Tools.(type) {
		case []string:
			args = append(args, "--tools", strings.Join(t, ","))
		case ToolPreset:
			// preset is the default, no flag needed
		}
	}

	// Thinking config — no dedicated CLI flag exists.
	// Thinking is controlled via settings (alwaysThinkingEnabled) or the API.
	// For --print mode, we skip it — the default behavior is adaptive thinking
	// for supported models.

	// Output format
	if opts.OutputFormat != nil {
		ofJSON, err := json.Marshal(opts.OutputFormat)
		if err == nil {
			args = append(args, "--output-format-json", string(ofJSON))
		}
	}

	// Settings
	if opts.Settings != nil {
		switch s := opts.Settings.(type) {
		case string:
			args = append(args, "--settings", s)
		default:
			settingsJSON, err := json.Marshal(s)
			if err == nil {
				args = append(args, "--settings", string(settingsJSON))
			}
		}
	}

	// MCP servers
	if len(opts.McpServers) > 0 {
		// --mcp-config accepts JSON files or strings.
		// Write to a temp file since the CLI may parse it as a file path.
		mcpConfig := map[string]interface{}{
			"mcpServers": opts.McpServers,
		}
		mcpJSON, err := json.Marshal(mcpConfig)
		if err == nil {
			tmpFile, err := os.CreateTemp("", "claude-mcp-*.json")
			if err == nil {
				tmpFile.Write(mcpJSON)
				tmpFile.Close()
				args = append(args, "--mcp-config", tmpFile.Name())
				// Note: temp file is cleaned up by OS on process exit
			}
		}
	}

	// Sandbox
	if opts.Sandbox != nil {
		sbJSON, err := json.Marshal(opts.Sandbox)
		if err == nil {
			args = append(args, "--sandbox", string(sbJSON))
		}
	}

	// Plugins
	for _, plugin := range opts.Plugins {
		args = append(args, "--plugin", plugin.Path)
	}

	// Agents
	if len(opts.Agents) > 0 {
		agentsJSON, err := json.Marshal(opts.Agents)
		if err == nil {
			args = append(args, "--agents", string(agentsJSON))
		}
	}

	// Tool config
	if opts.ToolConfig != nil {
		tcJSON, err := json.Marshal(opts.ToolConfig)
		if err == nil {
			args = append(args, "--tool-config", string(tcJSON))
		}
	}

	// Extra args
	for k, v := range opts.ExtraArgs {
		if v == nil {
			args = append(args, "--"+k)
		} else {
			args = append(args, "--"+k, *v)
		}
	}

	// Prompt as positional argument (must be last).
	// Use "--" to separate options from the positional prompt,
	// preventing variadic options like --mcp-config from consuming it.
	if prompt != "" {
		args = append(args, "--", prompt)
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
