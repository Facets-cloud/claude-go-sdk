package tools

import (
	"encoding/json"
	"testing"
)

func TestListMcpResourcesOutput_JSON(t *testing.T) {
	raw := `[
		{"uri": "file:///tmp/data", "name": "data", "mimeType": "text/plain", "server": "my-server"},
		{"uri": "file:///tmp/config", "name": "config", "server": "my-server"}
	]`
	var out ListMcpResourcesOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(out))
	}
	if out[0].URI != "file:///tmp/data" {
		t.Errorf("URI = %q", out[0].URI)
	}
	if out[1].MimeType != nil {
		t.Errorf("MimeType should be nil, got %v", out[1].MimeType)
	}
}

func TestReadMcpResourceOutput_JSON(t *testing.T) {
	raw := `{
		"contents": [
			{"uri": "file:///data", "mimeType": "text/plain", "text": "hello"},
			{"uri": "file:///blob", "blobSavedTo": "/tmp/saved"}
		]
	}`
	var out ReadMcpResourceOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Contents) != 2 {
		t.Fatalf("expected 2 contents, got %d", len(out.Contents))
	}
	if out.Contents[0].Text == nil || *out.Contents[0].Text != "hello" {
		t.Errorf("Text = %v", out.Contents[0].Text)
	}
	if out.Contents[1].BlobSavedTo == nil || *out.Contents[1].BlobSavedTo != "/tmp/saved" {
		t.Errorf("BlobSavedTo = %v", out.Contents[1].BlobSavedTo)
	}
}

func TestSubscribePollingInput_JSON(t *testing.T) {
	input := SubscribePollingInput{
		Type:       "tool",
		Server:     "my-server",
		ToolName:   strPtr("get-data"),
		IntervalMs: 5000,
		Reason:     strPtr("monitor changes"),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got SubscribePollingInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "tool" {
		t.Errorf("Type = %q", got.Type)
	}
	if got.IntervalMs != 5000 {
		t.Errorf("IntervalMs = %d", got.IntervalMs)
	}
}

func TestSubscribeMcpResourceOutput_JSON(t *testing.T) {
	raw := `{"subscribed": true, "subscriptionId": "sub-123"}`
	var out SubscribeMcpResourceOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Subscribed {
		t.Error("expected Subscribed=true")
	}
	if out.SubscriptionID != "sub-123" {
		t.Errorf("SubscriptionID = %q", out.SubscriptionID)
	}
}
