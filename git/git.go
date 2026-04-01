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
	Token    string // token for HTTPS authentication
	Provider string // "github" or "gitlab", determines token URL format
	SSHKey   string // path to SSH private key file
}

// Clone clones a repository, creating parent directories as needed.
// It tries authentication methods in priority order:
//  1. token auth (HTTPS + embedded token)
//  2. SSH with specified key
//  3. Default SSH (derived URL)
//  4. Bare HTTP
func Clone(path, sshURL, httpURL string, opts CloneOptions) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Build ordered list of auth strategies to try
	type strategy struct {
		label  string
		url    string
		sshKey string // non-empty means use GIT_SSH_COMMAND
	}

	var strategies []strategy

	// 1. Token auth
	if opts.Token != "" && httpURL != "" {
		tokenURL := buildTokenURL(httpURL, opts.Token, opts.Provider)
		strategies = append(strategies, strategy{
			label:  fmt.Sprintf("token 认证 (%s)", authLevel(opts)),
			url:    tokenURL,
			sshKey: "",
		})
	}

	// 2. SSH with specified key
	if opts.SSHKey != "" && sshURL != "" {
		strategies = append(strategies, strategy{
			label:  fmt.Sprintf("SSH key 认证 (%s)", authLevel(opts)),
			url:    sshURL,
			sshKey: opts.SSHKey,
		})
	}

	// 3. Default SSH (derived URL)
	if sshURL != "" {
		strategies = append(strategies, strategy{
			label:  "SSH 认证 (默认)",
			url:    sshURL,
			sshKey: "",
		})
	}

	// 4. Bare HTTP
	if httpURL != "" {
		strategies = append(strategies, strategy{
			label:  "HTTP 克隆",
			url:    httpURL,
			sshKey: "",
		})
	}

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

// authLevel returns a human-readable label for the auth source
func authLevel(opts CloneOptions) string {
	if opts.Token != "" && opts.SSHKey != "" {
		return "group/repo"
	}
	if opts.Token != "" {
		return "group/repo"
	}
	if opts.SSHKey != "" {
		return "group/repo"
	}
	return "resource"
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
