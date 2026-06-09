package config

import "fmt"

// VirtualGroup defines a named collection of real groups.
type VirtualGroup struct {
	Groups []string `yaml:"groups"`
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
func (c *Config) ResolveGroupSelection(group, vgroup string) ([]string, error) {
	if group == "" && vgroup == "" {
		return nil, nil
	}

	seen := make(map[string]bool)
	var selected []string

	add := func(name string) {
		if seen[name] {
			return
		}
		seen[name] = true
		selected = append(selected, name)
	}

	if group != "" {
		if _, _, err := c.FindGroup(group); err != nil {
			return nil, err
		}
		add(group)
	}

	if vgroup != "" {
		vg, err := c.FindVirtualGroup(vgroup)
		if err != nil {
			return nil, err
		}
		for _, member := range vg.Groups {
			if _, _, err := c.FindGroup(member); err != nil {
				return nil, fmt.Errorf("virtual group %q: %w", vgroup, err)
			}
			add(member)
		}
	}

	return selected, nil
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

// CountMemberRepos returns the total repo count across member real groups.
func (c *Config) CountMemberRepos(memberGroups []string) int {
	total := 0
	for _, name := range memberGroups {
		if _, g, err := c.FindGroup(name); err == nil {
			total += len(g.Repos)
		}
	}
	return total
}
