// Example: custom permission handler with Claude Code SDK.
package main

import (
	"context"
	"fmt"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	maxTurns := 2
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Create a file called hello.txt with 'Hello, World!' in it",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts claudeagent.CanUseToolOptions) (claudeagent.PermissionResult, error) {
				fmt.Printf("Permission request: tool=%s\n", toolName)

				// Allow Read, deny everything else
				if toolName == "Read" {
					return claudeagent.PermissionResultAllow{
						Behavior: claudeagent.PermissionBehaviorAllow,
					}, nil
				}

				return claudeagent.PermissionResultDeny{
					Behavior: claudeagent.PermissionBehaviorDeny,
					Message:  fmt.Sprintf("tool %s is not allowed in this example", toolName),
				}, nil
			},
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *claudeagent.SDKResultSuccess:
			fmt.Println("Done:", m.Result)
		case *claudeagent.SDKResultError:
			fmt.Println("Errors:", m.Errors)
		}
	}
}
