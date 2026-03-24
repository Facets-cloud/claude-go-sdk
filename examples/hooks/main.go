// Example: using hooks with Claude Code SDK.
//
// Hooks are configured as HookCallbackMatcher entries keyed by HookEvent.
// Each matcher can optionally filter by tool name pattern. The Hooks field
// on HookCallbackMatcher contains HookCallback functions that are invoked
// when the event fires.
package main

import (
	"context"
	"fmt"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func main() {
	maxTurns := 1
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "What time is it?",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
			Hooks: map[claudeagent.HookEvent][]claudeagent.HookCallbackMatcher{
				claudeagent.HookEventPreToolUse: {
					{
						Hooks: []claudeagent.HookCallback{
							func(ctx context.Context, input claudeagent.HookInput, toolUseID *string) (claudeagent.HookJSONOutput, error) {
								fmt.Println("Hook: PreToolUse triggered")
								return claudeagent.SyncHookJSONOutput{}, nil
							},
						},
					},
				},
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
