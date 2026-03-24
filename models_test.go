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

func TestAgentInfo_JSON(t *testing.T) {
	raw := `{"name":"Explore","description":"For exploration","model":"sonnet"}`
	var a AgentInfo
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatal(err)
	}
	if a.Name != "Explore" {
		t.Errorf("Name = %q", a.Name)
	}
}

func TestSlashCommand_JSON(t *testing.T) {
	raw := `{"name":"commit","description":"Create a git commit","argumentHint":"<message>"}`
	var s SlashCommand
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "commit" {
		t.Errorf("Name = %q", s.Name)
	}
	if s.ArgumentHint != "<message>" {
		t.Errorf("ArgumentHint = %q", s.ArgumentHint)
	}
}

func TestPromptRequest_JSON(t *testing.T) {
	raw := `{"prompt":"req1","message":"Choose an option","options":[{"key":"a","label":"Option A","description":"First option"}]}`
	var p PromptRequest
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	if p.Prompt != "req1" {
		t.Errorf("Prompt = %q", p.Prompt)
	}
	if len(p.Options) != 1 {
		t.Errorf("Options = %v", p.Options)
	}
	if p.Options[0].Key != "a" {
		t.Errorf("Options[0].Key = %q", p.Options[0].Key)
	}
}

func TestPromptResponse_JSON(t *testing.T) {
	raw := `{"prompt_response":"req1","selected":"a"}`
	var p PromptResponse
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatal(err)
	}
	if p.PromptResponse != "req1" {
		t.Errorf("PromptResponse = %q", p.PromptResponse)
	}
	if p.Selected != "a" {
		t.Errorf("Selected = %q", p.Selected)
	}
}
