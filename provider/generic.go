package provider

import "context"

func init() {
	Register("generic", func() Provider { return &GenericProvider{} })
}

type GenericProvider struct{}

func (g *GenericProvider) ListRepos(_ context.Context, _ ListReposParams) ([]Repo, error) {
	return nil, nil
}

func (g *GenericProvider) ListGroups(_ context.Context, _ ListGroupsParams) ([]RemoteGroup, error) {
	return nil, nil
}
