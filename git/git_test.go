package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// --- buildTokenURL tests ---

func TestBuildTokenURL_GitHub(t *testing.T) {
	result := buildTokenURL("https://github.com/org/repo.git", "ghp_abc123", "github")
	expected := "https://x-access-token:ghp_abc123@github.com/org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestBuildTokenURL_GitLab(t *testing.T) {
	result := buildTokenURL("https://gitlab.com/org/repo.git", "glpat-xyz", "gitlab")
	expected := "https://oauth2:glpat-xyz@gitlab.com/org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestBuildTokenURL_DefaultProvider(t *testing.T) {
	result := buildTokenURL("https://example.com/org/repo.git", "mytoken", "bitbucket")
	expected := "https://token:mytoken@example.com/org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestBuildTokenURL_EmptyToken(t *testing.T) {
	result := buildTokenURL("https://github.com/org/repo.git", "", "github")
	if result != "https://github.com/org/repo.git" {
		t.Errorf("empty token should return original URL, got %s", result)
	}
}

func TestBuildTokenURL_EmptyURL(t *testing.T) {
	result := buildTokenURL("", "mytoken", "github")
	if result != "" {
		t.Errorf("empty URL should return empty, got %s", result)
	}
}

func TestBuildTokenURL_NonHTTPS(t *testing.T) {
	result := buildTokenURL("http://github.com/org/repo.git", "mytoken", "github")
	// Non-https URL should be returned as-is
	if result != "http://github.com/org/repo.git" {
		t.Errorf("non-https URL should be returned as-is, got %s", result)
	}
}

// --- sanitizeError tests ---

func TestSanitizeError_RemovesURL(t *testing.T) {
	msg := "fatal: repository 'https://x-access-token:secret@github.com/org/repo.git/' not found"
	result := sanitizeError(msg)
	if strings.Contains(result, "secret") {
		t.Error("sanitized error should not contain token")
	}
	if strings.Contains(result, "x-access-token") {
		t.Error("sanitized error should not contain username")
	}
	if !strings.Contains(result, "<url redacted>") {
		t.Error("sanitized error should contain <url redacted>")
	}
}

func TestSanitizeError_NoURL(t *testing.T) {
	msg := "connection refused"
	result := sanitizeError(msg)
	if result != msg {
		t.Errorf("non-URL error should be unchanged, got %s", result)
	}
}

// --- expandTilde tests ---

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}
	result := expandTilde("~/test/path")
	expected := filepath.Join(home, "test/path")
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestExpandTilde_NoTilde(t *testing.T) {
	result := expandTilde("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("absolute path should be unchanged, got %s", result)
	}
}

func TestExpandTilde_JustTilde(t *testing.T) {
	result := expandTilde("~")
	if result != "~" {
		t.Errorf("bare tilde should be unchanged, got %s", result)
	}
}
