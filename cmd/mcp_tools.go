package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
	"github.com/wii/grepom/scanner"
)

// --- Input schemas ---

type grepomListInput struct {
	Group string `json:"group,omitempty" jsonschema_description:"optional group name to filter repos"`
}

type grepomStatusInput struct {
	Repo string `json:"repo,omitempty" jsonschema_description:"optional repo name; if omitted, returns status for all repos"`
}

type grepomSearchInput struct {
	Query string `json:"query" jsonschema_description:"search keyword (case-insensitive substring match)"`
}

type grepomDirInput struct {
	Repo string `json:"repo" jsonschema_description:"repo name to resolve local path for"`
}

type grepomPullInput struct {
	Repo string `json:"repo" jsonschema_description:"repo name to pull"`
}

type grepomCloneInput struct {
	Repo string `json:"repo" jsonschema_description:"repo name to clone"`
}

type grepomScanInput struct {
	Repo string `json:"repo" jsonschema_description:"repo name to scan for secrets"`
}

// --- Result helpers ---

type mcpEmptyOut struct{}

var mcpEmpty = mcpEmptyOut{}

func mcpTextResult(payload any) (*mcp.CallToolResult, mcpEmptyOut, error) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, mcpEmpty, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, mcpEmpty, nil
}

func mcpErrResult(err error) (*mcp.CallToolResult, mcpEmptyOut, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
	}, mcpEmpty, nil
}


// --- Handlers ---

func handleGrepomList(_ context.Context, _ *mcp.CallToolRequest, in grepomListInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	filter := repo.Filter{Group: in.Group}
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		return mcpErrResult(err)
	}

	type entry struct {
		Name     string `json:"name"`
		Group    string `json:"group"`
		Path     string `json:"path"`
		Resource string `json:"resource,omitempty"`
	}
	out := make([]entry, 0, len(repos))
	for _, r := range repos {
		out = append(out, entry{Name: r.Name, Group: r.GroupName, Path: r.Path, Resource: r.Resource})
	}
	return mcpTextResult(out)
}

func handleGrepomStatus(_ context.Context, _ *mcp.CallToolRequest, in grepomStatusInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	filter := repo.Filter{}
	if in.Repo != "" {
		filter.Name = in.Repo
	}
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		return mcpErrResult(err)
	}
	if in.Repo != "" && len(repos) == 0 {
		return mcpErrResult(fmt.Errorf("repo %q not found in config", in.Repo))
	}

	type entry struct {
		Name   string `json:"name"`
		Branch string `json:"branch,omitempty"`
		Status string `json:"status"`
		Ahead  int    `json:"ahead,omitempty"`
		Behind int    `json:"behind,omitempty"`
		Dirty  int    `json:"dirty,omitempty"`
	}
	out := make([]entry, 0, len(repos))
	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		st := git.GetStatus(fullPath)

		e := entry{Name: r.Name}
		switch {
		case !st.Cloned || st.NotARepo:
			e.Status = "not cloned"
		case !st.Clean:
			e.Status = "dirty"
			e.Dirty = st.Dirty
			e.Branch = st.Branch
		case st.Ahead > 0:
			e.Status = "ahead"
			e.Ahead = st.Ahead
			e.Branch = st.Branch
		case st.Behind > 0:
			e.Status = "behind"
			e.Behind = st.Behind
			e.Branch = st.Branch
		default:
			e.Status = "clean"
			e.Branch = st.Branch
		}
		out = append(out, e)
	}
	return mcpTextResult(out)
}

func handleGrepomSearch(_ context.Context, _ *mcp.CallToolRequest, in grepomSearchInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	if in.Query == "" {
		return mcpErrResult(fmt.Errorf("missing required parameter: query"))
	}

	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	allRepos, err := resolver.Resolve()
	if err != nil {
		return mcpErrResult(err)
	}

	results := repo.ApplySearchFilter(allRepos, in.Query, repo.Filter{})

	type entry struct {
		Name  string `json:"name"`
		Group string `json:"group"`
		Path  string `json:"path"`
	}
	out := make([]entry, 0, len(results))
	for _, r := range results {
		out = append(out, entry{Name: r.Name, Group: r.GroupName, Path: r.Path})
	}
	return mcpTextResult(out)
}

func handleGrepomDir(_ context.Context, _ *mcp.CallToolRequest, in grepomDirInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	if in.Repo == "" {
		return mcpErrResult(fmt.Errorf("missing required parameter: repo"))
	}

	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	allRepos, err := resolver.Resolve()
	if err != nil {
		return mcpErrResult(err)
	}

	results := repo.ApplyExactFirstSearch(allRepos, in.Repo, repo.Filter{})
	if len(results) == 0 {
		return mcpErrResult(fmt.Errorf("no repo found matching %q", in.Repo))
	}

	type entry struct {
		Name      string `json:"name"`
		LocalPath string `json:"local_path"`
		Cloned    bool   `json:"cloned"`
	}
	out := make([]entry, 0, len(results))
	for _, r := range results {
		fullPath := repo.FullPath(cfg.Base, r)
		out = append(out, entry{Name: r.Name, LocalPath: fullPath, Cloned: git.IsCloned(fullPath)})
	}
	return mcpTextResult(out)
}

func handleGrepomPull(_ context.Context, _ *mcp.CallToolRequest, in grepomPullInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	if in.Repo == "" {
		return mcpErrResult(fmt.Errorf("missing required parameter: repo"))
	}

	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	allRepos, err := resolver.Resolve()
	if err != nil {
		return mcpErrResult(err)
	}

	results := repo.ApplyExactFirstSearch(allRepos, in.Repo, repo.Filter{})
	if len(results) == 0 {
		return mcpErrResult(fmt.Errorf("no repo found matching %q", in.Repo))
	}

	r := results[0]
	fullPath := repo.FullPath(cfg.Base, r)
	if !git.IsCloned(fullPath) {
		return mcpErrResult(fmt.Errorf("repo %q is not cloned yet; use grepom_clone first", in.Repo))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", fullPath, "pull")
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return mcpErrResult(fmt.Errorf("git pull timed out after 60s for repo %q", in.Repo))
	}
	if err != nil {
		return mcpErrResult(fmt.Errorf("git pull failed for %q: %s", in.Repo, strings.TrimSpace(string(output))))
	}

	return mcpTextResult(map[string]string{
		"status": "ok",
		"repo":   r.Name,
		"output": strings.TrimSpace(string(output)),
	})
}

func handleGrepomClone(_ context.Context, _ *mcp.CallToolRequest, in grepomCloneInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	if in.Repo == "" {
		return mcpErrResult(fmt.Errorf("missing required parameter: repo"))
	}

	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	allRepos, err := resolver.Resolve()
	if err != nil {
		return mcpErrResult(err)
	}

	results := repo.ApplyExactFirstSearch(allRepos, in.Repo, repo.Filter{})
	if len(results) == 0 {
		return mcpErrResult(fmt.Errorf("repo %q not found in config", in.Repo))
	}

	r := results[0]
	fullPath := repo.FullPath(cfg.Base, r)
	if git.IsCloned(fullPath) {
		return mcpTextResult(map[string]string{
			"status":     "already cloned",
			"repo":       r.Name,
			"local_path": fullPath,
		})
	}

	err = git.Clone(fullPath, r.SSHURL, r.CloneURL, git.CloneOptions{
		Token:          r.Token,
		Provider:       r.Provider,
		SSHKey:         r.SSHKey,
		HasGroupToken:  r.HasGroupToken,
		HasGroupSSHKey: r.HasGroupSSHKey,
		PreferHTTPS:    r.PreferHTTPS,
	})
	if err != nil {
		return mcpErrResult(fmt.Errorf("clone failed for %q: %w", in.Repo, err))
	}

	return mcpTextResult(map[string]string{
		"status":     "cloned",
		"repo":       r.Name,
		"local_path": fullPath,
	})
}

func handleGrepomGroups(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, mcpEmptyOut, error) {
	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	type entry struct {
		Name      string `json:"name"`
		Resource  string `json:"resource,omitempty"`
		Path      string `json:"path,omitempty"`
		RepoCount int    `json:"repo_count"`
	}
	out := make([]entry, 0, len(cfg.Groups))
	for _, g := range cfg.Groups {
		out = append(out, entry{
			Name:      g.Name,
			Resource:  g.Resource,
			Path:      g.Path,
			RepoCount: len(g.Repos),
		})
	}
	return mcpTextResult(out)
}

func handleGrepomScan(_ context.Context, _ *mcp.CallToolRequest, in grepomScanInput) (*mcp.CallToolResult, mcpEmptyOut, error) {
	if in.Repo == "" {
		return mcpErrResult(fmt.Errorf("missing required parameter: repo"))
	}

	_, cfg, err := loadConfig()
	if err != nil {
		return mcpErrResult(fmt.Errorf("failed to load config: %w", err))
	}

	resolver := repo.NewResolver(cfg)
	allRepos, err := resolver.Resolve()
	if err != nil {
		return mcpErrResult(err)
	}

	results := repo.ApplyExactFirstSearch(allRepos, in.Repo, repo.Filter{})
	if len(results) == 0 {
		return mcpErrResult(fmt.Errorf("no repo found matching %q", in.Repo))
	}

	r := results[0]
	fullPath := repo.FullPath(cfg.Base, r)
	if !git.IsCloned(fullPath) {
		return mcpErrResult(fmt.Errorf("repo %q is not cloned; cannot scan", in.Repo))
	}

	s := scanner.NewScanner(scanner.Options{})
	findings, err := s.ScanDir(context.Background(), fullPath)
	if err != nil {
		return mcpErrResult(fmt.Errorf("scan failed for %q: %w", in.Repo, err))
	}

	type finding struct {
		File   string `json:"file"`
		Line   int    `json:"line"`
		RuleID string `json:"rule_id"`
	}
	out := make([]finding, 0, len(findings))
	for _, f := range findings {
		out = append(out, finding{File: f.File, Line: f.Line, RuleID: f.RuleID})
	}
	return mcpTextResult(map[string]any{
		"repo":     r.Name,
		"findings": out,
		"count":    len(out),
	})
}

// --- Registration ---

type grepomToolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func registerGrepomTools(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_list", Description: "List repositories from config, optionally filtered by group."}, handleGrepomList)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_status", Description: "Show git status for repos (branch, dirty, ahead/behind). Pass repo for a single repo."}, handleGrepomStatus)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_search", Description: "Search repositories by name (case-insensitive substring match)."}, handleGrepomSearch)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_dir", Description: "Get the local directory path for a repository."}, handleGrepomDir)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_pull", Description: "Run git pull for a specific repository (60s timeout)."}, handleGrepomPull)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_clone", Description: "Clone a repository to local filesystem."}, handleGrepomClone)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_groups", Description: "List all configured groups with repo counts."}, handleGrepomGroups)
	mcp.AddTool(s, &mcp.Tool{Name: "grepom_scan", Description: "Scan a repository for secrets (SSH keys, tokens, passwords)."}, handleGrepomScan)
}

func grepomToolCatalogue() []grepomToolDef {
	return []grepomToolDef{
		{"grepom_list", "List repositories from config, optionally filtered by group."},
		{"grepom_status", "Show git status for repos (branch, dirty, ahead/behind)."},
		{"grepom_search", "Search repositories by name (case-insensitive substring match)."},
		{"grepom_dir", "Get the local directory path for a repository."},
		{"grepom_pull", "Run git pull for a specific repository (60s timeout)."},
		{"grepom_clone", "Clone a repository to local filesystem."},
		{"grepom_groups", "List all configured groups with repo counts."},
		{"grepom_scan", "Scan a repository for secrets."},
	}
}
