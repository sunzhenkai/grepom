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
	syncSource string
	syncGroup  string
	syncOrg    string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize repository metadata and update configuration",
	Long: `Sync discovers new repositories and sub-groups from configured API sources,
and saves the discovered information to the config file.

This command only updates configuration metadata — it does NOT clone or pull
repositories. After running sync, use "grepom clone" to clone newly discovered
repos and "grepom pull" to update existing ones.

Only new groups and repos are added to the config; existing entries are never removed.`,
	Example: `  grepom sync                            # Sync all sources
  grepom sync --source my-gitlab         # Sync a specific source by name
  grepom sync --source 0                 # Sync a specific source by index
  grepom sync --group my-org/frontend    # Sync a specific group
  grepom sync --org my-org               # Sync a specific org`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		configPath, err := resolvedConfigPath()
		if err != nil {
			configPath = configFile
		}

		// Determine which sources to process
		type sourceInfo struct {
			index  int
			source config.Source
		}
		var sourcesToProcess []sourceInfo

		if syncSource != "" {
			idx, src, err := cfg.FindSource(syncSource)
			if err != nil {
				return err
			}
			sourcesToProcess = append(sourcesToProcess, sourceInfo{index: idx, source: *src})
		} else {
			for i, s := range cfg.Sources {
				sourcesToProcess = append(sourcesToProcess, sourceInfo{index: i, source: s})
			}
		}

		var totalRepos, totalNewRepos, totalNewGroups int

		for _, si := range sourcesToProcess {
			sourceIdx := si.index
			source := si.source

			p, err := provider.Get(source.Provider)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: source %d: %v\n", sourceIdx, err)
				continue
			}

			// Build filtered groups/orgs for this source
			groupsToSync := source.Groups
			if syncGroup != "" {
				groupsToSync = nil
				for _, g := range source.Groups {
					if g.Path == syncGroup || len(syncGroup) < len(g.Path) && g.Path[:len(syncGroup)] == syncGroup && g.Path[len(syncGroup)] == '/' {
						groupsToSync = append(groupsToSync, g)
					}
				}
			}

			orgsToSync := source.Orgs
			if syncOrg != "" {
				orgsToSync = nil
				for _, o := range source.Orgs {
					if o.Name == syncOrg {
						orgsToSync = append(orgsToSync, o)
					}
				}
			}

			// Discover repos for this source using filtered groups/orgs
			filteredSource := source
			filteredSource.Groups = groupsToSync
			filteredSource.Orgs = orgsToSync

			repos, err := p.ListRepos(context.Background(), filteredSource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: source %d: %v\n", sourceIdx, err)
				continue
			}

			totalRepos += len(repos)
			if verbose {
				sourceLabel := fmt.Sprintf("source %d", sourceIdx)
				if source.Name != "" {
					sourceLabel = fmt.Sprintf("source %q", source.Name)
				}
				fmt.Printf("%s: found %d repos\n", sourceLabel, len(repos))
			}

			// Convert discovered repos to RepoEntry list for config
			var newRepoEntries []config.RepoEntry
			for _, r := range repos {
				newRepoEntries = append(newRepoEntries, config.RepoEntry{
					Name: r.Name,
					URL:  r.CloneURL,
					Path: r.Path,
				})
			}

			// Save discovered repos to config
			if len(newRepoEntries) > 0 {
				added, err := config.SyncRepos(configPath, newRepoEntries)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error updating repos in config: %v\n", err)
				} else if added > 0 {
					totalNewRepos += added
					if verbose {
						fmt.Printf("added %d new repos to config\n", added)
					}
				}
			}

			// Discover sub-groups for recursive GitLab groups
			sgl, ok := p.(provider.SubGroupLister)
			if !ok {
				continue
			}

			var newGroups []config.GroupSource
			for _, g := range groupsToSync {
				if !g.Recursive {
					continue
				}

				subPaths, err := sgl.ListSubGroups(context.Background(), source, g.Path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: could not discover sub-groups for %s: %v\n", g.Path, err)
					continue
				}

				// Check which sub-groups are not in config
				for _, sp := range subPaths {
					found := false
					for _, eg := range source.Groups {
						if eg.Path == sp {
							found = true
							break
						}
					}
					if !found {
						newGroups = append(newGroups, config.GroupSource{
							Path:      sp,
							Recursive: true,
						})
					}
				}
			}

			if len(newGroups) > 0 {
				if err := config.SyncGroups(configPath, sourceIdx, newGroups); err != nil {
					fmt.Fprintf(os.Stderr, "error updating config: %v\n", err)
					continue
				}
				totalNewGroups += len(newGroups)
				if verbose {
					for _, ng := range newGroups {
						fmt.Printf("added group %s to config\n", ng.Path)
					}
				}
			}
		}

		// Summary
		fmt.Printf("sync complete: %d repos discovered, %d new repos saved, %d new groups\n", totalRepos, totalNewRepos, totalNewGroups)
		if totalNewRepos > 0 {
			fmt.Println("Run 'grepom clone' to clone new repositories.")
		}

		return nil
	},
}

func init() {
	syncCmd.Flags().StringVar(&syncSource, "source", "", "sync a specific source by name or index")
	syncCmd.Flags().StringVar(&syncGroup, "group", "", "sync repos under a specific group path")
	syncCmd.Flags().StringVar(&syncOrg, "org", "", "sync repos under a specific org name")
	rootCmd.AddCommand(syncCmd)
}
