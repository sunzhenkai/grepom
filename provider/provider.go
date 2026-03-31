package provider

import (
	"context"
	"fmt"

	"github.com/wii/grepom/config"
)

type Repo struct {
	Name     string
	CloneURL string
	SSHURL   string
	Path     string // relative path from base (includes group hierarchy)
	Provider string
}

type Provider interface {
	ListRepos(ctx context.Context, source config.Source) ([]Repo, error)
}

var registry = map[string]func() Provider{}

func Register(name string, factory func() Provider) {
	registry[name] = factory
}

func Get(name string) (Provider, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return factory(), nil
}

func AvailableProviders() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
