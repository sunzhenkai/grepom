package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/provider"
	"github.com/wii/grepom/repo"
)

var (
	syncGroup    string
	syncVGroup   string
	syncResource string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize repository metadata from remote groups",
	Long: `Sync discovers new repositories from configured remote groups
and saves the discovered information to the config file.

This command only updates configuration metadata — it does NOT clone or pull
repositories. After running sync, use "grepom clone" to clone newly discovered
repos and "grepom pull" to update existing ones.

Only new repos are added to the config; existing entries are never removed.`,
	Example: `  grepom sync                          # Sync all groups
  grepom sync --group frontend         # Sync a specific group by name
  grepom sync --resource work-gl       # Sync all groups using a specific resource`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := loadConfig()
		if err != nil {
			return err
		}

		configPath, err := resolvedConfigPath()
		if err != nil {
			configPath = defaultConfigPath()
		}

		groupSelection, err := cfg.ResolveGroupSelection(syncGroup, syncVGroup)
		if err != nil {
			return err
		}

		// Determine which groups to process
		var groupsToProcess []config.Group
		for _, g := range cfg.Groups {
			if !cfg.GroupInSelection(g.Name, groupSelection) {
				continue
			}
			if syncResource != "" && g.Resource != syncResource {
				continue
			}
			// Skip groups without resource (manual management)
			if g.Resource == "" {
				fmt.Printf("group %q: no resource bound, skipping sync\n", g.Name)
				continue
			}
			// Skip disabled groups (4.1)
			if !g.IsEnabled() {
				config.Verbose("skipping disabled group %q", g.Name)
				continue
			}
			// Skip groups whose resource is disabled (4.2)
			if res, ok := cfg.Resources[g.Resource]; ok && !res.IsEnabled() {
				config.Verbose("skipping group %q (resource %q is disabled)", g.Name, g.Resource)
				continue
			}
			groupsToProcess = append(groupsToProcess, g)
		}

		var totalRepos, totalNewRepos, totalExcluded int

		for _, g := range groupsToProcess {
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

			config.Verbose("syncing group %q (resource: %s, path: %s)", g.Name, g.Resource, g.Path)

			serverURL := res.APIURL()

			// GitLab/Codeup: use Groups query; GitHub: use Orgs
			var groups []provider.GroupQuery
			var orgs []string
			if res.Provider == "github" {
				orgs = []string{g.Path}
			} else {
				groups = []provider.GroupQuery{{Path: g.Path, Recursive: g.Recursive}}
			}

			resolvedToken, err := res.ResolvedToken()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: group %q (resource %q): %v\n", g.Name, g.Resource, err)
				continue
			}

			params := provider.ListReposParams{
				ServerURL:      serverURL,
				Token:          resolvedToken,
				Groups:         groups,
				Orgs:           orgs,
				OrganizationID: res.OrganizationID,
			}

			repos, err := p.ListRepos(context.Background(), params)
			if err != nil && res.Scheme() == "" && isConnectionError(err) {
				// auto 模式：HTTPS 连接失败，尝试 HTTP
				config.Verbose("resource %q: HTTPS connection failed, trying HTTP", g.Resource)
				params.ServerURL = buildHTTPURL(serverURL)
				repos, err = p.ListRepos(context.Background(), params)
				if err == nil {
					warnHTTPFallback(g.Resource)
				}
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
				continue
			}

			totalRepos += len(repos)
			if verbose {
				fmt.Printf("group %q: found %d repos\n", g.Name, len(repos))
			}

			// Convert discovered repos to GroupRepo entries, skipping excluded repos
			var newGroupRepos []config.GroupRepo
			var excludedCount int
			for _, r := range repos {
				if repo.IsExcluded(g.ExcludeRepos, r.Name, r.Path) {
					excludedCount++
					continue
				}
				// Check for duplicates within the current batch
				exists := false
				for _, existing := range newGroupRepos {
					if existing.URL == r.CloneURL {
						exists = true
						break
					}
				}
				if exists {
					continue
				}
				newGroupRepos = append(newGroupRepos, config.GroupRepo{
					Name: r.Name,
					URL:  r.CloneURL,
					Path: r.Path,
				})
			}
			totalExcluded += excludedCount
			if verbose && excludedCount > 0 {
				fmt.Printf("  skipped %d excluded repos in group %q\n", excludedCount, g.Name)
			}

			// Save discovered repos to this group in config
			if len(newGroupRepos) > 0 {
				added, err := config.SyncGroupRepos(configPath, g.Name, newGroupRepos)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
				} else if added > 0 {
					totalNewRepos += added
					if verbose {
						fmt.Printf("  added %d new repos to group %q\n", added, g.Name)
					}
				}
			}
		}

		fmt.Printf("sync complete: %d repos discovered, %d excluded, %d new repos saved\n", totalRepos, totalExcluded, totalNewRepos)
		if totalNewRepos > 0 {
			fmt.Println("Run 'grepom clone' to clone new repositories.")
		}

		return nil
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncGroup, "group", "g", "", "sync a specific group by name")
	syncCmd.Flags().StringVar(&syncVGroup, "vgroup", "", "sync groups in a virtual group")
	syncCmd.Flags().StringVarP(&syncResource, "resource", "R", "", "sync all groups using a specific resource")
	rootCmd.AddCommand(syncCmd)
}
