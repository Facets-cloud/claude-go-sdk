package tools

import (
	"encoding/json"
	"testing"
)

func TestAgentInput_JSON(t *testing.T) {
	input := AgentInput{
		Description: "Search codebase",
		Prompt:      "Find all TODO comments",
		Model:       strPtr("sonnet"),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got AgentInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Description != input.Description {
		t.Errorf("Description = %q, want %q", got.Description, input.Description)
	}
	if got.Prompt != input.Prompt {
		t.Errorf("Prompt = %q, want %q", got.Prompt, input.Prompt)
	}
	if got.Model == nil || *got.Model != "sonnet" {
		t.Errorf("Model = %v, want %q", got.Model, "sonnet")
	}
}

func TestAgentInput_OptionalFields(t *testing.T) {
	input := AgentInput{
		Description: "task",
		Prompt:      "do it",
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, field := range []string{"subagent_type", "model", "run_in_background", "name", "team_name", "mode", "isolation"} {
		if contains(s, field) {
			t.Errorf("optional field %q should be omitted, got %s", field, s)
		}
	}
}

func TestUnmarshalAgentOutput_Completed(t *testing.T) {
	raw := `{
		"agentId": "agent-1",
		"content": [{"type": "text", "text": "Done"}],
		"totalToolUseCount": 5,
		"totalDurationMs": 1000,
		"totalTokens": 500,
		"usage": {"input_tokens": 200, "output_tokens": 300, "cache_creation_input_tokens": null, "cache_read_input_tokens": null, "server_tool_use": null, "service_tier": null, "cache_creation": null},
		"status": "completed",
		"prompt": "find bugs"
	}`
	out, err := UnmarshalAgentOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	completed, ok := out.(AgentOutputCompleted)
	if !ok {
		t.Fatalf("expected AgentOutputCompleted, got %T", out)
	}
	if completed.AgentID != "agent-1" {
		t.Errorf("AgentID = %q, want %q", completed.AgentID, "agent-1")
	}
	if completed.Status != "completed" {
		t.Errorf("Status = %q, want %q", completed.Status, "completed")
	}
	if len(completed.Content) != 1 || completed.Content[0].Text != "Done" {
		t.Errorf("unexpected content: %+v", completed.Content)
	}
}

func TestUnmarshalAgentOutput_AsyncLaunched(t *testing.T) {
	raw := `{
		"status": "async_launched",
		"agentId": "agent-2",
		"description": "background task",
		"prompt": "run tests",
		"outputFile": "/tmp/out.json",
		"canReadOutputFile": true
	}`
	out, err := UnmarshalAgentOutput([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	async, ok := out.(AgentOutputAsyncLaunched)
	if !ok {
		t.Fatalf("expected AgentOutputAsyncLaunched, got %T", out)
	}
	if async.AgentID != "agent-2" {
		t.Errorf("AgentID = %q, want %q", async.AgentID, "agent-2")
	}
	if async.OutputFile != "/tmp/out.json" {
		t.Errorf("OutputFile = %q, want %q", async.OutputFile, "/tmp/out.json")
	}
}

// helpers
func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
func boolPtr(b bool) *bool    { return &b }

func contains(s, substr string) bool {
	return len(s) >= len(substr) && jsonContains(s, substr)
}

func jsonContains(s, key string) bool {
	return json.Valid([]byte(s)) && len(s) > 0 && stringContains(s, `"`+key+`"`)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
