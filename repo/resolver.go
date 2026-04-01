package repo

import (
	"path/filepath"
	"strings"

	"github.com/wii/grepom/config"
	"github.com/wii/grepom/provider"
)

// Filter defines criteria for filtering repos.
type Filter struct {
	Name     string
	Group    string // group name
	Resource string // resource name
}

// Resolver builds a list of provider.Repo from the config.
type Resolver struct {
	cfg *config.Config
}

func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

// Resolve builds the full repo list from all groups and standalone repos.
func (r *Resolver) Resolve() ([]provider.Repo, error) {
	var allRepos []provider.Repo

	for _, g := range r.cfg.Groups {
		res, ok := r.cfg.Resources[g.Resource]
		if !ok {
			continue
		}

		for _, gr := range g.Repos {
			localPath := config.ResolveGroupRepoPath(r.cfg.Base, g.LocalPath, g.Path, gr.Path)
			allRepos = append(allRepos, provider.Repo{
				Name:      gr.Name,
				CloneURL:  gr.URL,
				SSHURL:    deriveSSHURL(gr.URL, res.Provider),
				Path:      localPath,
				Provider:  res.Provider,
				Resource:  g.Resource,
				GroupName: g.Name,
			})
		}
	}

	for _, repo := range r.cfg.Repos {
		res, ok := r.cfg.Resources[repo.Resource]
		if !ok {
			continue
		}

		localPath := config.ResolveRepoPath(r.cfg.Base, repo.LocalPath)
		allRepos = append(allRepos, provider.Repo{
			Name:     repo.Name,
			CloneURL: repo.URL,
			SSHURL:   deriveSSHURL(repo.URL, res.Provider),
			Path:     localPath,
			Provider: res.Provider,
			Resource: repo.Resource,
		})
	}

	return allRepos, nil
}

// ResolveAndFilter builds the repo list and applies the given filter.
func (r *Resolver) ResolveAndFilter(filter Filter) ([]provider.Repo, error) {
	allRepos, err := r.Resolve()
	if err != nil {
		return nil, err
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

// FullPath returns the absolute local path for a repo.
// The repo's Path field already contains the full derived path.
func FullPath(base string, r provider.Repo) string {
	return filepath.Clean(r.Path)
}

// deriveSSHURL converts an HTTPS clone URL to SSH format based on provider.
func deriveSSHURL(cloneURL, prov string) string {
	switch prov {
	case "gitlab":
		if strings.HasPrefix(cloneURL, "https://") {
			return "git@" + strings.Replace(strings.TrimPrefix(cloneURL, "https://"), "/", ":", 1)
		}
		return cloneURL
	case "github":
		if strings.HasPrefix(cloneURL, "https://") {
			return "git@" + strings.Replace(strings.TrimPrefix(cloneURL, "https://"), "/", ":", 1)
		}
		return cloneURL
	default:
		return cloneURL
	}
}
