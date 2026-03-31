package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Status struct {
	Branch   string
	Clean    bool
	Ahead    int
	Behind   int
	Dirty    int
	Cloned   bool
	NotARepo bool
}

// IsCloned checks if a directory exists and contains a .git subdirectory.
func IsCloned(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// Clone clones a repository, creating parent directories as needed.
// It tries SSH URL first, then falls back to HTTP.
func Clone(path, sshURL, httpURL string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	url := sshURL
	if url == "" {
		url = httpURL
	}

	cmd := exec.Command("git", "clone", url, path)
	cmd.Stderr = &strings.Builder{}

	if err := cmd.Run(); err != nil {
		// If SSH fails, try HTTP
		if sshURL != "" && httpURL != "" {
			// Remove failed clone directory
			os.RemoveAll(path)
			os.MkdirAll(filepath.Dir(path), 0755)

			cmd = exec.Command("git", "clone", httpURL, path)
			cmd.Stderr = &strings.Builder{}
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}
			return nil
		}
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// Pull runs git pull in the given directory.
func Pull(path string) error {
	cmd := exec.Command("git", "-C", path, "pull")
	cmd.Stderr = &strings.Builder{}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

// GetStatus parses git status output and returns structured status info.
func GetStatus(path string) *Status {
	if !IsCloned(path) {
		return &Status{Cloned: false}
	}

	cmd := exec.Command("git", "-C", path, "status", "--porcelain=v2", "--branch")
	out, err := cmd.Output()
	if err != nil {
		// Not a git repo
		return &Status{NotARepo: true, Cloned: true}
	}

	return parseStatus(string(out))
}

func parseStatus(output string) *Status {
	status := &Status{
		Cloned: true,
		Clean:  true,
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# branch.head ") {
			status.Branch = strings.TrimPrefix(line, "# branch.head ")
		}
		if strings.HasPrefix(line, "# branch.ab ") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "+") {
					fmt.Sscanf(p, "+%d", &status.Ahead)
				}
				if strings.HasPrefix(p, "-") {
					fmt.Sscanf(p, "-%d", &status.Behind)
				}
			}
		}
		// Files with changes (not branch info lines)
		if line != "" && !strings.HasPrefix(line, "#") {
			status.Clean = false
			status.Dirty++
		}
	}

	return status
}
