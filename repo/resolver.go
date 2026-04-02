package repo

import (
	"path/filepath"
	"strings"

	"github.com/wii/grepom/config"
	"github.com/wii/grepom/provider"
)

// Filter defines criteria for filtering repos.
type Filter struct {
	Name            string
	Group           string // group name
	Resource        string // resource name
	IncludeDisabled bool   // when true, include disabled/excluded repos in results
}

// Resolver builds a list of provider.Repo from the config.
type Resolver struct {
	cfg *config.Config
}

func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

// resolveInternal builds the full repo list from all groups and standalone repos,
// setting DisabledReason for each repo that should be excluded.
func (r *Resolver) resolveInternal() []provider.Repo {
	var allRepos []provider.Repo

	for _, g := range r.cfg.Groups {
		res, ok := r.cfg.Resources[g.Resource]
		if !ok {
			continue
		}

		// Determine token: group override > resource default
		token := g.Token
		hasGroupToken := g.Token != ""
		if token == "" {
			token = res.Token
		}

		// Determine SSHKey: group override > resource default
		sshKey := g.SSHKey
		hasGroupSSHKey := g.SSHKey != ""
		if sshKey == "" {
			sshKey = res.SSHKey
		}

		for _, gr := range g.Repos {
			localPath := config.ResolveGroupRepoPath(r.cfg.Base, g.LocalPath, g.Path, gr.Path)

			pRepo := provider.Repo{
				Name:           gr.Name,
				CloneURL:       res.HTTPSURL(gr.Path),
				SSHURL:         deriveSSHURL(gr.Path, res.URL),
				Path:           localPath,
				Provider:       res.Provider,
				Resource:       g.Resource,
				GroupName:      g.Name,
				Token:          token,
				SSHKey:         sshKey,
				HasGroupToken:  hasGroupToken,
				HasGroupSSHKey: hasGroupSSHKey,
			}

			// Determine exclusion reason (priority: resource > group > exclude_repos)
			if !res.IsEnabled() {
				pRepo.DisabledReason = "disabled"
			} else if !g.IsEnabled() {
				pRepo.DisabledReason = "disabled"
			} else if isExcluded(g.ExcludeRepos, gr.Name) {
				pRepo.DisabledReason = "excluded"
			}

			allRepos = append(allRepos, pRepo)
		}
	}

	for _, repo := range r.cfg.Repos {
		res, ok := r.cfg.Resources[repo.Resource]
		if !ok {
			continue
		}

		// Determine token: repo override > resource default
		token := repo.Token
		hasGroupToken := repo.Token != ""
		if token == "" {
			token = res.Token
		}

		// Determine SSHKey: repo override > resource default
		sshKey := repo.SSHKey
		hasGroupSSHKey := repo.SSHKey != ""
		if sshKey == "" {
			sshKey = res.SSHKey
		}

		localPath := config.ResolveRepoPath(r.cfg.Base, repo.LocalPath)
		repoPath := extractRepoPath(repo.URL)

		pRepo := provider.Repo{
			Name:           repo.Name,
			CloneURL:       res.HTTPSURL(repoPath),
			SSHURL:         deriveSSHURL(repoPath, res.URL),
			Path:           localPath,
			Provider:       res.Provider,
			Resource:       repo.Resource,
			Token:          token,
			SSHKey:         sshKey,
			HasGroupToken:  hasGroupToken,
			HasGroupSSHKey: hasGroupSSHKey,
		}

		// Determine exclusion reason (priority: resource > repo)
		if !res.IsEnabled() {
			pRepo.DisabledReason = "disabled"
		} else if !repo.IsEnabled() {
			pRepo.DisabledReason = "disabled"
		}

		allRepos = append(allRepos, pRepo)
	}

	return allRepos
}

// isExcluded checks if a repo name is in the exclude list.
func isExcluded(excludeRepos []string, repoName string) bool {
	for _, name := range excludeRepos {
		if name == repoName {
			return true
		}
	}
	return false
}

// Resolve builds the full repo list, excluding disabled and excluded repos.
func (r *Resolver) Resolve() ([]provider.Repo, error) {
	allRepos := r.resolveInternal()

	// Filter out disabled/excluded repos
	var enabled []provider.Repo
	for _, repo := range allRepos {
		if repo.DisabledReason == "" {
			enabled = append(enabled, repo)
		}
	}
	return enabled, nil
}

// ResolveAndFilter builds the repo list and applies the given filter.
// When filter.IncludeDisabled is true, disabled/excluded repos are included in results.
func (r *Resolver) ResolveAndFilter(filter Filter) ([]provider.Repo, error) {
	allRepos := r.resolveInternal()

	// If IncludeDisabled is false, filter out disabled/excluded repos
	if !filter.IncludeDisabled {
		var enabled []provider.Repo
		for _, repo := range allRepos {
			if repo.DisabledReason == "" {
				enabled = append(enabled, repo)
			}
		}
		allRepos = enabled
	}

	return ApplyFilter(allRepos, filter), nil
}

// ApplyFilter filters a repo list by name, group, or resource.
func ApplyFilter(repos []provider.Repo, filter Filter) []provider.Repo {
	var result []provider.Repo

	for _, r := range repos {
		if filter.Name != "" && r.Name != filter.Name {
			continue
		}
		if filter.Group != "" && r.GroupName != filter.Group {
			continue
		}
		if filter.Resource != "" && r.Resource != filter.Resource {
			continue
		}
		result = append(result, r)
	}

	return result
}

// ApplySearchFilter filters repos by case-insensitive substring match on name,
// then applies exact group/resource filters.
func ApplySearchFilter(repos []provider.Repo, keyword string, filter Filter) []provider.Repo {
	keyword = strings.ToLower(keyword)
	var result []provider.Repo

	for _, r := range repos {
		if keyword != "" && !strings.Contains(strings.ToLower(r.Name), keyword) {
			continue
		}
		if filter.Group != "" && r.GroupName != filter.Group {
			continue
		}
		if filter.Resource != "" && r.Resource != filter.Resource {
			continue
		}
		result = append(result, r)
	}

	return result
}

// FullPath returns the absolute local path for a repo.
// The repo's Path field already contains the full derived path.
func FullPath(base string, r provider.Repo) string {
	return filepath.Clean(r.Path)
}

// deriveSSHURL 从 host:port 和 repo path 推导 SSH URL。
// 格式：git@<host>:<path>.git
func deriveSSHURL(repoPath, host string) string {
	return "git@" + host + ":" + repoPath + ".git"
}

// extractRepoPath 从克隆 URL 中提取 repo 路径部分。
// 例如 "https://gitlab.com/me/dotfiles.git" → "me/dotfiles"
// "git@gitlab.com:me/dotfiles.git" → "me/dotfiles"
// "me/dotfiles.git" → "me/dotfiles"
func extractRepoPath(cloneURL string) string {
	// 去掉 .git 后缀
	path := strings.TrimSuffix(cloneURL, ".git")

	// 处理 https:// 或 http:// 前缀
	for _, scheme := range []string{"https://", "http://"} {
		if strings.HasPrefix(path, scheme) {
			path = strings.TrimPrefix(path, scheme)
			// 取第一个 / 之后的部分（去掉 host:port）
			if idx := strings.Index(path, "/"); idx >= 0 {
				return path[idx+1:]
			}
			return path
		}
	}

	// 处理 git@host:path 格式
	if strings.HasPrefix(path, "git@") {
		path = strings.TrimPrefix(path, "git@")
		if idx := strings.Index(path, ":"); idx >= 0 {
			return path[idx+1:]
		}
		return path
	}

	// 已经是纯路径格式
	return path
}
