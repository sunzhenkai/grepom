package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML accepts resources as either a YAML map or a list of named items.
// List items MUST include a non-empty `name` field, which becomes the map key.
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	type plain struct {
		Base          string                 `yaml:"base"`
		Resources     yaml.Node              `yaml:"resources"`
		Groups        []Group                `yaml:"groups"`
		VirtualGroups map[string]VirtualGroup `yaml:"virtual_groups,omitempty"`
		Repos         []Repo                 `yaml:"repos"`
		Services      map[string]ServiceDef  `yaml:"services,omitempty"`
		YAMLIndent    int                    `yaml:"yaml_indent,omitempty"`
	}
	var p plain
	if err := value.Decode(&p); err != nil {
		return err
	}

	resources, err := decodeResourcesNode(&p.Resources)
	if err != nil {
		return err
	}

	c.Base = p.Base
	c.Resources = resources
	c.Groups = p.Groups
	c.VirtualGroups = p.VirtualGroups
	c.Repos = p.Repos
	c.Services = p.Services
	c.YAMLIndent = p.YAMLIndent
	return nil
}

func decodeResourcesNode(node *yaml.Node) (map[string]Resource, error) {
	if node == nil || node.Kind == 0 || node.Tag == "!!null" {
		return make(map[string]Resource), nil
	}
	if node.Kind == yaml.ScalarNode && (node.Value == "" || node.Value == "null" || node.Value == "~") {
		return make(map[string]Resource), nil
	}

	switch node.Kind {
	case yaml.MappingNode:
		var m map[string]Resource
		if err := node.Decode(&m); err != nil {
			return nil, fmt.Errorf("resources: %w", err)
		}
		if m == nil {
			m = make(map[string]Resource)
		}
		return m, nil

	case yaml.SequenceNode:
		m := make(map[string]Resource, len(node.Content))
		for i, itemNode := range node.Content {
			var named struct {
				Name string `yaml:"name"`
			}
			if err := itemNode.Decode(&named); err != nil {
				return nil, fmt.Errorf("resources[%d]: %w", i, err)
			}
			if named.Name == "" {
				return nil, fmt.Errorf("resources[%d]: 'name' field is required for list format", i)
			}
			if _, exists := m[named.Name]; exists {
				return nil, fmt.Errorf("resources: duplicate name %q", named.Name)
			}
			var res Resource
			if err := itemNode.Decode(&res); err != nil {
				return nil, fmt.Errorf("resources[%d]: %w", i, err)
			}
			m[named.Name] = res
		}
		return m, nil

	default:
		return nil, fmt.Errorf("resources must be a map or list")
	}
}
