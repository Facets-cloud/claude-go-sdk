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
