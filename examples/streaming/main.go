// Example: streaming messages from a Claude Code query.
package main

import (
	"fmt"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	maxTurns := 3
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "List 3 programming languages and why each is useful.",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		fmt.Printf("[%s] ", msg.MessageType())
		switch m := msg.(type) {
		case *claudeagent.SDKAssistantMessage:
			fmt.Printf("session=%s message=%s\n", m.SessionID, string(m.Message))
		case *claudeagent.SDKResultSuccess:
			fmt.Printf("session=%s turns=%d duration=%dms\n", m.SessionID, m.NumTurns, m.DurationMs)
		case *claudeagent.SDKResultError:
			fmt.Printf("errors=%v\n", m.Errors)
		default:
			fmt.Println()
		}
	}
}
