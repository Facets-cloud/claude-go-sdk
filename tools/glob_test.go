package tools

import (
	"encoding/json"
	"testing"
)

func TestGlobInput_JSON(t *testing.T) {
	input := GlobInput{
		Pattern: "**/*.go",
		Path:    strPtr("/src"),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got GlobInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Pattern != "**/*.go" {
		t.Errorf("Pattern = %q", got.Pattern)
	}
}

func TestGlobOutput_JSON(t *testing.T) {
	raw := `{
		"durationMs": 42.5,
		"numFiles": 3,
		"filenames": ["a.go", "b.go", "c.go"],
		"truncated": false
	}`
	var out GlobOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.NumFiles != 3 {
		t.Errorf("NumFiles = %d", out.NumFiles)
	}
	if len(out.Filenames) != 3 {
		t.Errorf("Filenames len = %d", len(out.Filenames))
	}
	if out.Truncated {
		t.Error("expected Truncated=false")
	}
}
