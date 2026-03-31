package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsCloned_True(t *testing.T) {
	dir := t.TempDir()
	gitDir := filepath.Join(dir, ".git")
	os.Mkdir(gitDir, 0755)

	if !IsCloned(dir) {
		t.Error("expected IsCloned to return true")
	}
}

func TestIsCloned_False(t *testing.T) {
	dir := t.TempDir()

	if IsCloned(dir) {
		t.Error("expected IsCloned to return false")
	}
}

func TestIsCloned_Nonexistent(t *testing.T) {
	if IsCloned("/nonexistent/path") {
		t.Error("expected IsCloned to return false for nonexistent path")
	}
}

func TestParseStatus_CleanOnMain(t *testing.T) {
	output := `# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0`
	status := parseStatus(output)

	if status.Branch != "main" {
		t.Errorf("expected branch main, got %s", status.Branch)
	}
	if !status.Clean {
		t.Error("expected clean")
	}
	if status.Ahead != 0 || status.Behind != 0 {
		t.Error("expected no ahead/behind")
	}
}

func TestParseStatus_Dirty(t *testing.T) {
	output := `# branch.head main
# branch.upstream origin/main
# branch.ab +0 -0
1 .M N... 100644 100644 100644 abcd1234 abcd5678 main.go`
	status := parseStatus(output)

	if status.Clean {
		t.Error("expected dirty")
	}
	if status.Dirty != 1 {
		t.Errorf("expected 1 dirty file, got %d", status.Dirty)
	}
}

func TestParseStatus_AheadBehind(t *testing.T) {
	output := `# branch.head main
# branch.upstream origin/main
# branch.ab +3 -2`
	status := parseStatus(output)

	if status.Ahead != 3 {
		t.Errorf("expected ahead 3, got %d", status.Ahead)
	}
	if status.Behind != 2 {
		t.Errorf("expected behind 2, got %d", status.Behind)
	}
}

func TestGetStatus_RealRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}

	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "test")
	run("commit", "--allow-empty", "-m", "init")

	status := GetStatus(dir)

	if !status.Cloned {
		t.Error("expected cloned=true")
	}
	if status.Branch != "main" && status.Branch != "master" {
		t.Errorf("expected main/master, got %s", status.Branch)
	}
	if !status.Clean {
		t.Error("expected clean for fresh repo")
	}
}

func TestGetStatus_NotCloned(t *testing.T) {
	status := GetStatus("/nonexistent")
	if status.Cloned {
		t.Error("expected not cloned")
	}
}
