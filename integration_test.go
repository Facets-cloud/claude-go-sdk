//go:build integration

package claudeagent_test

import (
	"context"
	"testing"
	"time"

	claudeagent "github.com/anthropics/claude-agent-sdk-go"
)

func TestIntegration_BasicQuery(t *testing.T) {
	maxTurns := 1
	mode := claudeagent.PermissionModeBypassPermissions
	skip := true
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Reply with exactly: HELLO",
		Options: &claudeagent.Options{
			MaxTurns:                        &maxTurns,
			PermissionMode:                  &mode,
			AllowDangerouslySkipPermissions: &skip,
			SystemPrompt:                    "You are a test assistant. Follow instructions exactly.",
		},
	})
	defer q.Close()

	var result string
	for msg := range q.Messages() {
		if r, ok := msg.(*claudeagent.SDKResultSuccess); ok {
			result = r.Result
		}
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestIntegration_InitializationResult(t *testing.T) {
	maxTurns := 1
	mode := claudeagent.PermissionModeBypassPermissions
	skip := true
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Say hi",
		Options: &claudeagent.Options{
			MaxTurns:                        &maxTurns,
			PermissionMode:                  &mode,
			AllowDangerouslySkipPermissions: &skip,
		},
	})
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := q.InitializationResult(ctx)
	if err != nil {
		t.Fatalf("InitializationResult: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil initialization result")
	}

	// Drain messages
	for range q.Messages() {
	}
}

func TestIntegration_MaxTurnsRespected(t *testing.T) {
	maxTurns := 1
	mode := claudeagent.PermissionModeBypassPermissions
	skip := true
	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "What is Go?",
		Options: &claudeagent.Options{
			MaxTurns:                        &maxTurns,
			PermissionMode:                  &mode,
			AllowDangerouslySkipPermissions: &skip,
		},
	})
	defer q.Close()

	var resultMsg *claudeagent.SDKResultSuccess
	for msg := range q.Messages() {
		if r, ok := msg.(*claudeagent.SDKResultSuccess); ok {
			resultMsg = r
		}
	}
	if resultMsg == nil {
		t.Fatal("expected a result message")
	}
	if resultMsg.NumTurns > 1 {
		t.Errorf("expected at most 1 turn, got %d", resultMsg.NumTurns)
	}
}

func TestIntegration_PermissionCallback(t *testing.T) {
	maxTurns := 2
	permCalled := false

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Read the file go.mod",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts claudeagent.CanUseToolOptions) (claudeagent.PermissionResult, error) {
				permCalled = true
				return claudeagent.PermissionResultAllow{
					Behavior: claudeagent.PermissionBehaviorAllow,
				}, nil
			},
		},
	})
	defer q.Close()

	for range q.Messages() {
	}

	if !permCalled {
		t.Error("expected permission callback to be called")
	}
}

func TestIntegration_PermissionDeny(t *testing.T) {
	maxTurns := 2

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Read the file go.mod",
		Options: &claudeagent.Options{
			MaxTurns: &maxTurns,
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts claudeagent.CanUseToolOptions) (claudeagent.PermissionResult, error) {
				return claudeagent.PermissionResultDeny{
					Behavior: claudeagent.PermissionBehaviorDeny,
					Message:  "all tools denied for test",
				}, nil
			},
		},
	})
	defer q.Close()

	var gotResult bool
	for msg := range q.Messages() {
		if claudeagent.IsResultMessage(msg) {
			gotResult = true
		}
	}
	if !gotResult {
		t.Error("expected a result message even when tools are denied")
	}
}

func TestIntegration_SystemPrompt(t *testing.T) {
	maxTurns := 1
	mode := claudeagent.PermissionModeBypassPermissions
	skip := true

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "What is your name?",
		Options: &claudeagent.Options{
			MaxTurns:                        &maxTurns,
			PermissionMode:                  &mode,
			AllowDangerouslySkipPermissions: &skip,
			SystemPrompt:                    "Your name is TestBot. Always introduce yourself as TestBot.",
		},
	})
	defer q.Close()

	var result string
	for msg := range q.Messages() {
		if r, ok := msg.(*claudeagent.SDKResultSuccess); ok {
			result = r.Result
		}
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestIntegration_Interrupt(t *testing.T) {
	maxTurns := 5
	mode := claudeagent.PermissionModeBypassPermissions
	skip := true

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Write a very long essay about the history of computing, at least 5000 words.",
		Options: &claudeagent.Options{
			MaxTurns:                        &maxTurns,
			PermissionMode:                  &mode,
			AllowDangerouslySkipPermissions: &skip,
		},
	})
	defer q.Close()

	// Wait for init to complete, then interrupt.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := q.InitializationResult(ctx)
	if err != nil {
		t.Fatalf("InitializationResult: %v", err)
	}

	// Give it a moment to start generating, then interrupt.
	time.Sleep(500 * time.Millisecond)
	if err := q.Interrupt(ctx); err != nil {
		t.Logf("Interrupt returned error (may be expected): %v", err)
	}

	// Drain remaining messages.
	for range q.Messages() {
	}
}

func TestIntegration_MessageTypes(t *testing.T) {
	maxTurns := 1
	mode := claudeagent.PermissionModeBypassPermissions
	skip := true

	q := claudeagent.NewQuery(claudeagent.QueryParams{
		Prompt: "Reply with exactly: OK",
		Options: &claudeagent.Options{
			MaxTurns:                        &maxTurns,
			PermissionMode:                  &mode,
			AllowDangerouslySkipPermissions: &skip,
		},
	})
	defer q.Close()

	var types []string
	for msg := range q.Messages() {
		types = append(types, msg.MessageType())
	}

	if len(types) == 0 {
		t.Fatal("expected at least one message")
	}

	// Should end with a result message.
	lastType := types[len(types)-1]
	if lastType != "result" {
		t.Errorf("last message type = %q, want 'result'", lastType)
	}
}
