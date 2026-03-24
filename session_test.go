package claudeagent

import (
	"encoding/json"
	"testing"
)

// --- SDKSessionInfo tests ---

func TestSDKSessionInfo_JSON(t *testing.T) {
	raw := `{
		"sessionId": "abc-123",
		"summary": "My session",
		"lastModified": 1711276800000,
		"fileSize": 4096,
		"customTitle": "Custom Title",
		"firstPrompt": "Hello",
		"gitBranch": "main",
		"cwd": "/home/user/project",
		"tag": "important",
		"createdAt": 1711276700000
	}`
	var info SDKSessionInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}
	if info.SessionID != "abc-123" {
		t.Errorf("got sessionId %q", info.SessionID)
	}
	if info.Summary != "My session" {
		t.Errorf("got summary %q", info.Summary)
	}
	if info.LastModified != 1711276800000 {
		t.Errorf("got lastModified %d", info.LastModified)
	}
	if info.FileSize == nil || *info.FileSize != 4096 {
		t.Error("expected fileSize 4096")
	}
	if info.CustomTitle == nil || *info.CustomTitle != "Custom Title" {
		t.Error("expected customTitle")
	}
	if info.FirstPrompt == nil || *info.FirstPrompt != "Hello" {
		t.Error("expected firstPrompt")
	}
	if info.GitBranch == nil || *info.GitBranch != "main" {
		t.Error("expected gitBranch")
	}
	if info.Cwd == nil || *info.Cwd != "/home/user/project" {
		t.Error("expected cwd")
	}
	if info.Tag == nil || *info.Tag != "important" {
		t.Error("expected tag")
	}
	if info.CreatedAt == nil || *info.CreatedAt != 1711276700000 {
		t.Error("expected createdAt")
	}
}

func TestSDKSessionInfo_MinimalFields(t *testing.T) {
	raw := `{"sessionId": "s-1", "summary": "test", "lastModified": 100}`
	var info SDKSessionInfo
	if err := json.Unmarshal([]byte(raw), &info); err != nil {
		t.Fatal(err)
	}
	if info.SessionID != "s-1" {
		t.Errorf("got sessionId %q", info.SessionID)
	}
	if info.FileSize != nil {
		t.Error("expected nil fileSize")
	}
	if info.CustomTitle != nil {
		t.Error("expected nil customTitle")
	}
}

func TestSDKSessionInfo_RoundTrip(t *testing.T) {
	fs := int64(2048)
	ct := "My Title"
	fp := "hello world"
	gb := "feature-branch"
	cwd := "/tmp"
	tag := "v1"
	ca := int64(1711276700000)
	info := SDKSessionInfo{
		SessionID:    "s-rt",
		Summary:      "round trip",
		LastModified: 1711276800000,
		FileSize:     &fs,
		CustomTitle:  &ct,
		FirstPrompt:  &fp,
		GitBranch:    &gb,
		Cwd:          &cwd,
		Tag:          &tag,
		CreatedAt:    &ca,
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKSessionInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SessionID != info.SessionID || got.Summary != info.Summary {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if *got.FileSize != *info.FileSize {
		t.Error("fileSize mismatch")
	}
}

// --- SessionMessage tests ---

func TestSessionMessage_JSON(t *testing.T) {
	raw := `{
		"type": "user",
		"uuid": "msg-1",
		"session_id": "s-1",
		"message": {"role": "user", "content": "hello"},
		"parent_tool_use_id": null
	}`
	var msg SessionMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != "user" {
		t.Errorf("got type %q", msg.Type)
	}
	if msg.UUID != "msg-1" {
		t.Errorf("got uuid %q", msg.UUID)
	}
	if msg.SessionID != "s-1" {
		t.Errorf("got session_id %q", msg.SessionID)
	}
	if msg.Message == nil {
		t.Error("expected non-nil message")
	}
}

func TestSessionMessage_AssistantType(t *testing.T) {
	raw := `{
		"type": "assistant",
		"uuid": "msg-2",
		"session_id": "s-1",
		"message": {"role": "assistant", "content": "hi"},
		"parent_tool_use_id": null
	}`
	var msg SessionMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != "assistant" {
		t.Errorf("got type %q", msg.Type)
	}
}

func TestSessionMessage_RoundTrip(t *testing.T) {
	msg := SessionMessage{
		Type:              "user",
		UUID:              "msg-rt",
		SessionID:         "s-rt",
		Message:           json.RawMessage(`{"content":"test"}`),
		ParentToolUseID:   nil,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var got SessionMessage
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != msg.Type || got.UUID != msg.UUID || got.SessionID != msg.SessionID {
		t.Errorf("round-trip mismatch")
	}
}

// --- Options types tests ---

func TestListSessionsOptions_JSON(t *testing.T) {
	opts := ListSessionsOptions{
		Dir:              strPtr("/project"),
		Limit:            intPtr(50),
		Offset:           intPtr(10),
		IncludeWorktrees: boolPtr(true),
	}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got ListSessionsOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Dir != "/project" {
		t.Errorf("got dir %q", *got.Dir)
	}
	if *got.Limit != 50 {
		t.Errorf("got limit %d", *got.Limit)
	}
	if *got.Offset != 10 {
		t.Errorf("got offset %d", *got.Offset)
	}
	if *got.IncludeWorktrees != true {
		t.Error("expected includeWorktrees true")
	}
}

func TestListSessionsOptions_Empty(t *testing.T) {
	opts := ListSessionsOptions{}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("expected empty JSON object, got %s", data)
	}
}

func TestGetSessionInfoOptions_JSON(t *testing.T) {
	opts := GetSessionInfoOptions{Dir: strPtr("/project")}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got GetSessionInfoOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Dir != "/project" {
		t.Errorf("got dir %q", *got.Dir)
	}
}

func TestGetSessionMessagesOptions_JSON(t *testing.T) {
	opts := GetSessionMessagesOptions{
		Dir:    strPtr("/project"),
		Limit:  intPtr(100),
		Offset: intPtr(5),
	}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got GetSessionMessagesOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Dir != "/project" {
		t.Errorf("got dir %q", *got.Dir)
	}
	if *got.Limit != 100 {
		t.Errorf("got limit %d", *got.Limit)
	}
	if *got.Offset != 5 {
		t.Errorf("got offset %d", *got.Offset)
	}
}

func TestSessionMutationOptions_JSON(t *testing.T) {
	opts := SessionMutationOptions{Dir: strPtr("/project")}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got SessionMutationOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Dir != "/project" {
		t.Errorf("got dir %q", *got.Dir)
	}
}

func TestSessionMutationOptions_Empty(t *testing.T) {
	opts := SessionMutationOptions{}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("expected empty JSON object, got %s", data)
	}
}

// --- Fork types tests ---

func TestForkSessionOptions_JSON(t *testing.T) {
	opts := ForkSessionOptions{
		SessionMutationOptions: SessionMutationOptions{Dir: strPtr("/project")},
		UpToMessageID:          strPtr("msg-5"),
		Title:                  strPtr("forked session"),
	}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got ForkSessionOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if *got.Dir != "/project" {
		t.Errorf("got dir %q", *got.Dir)
	}
	if *got.UpToMessageID != "msg-5" {
		t.Errorf("got upToMessageId %q", *got.UpToMessageID)
	}
	if *got.Title != "forked session" {
		t.Errorf("got title %q", *got.Title)
	}
}

func TestForkSessionResult_JSON(t *testing.T) {
	raw := `{"sessionId": "new-session-id"}`
	var result ForkSessionResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	if result.SessionID != "new-session-id" {
		t.Errorf("got sessionId %q", result.SessionID)
	}
}

func TestForkSessionResult_RoundTrip(t *testing.T) {
	result := ForkSessionResult{SessionID: "fork-1"}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var got ForkSessionResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SessionID != result.SessionID {
		t.Errorf("round-trip mismatch")
	}
}

// --- SDKSessionOptions (V2) tests ---

func TestSDKSessionOptions_JSON(t *testing.T) {
	opts := SDKSessionOptions{
		Model:             "claude-sonnet-4-6",
		PathToClaudeCode:  strPtr("/usr/local/bin/claude"),
		Executable:        strPtr("node"),
		ExecutableArgs:    []string{"--max-old-space-size=4096"},
		Env:               map[string]string{"FOO": "bar"},
		AllowedTools:      []string{"Bash", "Read"},
		DisallowedTools:   []string{"Write"},
		PermissionMode:    (*PermissionMode)(strPtr(string(PermissionModeDefault))),
	}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKSessionOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Model != "claude-sonnet-4-6" {
		t.Errorf("got model %q", got.Model)
	}
	if *got.PathToClaudeCode != "/usr/local/bin/claude" {
		t.Errorf("got pathToClaudeCodeExecutable %q", *got.PathToClaudeCode)
	}
	if len(got.AllowedTools) != 2 {
		t.Errorf("got %d allowed tools", len(got.AllowedTools))
	}
}

func TestSDKSessionOptions_Minimal(t *testing.T) {
	opts := SDKSessionOptions{Model: "claude-sonnet-4-6"}
	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	var got SDKSessionOptions
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Model != "claude-sonnet-4-6" {
		t.Errorf("got model %q", got.Model)
	}
	if got.PathToClaudeCode != nil {
		t.Error("expected nil pathToClaudeCodeExecutable")
	}
}
