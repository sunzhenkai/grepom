package repo

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/wii/grepom/config"
	"github.com/wii/grepom/provider"
)

type Filter struct {
	Name     string
	Group    string
	Provider string
}

type Resolver struct {
	cfg *config.Config
}

func NewResolver(cfg *config.Config) *Resolver {
	return &Resolver{cfg: cfg}
}

// Resolve fetches repos from all API sources and merges with explicit repo entries.
func (r *Resolver) Resolve(ctx context.Context) ([]provider.Repo, error) {
	var allRepos []provider.Repo

	// Fetch from API sources
	for _, source := range r.cfg.Sources {
		config.Verbose("fetching repos from %s source at %s", source.Provider, source.URL)
		p, err := provider.Get(source.Provider)
		if err != nil {
			return nil, err
		}

		repos, err := p.ListRepos(ctx, source)
		if err != nil {
			return nil, err
		}

		config.Verbose("found %d repos from %s", len(repos), source.URL)
		allRepos = append(allRepos, repos...)
	}

	// Add explicit repos
	for _, entry := range r.cfg.Repos {
		path := entry.Path
		if strings.HasPrefix(path, "./") {
			path = path[2:]
		}
		allRepos = append(allRepos, provider.Repo{
			Name:     entry.Name,
			CloneURL: entry.URL,
			SSHURL:   strings.Replace(entry.URL, "https://", "git@", 1),
			Path:     path,
			Provider: "explicit",
		})
	}

	config.Verbose("total %d repos resolved", len(allRepos))
	return allRepos, nil
}

// ResolveAndFilter fetches repos and applies the given filter.
func (r *Resolver) ResolveAndFilter(ctx context.Context, filter Filter) ([]provider.Repo, error) {
	allRepos, err := r.Resolve(ctx)
	if err != nil {
		return nil, err
	}

	return ApplyFilter(allRepos, filter), nil
}

// ApplyFilter filters a repo list by name, group, or provider.
func ApplyFilter(repos []provider.Repo, filter Filter) []provider.Repo {
	var result []provider.Repo

	for _, r := range repos {
		if filter.Name != "" && r.Name != filter.Name {
			continue
		}
		if filter.Group != "" && !strings.HasPrefix(r.Path, filter.Group) {
			continue
		}
		if filter.Provider != "" && r.Provider != filter.Provider {
			continue
		}
		result = append(result, r)
	}

	return result
}

// FullPath joins the base directory with a repo's relative path.
func FullPath(base string, repo provider.Repo) string {
	return filepath.Join(base, repo.Path)
}
