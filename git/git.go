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

// CloneOptions holds authentication options for cloning.
type CloneOptions struct {
	Token    string // token for HTTPS authentication (merged value)
	Provider string // "github" or "gitlab", determines token URL format
	SSHKey   string // path to SSH private key file (merged value)

	// Source tracking for 6-level auth priority chain
	HasGroupToken  bool // true if token was set at group/repo level
	HasGroupSSHKey bool // true if ssh_key was set at group/repo level
}

// authStrategy represents a single clone authentication method to try.
type authStrategy struct {
	label  string
	url    string
	sshKey string // non-empty means use GIT_SSH_COMMAND
}

// buildAuthStrategies builds an ordered list of clone strategies based on the
// priority chain: group/repo SSH → group/repo token → resource SSH →
// default SSH → resource token → bare HTTP.
func buildAuthStrategies(sshURL, httpURL string, opts CloneOptions) []authStrategy {
	var strategies []authStrategy

	// 1. group/repo SSH key (highest priority)
	if opts.HasGroupSSHKey && opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH key 认证 (group/repo)",
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}

	// 2. group/repo token
	if opts.HasGroupToken && opts.Token != "" && httpURL != "" {
		tokenURL := buildTokenURL(httpURL, opts.Token, opts.Provider)
		strategies = append(strategies, authStrategy{
			label:  "token 认证 (group/repo)",
			url:    tokenURL,
			sshKey: "",
		})
	}

	// 3. resource SSH key
	if !opts.HasGroupSSHKey && opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH key 认证 (resource)",
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}

	// 4. Default SSH (derived URL, uses system SSH agent/config)
	if sshURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "SSH 认证 (默认)",
			url:    sshURL,
			sshKey: "",
		})
	}

	// 5. resource token
	if !opts.HasGroupToken && opts.Token != "" && httpURL != "" {
		tokenURL := buildTokenURL(httpURL, opts.Token, opts.Provider)
		strategies = append(strategies, authStrategy{
			label:  "token 认证 (resource)",
			url:    tokenURL,
			sshKey: "",
		})
	}

	// 6. Bare HTTP
	if httpURL != "" {
		strategies = append(strategies, authStrategy{
			label:  "HTTP 克隆",
			url:    httpURL,
			sshKey: "",
		})
	}

	return strategies
}

// Clone clones a repository, creating parent directories as needed.
// It tries authentication methods in priority order using the 6-level chain:
// group/repo SSH → group/repo token → resource SSH → default SSH → resource token → HTTP.
func Clone(path, sshURL, httpURL string, opts CloneOptions) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	strategies := buildAuthStrategies(sshURL, httpURL, opts)

	if len(strategies) == 0 {
		return fmt.Errorf("no clone URL available")
	}

	total := len(strategies)
	for i, s := range strategies {
		fmt.Printf("  [%d/%d] 尝试 %s...\n", i+1, total, s.label)

		err := tryClone(path, s.url, s.sshKey)
		if err == nil {
			fmt.Printf("  [%d/%d] 成功\n", i+1, total)
			return nil
		}

		// Sanitize error message (remove any embedded token/URL)
		errMsg := sanitizeError(err.Error())
		fmt.Printf("  [%d/%d] 失败: %s\n", i+1, total, errMsg)

		// Clean up failed attempt
		os.RemoveAll(path)
		os.MkdirAll(filepath.Dir(path), 0755)
	}

	return fmt.Errorf("所有认证方式均失败")
}

// tryClone attempts a single git clone with the given URL and optional SSH key.
func tryClone(path, url, sshKey string) error {
	cmd := exec.Command("git", "clone", url, path)
	cmd.Stderr = &strings.Builder{}

	if sshKey != "" {
		expandedKey := expandTilde(sshKey)
		cmd.Env = append(os.Environ(),
			"GIT_SSH_COMMAND=ssh -i "+expandedKey+" -o IdentitiesOnly=yes -o StrictHostKeyChecking=no",
		)
	}

	return cmd.Run()
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
	// Remove any URL that contains user:pass@ pattern
	// Match https://<anything>:<anything>@
	idx := strings.Index(msg, "https://")
	if idx >= 0 {
		rest := msg[idx:]
		// Find the end of the URL (space or end of string)
		end := strings.IndexAny(rest, " \n")
		if end < 0 {
			end = len(rest)
		}
		msg = msg[:idx] + "<url redacted>" + rest[end:]
	}
	return msg
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
