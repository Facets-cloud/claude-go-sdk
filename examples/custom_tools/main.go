// Example: configuring MCP servers with Claude Code SDK.
//
// This demonstrates how to set up MCP (Model Context Protocol) servers
// that provide additional tools to Claude. It also shows AllowedTools
// and DisallowedTools for controlling which built-in tools are available.
package main

import (
	"fmt"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	maxTurns := 2

	// Configure an MCP server that provides custom tools.
	// In a real scenario, this would point to a running MCP server process.
	mcpServers := map[string]interface{}{
		"my-tools": claudeagent.McpStdioServerConfig{
			Command: "npx",
			Args:    []string{"-y", "@anthropic-ai/mcp-server-example"},
			Env:     map[string]string{"API_KEY": "example-key"},
		},
	}

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Use the available tools to read go.mod",
		Options: &claudeagent.Options{
			MaxTurns:        &maxTurns,
			AllowedTools:    []string{"Read", "Glob"},
			DisallowedTools: []string{"Bash", "Write"},
			McpServers:      mcpServers,
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *claudeagent.SDKResultSuccess:
			fmt.Println("Success:", m.Result)
		case *claudeagent.SDKResultError:
			fmt.Println("Errors:", m.Errors)
		}
	}
}
