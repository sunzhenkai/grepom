package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/provider"
	"github.com/wii/grepom/repo"
)

var (
	syncSource int
	syncGroup  string
	syncOrg    string
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize repositories and update configuration",
	Long: `Sync discovers new repositories from configured sources, clones new repos,
pulls existing repos, and appends newly discovered sub-groups to the config file.

Only new groups are added to the config; existing entries are never removed.`,
	Example: `  grepom sync                        # Sync all sources
  grepom sync --source 0              # Sync a specific source by index
  grepom sync --group my-org/frontend # Sync a specific group
  grepom sync --org my-org            # Sync a specific org`,
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
		sources := cfg.Sources
		if syncSource > 0 {
			if syncSource >= len(sources) {
				return fmt.Errorf("source index %d out of range (0-%d)", syncSource, len(sources)-1)
			}
			sources = sources[syncSource : syncSource+1]
		}

		var totalCloned, totalPulled, totalNewGroups int

		for si, source := range sources {
			sourceIdx := syncSource
			if syncSource <= 0 {
				sourceIdx = si
			}

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

			if verbose {
				fmt.Printf("source %d: found %d repos\n", sourceIdx, len(repos))
			}

			// Clone new repos, pull existing
			for _, r := range repos {
				fullPath := repo.FullPath(cfg.Base, r)

				if gitpkg.IsCloned(fullPath) {
					if verbose {
						fmt.Printf("pulling %s...\n", r.Path)
					}
					if err := gitpkg.Pull(fullPath); err != nil {
						fmt.Fprintf(os.Stderr, "error pulling %s: %v\n", r.Path, err)
						continue
					}
					totalPulled++
				} else {
					fmt.Printf("cloning %s...\n", r.Path)
					if err := gitpkg.Clone(fullPath, r.SSHURL, r.CloneURL); err != nil {
						fmt.Fprintf(os.Stderr, "error cloning %s: %v\n", r.Path, err)
						continue
					}
					totalCloned++
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
		fmt.Printf("sync complete: %d cloned, %d pulled, %d new groups\n", totalCloned, totalPulled, totalNewGroups)

		return nil
	},
}

func init() {
	syncCmd.Flags().IntVar(&syncSource, "source", -1, "sync a specific source by index")
	syncCmd.Flags().StringVar(&syncGroup, "group", "", "sync repos under a specific group path")
	syncCmd.Flags().StringVar(&syncOrg, "org", "", "sync repos under a specific org name")
	rootCmd.AddCommand(syncCmd)
}
