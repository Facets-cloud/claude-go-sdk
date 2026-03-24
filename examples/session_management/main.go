// Example: session continuation with Claude Code SDK.
package main

import (
	"context"
	"fmt"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	maxTurns := 1

	// First query
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Remember the number 42.",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
		},
	})

	// Get session ID from initialization
	initResult, err := q.InitializationResult(context.Background())
	if err != nil {
		fmt.Println("Init error:", err)
		return
	}
	fmt.Printf("Init complete: output_style=%s\n", initResult.OutputStyle)

	// Drain messages
	var sessionID string
	for msg := range q.Messages() {
		if r, ok := msg.(*claudeagent.SDKResultSuccess); ok {
			sessionID = r.SessionID
			fmt.Println("First query result:", r.Result)
		}
	}
	q.Close()

	if sessionID == "" {
		fmt.Println("No session ID received")
		return
	}

	// Resume the session
	continueSession := true
	q2 := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "What number did I ask you to remember?",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
			Resume:   &sessionID,
			Continue: &continueSession,
		},
	})
	defer q2.Close()

	for msg := range q2.Messages() {
		if r, ok := msg.(*claudeagent.SDKResultSuccess); ok {
			fmt.Println("Resumed query result:", r.Result)
		}
	}
}
