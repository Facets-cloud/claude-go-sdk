// Example: basic single-turn query with Claude Code SDK.
package main

import (
	"fmt"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	maxTurns := 1
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "What is 2 + 2?",
		Options: &claudeagent.Options{
			SystemPrompt: "You are a helpful math assistant.",
			MaxTurns:     &maxTurns,
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *claudeagent.SDKAssistantMessage:
			fmt.Printf("Assistant: %s\n", string(m.Message))
		case *claudeagent.SDKResultSuccess:
			fmt.Printf("Result: %s (cost: $%.4f)\n", m.Result, m.TotalCostUSD)
		case *claudeagent.SDKResultError:
			fmt.Printf("Error: %v\n", m.Errors)
		}
	}
}
