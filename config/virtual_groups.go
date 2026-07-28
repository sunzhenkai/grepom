package config

import "fmt"

// VirtualGroup defines a named collection of real groups and optional standalone repos.
type VirtualGroup struct {
	Groups []string `yaml:"groups"`
	Repos  []string `yaml:"repos,omitempty"`
}

// ScopeSelection is the resolved --group/--vgroup filter scope.
// Nil Groups with empty RepoNames means no restriction.
// Non-nil empty Groups with RepoNames means only those standalone repos.
type ScopeSelection struct {
	Groups    []string
	RepoNames []string
}

// FindVirtualGroup finds a virtual group by name.
func (c *Config) FindVirtualGroup(name string) (*VirtualGroup, error) {
	vg, ok := c.VirtualGroups[name]
	if !ok {
		return nil, fmt.Errorf("virtual group %q not found", name)
	}
	return &vg, nil
}

// ResolveGroupSelection expands --group and --vgroup into deduplicated real group names.
// Returns nil when both are empty, meaning no group restriction.
// Standalone repos referenced by virtual_groups.repos are ignored here (use ResolveScopeSelection).
func (c *Config) ResolveGroupSelection(group, vgroup string) ([]string, error) {
	scope, err := c.ResolveScopeSelection(group, vgroup)
	if err != nil {
		return nil, err
	}
	return scope.Groups, nil
}

// ResolveScopeSelection expands --group and --vgroup into real groups and standalone repo names.
func (c *Config) ResolveScopeSelection(group, vgroup string) (ScopeSelection, error) {
	if group == "" && vgroup == "" {
		return ScopeSelection{}, nil
	}

	seenGroups := make(map[string]bool)
	seenRepos := make(map[string]bool)
	var selectedGroups []string
	var selectedRepos []string

	addGroup := func(name string) {
		if seenGroups[name] {
			return
		}
		seenGroups[name] = true
		selectedGroups = append(selectedGroups, name)
	}
	addRepo := func(name string) {
		if seenRepos[name] {
			return
		}
		seenRepos[name] = true
		selectedRepos = append(selectedRepos, name)
	}

	if group != "" {
		if _, _, err := c.FindGroup(group); err != nil {
			return ScopeSelection{}, err
		}
		addGroup(group)
	}

	if vgroup != "" {
		vg, err := c.FindVirtualGroup(vgroup)
		if err != nil {
			return ScopeSelection{}, err
		}
		for _, member := range vg.Groups {
			if _, _, err := c.FindGroup(member); err != nil {
				return ScopeSelection{}, fmt.Errorf("virtual group %q: %w", vgroup, err)
			}
			addGroup(member)
		}
		for _, repoName := range vg.Repos {
			addRepo(repoName)
		}
	}

	return ScopeSelection{
		Groups:    selectedGroups,
		RepoNames: selectedRepos,
	}, nil
}

// GroupInSelection reports whether a real group name matches the resolved selection.
// An empty selection matches all groups.
func (c *Config) GroupInSelection(groupName string, selection []string) bool {
	if len(selection) == 0 {
		return true
	}
	for _, name := range selection {
		if name == groupName {
			return true
		}
	}
	return false
}

// FilterGroups returns config groups matching the resolved --group/--vgroup selection.
func (c *Config) FilterGroups(group, vgroup string) ([]Group, error) {
	selection, err := c.ResolveGroupSelection(group, vgroup)
	if err != nil {
		return nil, err
	}

	var result []Group
	for _, g := range c.Groups {
		if c.GroupInSelection(g.Name, selection) {
			result = append(result, g)
		}
	}
	return result, nil
}

// CountMemberRepos returns the total repo count across member real groups
// and optional standalone repos referenced by a virtual group.
func (c *Config) CountMemberRepos(memberGroups []string, memberRepos []string) int {
	total := 0
	for _, name := range memberGroups {
		if _, g, err := c.FindGroup(name); err == nil {
			total += len(g.Repos)
		}
	}
	total += len(memberRepos)
	return total
}
