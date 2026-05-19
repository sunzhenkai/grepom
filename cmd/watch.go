package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/cicd"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var watchID int

var watchCmd = &cobra.Command{
	Use:           "watch [repo-name]",
	Short:         "Watch the latest CI/CD pipeline",
	Long:          `Watch the latest CI/CD pipeline for a repository.
This is a shortcut for "grepom pipeline watch".

When no repo-name is provided, the command auto-detects the repository
from the current directory's git remote URL using a 3-level fallback:
  1. Match against repos in .grepom.yml config
  2. Match the remote host against config resources
  3. Use known public domains (github.com / gitlab.com) with env var tokens

Polls every 5 seconds until the pipeline reaches a terminal state.
Press Ctrl+C to stop early.`,
	Example: `  grepom watch                          # Auto-detect repo from current directory
  grepom watch web-app                  # Watch repo by name (like pipeline watch)
  grepom watch --id 1234                # Auto-detect repo, watch specific pipeline
  grepom watch web-app --id 1234        # Specify repo and pipeline ID`,
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runWatch,
}

func init() {
	watchCmd.Flags().IntVar(&watchID, "id", 0, "specific pipeline ID to watch (default: latest)")
	rootCmd.AddCommand(watchCmd)
}

// runWatch 是 watch 顶级命令的入口。
// 有 repo-name 参数时走 resolvePipelineInput 构造 WatchTarget；
// 无参数时走 resolveCurrentRepoPipeline 自动推断。
func runWatch(cmd *cobra.Command, args []string) error {
	var target WatchTarget
	var err error

	if len(args) > 0 {
		// 显式指定 repo-name：走与 pipeline watch 相同的路径
		_, cfg, cfgErr := loadConfig()
		if cfgErr != nil {
			return cfgErr
		}

		repoName := args[0]
		provider, serverURL, remotePath, token, resolveErr := resolvePipelineInput(cfg, repoName)
		if resolveErr != nil {
			return resolveErr
		}

		target = WatchTarget{
			Provider:  provider,
			ServerURL: serverURL,
			RepoPath:  remotePath,
			Token:     token,
			RepoName:  repoName,
		}
	} else {
		// 自动推断
		target, err = resolveCurrentRepoPipeline()
		if err != nil {
			return err
		}
	}

	return runWatchLoop(target, watchID, cmd)
}

// resolveCurrentRepoPipeline 通过三级 fallback 从当前 git 仓库推断 WatchTarget。
// 前置检查：当前目录必须是 git 仓库且配置了 remote origin。
// Level 1: 遍历 .grepom.yml 中的 repo 条目，用 remotePath 精确匹配。
// Level 2: 遍历 config resources，用 host 匹配，从 remote URL 推导 remotePath。
// Level 3: 已知公共域名（github.com / gitlab.com）+ 环境变量 token。
func resolveCurrentRepoPipeline() (WatchTarget, error) {
	// --- 前置检查 ---

	if !gitpkg.IsCloned(".") {
		return WatchTarget{}, fmt.Errorf(notGitRepoError)
	}

	remoteURL, err := gitpkg.GetRemoteURL(".", "origin")
	if err != nil {
		return WatchTarget{}, fmt.Errorf(noRemoteError)
	}

	remotePath := repo.ExtractRemotePath(remoteURL)
	host := extractHost(remoteURL)
	repoName := filepath.Base(remotePath) // 从 remotePath 取最后一段作为显示名

	// --- Level 1: 配置精确匹配 ---

	_, cfg, cfgErr := loadConfig()
	if cfgErr == nil && cfg != nil {
		resolver := repo.NewResolver(cfg)
		allRepos, resolveErr := resolver.Resolve()
		if resolveErr == nil {
			for _, r := range allRepos {
				rp := repo.ExtractRemotePath(r.CloneURL)
				if rp == "" {
					rp = repo.ExtractRemotePath(r.SSHURL)
				}
				if rp == remotePath {
					// 精确匹配：使用该 repo 的 resource 信息
					provider, serverURL, resolvedRemotePath, token, inputErr := resolvePipelineInput(cfg, r.Name)
					if inputErr != nil {
						return WatchTarget{}, inputErr
					}
					return WatchTarget{
						Provider:  provider,
						ServerURL: serverURL,
						RepoPath:  resolvedRemotePath,
						Token:     token,
						RepoName:  r.Name,
					}, nil
				}
			}
		}
	}

	// --- Level 2: Host 匹配 + Path 推导 ---

	if cfgErr == nil && cfg != nil {
		for name, res := range cfg.Resources {
			if res.URL == host {
				provider, providerErr := cicd.Get(res.Provider)
				if providerErr != nil {
					return WatchTarget{}, fmt.Errorf("resource %q: %w", name, providerErr)
				}

				resolvedToken, tokenErr := res.ResolvedToken()
				if tokenErr != nil {
					return WatchTarget{}, fmt.Errorf("resource %q: %w", name, tokenErr)
				}

				return WatchTarget{
					Provider:  provider,
					ServerURL: res.APIURL(),
					RepoPath:  remotePath,
					Token:     resolvedToken,
					RepoName:  repoName,
				}, nil
			}
		}
	}

	// --- Level 3: 已知公共域名 ---

	switch host {
	case "github.com":
		token := os.Getenv("GREPOM_GITHUB_TOKEN")
		if token == "" {
			return WatchTarget{}, fmt.Errorf(noTokenForKnownDomainError,
				repoName, remoteURL, "GitHub", "GREPOM_GITHUB_TOKEN",
				"github", "github", "github.com")
		}
		provider, _ := cicd.Get("github")
		return WatchTarget{
			Provider:  provider,
			ServerURL: "https://github.com",
			RepoPath:  remotePath,
			Token:     token,
			RepoName:  repoName,
		}, nil
	case "gitlab.com":
		token := os.Getenv("GREPOM_GITLAB_TOKEN")
		if token == "" {
			return WatchTarget{}, fmt.Errorf(noTokenForKnownDomainError,
				repoName, remoteURL, "GitLab", "GREPOM_GITLAB_TOKEN",
				"gitlab", "gitlab", "gitlab.com")
		}
		provider, _ := cicd.Get("gitlab")
		return WatchTarget{
			Provider:  provider,
			ServerURL: "https://gitlab.com",
			RepoPath:  remotePath,
			Token:     token,
			RepoName:  repoName,
		}, nil
	}

	// --- 所有 fallback 失败 ---

	if cfgErr != nil || cfg == nil {
		return WatchTarget{}, fmt.Errorf(noConfigMatchNoConfigError,
			repoName, remoteURL, remotePath, host,
			host, host, host)
	}

	// 有配置但找不到匹配
	return WatchTarget{}, fmt.Errorf(noConfigMatchWithConfigError,
		repoName, remoteURL, remotePath, host,
		remotePath, host, host, host)
}

// --- 详细错误信息模板 ---

const notGitRepoError = `current directory is not a git repository

grepom watch needs to run inside a git repository to auto-detect
the CI/CD pipeline from the remote URL.

Hints:
  • cd into your project directory and try again
  • or specify the repo name explicitly: grepom pipeline watch <repo-name>`

const noRemoteError = `current repository has no remote origin configured

grepom watch uses the git remote URL to find the CI/CD pipeline.

Hints:
  • add a remote first: git remote add origin <url>
  • or specify the repo name explicitly: grepom pipeline watch <repo-name>`

const noTokenForKnownDomainError = `found CI/CD provider but missing authentication token

  Current repo:   %s
  Remote URL:     %s
  Provider:       %s

Hints:
  • Set the environment variable: export %s=xxx
  • Or add a resource in .grepom.yml:
    resources:
      %s:
        provider: %s
        url: %s
        token: \${MY_TOKEN}`

const noConfigMatchNoConfigError = `cannot find CI/CD info for current repository

  Current repo:   %s
  Remote URL:     %s
  Remote path:    %s
  Host:           %s

Hints:
  • No .grepom.yml config found for this project
  • The host %q is not a known public domain (github.com / gitlab.com)

Suggestions:
  • Create a .grepom.yml config with a resource for %q:
    resources:
      mygitlab:
        provider: gitlab
        url: %s
        token: ${MY_GITLAB_TOKEN}
  • Or specify the repo name explicitly: grepom pipeline watch <repo-name>`

const noConfigMatchWithConfigError = `cannot find CI/CD info for current repository

  Current repo:   %s
  Remote URL:     %s
  Remote path:    %s
  Host:           %s

Hints:
  • .grepom.yml has no repo entry matching remote path %q
  • No resource in .grepom.yml matches host %q

Suggestions:
  • Add a resource for %q in .grepom.yml:
    resources:
      mygitlab:
        provider: gitlab
        url: %s
        token: ${MY_GITLAB_TOKEN}
  • Or add this repo to a group via: grepom add repo
  • Or specify the repo name explicitly: grepom pipeline watch <repo-name>`

// extractHost is defined in cmd/mr.go — reused here.
// isTerminal is defined in cmd/interactive.go — reused here.
