//go:build live

// Live tests — require Claude Code CLI installed and authenticated.
// Run with: go test -tags live -v -timeout 120s -run TestLive ./...
//
// These tests make real API calls and cost money.

package claudeagent

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestLive_BasicQuery(t *testing.T) {
	q := NewQuery(QueryParams{
		Prompt: "Reply with exactly the word PONG and nothing else.",
		Options: &Options{
			MaxTurns:       Int(1),
			Model:          String("sonnet"),
			PermissionMode: permModePtr(PermissionModeDontAsk),
			SystemPrompt:   "You are a test assistant. Follow instructions exactly. No extra text.",
		},
	})
	defer q.Close()

	var result string
	var gotInit bool
	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *SDKSystemMessage:
			gotInit = true
			t.Logf("Init: model=%s, tools=%v", m.Model, m.Tools[:min(3, len(m.Tools))])
		case *SDKAssistantMessage:
			t.Logf("Assistant message received (uuid=%s)", m.UUID)
		case *SDKResultSuccess:
			result = m.Result
			t.Logf("Result: %q (cost=$%.4f, turns=%d)", m.Result, m.TotalCostUSD, m.NumTurns)
		case *SDKResultError:
			t.Fatalf("Got error result: %v", m.Errors)
		}
	}

	if !gotInit {
		t.Error("never received init message")
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestLive_InitializationResult(t *testing.T) {
	q := NewQuery(QueryParams{
		Prompt: "Say hello",
		Options: &Options{
			MaxTurns:       Int(1),
			Model:          String("sonnet"),
			PermissionMode: permModePtr(PermissionModeDontAsk),
		},
	})
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	initResp, err := q.InitializationResult(ctx)
	if err != nil {
		t.Fatalf("InitializationResult: %v", err)
	}
	if initResp == nil {
		t.Fatal("initResp is nil")
	}

	t.Logf("Commands: %d", len(initResp.Commands))
	t.Logf("Models: %d", len(initResp.Models))
	t.Logf("Agents: %d", len(initResp.Agents))
	t.Logf("OutputStyle: %s", initResp.OutputStyle)

	if len(initResp.Models) == 0 {
		t.Error("expected at least one model")
	}

	// Drain messages
	for range q.Messages() {
	}
}

func TestLive_SupportedModels(t *testing.T) {
	q := NewQuery(QueryParams{
		Prompt: "Say hi",
		Options: &Options{
			MaxTurns:       Int(1),
			Model:          String("sonnet"),
			PermissionMode: permModePtr(PermissionModeDontAsk),
		},
	})
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	models, err := q.SupportedModels(ctx)
	if err != nil {
		t.Fatalf("SupportedModels: %v", err)
	}

	t.Logf("Got %d models:", len(models))
	for _, m := range models {
		t.Logf("  %s (%s) - effort=%v, fastMode=%v", m.Value, m.DisplayName, m.SupportsEffort, m.SupportsFastMode)
	}

	if len(models) == 0 {
		t.Error("expected at least one model")
	}

	// Drain
	for range q.Messages() {
	}
}

func TestLive_SupportedAgents(t *testing.T) {
	q := NewQuery(QueryParams{
		Prompt: "Say hi",
		Options: &Options{
			MaxTurns:       Int(1),
			Model:          String("sonnet"),
			PermissionMode: permModePtr(PermissionModeDontAsk),
		},
	})
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	agents, err := q.SupportedAgents(ctx)
	if err != nil {
		t.Fatalf("SupportedAgents: %v", err)
	}

	t.Logf("Got %d agents:", len(agents))
	for _, a := range agents {
		t.Logf("  %s: %s", a.Name, a.Description[:min(80, len(a.Description))])
	}

	// Drain
	for range q.Messages() {
	}
}

func TestLive_AccountInfo(t *testing.T) {
	q := NewQuery(QueryParams{
		Prompt: "Say hi",
		Options: &Options{
			MaxTurns:       Int(1),
			Model:          String("sonnet"),
			PermissionMode: permModePtr(PermissionModeDontAsk),
		},
	})
	defer q.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	account, err := q.AccountInfo(ctx)
	if err != nil {
		t.Fatalf("AccountInfo: %v", err)
	}

	if account != nil {
		t.Logf("Account: email=%v, org=%v, provider=%v", account.Email, account.Organization, account.ApiProvider)
	} else {
		t.Log("Account info is nil (may be expected for some auth types)")
	}

	// Drain
	for range q.Messages() {
	}
}

func TestLive_MessageTypes(t *testing.T) {
	q := NewQuery(QueryParams{
		Prompt: "What is 2+2? Reply with just the number.",
		Options: &Options{
			MaxTurns:       Int(1),
			Model:          String("sonnet"),
			PermissionMode: permModePtr(PermissionModeDontAsk),
		},
	})
	defer q.Close()

	typeCounts := make(map[string]int)
	for msg := range q.Messages() {
		typeCounts[msg.MessageType()]++

		// Verify we can type-switch on all message types
		switch m := msg.(type) {
		case *SDKSystemMessage:
			if m.ClaudeCodeVersion == "" {
				t.Error("missing claude_code_version in init")
			}
		case *SDKAssistantMessage:
			var parsed struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			}
			if err := json.Unmarshal(m.Message, &parsed); err == nil {
				for _, block := range parsed.Content {
					if block.Type == "text" {
						t.Logf("Claude: %s", block.Text)
					}
				}
			}
		case *SDKResultSuccess:
			t.Logf("Success: %q", m.Result)
		case *SDKResultError:
			t.Logf("Error: %v", m.Errors)
		}
	}

	t.Logf("Message types received: %v", typeCounts)
	if typeCounts["system"] == 0 {
		t.Error("expected at least one system message")
	}
	if typeCounts["result"] == 0 {
		t.Error("expected a result message")
	}
}

func TestLive_PermissionCallback(t *testing.T) {
	permCalled := false
	q := NewQuery(QueryParams{
		Prompt: "Run the command: echo HELLO",
		Options: &Options{
			MaxTurns: Int(3),
			Model:    String("sonnet"),
			SystemPrompt: "You must use the Bash tool to run commands. Always use Bash.",
			CanUseTool: func(ctx context.Context, toolName string, input map[string]interface{}, opts CanUseToolOptions) (PermissionResult, error) {
				permCalled = true
				t.Logf("Permission requested: tool=%s, input=%v", toolName, input)
				return PermissionResultAllow{Behavior: PermissionBehaviorAllow}, nil
			},
		},
	})
	defer q.Close()

	for msg := range q.Messages() {
		switch m := msg.(type) {
		case *SDKResultSuccess:
			t.Logf("Result: %q", m.Result)
		case *SDKResultError:
			t.Logf("Error: %v", m.Errors)
		}
	}

	if !permCalled {
		t.Log("Warning: permission callback was not invoked (model may not have used Bash)")
	}
}

func TestLive_V2Session_BasicSendStream(t *testing.T) {
	sess, err := CreateSession(SDKSessionOptions{
		Model: "sonnet",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	defer sess.Close()

	sess.Send("What is 3+3? Reply with just the number.")

	var gotResult bool
	for msg := range sess.Stream() {
		switch m := msg.(type) {
		case *SDKResultSuccess:
			gotResult = true
			t.Logf("V2 Result: %q (cost=$%.4f)", m.Result, m.TotalCostUSD)
		case *SDKResultError:
			t.Logf("V2 Error: %v", m.Errors)
		}
	}

	if !gotResult {
		t.Error("expected a result from V2 session")
	}
}

func TestLive_Prompt_OneShot(t *testing.T) {
	result, err := Prompt("What is the capital of France? One word only.", SDKSessionOptions{
		Model: "sonnet",
	})
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}

	switch m := result.(type) {
	case *SDKResultSuccess:
		t.Logf("Prompt result: %q (cost=$%.4f)", m.Result, m.TotalCostUSD)
		if m.Result == "" {
			t.Error("expected non-empty result")
		}
	case *SDKResultError:
		t.Logf("Prompt error: %v", m.Errors)
	default:
		t.Errorf("unexpected result type: %T", result)
	}
}

// helpers

func permModePtr(m PermissionMode) *PermissionMode { return &m }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	// Suppress unused import warning
	_ = fmt.Sprintf
}
