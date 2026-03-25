package claudeagent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// QueryParams configures a query to Claude Code.
type QueryParams struct {
	// Prompt is either a string or a <-chan SDKUserMessage for streaming input.
	Prompt interface{}

	// Options configures the query behavior.
	Options *Options
}

// RewindFilesOptions configures a rewind operation.
type RewindFilesOptions struct {
	DryRun bool `json:"dryRun,omitempty"`
}

// RewindFilesResult contains the result of a rewind operation.
type RewindFilesResult struct {
	CanRewind    bool     `json:"canRewind"`
	Error        *string  `json:"error,omitempty"`
	FilesChanged []string `json:"filesChanged,omitempty"`
	Insertions   *int     `json:"insertions,omitempty"`
	Deletions    *int     `json:"deletions,omitempty"`
}

// Query manages a conversation with the Claude Code subprocess.
type Query struct {
	messages    chan SDKMessage
	process     SpawnedProcess
	correlation *CorrelationEngine
	opts        *Options
	initResp    *SDKControlInitializeResponse
	initOnce    sync.Once
	initCh      chan struct{}
	done        chan struct{}
	closeOnce   sync.Once
	err         error
}

// NewQuery creates and starts a new query to Claude Code.
func NewQuery(params QueryParams) *Query {
	q := &Query{
		messages:    make(chan SDKMessage, 64),
		correlation: NewCorrelationEngine(),
		opts:        params.Options,
		initCh:      make(chan struct{}),
		done:        make(chan struct{}),
	}

	prompt := ""
	if s, ok := params.Prompt.(string); ok {
		prompt = s
	}

	go q.run(prompt)
	return q
}

// Messages returns the channel of messages from the agent.
func (q *Query) Messages() <-chan SDKMessage {
	return q.messages
}

// Interrupt stops the current query execution.
func (q *Query) Interrupt(ctx context.Context) error {
	return q.sendControlRequest(ctx, SDKControlInterruptRequest{
		Subtype: "interrupt",
	})
}

// SetPermissionMode changes the permission mode mid-session.
func (q *Query) SetPermissionMode(ctx context.Context, mode PermissionMode) error {
	return q.sendControlRequest(ctx, SDKControlSetPermissionModeRequest{
		Subtype: "set_permission_mode",
		Mode:    mode,
	})
}

// SetModel changes the model mid-session. Pass nil to revert to default.
func (q *Query) SetModel(ctx context.Context, model *string) error {
	return q.sendControlRequest(ctx, SDKControlSetModelRequest{
		Subtype: "set_model",
		Model:   model,
	})
}

// SetMaxThinkingTokens sets the max thinking token budget.
// Pass nil to clear the limit and revert to the model default.
func (q *Query) SetMaxThinkingTokens(ctx context.Context, tokens *int) error {
	return q.sendControlRequest(ctx, SDKControlSetMaxThinkingTokensRequest{
		Subtype:           "set_max_thinking_tokens",
		MaxThinkingTokens: tokens,
	})
}

// InitializationResult returns the full init response after the handshake completes.
func (q *Query) InitializationResult(ctx context.Context) (*SDKControlInitializeResponse, error) {
	select {
	case <-q.initCh:
		return q.initResp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// McpServerStatus returns MCP server connection statuses.
func (q *Query) McpServerStatus(ctx context.Context) ([]McpServerStatus, error) {
	resp, err := q.sendControlRequestWithResponse(ctx, SDKControlMcpStatusRequest{
		Subtype: "mcp_status",
	})
	if err != nil {
		return nil, err
	}
	var result struct {
		Servers []McpServerStatus `json:"servers"`
	}
	if err := json.Unmarshal(resp.Response, &result); err != nil {
		return nil, fmt.Errorf("parse mcp status response: %w", err)
	}
	return result.Servers, nil
}

// ReconnectMcpServer reconnects a disconnected MCP server.
func (q *Query) ReconnectMcpServer(ctx context.Context, serverName string) error {
	return q.sendControlRequest(ctx, SDKControlMcpReconnectRequest{
		Subtype:    "mcp_reconnect",
		ServerName: serverName,
	})
}

// ToggleMcpServer enables or disables an MCP server.
func (q *Query) ToggleMcpServer(ctx context.Context, serverName string, enabled bool) error {
	return q.sendControlRequest(ctx, SDKControlMcpToggleRequest{
		Subtype:    "mcp_toggle",
		ServerName: serverName,
		Enabled:    enabled,
	})
}

// RewindFiles rewinds file changes to a specific message.
func (q *Query) RewindFiles(ctx context.Context, userMessageID string, opts *RewindFilesOptions) (*RewindFilesResult, error) {
	req := SDKControlRewindFilesRequest{
		Subtype:       "rewind_files",
		UserMessageID: userMessageID,
	}
	if opts != nil && opts.DryRun {
		req.DryRun = Bool(true)
	}
	resp, err := q.sendControlRequestWithResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	var result RewindFilesResult
	if err := json.Unmarshal(resp.Response, &result); err != nil {
		return nil, fmt.Errorf("parse rewind files response: %w", err)
	}
	return &result, nil
}

// ApplyFlagSettings merges settings into the flag settings layer mid-session.
func (q *Query) ApplyFlagSettings(ctx context.Context, settings Settings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	var settingsMap map[string]interface{}
	if err := json.Unmarshal(settingsJSON, &settingsMap); err != nil {
		return fmt.Errorf("convert settings to map: %w", err)
	}
	return q.sendControlRequest(ctx, SDKControlApplyFlagSettingsRequest{
		Subtype:  "apply_flag_settings",
		Settings: settingsMap,
	})
}

// SupportedCommands returns available slash commands from the init response.
func (q *Query) SupportedCommands(ctx context.Context) ([]SlashCommand, error) {
	initResp, err := q.InitializationResult(ctx)
	if err != nil {
		return nil, err
	}
	return initResp.Commands, nil
}

// SupportedModels returns available models from the init response.
func (q *Query) SupportedModels(ctx context.Context) ([]ModelInfo, error) {
	initResp, err := q.InitializationResult(ctx)
	if err != nil {
		return nil, err
	}
	return initResp.Models, nil
}

// SupportedAgents returns available subagents from the init response.
func (q *Query) SupportedAgents(ctx context.Context) ([]AgentInfo, error) {
	initResp, err := q.InitializationResult(ctx)
	if err != nil {
		return nil, err
	}
	return initResp.Agents, nil
}

// AccountInfo returns authenticated account info from the init response.
func (q *Query) AccountInfo(ctx context.Context) (*AccountInfo, error) {
	initResp, err := q.InitializationResult(ctx)
	if err != nil {
		return nil, err
	}
	return initResp.Account, nil
}

// SetMcpServers replaces the set of dynamically-managed MCP servers.
func (q *Query) SetMcpServers(ctx context.Context, servers map[string]interface{}) (*McpSetServersResult, error) {
	resp, err := q.sendControlRequestWithResponse(ctx, SDKControlMcpSetServersRequest{
		Subtype: "mcp_set_servers",
		Servers: servers,
	})
	if err != nil {
		return nil, err
	}
	var result McpSetServersResult
	if err := json.Unmarshal(resp.Response, &result); err != nil {
		return nil, fmt.Errorf("parse mcp set servers response: %w", err)
	}
	return &result, nil
}

// StreamInput sends user messages to the query post-construction.
func (q *Query) StreamInput(messages <-chan SDKUserMessage) error {
	go func() {
		for msg := range messages {
			data, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			data = append(data, '\n')
			_, _ = q.process.Stdin().Write(data)
		}
	}()
	return nil
}

// StopTask stops a running background task.
func (q *Query) StopTask(ctx context.Context, taskID string) error {
	return q.sendControlRequest(ctx, SDKControlStopTaskRequest{
		Subtype: "stop_task",
		TaskID:  taskID,
	})
}

// Close terminates the query and cleans up resources.
func (q *Query) Close() {
	q.closeOnce.Do(func() {
		close(q.done)
		q.correlation.Close()
		if q.process != nil {
			_ = q.process.Kill()
		}
	})
}

// --- Internal methods ---

func (q *Query) run(prompt string) {
	defer close(q.messages)

	cliPath, err := CLIPath(q.cliPath())
	if err != nil {
		q.err = err
		return
	}

	args := buildProcessArgs(q.opts, prompt)

	spawnOpts := SpawnOptions{
		Command: cliPath,
		Args:    args,
		Cwd:     q.cwd(),
		Env:     q.env(),
	}

	spawnFn := defaultSpawn
	if q.opts != nil && q.opts.SpawnClaudeCodeProcess != nil {
		spawnFn = q.opts.SpawnClaudeCodeProcess
	}

	q.process = spawnFn(spawnOpts)

	// Start the process
	if dp, ok := q.process.(*defaultSpawnedProcess); ok {
		if err := dp.cmd.Start(); err != nil {
			q.err = fmt.Errorf("start claude process: %w", err)
			return
		}
	}

	// Send initialization control request
	q.sendInitialize()

	// Read stdout JSON lines
	scanner := bufio.NewScanner(q.process.Stdout())
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line

	for scanner.Scan() {
		select {
		case <-q.done:
			return
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Check if this is a control response
		if q.tryHandleControlResponse(line) {
			continue
		}

		// Check if this is a control request (permission, hook callback)
		if q.tryHandleControlRequest(line) {
			continue
		}

		// Parse as SDK message
		msg, err := ParseSDKMessage(line)
		if err != nil {
			continue // skip unparseable lines
		}

		select {
		case q.messages <- msg:
		case <-q.done:
			return
		}
	}

	// Wait for process exit
	if q.process != nil {
		_ = q.process.Wait()
	}
}

func (q *Query) sendInitialize() {
	initReq := SDKControlInitializeRequest{
		Subtype:        "initialize",
		ProtocolVersion: 1,
	}

	// Add capabilities
	if q.opts != nil {
		if q.opts.CanUseTool != nil {
			initReq.CanUseTool = true
		}
		if q.opts.Hooks != nil {
			initReq.HasHooks = true
		}
		if q.opts.OnElicitation != nil {
			initReq.HasElicitation = true
		}
	}

	reqID := uuid.New().String()
	ch := q.correlation.Register(reqID)

	data, err := q.marshalControlRequest(reqID, initReq)
	if err != nil {
		return
	}
	_, _ = q.process.Stdin().Write(append(data, '\n'))

	// Wait for response in background
	go func() {
		resp, err := WaitForResponse(context.Background(), ch)
		if err != nil || resp == nil {
			q.initOnce.Do(func() { close(q.initCh) })
			return
		}

		var initResp SDKControlInitializeResponse
		if err := json.Unmarshal(resp.Response, &initResp); err == nil {
			q.initResp = &initResp
		}
		q.initOnce.Do(func() { close(q.initCh) })
	}()
}

func (q *Query) tryHandleControlResponse(line []byte) bool {
	var envelope struct {
		Type     string          `json:"type"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		return false
	}
	if envelope.Type != "control_response" {
		return false
	}

	var respEnv struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(envelope.Response, &respEnv); err != nil {
		return false
	}

	resp := &SDKControlResponse{
		Type:     envelope.Type,
		Response: envelope.Response,
	}
	q.correlation.Deliver(respEnv.RequestID, resp)
	return true
}

func (q *Query) tryHandleControlRequest(line []byte) bool {
	var envelope struct {
		Type      string          `json:"type"`
		RequestID string          `json:"request_id"`
		Request   json.RawMessage `json:"request"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		return false
	}

	if envelope.Type == "control_request" {
		go q.handleControlRequest(envelope.RequestID, envelope.Request)
		return true
	}

	return false
}

func (q *Query) handleControlRequest(requestID string, request json.RawMessage) {
	subtype, err := ParseControlRequestSubtype(request)
	if err != nil {
		return
	}

	switch subtype {
	case "permission":
		q.handlePermissionRequest(requestID, request)
	case "hook_callback":
		q.handleHookCallback(requestID, request)
	case "elicitation":
		q.handleElicitation(requestID, request)
	}
}

func (q *Query) handlePermissionRequest(requestID string, request json.RawMessage) {
	if q.opts == nil || q.opts.CanUseTool == nil {
		// No handler — deny by default
		q.sendControlResponse(requestID, PermissionResultDeny{
			Behavior: PermissionBehaviorDeny,
			Message:  "no permission handler configured",
		})
		return
	}

	var req SDKControlPermissionRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return
	}

	ctx := context.Background()
	if q.opts.AbortContext != nil {
		ctx = q.opts.AbortContext
	}

	result, err := q.opts.CanUseTool(ctx, req.ToolName, req.Input, CanUseToolOptions{
		ToolUseID: req.ToolUseID,
	})
	if err != nil {
		q.sendControlResponse(requestID, PermissionResultDeny{
			Behavior: PermissionBehaviorDeny,
			Message:  err.Error(),
		})
		return
	}

	q.sendControlResponse(requestID, result)
}

func (q *Query) handleHookCallback(requestID string, request json.RawMessage) {
	// Hook callback handling — send empty response for now
	q.sendControlResponse(requestID, SyncHookJSONOutput{})
}

func (q *Query) handleElicitation(requestID string, request json.RawMessage) {
	if q.opts == nil || q.opts.OnElicitation == nil {
		q.sendControlResponse(requestID, map[string]string{"action": "decline"})
		return
	}

	var req SDKControlElicitationRequest
	if err := json.Unmarshal(request, &req); err != nil {
		q.sendControlResponse(requestID, map[string]string{"action": "decline"})
		return
	}

	ctx := context.Background()
	if q.opts.AbortContext != nil {
		ctx = q.opts.AbortContext
	}

	result, err := q.opts.OnElicitation(ctx, ElicitationRequest{
		ServerName: req.McpServerName,
		Message:    req.Message,
	})
	if err != nil || result == nil {
		q.sendControlResponse(requestID, map[string]string{"action": "decline"})
		return
	}

	q.sendControlResponse(requestID, result)
}

func (q *Query) sendControlRequest(ctx context.Context, request interface{}) error {
	_, err := q.sendControlRequestWithResponse(ctx, request)
	return err
}

func (q *Query) sendControlRequestWithResponse(ctx context.Context, request interface{}) (*SDKControlResponse, error) {
	reqID := uuid.New().String()
	ch := q.correlation.Register(reqID)

	data, err := q.marshalControlRequest(reqID, request)
	if err != nil {
		return nil, err
	}

	if _, err := q.process.Stdin().Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("write control request: %w", err)
	}

	return WaitForResponse(ctx, ch)
}

func (q *Query) marshalControlRequest(requestID string, request interface{}) ([]byte, error) {
	reqData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal control request: %w", err)
	}
	envelope := SDKControlRequest{
		Type:      "control_request",
		RequestID: requestID,
		Request:   reqData,
	}
	return json.Marshal(envelope)
}

func (q *Query) sendControlResponse(requestID string, response interface{}) {
	respData, err := json.Marshal(response)
	if err != nil {
		return
	}
	envelope := map[string]interface{}{
		"type":       "control_response",
		"request_id": requestID,
		"response":   json.RawMessage(respData),
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return
	}
	_, _ = q.process.Stdin().Write(append(data, '\n'))
}

func (q *Query) cliPath() *string {
	if q.opts != nil {
		return q.opts.PathToClaudeCodeExecutable
	}
	return nil
}

func (q *Query) cwd() string {
	if q.opts != nil && q.opts.Cwd != nil {
		return *q.opts.Cwd
	}
	return ""
}

func (q *Query) env() map[string]string {
	if q.opts != nil {
		return q.opts.Env
	}
	return nil
}
