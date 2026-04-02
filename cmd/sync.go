package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/provider"
)

var (
	syncGroup    string
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
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		configPath, err := resolvedConfigPath()
		if err != nil {
			configPath = configFile
		}

		// Determine which groups to process
		var groupsToProcess []config.Group
		for _, g := range cfg.Groups {
			if syncGroup != "" && g.Name != syncGroup {
				continue
			}
			if syncResource != "" && g.Resource != syncResource {
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

		var totalRepos, totalNewRepos int

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

			params := provider.ListReposParams{
				ServerURL: res.APIURL(),
				Token:     res.Token,
			}

			// GitLab: use Groups query; GitHub: use Orgs
			if res.Provider == "github" {
				params.Orgs = []string{g.Path}
			} else {
				params.Groups = []provider.GroupQuery{{Path: g.Path, Recursive: g.Recursive}}
			}

			repos, err := p.ListRepos(context.Background(), params)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: group %q: %v\n", g.Name, err)
				continue
			}

			totalRepos += len(repos)
			if verbose {
				fmt.Printf("group %q: found %d repos\n", g.Name, len(repos))
			}

			// Convert discovered repos to GroupRepo entries
			var newGroupRepos []config.GroupRepo
			for _, r := range repos {
				newGroupRepos = append(newGroupRepos, config.GroupRepo{
					Name: r.Name,
					URL:  r.CloneURL,
					Path: r.Path,
				})
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

		fmt.Printf("sync complete: %d repos discovered, %d new repos saved\n", totalRepos, totalNewRepos)
		if totalNewRepos > 0 {
			fmt.Println("Run 'grepom clone' to clone new repositories.")
		}

		return nil
	},
}

func init() {
	syncCmd.Flags().StringVarP(&syncGroup, "group", "g", "", "sync a specific group by name")
	syncCmd.Flags().StringVarP(&syncResource, "resource", "R", "", "sync all groups using a specific resource")
	rootCmd.AddCommand(syncCmd)
}
