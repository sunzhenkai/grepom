package git

import (
	"fmt"
	"io"
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

// CloneOptions holds authentication options for cloning.
type CloneOptions struct {
	Token    string // token for HTTPS authentication (merged value)
	Provider string // "github" or "gitlab", determines token URL format
	SSHKey   string // path to SSH private key file (merged value)

	// Source tracking for 6-level auth priority chain
	HasGroupToken  bool // true if token was set at group/repo level
	HasGroupSSHKey bool // true if ssh_key was set at group/repo level

	// PreferHTTPS puts token/anonymous HTTPS before SSH attempts.
	PreferHTTPS bool

	// LogWriter controls where clone attempt logs are written.
	// When nil, logs are written to stdout via fmt.Printf (backward compatible).
	// When set (e.g. &bytes.Buffer{}), logs are written there for collection.
	LogWriter io.Writer
}

// authStrategy represents a single clone authentication method to try.
type authStrategy struct {
	label  string
	url    string
	sshKey string // non-empty means use GIT_SSH_COMMAND
}

// defaultSSHIdentityFiles are tried after default SSH when no ssh_key is configured.
var defaultSSHIdentityFiles = []string{
	"~/.ssh/id_ed25519",
	"~/.ssh/id_rsa",
	"~/.ssh/id_ecdsa",
	"~/.ssh/id_ed25519_sk",
}

// buildAuthStrategies builds an ordered list of clone strategies.
// PreferHTTPS (absolute http(s) URL) tries token/anonymous HTTPS before SSH.
// Otherwise (relative path / absolute SSH) keeps SSH-first ordering.
func buildAuthStrategies(sshURL, httpURL string, opts CloneOptions) []authStrategy {
	if opts.PreferHTTPS {
		return buildHTTPSPreferredStrategies(sshURL, httpURL, opts)
	}
	return buildSSHPreferredStrategies(sshURL, httpURL, opts)
}

func buildSSHPreferredStrategies(sshURL, httpURL string, opts CloneOptions) []authStrategy {
	var strategies []authStrategy

	// 1. group/repo SSH key (highest priority)
	if opts.HasGroupSSHKey && opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH key auth (group/repo)",
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}

	// 2. group/repo token
	if opts.HasGroupToken && opts.Token != "" && httpURL != "" {
		tokenURL := buildTokenURL(httpURL, opts.Token, opts.Provider)
		strategies = append(strategies, authStrategy{
			label:  "token auth (group/repo)",
			url:    tokenURL,
			sshKey: "",
		})
	}

	// 3. resource SSH key
	if !opts.HasGroupSSHKey && opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH key auth (resource)",
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}

	// 4. Default SSH (derived URL, uses system SSH agent/config)
	if sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH auth (default)",
			url:    sshURL,
			sshKey: "",
		})
	}

	// 5. Default identity files (only when no ssh_key configured)
	if opts.SSHKey == "" && sshURL != "" {
		strategies = append(strategies, defaultIdentityStrategies(sshURL)...)
	}

	// 6. resource token
	if !opts.HasGroupToken && opts.Token != "" && httpURL != "" {
		tokenURL := buildTokenURL(httpURL, opts.Token, opts.Provider)
		strategies = append(strategies, authStrategy{
			label:  "token auth (resource)",
			url:    tokenURL,
			sshKey: "",
		})
	}

	return strategies
}

func buildHTTPSPreferredStrategies(sshURL, httpURL string, opts CloneOptions) []authStrategy {
	var strategies []authStrategy

	// Token HTTPS before anonymous (group/repo or resource).
	if opts.Token != "" && httpURL != "" {
		label := "token auth (resource)"
		if opts.HasGroupToken {
			label = "token auth (group/repo)"
		}
		tokenURL := buildTokenURL(httpURL, opts.Token, opts.Provider)
		strategies = append(strategies, authStrategy{
			label:  label,
			url:    tokenURL,
			sshKey: "",
		})
	}

	// Anonymous HTTPS (no userinfo)
	if httpURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "HTTPS auth (anonymous)",
			url:    httpURL,
			sshKey: "",
		})
	}

	// SSH fallbacks: group/repo key → resource key → default → default identity files
	if opts.HasGroupSSHKey && opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH key auth (group/repo)",
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}
	if !opts.HasGroupSSHKey && opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH key auth (resource)",
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}
	if sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH auth (default)",
			url:    sshURL,
			sshKey: "",
		})
	}
	if opts.SSHKey == "" && sshURL != "" {
		strategies = append(strategies, defaultIdentityStrategies(sshURL)...)
	}

	return strategies
}

// identityFileStat checks whether a default SSH identity file exists.
// Tests may override this to control probe behavior.
var identityFileStat = func(path string) error {
	_, err := os.Stat(path)
	return err
}

func defaultIdentityStrategies(sshURL string) []authStrategy {
	var strategies []authStrategy
	for _, keyPath := range defaultSSHIdentityFiles {
		expanded := expandTilde(keyPath)
		if err := identityFileStat(expanded); err != nil {
			continue
		}
		strategies = append(strategies, authStrategy{
			label:  "SSH key auth (default identity: " + filepath.Base(expanded) + ")",
			url:    sshURL,
			sshKey: keyPath,
		})
	}
	return strategies
}

// Clone clones a repository, creating parent directories as needed.
// It tries authentication methods in priority order (SSH-first or HTTPS-first).
func Clone(path, sshURL, httpURL string, opts CloneOptions) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	strategies := buildAuthStrategies(sshURL, httpURL, opts)

	if len(strategies) == 0 {
		return fmt.Errorf("no clone URL available")
	}

	w := opts.LogWriter
	total := len(strategies)
	var failureReasons []string
	for i, s := range strategies {
		displayURL := maskTokenURL(s.url)
		logf(w, "  [%d/%d] trying %s... %s\n", i+1, total, s.label, displayURL)

		err := tryClone(path, s.url, s.sshKey)
		if err == nil {
			logf(w, "  [%d/%d] ok\n", i+1, total)
			return nil
		}

		// Sanitize error message (remove any embedded token/URL)
		errMsg := sanitizeError(err.Error())
		logf(w, "  [%d/%d] failed: %s\n", i+1, total, errMsg)
		failureReasons = append(failureReasons, summarizeFailure(s.label, errMsg))

		// Clean up failed attempt
		os.RemoveAll(path)
		os.MkdirAll(filepath.Dir(path), 0755)
	}

	if len(failureReasons) == 0 {
		return fmt.Errorf("all authentication methods failed")
	}
	return fmt.Errorf("all authentication methods failed: %s", strings.Join(failureReasons, "; "))
}

// summarizeFailure keeps a short, single-line reason for the final error summary.
func summarizeFailure(label, errMsg string) string {
	msg := strings.TrimSpace(errMsg)
	if msg == "" {
		return label
	}
	// Prefer last non-empty line (git often prints details on the last stderr line).
	lines := strings.Split(msg, "\n")
	last := ""
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			last = line
		}
	}
	if last == "" {
		last = msg
	}
	if len(last) > 160 {
		last = last[:160] + "..."
	}
	return label + ": " + last
}

// logf writes to w if non-nil, otherwise to stdout via fmt.Printf.
func logf(w io.Writer, format string, args ...interface{}) {
	if w != nil {
		fmt.Fprintf(w, format, args...)
	} else {
		fmt.Printf(format, args...)
	}
}

// tryClone attempts a single git clone with the given URL and optional SSH key.
func tryClone(path, url, sshKey string) error {
	cmd := exec.Command("git", "clone", url, path)
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if sshKey != "" {
		expandedKey := expandTilde(sshKey)
		cmd.Env = append(os.Environ(),
			"GIT_SSH_COMMAND=ssh -i "+expandedKey+" -o IdentitiesOnly=yes -o StrictHostKeyChecking=no",
		)
	}

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("git clone: %s", sanitizeError(msg))
		}
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// buildTokenURL constructs a token-authenticated HTTPS URL.
// GitHub: https://x-access-token:<token>@<host>/<path>.git
// GitLab: https://oauth2:<token>@<host>/<path>.git
func buildTokenURL(httpURL, token, provider string) string {
	if token == "" || httpURL == "" {
		return httpURL
	}

	var user string
	switch provider {
	case "github":
		user = "x-access-token"
	case "gitlab":
		user = "oauth2"
	default:
		user = "token"
	}

	// Parse: https://<host>/<path>
	if !strings.HasPrefix(httpURL, "https://") {
		return httpURL
	}
	rest := strings.TrimPrefix(httpURL, "https://")
	return "https://" + user + ":" + token + "@" + rest
}

// sanitizeError removes sensitive information (tokens, URLs) from error messages.
func sanitizeError(msg string) string {
	var result strings.Builder
	remaining := msg
	for {
		idx := -1
		scheme := ""
		for _, s := range []string{"https://", "http://"} {
			if i := strings.Index(remaining, s); i >= 0 && (idx < 0 || i < idx) {
				idx = i
				scheme = s
			}
		}
		if idx < 0 {
			result.WriteString(remaining)
			break
		}
		result.WriteString(remaining[:idx])
		rest := remaining[idx:]
		end := strings.IndexAny(rest, " \n\t")
		if end < 0 {
			end = len(rest)
		}
		urlPart := rest[:end]
		if at := strings.Index(urlPart, "@"); at > len(scheme) {
			cred := urlPart[len(scheme):at]
			if colon := strings.Index(cred, ":"); colon >= 0 {
				result.WriteString(scheme + cred[:colon] + ":***@" + urlPart[at+1:])
			} else {
				result.WriteString("<url redacted>")
			}
		} else {
			result.WriteString("<url redacted>")
		}
		remaining = rest[end:]
	}
	return result.String()
}

// maskTokenURL 将 token URL 中的 token 部分替换为 ***。
// 例如 "https://oauth2:secret@gitlab.com/path.git" → "https://oauth2:***@gitlab.com/path.git"
func maskTokenURL(url string) string {
	// 匹配 https://<user>:<token>@ 格式
	if idx := strings.Index(url, "://"); idx >= 0 {
		rest := url[idx+3:]
		if atIdx := strings.Index(rest, "@"); atIdx >= 0 {
			// 找到 user:token 部分
			credPart := rest[:atIdx]
			if colonIdx := strings.Index(credPart, ":"); colonIdx >= 0 {
				user := credPart[:colonIdx]
				return url[:idx+3] + user + ":***@" + rest[atIdx+1:]
			}
		}
	}
	return url
}

// expandTilde expands ~/ to theuser's home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// Pull runs git pull in the given directory.
func Pull(path string) error {
	cmd := exec.Command("git", "-C", path, "pull")
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("git pull: %s", sanitizeError(msg))
		}
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

// Push runs git push in the given directory with optional extra arguments.
// The extra arguments are passed directly to git push (e.g. "--", "origin", "main").
func Push(path string, args ...string) error {
	pushArgs := append([]string{"-C", path, "push"}, args...)
	cmd := exec.Command("git", pushArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

// GetCurrentBranch returns the name of the current branch (or detached HEAD commit).
// Uses `git rev-parse --abbrev-ref HEAD`.
func GetCurrentBranch(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetRemoteURL returns the URL of the specified remote (e.g. "origin").
func GetRemoteURL(path string, remote string) (string, error) {
	cmd := exec.Command("git", "-C", path, "remote", "get-url", remote)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote %s URL: %w", remote, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// HasUnpushedCommits checks whether the given branch has commits not yet pushed to origin.
// Returns (true, N, nil) if there are N unpushed commits.
// Returns (false, 0, nil) if everything is pushed.
func HasUnpushedCommits(path string, branch string) (bool, int, error) {
	// Try to count commits between origin/{branch} and HEAD
	cmd := exec.Command("git", "-C", path, "log", "origin/"+branch+"..HEAD", "--oneline")
	out, err := cmd.Output()
	if err != nil {
		// origin/{branch} may not exist (new branch never pushed)
		// Fall back: count all commits on this branch vs empty tree
		cmd2 := exec.Command("git", "-C", path, "rev-list", "--count", "HEAD")
		out2, err2 := cmd2.Output()
		if err2 != nil {
			return false, 0, fmt.Errorf("failed to check unpushed commits: %w", err)
		}
		count := strings.TrimSpace(string(out2))
		n := 0
		fmt.Sscanf(count, "%d", &n)
		if n > 0 {
			return true, n, nil
		}
		return false, 0, nil
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return false, 0, nil
	}
	return true, len(lines), nil
}

// GetHeadCommitMessage returns the full commit message of HEAD.
// Uses `git log -1 --format=%B`.
func GetHeadCommitMessage(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "log", "-1", "--format=%B")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit message: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// PushBranch pushes a specific branch to origin.
func PushBranch(path string, branch string) error {
	cmd := exec.Command("git", "-C", path, "push", "-u", "origin", branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push origin %s: %w", branch, err)
	}
	return nil
}

// GetDefaultBranch returns the short name of origin's default branch.
// It prefers querying the remote via ls-remote (accurate even when local origin/HEAD is stale),
// then falls back to the local origin/HEAD symbolic ref.
func GetDefaultBranch(path string) (string, error) {
	if branch, err := defaultBranchFromRemote(path); err == nil {
		return branch, nil
	}
	return defaultBranchFromLocalOriginHEAD(path)
}

func defaultBranchFromRemote(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "ls-remote", "--symref", "origin", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if branch, ok := parseRemoteDefaultBranch(string(out)); ok {
		return branch, nil
	}
	return "", fmt.Errorf("no default branch in ls-remote output")
}

func parseRemoteDefaultBranch(lsRemoteOut string) (string, bool) {
	for _, line := range strings.Split(lsRemoteOut, "\n") {
		line = strings.TrimSpace(line)
		const prefix = "ref: refs/heads/"
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		ref := strings.TrimPrefix(line, prefix)
		if idx := strings.IndexAny(ref, " \t"); idx >= 0 {
			ref = ref[:idx]
		}
		if ref != "" {
			return ref, true
		}
	}
	return "", false
}

func defaultBranchFromLocalOriginHEAD(path string) (string, error) {
	cmd := exec.Command("git", "-C", path, "symbolic-ref", "refs/remotes/origin/HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to detect default branch: %w", err)
	}

	// Output is like "refs/remotes/origin/main\n", extract "main"
	ref := strings.TrimSpace(string(out))
	const prefix = "refs/remotes/origin/"
	if !strings.HasPrefix(ref, prefix) {
		return "", fmt.Errorf("unexpected symbolic-ref format: %s", ref)
	}
	return strings.TrimPrefix(ref, prefix), nil
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
