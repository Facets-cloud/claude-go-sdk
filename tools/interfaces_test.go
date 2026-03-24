package tools

import (
	"encoding/json"
	"testing"
)

// Compile-time interface satisfaction checks for AgentOutput variants.
var _ AgentOutput = AgentOutputCompleted{}
var _ AgentOutput = AgentOutputAsyncLaunched{}

// Compile-time interface satisfaction checks for FileReadOutput variants.
var _ FileReadOutput = FileReadText{}
var _ FileReadOutput = FileReadImage{}
var _ FileReadOutput = FileReadNotebook{}
var _ FileReadOutput = FileReadPDF{}
var _ FileReadOutput = FileReadParts{}

func TestUnmarshalAgentOutput_InvalidJSON(t *testing.T) {
	_, err := UnmarshalAgentOutput([]byte(`{invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestUnmarshalAgentOutput_InvalidCompletedPayload(t *testing.T) {
	// Valid JSON but the status probe succeeds, then the full unmarshal
	// should still work for valid-shaped data. Test with bad nested types.
	_, err := UnmarshalAgentOutput([]byte(`{"status": "completed", "totalToolUseCount": "not-a-number"}`))
	if err == nil {
		t.Error("expected error for invalid field type")
	}
}

func TestUnmarshalAgentOutput_InvalidAsyncPayload(t *testing.T) {
	_, err := UnmarshalAgentOutput([]byte(`{"status": "async_launched", "agentId": 123}`))
	if err == nil {
		t.Error("expected error for invalid field type")
	}
}

func TestUnmarshalFileReadOutput_InvalidJSON(t *testing.T) {
	_, err := UnmarshalFileReadOutput([]byte(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestUnmarshalFileReadOutput_InvalidTextPayload(t *testing.T) {
	_, err := UnmarshalFileReadOutput([]byte(`{"type": "text", "file": "not-an-object"}`))
	if err == nil {
		t.Error("expected error for invalid file field")
	}
}

func TestUnmarshalFileReadOutput_InvalidImagePayload(t *testing.T) {
	_, err := UnmarshalFileReadOutput([]byte(`{"type": "image", "file": "bad"}`))
	if err == nil {
		t.Error("expected error for invalid image payload")
	}
}

func TestUnmarshalFileReadOutput_InvalidNotebookPayload(t *testing.T) {
	_, err := UnmarshalFileReadOutput([]byte(`{"type": "notebook", "file": 42}`))
	if err == nil {
		t.Error("expected error for invalid notebook payload")
	}
}

func TestUnmarshalFileReadOutput_InvalidPDFPayload(t *testing.T) {
	_, err := UnmarshalFileReadOutput([]byte(`{"type": "pdf", "file": false}`))
	if err == nil {
		t.Error("expected error for invalid PDF payload")
	}
}

func TestUnmarshalFileReadOutput_InvalidPartsPayload(t *testing.T) {
	_, err := UnmarshalFileReadOutput([]byte(`{"type": "parts", "file": true}`))
	if err == nil {
		t.Error("expected error for invalid parts payload")
	}
}

// Edge case tests: nil/null fields, empty slices, zero values.

func TestAgentInput_MinimalFields(t *testing.T) {
	input := AgentInput{Description: "d", Prompt: "p"}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got AgentInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.SubagentType != nil {
		t.Error("SubagentType should be nil")
	}
	if got.RunInBackground != nil {
		t.Error("RunInBackground should be nil")
	}
}

func TestBashOutput_NullOptionalFields(t *testing.T) {
	raw := `{"stdout": "", "stderr": "", "interrupted": false}`
	var out BashOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.IsImage != nil {
		t.Error("IsImage should be nil")
	}
	if out.BackgroundTaskID != nil {
		t.Error("BackgroundTaskID should be nil")
	}
	if out.StructuredContent != nil {
		t.Error("StructuredContent should be nil")
	}
}

func TestGlobOutput_EmptyFilenames(t *testing.T) {
	raw := `{"durationMs": 1.0, "numFiles": 0, "filenames": [], "truncated": false}`
	var out GlobOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.NumFiles != 0 {
		t.Errorf("NumFiles = %d", out.NumFiles)
	}
	if len(out.Filenames) != 0 {
		t.Errorf("Filenames should be empty, got %d", len(out.Filenames))
	}
}

func TestGrepOutput_ContentMode(t *testing.T) {
	content := "line1\nline2"
	raw := `{"mode": "content", "numFiles": 1, "filenames": ["a.go"], "content": "line1\nline2", "numLines": 2, "numMatches": 2, "appliedLimit": 10, "appliedOffset": 0}`
	var out GrepOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Content == nil || *out.Content != content {
		t.Errorf("Content = %v", out.Content)
	}
	if out.AppliedLimit == nil || *out.AppliedLimit != 10 {
		t.Errorf("AppliedLimit = %v", out.AppliedLimit)
	}
}

func TestFileWriteOutput_UpdateWithOriginal(t *testing.T) {
	orig := "old content"
	raw := `{
		"type": "update",
		"filePath": "/tmp/existing.go",
		"content": "new content",
		"structuredPatch": [{"oldStart": 1, "oldLines": 1, "newStart": 1, "newLines": 1, "lines": ["-old content", "+new content"]}],
		"originalFile": "old content",
		"gitDiff": {"filename": "existing.go", "status": "modified", "additions": 1, "deletions": 1, "changes": 2, "patch": "@@..."}
	}`
	var out FileWriteOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Type != "update" {
		t.Errorf("Type = %q", out.Type)
	}
	if out.OriginalFile == nil || *out.OriginalFile != orig {
		t.Errorf("OriginalFile = %v", out.OriginalFile)
	}
	if out.GitDiff == nil {
		t.Fatal("GitDiff should not be nil")
	}
	if out.GitDiff.Status != "modified" {
		t.Errorf("GitDiff.Status = %q", out.GitDiff.Status)
	}
}

func TestFileEditOutput_WithGitDiff(t *testing.T) {
	raw := `{
		"filePath": "/tmp/f.go",
		"oldString": "a",
		"newString": "b",
		"originalFile": "a",
		"structuredPatch": [],
		"userModified": true,
		"replaceAll": true,
		"gitDiff": {"filename": "f.go", "status": "modified", "additions": 1, "deletions": 1, "changes": 2, "patch": "@@", "repository": "owner/repo"}
	}`
	var out FileEditOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if !out.UserModified {
		t.Error("expected UserModified=true")
	}
	if out.GitDiff == nil || out.GitDiff.Repository == nil || *out.GitDiff.Repository != "owner/repo" {
		t.Error("expected GitDiff.Repository = owner/repo")
	}
}

func TestFileReadImage_NoDimensions(t *testing.T) {
	raw := `{"type": "image", "file": {"base64": "abc", "type": "image/jpeg", "originalSize": 100}}`
	out, err := UnmarshalFileReadOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	img := out.(FileReadImage)
	if img.File.Dimensions != nil {
		t.Error("Dimensions should be nil when omitted")
	}
}

func TestMcpInput_ArbitraryJSON(t *testing.T) {
	raw := `{"tool": "my-tool", "args": {"key": "value"}, "nested": [1, 2, 3]}`
	var input McpInput
	if err := json.Unmarshal([]byte(raw), &input); err != nil {
		t.Fatal(err)
	}
	if input["tool"] != "my-tool" {
		t.Errorf("tool = %v", input["tool"])
	}
}

func TestListMcpResourcesOutput_Empty(t *testing.T) {
	var out ListMcpResourcesOutput
	if err := json.Unmarshal([]byte(`[]`), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Errorf("expected empty, got %d", len(out))
	}
}

func TestTodoWriteOutput_WithVerificationNudge(t *testing.T) {
	raw := `{
		"oldTodos": [],
		"newTodos": [{"content": "test", "status": "pending", "activeForm": "Testing"}],
		"verificationNudgeNeeded": true
	}`
	var out TodoWriteOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.VerificationNudgeNeeded == nil || !*out.VerificationNudgeNeeded {
		t.Error("expected VerificationNudgeNeeded=true")
	}
}

func TestConfigInput_BooleanValue(t *testing.T) {
	input := ConfigInput{Setting: "darkMode", Value: true}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got ConfigInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Value != true {
		t.Errorf("Value = %v", got.Value)
	}
}

func TestConfigInput_NumericValue(t *testing.T) {
	input := ConfigInput{Setting: "maxTokens", Value: float64(4096)}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got ConfigInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Value != float64(4096) {
		t.Errorf("Value = %v", got.Value)
	}
}

func TestConfigOutput_ErrorCase(t *testing.T) {
	raw := `{"success": false, "error": "unknown setting"}`
	var out ConfigOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Success {
		t.Error("expected Success=false")
	}
	if out.Error == nil || *out.Error != "unknown setting" {
		t.Errorf("Error = %v", out.Error)
	}
}

func TestExitWorktreeInput_JSON(t *testing.T) {
	input := ExitWorktreeInput{
		Action:         "remove",
		DiscardChanges: boolPtr(true),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got ExitWorktreeInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Action != "remove" {
		t.Errorf("Action = %q", got.Action)
	}
	if got.DiscardChanges == nil || !*got.DiscardChanges {
		t.Error("expected DiscardChanges=true")
	}
}

func TestExitPlanModeOutput_NullPlan(t *testing.T) {
	raw := `{"plan": null, "isAgent": false}`
	var out ExitPlanModeOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Plan != nil {
		t.Error("Plan should be nil")
	}
	if out.IsAgent {
		t.Error("expected IsAgent=false")
	}
}

func TestWebSearchInput_NoDomainFilters(t *testing.T) {
	input := WebSearchInput{Query: "test"}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got WebSearchInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.AllowedDomains != nil {
		t.Error("AllowedDomains should be nil")
	}
	if got.BlockedDomains != nil {
		t.Error("BlockedDomains should be nil")
	}
}

func TestTaskStopInput_JSON(t *testing.T) {
	input := TaskStopInput{TaskID: strPtr("t-1")}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got TaskStopInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.TaskID == nil || *got.TaskID != "t-1" {
		t.Errorf("TaskID = %v", got.TaskID)
	}
	if got.ShellID != nil {
		t.Error("ShellID should be nil")
	}
}
