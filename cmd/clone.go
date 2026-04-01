package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	cloneGroup    string
	cloneResource string
)

var cloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone repositories to local filesystem",
	Long:  "Clone repositories from configured groups and standalone repos. Repositories are cloned preserving directory hierarchy.",
	Example: `  grepom clone                           # Clone all repos
  grepom clone web-app                   # Clone a specific repo
  grepom clone --group frontend          # Clone all repos in a group
  grepom clone --resource work-gl        # Clone all repos from a resource`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{Group: cloneGroup, Resource: cloneResource}
		if len(args) > 0 {
			filter.Name = args[0]
		}

		resolver := repo.NewResolver(cfg)
		repos, err := resolver.ResolveAndFilter(filter)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories to clone.")
			return nil
		}

		for _, r := range repos {
			fullPath := repo.FullPath(cfg.Base, r)

			if gitpkg.IsCloned(fullPath) {
				if verbose {
					fmt.Printf("skip %s (already cloned)\n", r.Path)
				}
				continue
			}

			fmt.Printf("cloning %s...\n", r.Path)
			opts := gitpkg.CloneOptions{
				Token:          r.Token,
				Provider:       r.Provider,
				SSHKey:         r.SSHKey,
				HasGroupToken:  r.HasGroupToken,
				HasGroupSSHKey: r.HasGroupSSHKey,
			}
			if err := gitpkg.Clone(fullPath, r.SSHURL, r.CloneURL, opts); err != nil {
				fmt.Fprintf(os.Stderr, "error cloning %s: %v\n", r.Path, err)
				continue
			}
			fmt.Printf("  %s done\n", r.Name)
		}

		return nil
	},
}

func init() {
	cloneCmd.Flags().StringVar(&cloneGroup, "group", "", "clone all repos under a group")
	cloneCmd.Flags().StringVar(&cloneResource, "resource", "", "clone all repos from a resource")
	rootCmd.AddCommand(cloneCmd)
}
