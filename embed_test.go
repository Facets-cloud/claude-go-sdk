package claudeagent

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCLIPath_CustomPath(t *testing.T) {
	// Create a temporary executable file to use as custom path.
	dir := t.TempDir()
	fakeCLI := filepath.Join(dir, "my-claude")
	if runtime.GOOS == "windows" {
		fakeCLI += ".exe"
	}
	if err := os.WriteFile(fakeCLI, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	path, err := CLIPath(&fakeCLI)
	if err != nil {
		t.Fatalf("CLIPath returned error: %v", err)
	}
	if path != fakeCLI {
		t.Errorf("expected %q, got %q", fakeCLI, path)
	}
}

func TestCLIPath_CustomPathNotFound(t *testing.T) {
	bad := "/nonexistent/path/to/claude"
	_, err := CLIPath(&bad)
	if err == nil {
		t.Fatal("expected error for nonexistent custom path")
	}
}

func TestCLIPath_BundledCLI(t *testing.T) {
	// Create a fake bundled CLI next to a fake executable.
	dir := t.TempDir()
	bundled := filepath.Join(dir, "claude")
	if runtime.GOOS == "windows" {
		bundled += ".exe"
	}
	if err := os.WriteFile(bundled, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	path, err := cliPathWithExeDir(nil, dir)
	if err != nil {
		t.Fatalf("cliPathWithExeDir returned error: %v", err)
	}
	if path != bundled {
		t.Errorf("expected bundled path %q, got %q", bundled, path)
	}
}

func TestCLIPath_FallbackToPATH(t *testing.T) {
	// Only run this test if "claude" is actually on PATH.
	which, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude not on PATH, skipping fallback test")
	}

	// Use an empty dir as exeDir so bundled lookup fails.
	emptyDir := t.TempDir()
	path, err := cliPathWithExeDir(nil, emptyDir)
	if err != nil {
		t.Fatalf("cliPathWithExeDir returned error: %v", err)
	}
	if path != which {
		t.Errorf("expected PATH result %q, got %q", which, path)
	}
}

func TestCLIPath_NothingFound(t *testing.T) {
	// Override PATH to empty so LookPath fails, and use empty exeDir.
	emptyDir := t.TempDir()
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", "")
	defer os.Setenv("PATH", origPath)

	_, err := cliPathWithExeDir(nil, emptyDir)
	if err == nil {
		t.Fatal("expected error when no CLI found anywhere")
	}
}
