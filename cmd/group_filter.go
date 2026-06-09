package cmd

import (
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/repo"
)

func buildRepoFilter(cfg *config.Config, group, vgroup, resource string, includeDisabled bool) (repo.Filter, error) {
	groups, err := cfg.ResolveGroupSelection(group, vgroup)
	if err != nil {
		return repo.Filter{}, err
	}

	filter := repo.Filter{
		Resource:        resource,
		IncludeDisabled: includeDisabled,
	}
	if len(groups) == 1 {
		filter.Group = groups[0]
	}
	if len(groups) > 0 {
		filter.Groups = groups
	}
	return filter, nil
}
