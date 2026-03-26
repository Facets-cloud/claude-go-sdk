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

	// Bidirectional protocol fields
	prompt          interface{} // string or <-chan SDKUserMessage
	isSingleTurn    bool
	hookCallbacks   map[string]HookCallback // callback_id -> handler
	nextHookID      int
	cancelControllers sync.Map // request_id -> context.CancelFunc
}

// NewQuery creates and starts a new query to Claude Code.
func NewQuery(params QueryParams) *Query {
	q := &Query{
		messages:      make(chan SDKMessage, 64),
		correlation:   NewCorrelationEngine(),
		opts:          params.Options,
		initCh:        make(chan struct{}),
		done:          make(chan struct{}),
		prompt:        params.Prompt,
		hookCallbacks: make(map[string]HookCallback),
	}

	// Single-turn if prompt is a string (not a channel)
	if _, ok := params.Prompt.(string); ok {
		q.isSingleTurn = true
	}

	go q.run()
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

func (q *Query) run() {
	defer close(q.messages)

	cliPath, err := CLIPath(q.cliPath())
	if err != nil {
		q.err = err
		return
	}

	args := buildProcessArgs(q.opts)

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

	// --- Bidirectional protocol: send user message via stdin ---
	if promptStr, ok := q.prompt.(string); ok && promptStr != "" {
		q.writeUserMessage(promptStr)
	}

	// Send initialize control request
	q.sendInitialize()

	// If prompt is a channel, start streaming input
	if ch, ok := q.prompt.(<-chan SDKUserMessage); ok {
		go q.streamInputFromChannel(ch)
	}

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

		// Check if this is a control request (permission, hook callback, mcp)
		if q.tryHandleControlRequest(line) {
			continue
		}

		// Check if this is a control cancel request
		if q.tryHandleControlCancelRequest(line) {
			continue
		}

		// Skip keep_alive, streamlined_text, streamlined_tool_use_summary
		if q.isSkippableMessageType(line) {
			continue
		}

		// Parse as SDK message
		msg, err := ParseSDKMessage(line)
		if err != nil {
			continue // skip unparseable lines
		}

		// On result message + single turn: close stdin to signal we're done
		if msg.MessageType() == "result" && q.isSingleTurn {
			q.process.Stdin().Close()
		}

		select {
		case q.messages <- msg:
		case <-q.done:
			return
		}
	}

	// Ensure initCh is closed even if no init message arrived
	q.initOnce.Do(func() { close(q.initCh) })

	// Wait for process exit
	if q.process != nil {
		_ = q.process.Wait()
	}
}

// writeUserMessage sends the initial user message to stdin as JSON.
func (q *Query) writeUserMessage(text string) {
	content := []map[string]string{{"type": "text", "text": text}}
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return
	}
	msg := SDKUserMessage{
		Type:            "user",
		SessionID:       "",
		Message:         contentJSON,
		ParentToolUseID: nil,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_, _ = q.process.Stdin().Write(append(data, '\n'))
}

// streamInputFromChannel reads user messages from a channel and writes them to stdin.
func (q *Query) streamInputFromChannel(ch <-chan SDKUserMessage) {
	for msg := range ch {
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		_, _ = q.process.Stdin().Write(append(data, '\n'))
	}
}

// isSkippableMessageType checks if a JSON line is a type we should skip silently.
func (q *Query) isSkippableMessageType(line []byte) bool {
	var env struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &env); err != nil {
		return false
	}
	switch env.Type {
	case "keep_alive", "streamlined_text", "streamlined_tool_use_summary":
		return true
	}
	return false
}

// tryHandleControlCancelRequest handles control_cancel_request messages.
func (q *Query) tryHandleControlCancelRequest(line []byte) bool {
	var env struct {
		Type      string `json:"type"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(line, &env); err != nil {
		return false
	}
	if env.Type != "control_cancel_request" {
		return false
	}
	if cancel, ok := q.cancelControllers.LoadAndDelete(env.RequestID); ok {
		cancel.(context.CancelFunc)()
	}
	return true
}

func (q *Query) sendInitialize() {
	initReq := SDKControlInitializeRequest{
		Subtype: "initialize",
	}

	// Build hooks with callback IDs
	if q.opts != nil && q.opts.Hooks != nil {
		hooksWire := make(map[string][]SDKHookCallbackMatcherWire)
		for event, matchers := range q.opts.Hooks {
			var wireMatchers []SDKHookCallbackMatcherWire
			for _, matcher := range matchers {
				var callbackIDs []string
				for _, hook := range matcher.Hooks {
					callbackID := fmt.Sprintf("hook_%d", q.nextHookID)
					q.nextHookID++
					q.hookCallbacks[callbackID] = hook
					callbackIDs = append(callbackIDs, callbackID)
				}
				wireMatchers = append(wireMatchers, SDKHookCallbackMatcherWire{
					Matcher:         matcher.Matcher,
					HookCallbackIDs: callbackIDs,
					Timeout:         matcher.Timeout,
				})
			}
			hooksWire[string(event)] = wireMatchers
		}
		if len(hooksWire) > 0 {
			initReq.Hooks = hooksWire
		}
	}

	// Build sdkMcpServers list (names of type:"sdk" MCP servers)
	if q.opts != nil && len(q.opts.McpServers) > 0 {
		var sdkNames []string
		for name, cfg := range q.opts.McpServers {
			if m, ok := cfg.(map[string]interface{}); ok {
				if t, _ := m["type"].(string); t == "sdk" {
					sdkNames = append(sdkNames, name)
				}
			}
		}
		if len(sdkNames) > 0 {
			initReq.SdkMcpServers = sdkNames
		}
	}

	// JSON schema for structured output
	if q.opts != nil && q.opts.OutputFormat != nil {
		initReq.JSONSchema = q.opts.OutputFormat.Schema
	}

	// System prompt
	if q.opts != nil && q.opts.SystemPrompt != nil {
		switch sp := q.opts.SystemPrompt.(type) {
		case string:
			initReq.SystemPrompt = &sp
		case SystemPromptPreset:
			initReq.AppendSystemPrompt = sp.Append
		}
	}

	// Agents
	if q.opts != nil && len(q.opts.Agents) > 0 {
		agentsMap := make(map[string]interface{}, len(q.opts.Agents))
		for k, v := range q.opts.Agents {
			agentsMap[k] = v
		}
		initReq.Agents = agentsMap
	}

	// Prompt suggestions
	if q.opts != nil && q.opts.PromptSuggestions != nil {
		initReq.PromptSuggestions = q.opts.PromptSuggestions
	}

	// Agent progress summaries
	if q.opts != nil && q.opts.AgentProgressSummaries != nil {
		initReq.AgentProgressSummaries = q.opts.AgentProgressSummaries
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

		// Parse the inner response from the success envelope
		var successEnv struct {
			Subtype  string          `json:"subtype"`
			Response json.RawMessage `json:"response"`
		}
		if err := json.Unmarshal(resp.Response, &successEnv); err == nil && successEnv.Subtype == "success" {
			var initResp SDKControlInitializeResponse
			if err := json.Unmarshal(successEnv.Response, &initResp); err == nil {
				q.initResp = &initResp
			}
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
	ctx, cancel := context.WithCancel(context.Background())
	if q.opts != nil && q.opts.AbortContext != nil {
		ctx, cancel = context.WithCancel(q.opts.AbortContext)
	}
	q.cancelControllers.Store(requestID, cancel)
	defer func() {
		q.cancelControllers.Delete(requestID)
		cancel()
	}()

	subtype, err := ParseControlRequestSubtype(request)
	if err != nil {
		q.sendControlErrorResponse(requestID, "failed to parse control request subtype")
		return
	}

	switch subtype {
	case "can_use_tool":
		q.handlePermissionRequest(ctx, requestID, request)
	case "hook_callback":
		q.handleHookCallback(ctx, requestID, request)
	case "elicitation":
		q.handleElicitation(ctx, requestID, request)
	case "mcp_message":
		q.handleMcpMessage(ctx, requestID, request)
	default:
		q.sendControlErrorResponse(requestID, "unsupported control request subtype: "+subtype)
	}
}

func (q *Query) handlePermissionRequest(ctx context.Context, requestID string, request json.RawMessage) {
	if q.opts == nil || q.opts.CanUseTool == nil {
		q.sendControlErrorResponse(requestID, "canUseTool callback is not provided.")
		return
	}

	var req SDKControlPermissionRequest
	if err := json.Unmarshal(request, &req); err != nil {
		q.sendControlErrorResponse(requestID, "failed to parse permission request: "+err.Error())
		return
	}

	result, err := q.opts.CanUseTool(ctx, req.ToolName, req.Input, CanUseToolOptions{
		ToolUseID: req.ToolUseID,
	})
	if err != nil {
		q.sendControlErrorResponse(requestID, err.Error())
		return
	}

	// Include toolUseID in response (matches TS SDK)
	type permResult struct {
		ToolUseID string `json:"toolUseID"`
	}
	// Merge result with toolUseID
	resultJSON, _ := json.Marshal(result)
	var resultMap map[string]interface{}
	json.Unmarshal(resultJSON, &resultMap)
	if resultMap == nil {
		resultMap = make(map[string]interface{})
	}
	resultMap["toolUseID"] = req.ToolUseID
	q.sendControlResponse(requestID, resultMap)
}

func (q *Query) handleHookCallback(ctx context.Context, requestID string, request json.RawMessage) {
	var req SDKHookCallbackRequest
	if err := json.Unmarshal(request, &req); err != nil {
		q.sendControlErrorResponse(requestID, "failed to parse hook callback request: "+err.Error())
		return
	}

	cb, ok := q.hookCallbacks[req.CallbackID]
	if !ok {
		q.sendControlErrorResponse(requestID, "unknown hook callback ID: "+req.CallbackID)
		return
	}

	// Parse the input as a HookInput (generic interface{})
	var hookInput HookInput
	if req.Input != nil {
		var parsed map[string]interface{}
		if err := json.Unmarshal(req.Input, &parsed); err == nil {
			hookInput = parsed
		}
	}

	result, err := cb(ctx, hookInput, req.ToolUseID)
	if err != nil {
		q.sendControlErrorResponse(requestID, err.Error())
		return
	}

	q.sendControlResponse(requestID, result)
}

func (q *Query) handleElicitation(ctx context.Context, requestID string, request json.RawMessage) {
	if q.opts == nil || q.opts.OnElicitation == nil {
		q.sendControlResponse(requestID, map[string]string{"action": "decline"})
		return
	}

	var req SDKControlElicitationRequest
	if err := json.Unmarshal(request, &req); err != nil {
		q.sendControlResponse(requestID, map[string]string{"action": "decline"})
		return
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

// handleMcpMessage handles mcp_message control requests (placeholder for Step 5).
func (q *Query) handleMcpMessage(ctx context.Context, requestID string, request json.RawMessage) {
	var req SDKControlMcpMessageRequest
	if err := json.Unmarshal(request, &req); err != nil {
		q.sendControlErrorResponse(requestID, "failed to parse mcp message request: "+err.Error())
		return
	}
	// MCP message handling will be implemented in Step 5 (in-process MCP server bridge)
	q.sendControlErrorResponse(requestID, "SDK MCP server not found: "+req.ServerName)
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
	resp, err := BuildControlSuccessResponse(requestID, response)
	if err != nil {
		return
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}
	_, _ = q.process.Stdin().Write(append(data, '\n'))
}

func (q *Query) sendControlErrorResponse(requestID string, errMsg string) {
	resp, err := BuildControlErrorResponse(requestID, errMsg)
	if err != nil {
		return
	}
	data, err := json.Marshal(resp)
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
