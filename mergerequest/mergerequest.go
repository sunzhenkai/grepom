package mergerequest

import (
	"context"
	"fmt"
)

// MergeRequest represents a created Merge Request / Pull Request.
type MergeRequest struct {
	ID           int
	Number       int    // GitHub: number, GitLab: iid
	Title        string
	Description  string
	URL          string // Web URL
	State        string // "open", "closed", "merged"
	SourceBranch string
	TargetBranch string
	Draft        bool
}

// CreateMergeRequestParams contains the parameters for creating a MR/PR.
type CreateMergeRequestParams struct {
	ServerURL    string
	Token        string
	RepoPath     string // "owner/repo" or "group/project"
	Title        string
	Description  string
	SourceBranch string
	TargetBranch string
	Draft        bool
}

// WebURLParams contains parameters for building a browser-based creation URL.
type WebURLParams struct {
	ServerURL    string
	RepoPath     string // "owner/repo" or "group/project"
	SourceBranch string
	TargetBranch string
	Title        string
	Draft        bool
}

// MergeRequestProvider is the interface for creating MR/PR on different platforms.
type MergeRequestProvider interface {
	CreateMergeRequest(ctx context.Context, params CreateMergeRequestParams) (*MergeRequest, error)
	BuildWebURL(params WebURLParams) string
}

// --- 注册表 ---

var registry = map[string]func() MergeRequestProvider{}

// Register 注册一个 MergeRequestProvider 工厂函数。
func Register(name string, factory func() MergeRequestProvider) {
	registry[name] = factory
}

// Get 通过名称获取 MergeRequestProvider 实例。
func Get(name string) (MergeRequestProvider, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported merge request provider: %s", name)
	}
	return factory(), nil
}
