package tools

import (
	"encoding/json"
	"testing"
)

func TestTodoWriteInput_JSON(t *testing.T) {
	input := TodoWriteInput{
		Todos: []TodoItem{
			{Content: "Write tests", Status: "in_progress", ActiveForm: "Writing tests"},
			{Content: "Deploy", Status: "pending", ActiveForm: "Deploying"},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got TodoWriteInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Todos) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(got.Todos))
	}
	if got.Todos[0].Status != "in_progress" {
		t.Errorf("Status = %q", got.Todos[0].Status)
	}
}

func TestTodoWriteOutput_JSON(t *testing.T) {
	raw := `{
		"oldTodos": [{"content": "task", "status": "pending", "activeForm": "Working"}],
		"newTodos": [{"content": "task", "status": "completed", "activeForm": "Working"}]
	}`
	var out TodoWriteOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.OldTodos[0].Status != "pending" {
		t.Errorf("OldTodos status = %q", out.OldTodos[0].Status)
	}
	if out.NewTodos[0].Status != "completed" {
		t.Errorf("NewTodos status = %q", out.NewTodos[0].Status)
	}
}

func TestAskUserQuestionInput_JSON(t *testing.T) {
	input := AskUserQuestionInput{
		Questions: []AskQuestion{
			{
				Question:    "Which framework?",
				Header:      "Framework",
				MultiSelect: false,
				Options: []AskQuestionOption{
					{Label: "React", Description: "Popular UI library"},
					{Label: "Vue", Description: "Progressive framework"},
				},
			},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got AskUserQuestionInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Questions) != 1 {
		t.Fatalf("expected 1 question, got %d", len(got.Questions))
	}
	if got.Questions[0].Header != "Framework" {
		t.Errorf("Header = %q", got.Questions[0].Header)
	}
	if len(got.Questions[0].Options) != 2 {
		t.Errorf("Options len = %d", len(got.Questions[0].Options))
	}
}

func TestAskUserQuestionOutput_JSON(t *testing.T) {
	raw := `{
		"questions": [{"question": "Which?", "header": "Pick", "options": [], "multiSelect": false}],
		"answers": {"Which?": "Option A"},
		"annotations": {"Which?": {"preview": "some preview", "notes": "user note"}}
	}`
	var out AskUserQuestionOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Answers["Which?"] != "Option A" {
		t.Errorf("Answer = %q", out.Answers["Which?"])
	}
	if out.Annotations["Which?"].Preview == nil || *out.Annotations["Which?"].Preview != "some preview" {
		t.Error("expected annotation preview")
	}
}

func TestConfigInput_JSON(t *testing.T) {
	input := ConfigInput{
		Setting: "theme",
		Value:   "dark",
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got ConfigInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Setting != "theme" {
		t.Errorf("Setting = %q", got.Setting)
	}
}

func TestConfigOutput_JSON(t *testing.T) {
	raw := `{"success": true, "operation": "set", "setting": "theme", "previousValue": "light", "newValue": "dark"}`
	var out ConfigOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Success {
		t.Error("expected Success=true")
	}
	if out.Operation == nil || *out.Operation != "set" {
		t.Errorf("Operation = %v", out.Operation)
	}
}

func TestEnterWorktreeInput_JSON(t *testing.T) {
	input := EnterWorktreeInput{Name: strPtr("feature-branch")}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got EnterWorktreeInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name == nil || *got.Name != "feature-branch" {
		t.Errorf("Name = %v", got.Name)
	}
}

func TestExitWorktreeOutput_JSON(t *testing.T) {
	raw := `{
		"action": "remove",
		"originalCwd": "/home/user/project",
		"worktreePath": "/tmp/wt-123",
		"discardedFiles": 3,
		"message": "Worktree removed"
	}`
	var out ExitWorktreeOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Action != "remove" {
		t.Errorf("Action = %q", out.Action)
	}
	if out.DiscardedFiles == nil || *out.DiscardedFiles != 3 {
		t.Errorf("DiscardedFiles = %v", out.DiscardedFiles)
	}
}

func TestExitPlanModeInput_JSON(t *testing.T) {
	input := ExitPlanModeInput{
		AllowedPrompts: []AllowedPrompt{
			{Tool: "Bash", Prompt: "run tests"},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got ExitPlanModeInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.AllowedPrompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(got.AllowedPrompts))
	}
	if got.AllowedPrompts[0].Tool != "Bash" {
		t.Errorf("Tool = %q", got.AllowedPrompts[0].Tool)
	}
}

func TestExitPlanModeOutput_JSON(t *testing.T) {
	plan := "Step 1: Do X\nStep 2: Do Y"
	raw := `{"plan": "Step 1: Do X\nStep 2: Do Y", "isAgent": true, "filePath": "/tmp/plan.md"}`
	var out ExitPlanModeOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Plan == nil || *out.Plan != plan {
		t.Errorf("Plan = %v", out.Plan)
	}
	if !out.IsAgent {
		t.Error("expected IsAgent=true")
	}
}

func TestTaskOutputInput_JSON(t *testing.T) {
	input := TaskOutputInput{
		TaskID:  "task-42",
		Block:   true,
		Timeout: 30000,
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got TaskOutputInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.TaskID != "task-42" {
		t.Errorf("TaskID = %q", got.TaskID)
	}
	if !got.Block {
		t.Error("expected Block=true")
	}
}

func TestTaskStopOutput_JSON(t *testing.T) {
	raw := `{"message": "Stopped", "task_id": "t-1", "task_type": "bash", "command": "sleep 100"}`
	var out TaskStopOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.TaskID != "t-1" {
		t.Errorf("TaskID = %q", out.TaskID)
	}
	if out.Command == nil || *out.Command != "sleep 100" {
		t.Errorf("Command = %v", out.Command)
	}
}

func TestNotebookEditInput_JSON(t *testing.T) {
	input := NotebookEditInput{
		NotebookPath: "/tmp/nb.ipynb",
		NewSource:    "print('hello')",
		CellType:     strPtr("code"),
		EditMode:     strPtr("replace"),
	}
	data, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var got NotebookEditInput
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.NotebookPath != "/tmp/nb.ipynb" {
		t.Errorf("NotebookPath = %q", got.NotebookPath)
	}
}

func TestNotebookEditOutput_JSON(t *testing.T) {
	raw := `{
		"new_source": "print('world')",
		"cell_type": "code",
		"language": "python",
		"edit_mode": "replace",
		"notebook_path": "/tmp/nb.ipynb",
		"original_file": "{}",
		"updated_file": "{}"
	}`
	var out NotebookEditOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.CellType != "code" {
		t.Errorf("CellType = %q", out.CellType)
	}
	if out.Language != "python" {
		t.Errorf("Language = %q", out.Language)
	}
}
