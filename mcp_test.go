package claudeagent

import (
	"encoding/json"
	"testing"
)

func TestMcpStdioServerConfig_JSON(t *testing.T) {
	raw := `{"command":"node","args":["server.js"],"env":{"PORT":"3000"}}`
	var c McpStdioServerConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.Command != "node" {
		t.Errorf("Command = %q", c.Command)
	}
}

func TestMcpSSEServerConfig_JSON(t *testing.T) {
	raw := `{"type":"sse","url":"http://localhost:3000","headers":{"Authorization":"Bearer tok"}}`
	var c McpSSEServerConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.URL != "http://localhost:3000" {
		t.Errorf("URL = %q", c.URL)
	}
}

func TestMcpHttpServerConfig_JSON(t *testing.T) {
	raw := `{"type":"http","url":"http://localhost:8080"}`
	var c McpHttpServerConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.Type != "http" {
		t.Errorf("Type = %q", c.Type)
	}
}

func TestMcpSdkServerConfig_JSON(t *testing.T) {
	raw := `{"type":"sdk","name":"my-server"}`
	var c McpSdkServerConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "my-server" {
		t.Errorf("Name = %q", c.Name)
	}
}

func TestMcpClaudeAIProxyServerConfig_JSON(t *testing.T) {
	raw := `{"type":"claudeai-proxy","url":"https://proxy.claude.ai","id":"srv1"}`
	var c McpClaudeAIProxyServerConfig
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		t.Fatal(err)
	}
	if c.ID != "srv1" {
		t.Errorf("ID = %q", c.ID)
	}
}

func TestMcpServerStatus_JSON(t *testing.T) {
	raw := `{"name":"my-server","status":"connected","serverInfo":{"name":"test","version":"1.0"},"tools":[{"name":"mytool","description":"a tool"}]}`
	var s McpServerStatus
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Name != "my-server" {
		t.Errorf("Name = %q", s.Name)
	}
	if len(s.Tools) != 1 {
		t.Errorf("Tools len = %d", len(s.Tools))
	}
}

func TestMcpSetServersResult_JSON(t *testing.T) {
	raw := `{"added":["s1"],"removed":["s2"],"errors":{"s3":"failed to connect"}}`
	var r McpSetServersResult
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		t.Fatal(err)
	}
	if len(r.Added) != 1 {
		t.Errorf("Added = %v", r.Added)
	}
	if r.Errors["s3"] != "failed to connect" {
		t.Errorf("Errors = %v", r.Errors)
	}
}
