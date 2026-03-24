package tools

import (
	"encoding/json"
	"testing"
)

func TestBashInput_JSON(t *testing.T) {
	input := BashInput{
		Command:     "ls -la",
		Timeout:     intPtr(5000),
		Description: strPtr("List files"),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got BashInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Command != "ls -la" {
		t.Errorf("Command = %q, want %q", got.Command, "ls -la")
	}
	if got.Timeout == nil || *got.Timeout != 5000 {
		t.Errorf("Timeout = %v, want 5000", got.Timeout)
	}
}

func TestBashOutput_JSON(t *testing.T) {
	raw := `{
		"stdout": "file1.go\nfile2.go\n",
		"stderr": "",
		"interrupted": false,
		"backgroundTaskId": "task-123"
	}`
	var out BashOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Stdout != "file1.go\nfile2.go\n" {
		t.Errorf("Stdout = %q", out.Stdout)
	}
	if out.Interrupted {
		t.Error("expected Interrupted=false")
	}
	if out.BackgroundTaskID == nil || *out.BackgroundTaskID != "task-123" {
		t.Errorf("BackgroundTaskID = %v", out.BackgroundTaskID)
	}
}

func TestBashOutput_AllFields(t *testing.T) {
	out := BashOutput{
		Stdout:                    "output",
		Stderr:                    "err",
		Interrupted:               true,
		IsImage:                   boolPtr(false),
		DangerouslyDisableSandbox: boolPtr(true),
		PersistedOutputPath:       strPtr("/tmp/out"),
		PersistedOutputSize:       intPtr(1024),
		TokenSaverOutput:          strPtr("compressed"),
	}
	data, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	var got BashOutput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.Interrupted {
		t.Error("expected Interrupted=true")
	}
	if got.PersistedOutputSize == nil || *got.PersistedOutputSize != 1024 {
		t.Errorf("PersistedOutputSize = %v", got.PersistedOutputSize)
	}
}
