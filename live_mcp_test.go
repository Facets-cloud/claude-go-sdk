//go:build live

package claudeagent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLive_McpInlineTools(t *testing.T) {
	cwd := "."
	q := NewQuery(QueryParams{
		Prompt: `Use the greet tool to greet "TestUser". Reply with ONLY the exact greeting text returned by the tool, nothing else.`,
		Options: &Options{
			MaxTurns: Int(3),
			Model:    String("sonnet"),
			Cwd:      &cwd,
			McpServers: map[string]interface{}{
				"go-tools": McpStdioServerConfig{
					Command: "go",
					Args:    []string{"run", "./demos/mcp-inline/server"},
				},
			},
			PermissionMode: permModePtr(PermissionModeDontAsk),
			AllowedTools:   []string{"mcp__go-tools__greet", "mcp__go-tools__calculate", "mcp__go-tools__reverse_text"},
			SystemPrompt:   "You are a test assistant. When asked to use a tool, use it and return ONLY the tool's output text. No extra commentary.",
		},
	})
	defer q.Close()

	var result string
	var toolUsed bool
	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *SDKAssistantMessage:
			var parsed struct {
				Content []struct {
					Type string `json:"type"`
					Name string `json:"name"`
					Text string `json:"text"`
				} `json:"content"`
			}
			if err := json.Unmarshal(m.Message, &parsed); err == nil {
				for _, block := range parsed.Content {
					if block.Type == "tool_use" {
						t.Logf("Tool used: %s", block.Name)
						if strings.Contains(block.Name, "greet") {
							toolUsed = true
						}
					}
				}
			}
		case *SDKResultSuccess:
			result = m.Result
			t.Logf("Result: %q (cost=$%.4f, turns=%d)", m.Result, m.TotalCostUSD, m.NumTurns)
		case *SDKResultError:
			t.Logf("Error: %v", m.Errors)
		}
	}

	if !toolUsed {
		t.Error("expected the greet MCP tool to be used")
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
	if !strings.Contains(strings.ToLower(result), "hello") && !strings.Contains(strings.ToLower(result), "testuser") {
		t.Errorf("result should contain greeting for TestUser, got: %q", result)
	}
}

func TestLive_McpCalculateTool(t *testing.T) {
	cwd := "."
	q := NewQuery(QueryParams{
		Prompt: `Use the calculate tool to multiply 7 by 8. Reply with ONLY the number, nothing else.`,
		Options: &Options{
			MaxTurns: Int(3),
			Model:    String("sonnet"),
			Cwd:      &cwd,
			McpServers: map[string]interface{}{
				"go-tools": McpStdioServerConfig{
					Command: "go",
					Args:    []string{"run", "./demos/mcp-inline/server"},
				},
			},
			PermissionMode: permModePtr(PermissionModeDontAsk),
			AllowedTools:   []string{"mcp__go-tools__calculate"},
			SystemPrompt:   "You are a test assistant. Use tools when asked. Reply with ONLY the result.",
		},
	})
	defer q.Close()

	var result string
	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *SDKResultSuccess:
			result = m.Result
			t.Logf("Result: %q (cost=$%.4f)", m.Result, m.TotalCostUSD)
		case *SDKResultError:
			t.Logf("Error: %v", m.Errors)
		}
	}

	if !strings.Contains(result, "56") {
		t.Errorf("expected result to contain '56' (7*8), got: %q", result)
	}
}

func TestLive_McpReverseTextTool(t *testing.T) {
	cwd := "."
	q := NewQuery(QueryParams{
		Prompt: `Use the reverse_text tool to reverse "GoSDK". Reply with ONLY the reversed text, nothing else.`,
		Options: &Options{
			MaxTurns: Int(3),
			Model:    String("sonnet"),
			Cwd:      &cwd,
			McpServers: map[string]interface{}{
				"go-tools": McpStdioServerConfig{
					Command: "go",
					Args:    []string{"run", "./demos/mcp-inline/server"},
				},
			},
			PermissionMode: permModePtr(PermissionModeDontAsk),
			AllowedTools:   []string{"mcp__go-tools__reverse_text"},
			SystemPrompt:   "You are a test assistant. Use tools when asked. Reply with ONLY the result.",
		},
	})
	defer q.Close()

	var result string
	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *SDKResultSuccess:
			result = m.Result
			t.Logf("Result: %q (cost=$%.4f)", m.Result, m.TotalCostUSD)
		case *SDKResultError:
			t.Logf("Error: %v", m.Errors)
		}
	}

	if !strings.Contains(result, "KDSoG") {
		t.Errorf("expected result to contain 'KDSoG' (reverse of GoSDK), got: %q", result)
	}
}
