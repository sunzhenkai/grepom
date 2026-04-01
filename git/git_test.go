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

// --- buildAuthStrategies tests (SSH-priority chain) ---

func TestBuildAuthStrategies_GroupRepoSSHFirst(t *testing.T) {
	opts := CloneOptions{
		Token:          "group-token",
		Provider:       "github",
		SSHKey:         "/path/to/group-key",
		HasGroupToken:  true,
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	if len(strategies) < 2 {
		t.Fatalf("expected at least 2 strategies, got %d", len(strategies))
	}
	// First strategy should be group/repo SSH key
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("expected first strategy to be group/repo SSH key, got: %s", strategies[0].label)
	}
	if strategies[0].sshKey != "/path/to/group-key" {
		t.Errorf("expected SSH key in first strategy, got sshKey=%s", strategies[0].sshKey)
	}
	// Second should be group/repo token
	if !strings.Contains(strategies[1].label, "token") || !strings.Contains(strategies[1].label, "group/repo") {
		t.Errorf("expected second strategy to be group/repo token, got: %s", strategies[1].label)
	}
}

func TestBuildAuthStrategies_ResourceSSHPriorityOverToken(t *testing.T) {
	opts := CloneOptions{
		Token:          "res-token",
		Provider:       "gitlab",
		SSHKey:         "/path/to/res-key",
		HasGroupToken:  false,
		HasGroupSSHKey: false,
	}
	strategies := buildAuthStrategies("git@gitlab.com:org/repo.git", "https://gitlab.com/org/repo.git", opts)

	// First should be resource SSH key (not token)
	if len(strategies) < 2 {
		t.Fatalf("expected at least 2 strategies, got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "resource") {
		t.Errorf("expected first strategy to be resource SSH key, got: %s", strategies[0].label)
	}
	// Second should be resource token
	if !strings.Contains(strategies[1].label, "token") || !strings.Contains(strategies[1].label, "resource") {
		t.Errorf("expected second strategy to be resource token, got: %s", strategies[1].label)
	}
}

func TestBuildAuthStrategies_GroupRepoAuthBeforeResource(t *testing.T) {
	// When group has SSH key only, and resource has both token and SSH key
	opts := CloneOptions{
		Token:          "res-token", // from resource (not overridden)
		Provider:       "github",
		SSHKey:         "/path/to/group-key", // from group (overridden)
		HasGroupToken:  false,
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// 1. group/repo SSH key
	// 2. resource token (not resource SSH since HasGroupSSHKey=true means SSH is from group)
	// 3. default SSH
	// 4. HTTP
	if len(strategies) < 2 {
		t.Fatalf("expected at least 2 strategies, got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("expected first strategy to be group/repo SSH key, got: %s", strategies[0].label)
	}
	// Since HasGroupToken=false, second should be resource token (NOT resource SSH, because HasGroupSSHKey=true)
	if !strings.Contains(strategies[1].label, "token") || !strings.Contains(strategies[1].label, "resource") {
		t.Errorf("expected second strategy to be resource token, got: %s", strategies[1].label)
	}
}

func TestBuildAuthStrategies_NoAuth(t *testing.T) {
	opts := CloneOptions{Provider: "github"}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	if len(strategies) != 2 {
		t.Fatalf("expected 2 strategies (default SSH + HTTP), got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH 认证 (默认)") {
		t.Errorf("expected first to be default SSH, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[1].label, "HTTP") {
		t.Errorf("expected second to be HTTP, got: %s", strategies[1].label)
	}
}

func TestBuildAuthStrategies_OnlyToken(t *testing.T) {
	opts := CloneOptions{
		Token:         "mytoken",
		Provider:      "gitlab",
		HasGroupToken: false,
	}
	strategies := buildAuthStrategies("git@gitlab.com:org/repo.git", "https://gitlab.com/org/repo.git", opts)

	// Should have: resource token → default SSH → HTTP
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "token") || !strings.Contains(strategies[0].label, "resource") {
		t.Errorf("expected first to be resource token, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[0].url, "oauth2:mytoken@") {
		t.Errorf("expected token in URL, got: %s", strategies[0].url)
	}
}

func TestBuildAuthStrategies_OnlySSHKey(t *testing.T) {
	opts := CloneOptions{
		SSHKey:         "/path/to/key",
		Provider:       "github",
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// Should have: group/repo SSH key → default SSH → HTTP
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("expected first to be group/repo SSH key, got: %s", strategies[0].label)
	}
	if strategies[0].sshKey != "/path/to/key" {
		t.Errorf("expected ssh key in strategy, got: %s", strategies[0].sshKey)
	}
}

func TestBuildAuthStrategies_Full6LevelChain(t *testing.T) {
	// Group has both SSH key and token, and resource also has both
	// After resolver merge: SSHKey=group key, Token=group token (overrides)
	// But we track both levels. In practice, resolver only sets the merged values.
	// For the 6-level test, group/repo overrides both:
	opts := CloneOptions{
		Token:          "group-token",
		Provider:       "github",
		SSHKey:         "/path/to/group-key",
		HasGroupToken:  true,
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// Should have: group/repo SSH → group/repo token → default SSH → HTTP = 4
	if len(strategies) != 4 {
		t.Fatalf("expected 4 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	if !strings.Contains(strategies[0].label, "group/repo") || !strings.Contains(strategies[0].label, "SSH") {
		t.Errorf("strategy 0: expected group/repo SSH, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[1].label, "group/repo") || !strings.Contains(strategies[1].label, "token") {
		t.Errorf("strategy 1: expected group/repo token, got: %s", strategies[1].label)
	}
	if !strings.Contains(strategies[2].label, "默认") {
		t.Errorf("strategy 2: expected default SSH, got: %s", strategies[2].label)
	}
	if !strings.Contains(strategies[3].label, "HTTP") {
		t.Errorf("strategy 3: expected HTTP, got: %s", strategies[3].label)
	}
}

func strategyLabels(strategies []authStrategy) []string {
	labels := make([]string, len(strategies))
	for i, s := range strategies {
		labels[i] = s.label
	}
	return labels
}
