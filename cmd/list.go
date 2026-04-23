package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/provider"
	"github.com/wii/grepom/repo"
)

var (
	listGroup    string
	listResource string
	listType     string
	listRemote   bool
	listAll      bool
	listNoPush   bool
	listNoCommit bool
)

var listCmd = &cobra.Command{
	Use:   "list [name|groups|resources]",
	Short: "List resources, groups, or repositories",
	Long: `List resources, groups, or repositories based on the --type flag.

By default, lists only repos that need attention (unpushed commits or uncommitted changes).
Use --all to list all repos including clean ones.
Use --type resources to list configured resources, or --type groups to list configured groups.
Use --remote to query repos directly from the provider API.

Shortcut: use "groups" or "resources" as a positional argument instead of --type.
E.g. "grepom list groups" is equivalent to "grepom list --type groups".`,
	Example: `  grepom list                            # List repos needing attention (unpushed/uncommitted)
  grepom list --all                      # List all repos with status
  grepom list --no-push                  # List only repos with unpushed commits
  grepom list --no-commit                # List only repos with uncommitted changes
  grepom list web-app                     # List a specific repo
  grepom list --group frontend            # List attention repos in a group
  grepom list --resource work-gl          # List attention repos from a resource
  grepom list groups                      # List all configured groups
  grepom list resources                   # List all configured resources
  grepom list --remote                    # List remote repos from provider API
  grepom list --all --remote --group fe   # List all remote repos for a group`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := loadConfig()
		if err != nil {
			return err
		}

		// 位置参数关键字 groups/resources 快捷方式（仅当 --type 未被显式设置时生效）
		if len(args) > 0 && !cmd.Flags().Changed("type") {
			switch args[0] {
			case "groups":
				listType = "groups"
			case "resources":
				listType = "resources"
			}
		}

		// --remote 不支持 --type resources
		if listRemote && listType == "resources" {
			return fmt.Errorf("--remote is not supported with --type resources")
		}

		switch listType {
		case "resources":
			return runListResources(cfg)
		case "groups":
			if listRemote {
				return runListRemoteGroups(cfg)
			}
			return runListGroups(cfg)
		default:
			if listRemote {
				return runListRemoteRepos(cfg)
			}
			return runListRepos(cfg, args)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&listGroup, "group", "g", "", "filter by group name")
	listCmd.Flags().StringVarP(&listResource, "resource", "R", "", "filter by resource name")
	listCmd.Flags().StringVarP(&listType, "type", "t", "repos", "type to list: repos, resources, groups")
	listCmd.Flags().BoolVarP(&listRemote, "remote", "r", false, "list remote repos from provider API instead of local config")
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "list all repos (disable attention filter, include disabled/excluded)")
	listCmd.Flags().BoolVar(&listNoPush, "no-push", false, "only show repos with unpushed commits (ahead > 0)")
	listCmd.Flags().BoolVar(&listNoCommit, "no-commit", false, "only show repos with uncommitted changes (dirty > 0)")
	rootCmd.AddCommand(listCmd)
}

func runListRepos(cfg *config.Config, args []string) error {
	filter := repo.Filter{
		Group:           listGroup,
		Resource:        listResource,
		IncludeDisabled: listAll,
	}
	// 位置参数：当为 groups/resources 时已被 RunE 处理，此处仅处理仓库名过滤
	if len(args) > 0 && args[0] != "groups" && args[0] != "resources" {
		filter.Name = args[0]
	}

	resolver := repo.NewResolver(cfg)
	repos, err := resolver.ResolveAndFilter(filter)
	if err != nil {
		return err
	}

	// 为每个 repo 计算 git 状态
	type repoWithStatus struct {
		repo   provider.Repo
		status string
	}

	var entries []repoWithStatus
	for _, r := range repos {
		fullPath := repo.FullPath(cfg.Base, r)
		st := gitpkg.GetStatus(fullPath)

		var label string
		switch {
		case !st.Cloned || st.NotARepo:
			label = "not cloned"
		case !st.Clean:
			label = fmt.Sprintf("dirty(%d)", st.Dirty)
		case st.Ahead > 0:
			label = fmt.Sprintf("ahead %d", st.Ahead)
		case st.Behind > 0:
			label = fmt.Sprintf("behind %d", st.Behind)
		default:
			label = "clean"
		}

		entries = append(entries, repoWithStatus{repo: r, status: label})
	}

	// 状态过滤
	if !listAll {
		var filtered []repoWithStatus
		for _, e := range entries {
			if listNoPush || listNoCommit {
				// 显式指定了 --no-push / --no-commit：按具体条件过滤
				if e.status == "not cloned" {
					continue
				}
				match := false
				if listNoPush && strings.HasPrefix(e.status, "ahead") {
					match = true
				}
				if listNoCommit && strings.HasPrefix(e.status, "dirty") {
					match = true
				}
				if match {
					filtered = append(filtered, e)
				}
			} else {
				// 默认 attention 模式：只展示 dirty 或 ahead 的 repo
				if strings.HasPrefix(e.status, "dirty") || strings.HasPrefix(e.status, "ahead") {
					filtered = append(filtered, e)
				}
			}
		}
		entries = filtered
	}

	if len(entries) == 0 {
		fmt.Println("No repositories found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tGROUP\tRESOURCE\tSTATUS")

	for _, e := range entries {
		name := e.repo.Name
		if listAll && e.repo.DisabledReason != "" {
			switch e.repo.DisabledReason {
			case "disabled":
				name = name + " [disabled]"
			case "excluded":
				name = name + " [excluded]"
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, e.repo.Path, e.repo.GroupName, e.repo.Resource, e.status)
	}
	w.Flush()

	return nil
}

func runListResources(cfg *config.Config) error {
	if len(cfg.Resources) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPROVIDER\tURL\tSSH_KEY")

	for name, r := range cfg.Resources {
		sshKey := "-"
		if r.SSHKey != "" {
			sshKey = r.SSHKey
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, r.Provider, r.URL, sshKey)
	}
	w.Flush()

	return nil
}

func runListGroups(cfg *config.Config) error {
	if len(cfg.Groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tRESOURCE\tPATH\tLOCAL_PATH\tRECURSIVE\tREPOS")

	for _, g := range cfg.Groups {
		recursive := "no"
		if g.Recursive {
			recursive = "yes"
		}
		localPath := g.LocalPath
		if localPath == "" {
			localPath = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n", g.Name, g.Resource, g.Path, localPath, recursive, len(g.Repos))
	}
	w.Flush()

	return nil
}

// runListRemoteRepos 通过 provider API 实时查询远程仓库列表并展示
func runListRemoteRepos(cfg *config.Config) error {
	// 收集需要查询的 groups
	type groupInfo struct {
		group      config.Group
		isDisabled bool // group 或 resource 被禁用
	}
	var groupsToProcess []groupInfo
	for _, g := range cfg.Groups {
		if listGroup != "" && g.Name != listGroup {
			continue
		}
		if listResource != "" && g.Resource != listResource {
			continue
		}
		// Skip groups without resource for remote listing
		if g.Resource == "" {
			fmt.Printf("group %q: no resource bound, skipping remote listing\n", g.Name)
			continue
		}
		// 非 --all 模式下跳过禁用的 group 和 resource
		if !listAll {
			if !g.IsEnabled() {
				continue
			}
			if res, ok := cfg.Resources[g.Resource]; ok && !res.IsEnabled() {
				continue
			}
		}
		isDisabled := !g.IsEnabled()
		if res, ok := cfg.Resources[g.Resource]; ok && !res.IsEnabled() {
			isDisabled = true
		}
		groupsToProcess = append(groupsToProcess, groupInfo{group: g, isDisabled: isDisabled})
	}

	if len(groupsToProcess) == 0 {
		fmt.Println("No remote repositories found.")
		return nil
	}

	// 查询远程仓库
	type remoteEntry struct {
		name       string
		path       string
		group      string
		resource   string
		cloneURL   string
		isExcluded bool
		isDisabled bool
	}

	var entries []remoteEntry

	for _, gi := range groupsToProcess {
		g := gi.group
		res, ok := cfg.Resources[g.Resource]
		if !ok {
			fmt.Fprintf(os.Stderr, "error: group %q: resource %q not found\n", g.Name, g.Resource)
			continue
		}

		p, err := provider.Get(res.Provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
			continue
		}

		resolvedToken, err := res.ResolvedToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: group %q (resource %q): %v\n", g.Name, g.Resource, err)
			continue
		}

		params := provider.ListReposParams{
			ServerURL:      res.APIURL(),
			Token:          resolvedToken,
			OrganizationID: res.OrganizationID,
		}

		// GitLab/Codeup: 使用 Groups 查询；GitHub: 使用 Orgs
		if res.Provider == "github" {
			params.Orgs = []string{g.Path}
		} else {
			params.Groups = []provider.GroupQuery{{Path: g.Path, Recursive: g.Recursive}}
		}

		repos, err := p.ListRepos(context.Background(), params)
		if err != nil && res.Scheme() == "" && isConnectionError(err) {
			config.Verbose("resource %q: HTTPS connection failed, trying HTTP", g.Resource)
			params.ServerURL = buildHTTPURL(res.APIURL())
			repos, err = p.ListRepos(context.Background(), params)
			if err == nil {
				warnHTTPFallback(g.Resource)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
			continue
		}

		for _, r := range repos {
			excluded := repo.IsExcluded(g.ExcludeRepos, r.Name, r.Path)
			// 非 --all 模式下跳过被排除的仓库
			if !listAll && excluded {
				continue
			}
			cloneURL := r.CloneURL
			if cloneURL == "" {
				cloneURL = r.SSHURL
			}
			entries = append(entries, remoteEntry{
				name:       r.Name,
				path:       r.Path,
				group:      g.Name,
				resource:   g.Resource,
				cloneURL:   cloneURL,
				isExcluded: excluded,
				isDisabled: gi.isDisabled,
			})
		}
	}

	if len(entries) == 0 {
		fmt.Println("No remote repositories found.")
		return nil
	}

	// 表格输出
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPATH\tGROUP\tRESOURCE\tCLONE_URL")

	for _, e := range entries {
		name := e.name
		if listAll {
			if e.isDisabled {
				name = name + " [disabled]"
			} else if e.isExcluded {
				name = name + " [excluded]"
			}
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, e.path, e.group, e.resource, e.cloneURL)
	}
	w.Flush()

	return nil
}

// runListRemoteGroups 通过 provider API 实时查询远程 groups/orgs 列表并展示
func runListRemoteGroups(cfg *config.Config) error {
	// 收集需要查询的 resources
	var resourcesToQuery []struct {
		name string
		res  config.Resource
	}
	for name, res := range cfg.Resources {
		if listResource != "" && name != listResource {
			continue
		}
		resourcesToQuery = append(resourcesToQuery, struct {
			name string
			res  config.Resource
		}{name: name, res: res})
	}

	if len(resourcesToQuery) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	// 查询远程 groups
	type remoteGroupEntry struct {
		name     string
		resource string
		path     string
	}

	var entries []remoteGroupEntry

	for _, rq := range resourcesToQuery {
		p, err := provider.Get(rq.res.Provider)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: resource %q: %v\n", rq.name, err)
			continue
		}

		resolvedToken, err := rq.res.ResolvedToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: resource %q: %v\n", rq.name, err)
			continue
		}

		params := provider.ListGroupsParams{
			ServerURL:      rq.res.APIURL(),
			Token:          resolvedToken,
			OrganizationID: rq.res.OrganizationID,
		}

		groups, err := p.ListGroups(context.Background(), params)
		if err != nil && rq.res.Scheme() == "" && isConnectionError(err) {
			config.Verbose("resource %q: HTTPS connection failed, trying HTTP", rq.name)
			params.ServerURL = buildHTTPURL(rq.res.APIURL())
			groups, err = p.ListGroups(context.Background(), params)
			if err == nil {
				warnHTTPFallback(rq.name)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: resource %q: %v\n", rq.name, err)
			continue
		}

		for _, g := range groups {
			entries = append(entries, remoteGroupEntry{
				name:     g.Name,
				resource: rq.name,
				path:     g.Path,
			})
		}
	}

	// 按 --group 过滤
	if listGroup != "" {
		var filtered []remoteGroupEntry
		for _, e := range entries {
			if e.name == listGroup {
				filtered = append(filtered, e)
			}
		}
		entries = filtered
	}

	if len(entries) == 0 {
		fmt.Println("No remote groups found.")
		return nil
	}

	// 表格输出
	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tRESOURCE\tPATH")

	for _, e := range entries {
		fmt.Fprintf(w, "%s\t%s\t%s\n", e.name, e.resource, e.path)
	}
	w.Flush()

	return nil
}
