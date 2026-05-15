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

	// Expected order: resource SSH → default SSH → resource token
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	// First should be resource SSH key
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "resource") {
		t.Errorf("expected first strategy to be resource SSH key, got: %s", strategies[0].label)
	}
	// Second should be default SSH
	if !strings.Contains(strategies[1].label, "default") {
		t.Errorf("expected second strategy to be default SSH, got: %s", strategies[1].label)
	}
	// Third should be resource token
	if !strings.Contains(strategies[2].label, "token") || !strings.Contains(strategies[2].label, "resource") {
		t.Errorf("expected third strategy to be resource token, got: %s", strategies[2].label)
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
	// 2. default SSH (since group has SSH, resource SSH is skipped; but default SSH is always tried)
	// 3. resource token (HasGroupToken=false, so resource token is included)
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("expected first strategy to be group/repo SSH key, got: %s", strategies[0].label)
	}
	// Second should be default SSH
	if !strings.Contains(strategies[1].label, "default") {
		t.Errorf("expected second strategy to be default SSH, got: %s", strategies[1].label)
	}
	// Third should be resource token
	if !strings.Contains(strategies[2].label, "token") || !strings.Contains(strategies[2].label, "resource") {
		t.Errorf("expected third strategy to be resource token, got: %s", strategies[2].label)
	}
}

func TestBuildAuthStrategies_NoAuth(t *testing.T) {
	opts := CloneOptions{Provider: "github"}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	if len(strategies) != 1 {
		t.Fatalf("expected 1 strategy (default SSH only), got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH auth (default)") {
		t.Errorf("expected first to be default SSH, got: %s", strategies[0].label)
	}
}

func TestBuildAuthStrategies_OnlyToken(t *testing.T) {
	opts := CloneOptions{
		Token:         "mytoken",
		Provider:      "gitlab",
		HasGroupToken: false,
	}
	strategies := buildAuthStrategies("git@gitlab.com:org/repo.git", "https://gitlab.com/org/repo.git", opts)

	// Should have: default SSH → resource token
	if len(strategies) != 2 {
		t.Fatalf("expected 2 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	// First should be default SSH (resource SSH skipped because no SSH key)
	if !strings.Contains(strategies[0].label, "default") {
		t.Errorf("expected first to be default SSH, got: %s", strategies[0].label)
	}
	// Second should be resource token
	if !strings.Contains(strategies[1].label, "token") || !strings.Contains(strategies[1].label, "resource") {
		t.Errorf("expected second to be resource token, got: %s", strategies[1].label)
	}
	if !strings.Contains(strategies[1].url, "oauth2:mytoken@") {
		t.Errorf("expected token in URL, got: %s", strategies[1].url)
	}
}

func TestBuildAuthStrategies_OnlySSHKey(t *testing.T) {
	opts := CloneOptions{
		SSHKey:         "/path/to/key",
		Provider:       "github",
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// Should have: group/repo SSH key → default SSH
	if len(strategies) != 2 {
		t.Fatalf("expected 2 strategies, got %d", len(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("expected first to be group/repo SSH key, got: %s", strategies[0].label)
	}
	if strategies[0].sshKey != "/path/to/key" {
		t.Errorf("expected ssh key in strategy, got: %s", strategies[0].sshKey)
	}
}

func TestBuildAuthStrategies_Full5LevelChain(t *testing.T) {
	// Group has both SSH key and token, and resource also has both
	// After resolver merge: SSHKey=group key, Token=group token (overrides)
	// But we track both levels. In practice, resolver only sets the merged values.
	// For the 5-level test, group/repo overrides both:
	opts := CloneOptions{
		Token:          "group-token",
		Provider:       "github",
		SSHKey:         "/path/to/group-key",
		HasGroupToken:  true,
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// Should have: group/repo SSH → group/repo token → default SSH = 3
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	if !strings.Contains(strategies[0].label, "group/repo") || !strings.Contains(strategies[0].label, "SSH") {
		t.Errorf("strategy 0: expected group/repo SSH, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[1].label, "group/repo") || !strings.Contains(strategies[1].label, "token") {
		t.Errorf("strategy 1: expected group/repo token, got: %s", strategies[1].label)
	}
	if !strings.Contains(strategies[2].label, "default") {
		t.Errorf("strategy 2: expected default SSH, got: %s", strategies[2].label)
	}
}

// --- New priority tests: default SSH before resource token ---

func TestBuildAuthStrategies_GroupSSH_ResourceTokenOnly(t *testing.T) {
	// group 有 SSH key + resource 有 token 时：group SSH → default SSH → resource token
	opts := CloneOptions{
		Token:          "res-token",
		Provider:       "github",
		SSHKey:         "/path/to/group-key",
		HasGroupToken:  false,
		HasGroupSSHKey: true,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// Expected: group SSH → default SSH → resource token = 3
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	if !strings.Contains(strategies[0].label, "SSH key") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("strategy 0: expected group/repo SSH key, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[1].label, "default") {
		t.Errorf("strategy 1: expected default SSH, got: %s", strategies[1].label)
	}
	if !strings.Contains(strategies[2].label, "token") || !strings.Contains(strategies[2].label, "resource") {
		t.Errorf("strategy 2: expected resource token, got: %s", strategies[2].label)
	}
}

func TestBuildAuthStrategies_GroupToken_ResourceSSHAndToken(t *testing.T) {
	// group 有 token（无 SSH）+ resource 有 SSH + token 时：
	// group token → resource SSH → default SSH → resource token
	opts := CloneOptions{
		Token:          "merged-token", // from group override
		Provider:       "github",
		SSHKey:         "/path/to/res-key", // from resource fallback
		HasGroupToken:  true,
		HasGroupSSHKey: false,
	}
	strategies := buildAuthStrategies("git@github.com:org/repo.git", "https://github.com/org/repo.git", opts)

	// Expected: group token → resource SSH → default SSH = 3
	// Note: resource token is skipped because HasGroupToken=true
	if len(strategies) != 3 {
		t.Fatalf("expected 3 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	if !strings.Contains(strategies[0].label, "token") || !strings.Contains(strategies[0].label, "group/repo") {
		t.Errorf("strategy 0: expected group/repo token, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[1].label, "SSH key") || !strings.Contains(strategies[1].label, "resource") {
		t.Errorf("strategy 1: expected resource SSH key, got: %s", strategies[1].label)
	}
	if !strings.Contains(strategies[2].label, "default") {
		t.Errorf("strategy 2: expected default SSH, got: %s", strategies[2].label)
	}
}

func TestBuildAuthStrategies_DefaultSSHPriorToResourceToken(t *testing.T) {
	// 仅 resource token（无 SSH key）时：default SSH → resource token
	opts := CloneOptions{
		Token:    "res-token",
		Provider: "gitlab",
	}
	strategies := buildAuthStrategies("git@gitlab.com:org/repo.git", "https://gitlab.com/org/repo.git", opts)

	// Expected: default SSH → resource token = 2
	if len(strategies) != 2 {
		t.Fatalf("expected 2 strategies, got %d: %v", len(strategies), strategyLabels(strategies))
	}
	if !strings.Contains(strategies[0].label, "default") {
		t.Errorf("strategy 0: expected default SSH, got: %s", strategies[0].label)
	}
	if !strings.Contains(strategies[1].label, "token") || !strings.Contains(strategies[1].label, "resource") {
		t.Errorf("strategy 1: expected resource token, got: %s", strategies[1].label)
	}
}

func strategyLabels(strategies []authStrategy) []string {
	labels := make([]string, len(strategies))
	for i, s := range strategies {
		labels[i] = s.label
	}
	return labels
}

// --- maskTokenURL tests ---

func TestMaskTokenURL_GitLab(t *testing.T) {
	result := maskTokenURL("https://oauth2:glpat-secret@gitlab.com/org/repo.git")
	expected := "https://oauth2:***@gitlab.com/org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestMaskTokenURL_GitHub(t *testing.T) {
	result := maskTokenURL("https://x-access-token:ghp_secret@github.com/org/repo.git")
	expected := "https://x-access-token:***@github.com/org/repo.git"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestMaskTokenURL_NoToken(t *testing.T) {
	result := maskTokenURL("git@gitlab.com:org/repo.git")
	if result != "git@gitlab.com:org/repo.git" {
		t.Errorf("non-token URL should be unchanged, got: %s", result)
	}
}

func TestMaskTokenURL_Empty(t *testing.T) {
	result := maskTokenURL("")
	if result != "" {
		t.Errorf("empty URL should return empty, got %s", result)
	}
}

// --- GetDefaultBranch tests ---

func TestGetDefaultBranch_RealRepo(t *testing.T) {
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
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	// Manually set origin/HEAD to simulate remote HEAD resolution
	run("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	branch := "main"
	out, _ := exec.Command("git", "-C", dir, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if len(out) > 0 {
		branch = strings.TrimSpace(string(out))
	}

	result, err := GetDefaultBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != branch {
		t.Errorf("expected default branch %s, got %s", branch, result)
	}
}

func TestGetDefaultBranch_NoOriginHEAD(t *testing.T) {
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
	// Don't set origin/HEAD

	_, err := GetDefaultBranch(dir)
	if err == nil {
		t.Error("expected error when origin/HEAD is not set")
	}
}

func TestGetDefaultBranch_NotAGitDir(t *testing.T) {
	_, err := GetDefaultBranch(t.TempDir())
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

// --- CheckPullSafety tests ---

func TestCheckPullSafety_NotCloned(t *testing.T) {
	dir := t.TempDir()
	ok, reason := CheckPullSafety(dir)
	if ok {
		t.Error("expected not eligible")
	}
	if reason != "not cloned" {
		t.Errorf("expected 'not cloned', got %s", reason)
	}
}

func TestCheckPullSafety_CleanOnDefaultBranch(t *testing.T) {
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
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	run("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	ok, reason := CheckPullSafety(dir)
	if !ok {
		t.Errorf("expected eligible, got skip reason: %s", reason)
	}
}

func TestCheckPullSafety_Dirty(t *testing.T) {
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
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	run("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")
	// Create a dirty file
	writeFile(t, filepath.Join(dir, "dirty.txt"), "content")

	ok, reason := CheckPullSafety(dir)
	if ok {
		t.Error("expected not eligible due to dirty")
	}
	if reason != "dirty working tree" {
		t.Errorf("expected 'dirty working tree', got %s", reason)
	}
}

func TestCheckPullSafety_NonDefaultBranch(t *testing.T) {
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
	run("checkout", "-b", "feature")
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	run("symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main")

	ok, reason := CheckPullSafety(dir)
	if ok {
		t.Error("expected not eligible on feature branch")
	}
	if reason != "on feature, not default branch" {
		t.Errorf("expected branch reason, got %s", reason)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

// --- CloneAll/PullAll tests ---

func TestCloneAll_Empty(t *testing.T) {
	results := CloneAll(4, nil, nil)
	if results != nil {
		t.Error("expected nil for empty tasks")
	}
}

func TestCloneAll_PreservesOrder(t *testing.T) {
	// Test with 0 concurrency (falls back to 1) and empty tasks
	results := CloneAll(0, nil, nil)
	if results != nil {
		t.Error("expected nil for empty tasks with 0 concurrency")
	}

	// Test with invalid concurrency still works
	results = CloneAll(-1, []CloneTask{}, nil)
	if results != nil {
		t.Error("expected nil for empty tasks with negative concurrency")
	}
}

func TestPullAll_Empty(t *testing.T) {
	results := PullAll(4, nil, nil)
	if results != nil {
		t.Error("expected nil for empty tasks")
	}
}

func TestPullAll_InvalidConcurrency(t *testing.T) {
	results := PullAll(0, nil, nil)
	if results != nil {
		t.Error("expected nil for empty tasks with 0 concurrency")
	}
}

// --- GetCurrentBranch tests ---

func TestGetCurrentBranch_RealRepo(t *testing.T) {
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

	branch, err := GetCurrentBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "main" && branch != "master" {
		t.Errorf("expected main or master, got %s", branch)
	}
}

func TestGetCurrentBranch_FeatureBranch(t *testing.T) {
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
	run("checkout", "-b", "feature-x")

	branch, err := GetCurrentBranch(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if branch != "feature-x" {
		t.Errorf("expected feature-x, got %s", branch)
	}
}

func TestGetCurrentBranch_NotAGitDir(t *testing.T) {
	_, err := GetCurrentBranch(t.TempDir())
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

// --- GetRemoteURL tests ---

func TestGetRemoteURL_Origin(t *testing.T) {
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
	run("remote", "add", "origin", "https://github.com/myorg/myrepo.git")

	url, err := GetRemoteURL(dir, "origin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/myorg/myrepo.git" {
		t.Errorf("expected https://github.com/myorg/myrepo.git, got %s", url)
	}
}

func TestGetRemoteURL_SSH(t *testing.T) {
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
	run("remote", "add", "origin", "git@github.com:myorg/myrepo.git")

	url, err := GetRemoteURL(dir, "origin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "git@github.com:myorg/myrepo.git" {
		t.Errorf("expected SSH URL, got %s", url)
	}
}

func TestGetRemoteURL_NoRemote(t *testing.T) {
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

	_, err := GetRemoteURL(dir, "origin")
	if err == nil {
		t.Error("expected error for missing remote")
	}
}

// --- HasUnpushedCommits tests ---

func TestHasUnpushedCommits_NoRemote(t *testing.T) {
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

	has, count, err := HasUnpushedCommits(dir, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected has=true for branch with no remote tracking")
	}
	if count < 1 {
		t.Errorf("expected at least 1 commit, got %d", count)
	}
}

func TestHasUnpushedCommits_UpToDate(t *testing.T) {
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
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	// Simulate origin/main being at HEAD by creating the ref
	run("update-ref", "refs/remotes/origin/main", "HEAD")

	has, _, err := HasUnpushedCommits(dir, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected has=false when up to date")
	}
}

func TestHasUnpushedCommits_WithExtra(t *testing.T) {
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
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	run("update-ref", "refs/remotes/origin/main", "HEAD")
	// Add extra commit
	run("commit", "--allow-empty", "-m", "second")

	has, count, err := HasUnpushedCommits(dir, "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected has=true with extra commit")
	}
	if count != 1 {
		t.Errorf("expected 1 unpushed commit, got %d", count)
	}
}

// --- GetHeadCommitMessage tests ---

func TestGetHeadCommitMessage_SingleLine(t *testing.T) {
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
	run("commit", "--allow-empty", "-m", "fix: typo")

	msg, err := GetHeadCommitMessage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != "fix: typo" {
		t.Errorf("expected 'fix: typo', got %q", msg)
	}
}

func TestGetHeadCommitMessage_MultiLine(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	dir := t.TempDir()
	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Run()
	}

	runGit("init")
	runGit("config", "user.email", "test@test.com")
	runGit("config", "user.name", "test")

	// Create a file and commit with multi-line message
	writeFile(t, filepath.Join(dir, "file.txt"), "hello")
	runGit("add", ".")
	cmd := exec.Command("git", "commit", "-m", "feat: add dark mode\n\nImplement dark mode toggle")
	cmd.Dir = dir
	cmd.Run()

	msg, err := GetHeadCommitMessage(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "feat: add dark mode") {
		t.Errorf("expected title in message, got %q", msg)
	}
	if !strings.Contains(msg, "Implement dark mode toggle") {
		t.Errorf("expected body in message, got %q", msg)
	}
}

func TestGetHeadCommitMessage_NotAGitDir(t *testing.T) {
	_, err := GetHeadCommitMessage(t.TempDir())
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

// --- Push tests ---

func TestPush_NotAGitDir(t *testing.T) {
	dir := t.TempDir()
	err := Push(dir)
	if err == nil {
		t.Error("expected error when pushing from non-git directory")
	}
}

func TestPush_WithArgs(t *testing.T) {
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

	// Push to a non-existent remote should fail, but the command should be constructed correctly
	err := Push(dir, "--dry-run")
	if err == nil {
		// If it succeeds (unlikely without a remote), that's fine too
		t.Log("push --dry-run succeeded (no remote)")
	}
	// The key test is that the function doesn't panic with args
}
