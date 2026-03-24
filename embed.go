package claudeagent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// CLIPath returns the path to the Claude Code CLI executable.
// It checks the customPath option first, then looks for a bundled CLI
// adjacent to the running executable, then falls back to PATH lookup.
func CLIPath(customPath *string) (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		// If we can't determine our own path, skip bundled lookup.
		return cliPathWithExeDir(customPath, "")
	}
	exeDir := filepath.Dir(exePath)
	return cliPathWithExeDir(customPath, exeDir)
}

// cliPathWithExeDir is the internal implementation that accepts an explicit
// directory to search for a bundled CLI. This enables testing without
// depending on the real executable location.
func cliPathWithExeDir(customPath *string, exeDir string) (string, error) {
	// 1. Custom path takes priority.
	if customPath != nil && *customPath != "" {
		if _, err := os.Stat(*customPath); err != nil {
			return "", fmt.Errorf("custom CLI path not found: %w", err)
		}
		return *customPath, nil
	}

	// 2. Look for bundled CLI adjacent to the executable.
	if exeDir != "" {
		name := "claude"
		if runtime.GOOS == "windows" {
			name = "claude.exe"
		}
		bundled := filepath.Join(exeDir, name)
		if _, err := os.Stat(bundled); err == nil {
			return bundled, nil
		}
	}

	// 3. Fall back to PATH lookup.
	path, err := exec.LookPath("claude")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("claude CLI not found: checked custom path, bundled location, and PATH")
}
