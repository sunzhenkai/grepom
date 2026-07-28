package cmd

import (
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/repo"
)

func buildRepoFilter(cfg *config.Config, group, vgroup, resource string, includeDisabled bool) (repo.Filter, error) {
	scope, err := cfg.ResolveScopeSelection(group, vgroup)
	if err != nil {
		return repo.Filter{}, err
	}

	filter := repo.Filter{
		Resource:        resource,
		IncludeDisabled: includeDisabled,
		RepoNames:       scope.RepoNames,
	}
	if len(scope.Groups) == 1 {
		filter.Group = scope.Groups[0]
	}
	if len(scope.Groups) > 0 {
		filter.Groups = scope.Groups
	}
	return filter, nil
}
