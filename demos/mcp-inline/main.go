// MCP Inline Demo — Claude Agent SDK for Go
//
// Demonstrates using custom in-process MCP tools with Claude.
// The MCP server is built as a separate binary and registered via McpStdioServerConfig.
//
// Usage: go run ./demos/mcp-inline
package main

import (
	"encoding/json"
	"fmt"
	"os"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	// Build the MCP server binary path
	// In production, you'd distribute this as a compiled binary
	goExe := "go"

	cwd, _ := os.Getwd()
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: `Use the greet tool to greet "Alice", then use the calculate tool to multiply 7 by 8, then use the reverse_text tool to reverse "Hello World". Report all three results.`,
		Options: &claudeagent.Options{
			MaxTurns: claudeagent.Int(5),
			Model:    claudeagent.String("sonnet"),
			Cwd:      &cwd,
			McpServers: map[string]interface{}{
				"go-tools": claudeagent.McpStdioServerConfig{
					Command: goExe,
					Args:    []string{"run", "./demos/mcp-inline/server"},
				},
			},
			PermissionMode: permModePtr(claudeagent.PermissionModeDontAsk),
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *claudeagent.SDKSystemMessage:
			fmt.Printf("[init] model=%s, mcp_servers=%v\n", m.Model, m.McpServers)

		case *claudeagent.SDKAssistantMessage:
			var parsed struct {
				Content []struct {
					Type  string                 `json:"type"`
					Text  string                 `json:"text"`
					Name  string                 `json:"name"`
					Input map[string]interface{} `json:"input"`
				} `json:"content"`
			}
			if err := json.Unmarshal(m.Message, &parsed); err == nil {
				for _, block := range parsed.Content {
					switch block.Type {
					case "text":
						fmt.Printf("\nClaude: %s\n", block.Text)
					case "tool_use":
						fmt.Printf("\n[tool call: %s(%v)]\n", block.Name, block.Input)
					}
				}
			}

		case *claudeagent.SDKResultSuccess:
			fmt.Printf("\n---\nResult: %s\nCost: $%.4f | Turns: %d\n", m.Result, m.TotalCostUSD, m.NumTurns)

		case *claudeagent.SDKResultError:
			fmt.Printf("\nError: %v\n", m.Errors)
		}
	}
}

func permModePtr(m claudeagent.PermissionMode) *claudeagent.PermissionMode {
	return &m
}
