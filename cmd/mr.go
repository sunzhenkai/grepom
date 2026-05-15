package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/mergerequest"
	"github.com/wii/grepom/repo"
)

var (
	mrFrom     string
	mrTo       string
	mrTitle    string
	mrBody     string
	mrBodyFile string
	mrDraft    bool
	mrWeb      bool
)

var mrCmd = &cobra.Command{
	Use:   "mr",
	Short: "Create a Merge Request / Pull Request",
	Long: `Create a Merge Request (GitLab) or Pull Request (GitHub) from the command line.

Supports GitHub and GitLab. Codeup is not supported and will show a friendly message.

Without arguments, the command auto-detects:
  - Current branch as source (--from)
  - Default branch as target (--to)
  - Provider type from the remote URL
  - Title/body from the latest commit message

If the branch has unpushed commits, you will be prompted to push first.`,
	Example: `  grepom mr                                    # Auto-detect and create MR/PR
  grepom mr --from feature-x --to main         # Specify branches
  grepom mr --title "Add dark mode"             # Custom title
  grepom mr --draft                             # Create as draft
  grepom mr --web                               # Open browser to create
  grepom pr                                     # Alias for 'mr'`,
	RunE: runMR,
}

func init() {
	mrCmd.Flags().StringVar(&mrFrom, "from", "", "source branch (default: current branch)")
	mrCmd.Flags().StringVar(&mrTo, "to", "", "target branch (default: repository default branch)")
	mrCmd.Flags().StringVar(&mrTitle, "title", "", "MR/PR title (default: HEAD commit subject)")
	mrCmd.Flags().StringVar(&mrBody, "body", "", "MR/PR description (default: HEAD commit body)")
	mrCmd.Flags().StringVar(&mrBodyFile, "body-file", "", "read MR/PR description from file")
	mrCmd.Flags().BoolVar(&mrDraft, "draft", false, "create as draft MR/PR")
	mrCmd.Flags().BoolVar(&mrWeb, "web", false, "open browser to create MR/PR")

	rootCmd.AddCommand(mrCmd)

	// Register 'pr' as an alias for 'mr'
	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Create a Pull Request (alias for 'mr')",
		Long:  mrCmd.Long,
		Example: `  grepom pr                                    # Alias for 'grepom mr'
  grepom pr --from feature-x --to main         # Same as 'mr --from ... --to ...'
  grepom pr --draft                             # Create as draft PR`,
		RunE: runMR,
	}
	prCmd.Flags().StringVar(&mrFrom, "from", "", "source branch (default: current branch)")
	prCmd.Flags().StringVar(&mrTo, "to", "", "target branch (default: repository default branch)")
	prCmd.Flags().StringVar(&mrTitle, "title", "", "PR title (default: HEAD commit subject)")
	prCmd.Flags().StringVar(&mrBody, "body", "", "PR description (default: HEAD commit body)")
	prCmd.Flags().StringVar(&mrBodyFile, "body-file", "", "read PR description from file")
	prCmd.Flags().BoolVar(&mrDraft, "draft", false, "create as draft PR")
	prCmd.Flags().BoolVar(&mrWeb, "web", false, "open browser to create PR")
	rootCmd.AddCommand(prCmd)
}

func runMR(cmd *cobra.Command, args []string) error {
	// Step 1: Check current directory is a git repo
	if !gitpkg.IsCloned(".") {
		return fmt.Errorf("current directory is not a git repository")
	}

	// Step 2: Get current branch and default branch
	currentBranch, err := gitpkg.GetCurrentBranch(".")
	if err != nil {
		return fmt.Errorf("failed to detect current branch: %w", err)
	}

	fromBranch := mrFrom
	if fromBranch == "" {
		fromBranch = currentBranch
	}

	toBranch := mrTo
	if toBranch == "" {
		toBranch, err = gitpkg.GetDefaultBranch(".")
		if err != nil {
			// Fallback: try 'main' then 'master'
			if branchExists("main") {
				toBranch = "main"
			} else if branchExists("master") {
				toBranch = "master"
			} else {
				return fmt.Errorf("cannot detect default branch: %w\nUse --to to specify the target branch", err)
			}
		}
	}

	// Step 3: from == to protection
	if fromBranch == toBranch {
		return fmt.Errorf("source branch and target branch are the same (%q).\nUse --from to specify a different source branch.", fromBranch)
	}

	// Step 4: Detect provider from remote URL
	remoteURL, err := gitpkg.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	providerName, serverURL, token, err := detectProvider(remoteURL)
	if err != nil {
		return err
	}

	// Step 5: Extract repo path from remote URL
	repoPath := repo.ExtractRemotePath(remoteURL)
	if repoPath == "" {
		return fmt.Errorf("cannot extract repository path from remote URL: %s", remoteURL)
	}

	// Step 6: Handle Codeup (not supported)
	if providerName == "codeup" {
		webURL := buildCodeupWebURL(serverURL, repoPath)
		fmt.Fprintf(os.Stderr, "\n  \u274C Codeup \u6682\u4E0D\u652F\u6301\u901A\u8FC7 API \u521B\u5EFA Merge Request\u3002\n")
		fmt.Fprintf(os.Stderr, "  \u8BF7\u5728\u6D4F\u89C8\u5668\u4E2D\u624B\u52A8\u521B\u5EFA:\n  %s\n\n", webURL)
		return nil
	}

	// Step 7: Get MR/PR provider
	mrProvider, err := mergerequest.Get(providerName)
	if err != nil {
		return err
	}

	// Step 8: Resolve title and body
	title, body, err := resolveTitleAndBody()
	if err != nil {
		return err
	}

	// Step 9: Check for unpushed commits and prompt
	if !mrWeb {
		hasUnpushed, count, err := gitpkg.HasUnpushedCommits(".", fromBranch)
		if err != nil {
			return fmt.Errorf("failed to check unpushed commits: %w", err)
		}
		if hasUnpushed {
			if !isTerminal() {
				return fmt.Errorf("branch %q has %d unpushed commit(s). Please push first or run in a TTY environment.", fromBranch, count)
			}

			confirm := false
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Branch %q has %d unpushed commit(s). Push to origin first?", fromBranch, count),
				Default: true,
			}
			if err := survey.AskOne(prompt, &confirm); err != nil {
				if err == terminal.InterruptErr {
					fmt.Println("\nCancelled.")
					return nil
				}
				return err
			}

			if !confirm {
				return fmt.Errorf("please push branch %q before creating MR/PR", fromBranch)
			}

			fmt.Printf("Pushing to origin/%s...\n", fromBranch)
			if err := gitpkg.PushBranch(".", fromBranch); err != nil {
				return fmt.Errorf("push failed: %w", err)
			}
			fmt.Println("Push succeeded.")
		}
	}

	// Step 10: --web mode or API mode
	if mrWeb {
		webURL := mrProvider.BuildWebURL(mergerequest.WebURLParams{
			ServerURL:    serverURL,
			RepoPath:     repoPath,
			SourceBranch: fromBranch,
			TargetBranch: toBranch,
			Title:        title,
			Draft:        mrDraft,
		})
		fmt.Println(webURL)
		return openBrowser(webURL)
	}

	// Step 11: Create MR/PR via API
	fmt.Printf("Creating %s: %s → %s...\n", mrLabel(providerName), fromBranch, toBranch)

	mr, err := mrProvider.CreateMergeRequest(context.Background(), mergerequest.CreateMergeRequestParams{
		ServerURL:    serverURL,
		Token:        token,
		RepoPath:     repoPath,
		Title:        title,
		Description:  body,
		SourceBranch: fromBranch,
		TargetBranch: toBranch,
		Draft:        mrDraft,
	})
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", mrLabel(providerName), err)
	}

	// Step 12: Output result
	fmt.Printf("\n  \u2705 %s #%d: %s\n  %s\n\n",
		strings.ToUpper(mrLabel(providerName)), mr.Number, mr.Title, mr.URL)

	return nil
}

// --- Provider detection ---

// detectProvider determines the provider type, server URL, and token from a remote URL.
func detectProvider(remoteURL string) (providerName string, serverURL string, token string, err error) {
	host := extractHost(remoteURL)

	// Strategy 1: Try to match config resources
	cfg, cfgErr := tryLoadConfig()
	if cfgErr == nil && cfg != nil {
		for name, res := range cfg.Resources {
			if res.URL == host {
				resolvedToken, tokenErr := res.ResolvedToken()
				if tokenErr != nil {
					return "", "", "", fmt.Errorf("resource %q: %w", name, tokenErr)
				}
				scheme := "https://"
				if res.Scheme() == "http" {
					scheme = "http://"
				}
				return res.Provider, scheme + res.URL, resolvedToken, nil
			}
		}
	}

	// Strategy 2: Known domain matching
	switch host {
	case "github.com":
		token = os.Getenv("GREPOM_GITHUB_TOKEN")
		return "github", "https://github.com", token, nil
	case "gitlab.com":
		token = os.Getenv("GREPOM_GITLAB_TOKEN")
		return "gitlab", "https://gitlab.com", token, nil
	case "codeup.aliyun.com":
		return "codeup", "https://codeup.aliyun.com", "", nil
	}

	return "", "", "", fmt.Errorf("cannot determine provider for host %q\nAdd a resource in your .grepom.yml config file, or set GREPOM_GITHUB_TOKEN / GREPOM_GITLAB_TOKEN environment variable", host)
}

// resolveToken tries to get a token for the given provider and host.
func resolveToken(providerName string, host string, cfg *config.Config) (string, error) {
	// Strategy 1: From config
	if cfg != nil {
		for _, res := range cfg.Resources {
			if res.URL == host {
				token, err := res.ResolvedToken()
				if err != nil {
					return "", err
				}
				if token != "" {
					return token, nil
				}
			}
		}
	}

	// Strategy 2: From environment variable
	switch providerName {
	case "github":
		if token := os.Getenv("GREPOM_GITHUB_TOKEN"); token != "" {
			return token, nil
		}
	case "gitlab":
		if token := os.Getenv("GREPOM_GITLAB_TOKEN"); token != "" {
			return token, nil
		}
	}

	return "", fmt.Errorf("no token found for %s\nSet GREPOM_%s_TOKEN environment variable or add a resource in .grepom.yml",
		providerName, strings.ToUpper(providerName))
}

// --- Helper functions ---

// tryLoadConfig attempts to load the config, returning nil if not found.
func tryLoadConfig() (*config.Config, error) {
	path, err := config.FindConfig(configFile)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// extractHost extracts the host from a git remote URL.
func extractHost(remoteURL string) string {
	// Handle git@host:path format
	if strings.HasPrefix(remoteURL, "git@") {
		rest := strings.TrimPrefix(remoteURL, "git@")
		if idx := strings.Index(rest, ":"); idx >= 0 {
			return rest[:idx]
		}
		// Try with /
		if idx := strings.Index(rest, "/"); idx >= 0 {
			return rest[:idx]
		}
		return rest
	}

	// Handle https://host/path format
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(remoteURL, scheme) {
			rest := strings.TrimPrefix(remoteURL, scheme)
			if idx := strings.Index(rest, "/"); idx >= 0 {
				return rest[:idx]
			}
			return rest
		}
	}

	// Handle ssh://host/path format
	if strings.HasPrefix(remoteURL, "ssh://") {
		rest := strings.TrimPrefix(remoteURL, "ssh://")
		// Strip user@ if present
		if atIdx := strings.Index(rest, "@"); atIdx >= 0 {
			rest = rest[atIdx+1:]
		}
		if idx := strings.Index(rest, "/"); idx >= 0 {
			return rest[:idx]
		}
		return rest
	}

	return remoteURL
}

// resolveTitleAndBody resolves the MR/PR title and body from flags or HEAD commit.
func resolveTitleAndBody() (string, string, error) {
	title := mrTitle
	body := mrBody

	// Read body from file if specified
	if mrBodyFile != "" {
		data, err := os.ReadFile(mrBodyFile)
		if err != nil {
			return "", "", fmt.Errorf("failed to read body file: %w", err)
		}
		body = string(data)
	}

	// Auto-detect from commit message if not specified
	if title == "" && body == "" {
		msg, err := gitpkg.GetHeadCommitMessage(".")
		if err != nil {
			return "", "", fmt.Errorf("failed to get commit message: %w", err)
		}

		parts := strings.SplitN(msg, "\n", 2)
		title = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			body = strings.TrimSpace(parts[1])
		}
	} else if title == "" {
		// Body specified but no title — get title from commit
		msg, err := gitpkg.GetHeadCommitMessage(".")
		if err != nil {
			return "", "", fmt.Errorf("failed to get commit message: %w", err)
		}
		parts := strings.SplitN(msg, "\n", 2)
		title = strings.TrimSpace(parts[0])
	}

	if title == "" {
		return "", "", fmt.Errorf("no title provided and no commit message found")
	}

	return title, body, nil
}

// branchExists checks if a branch exists (local or remote).
func branchExists(name string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", name)
	cmd.Dir = "."
	return cmd.Run() == nil
}

// openBrowser opens a URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// mrLabel returns "PR" for github, "MR" for others.
func mrLabel(providerName string) string {
	if providerName == "github" {
		return "PR"
	}
	return "MR"
}

// buildCodeupWebURL builds a Codeup MR creation page URL.
func buildCodeupWebURL(serverURL string, repoPath string) string {
	base := strings.TrimRight(serverURL, "/")
	return fmt.Sprintf("%s/%s/merge_requests/new", base, repoPath)
}
