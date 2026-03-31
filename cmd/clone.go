package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	cloneGroup string
)

var cloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone repositories to local filesystem",
	Long:  "Clone repositories from configured sources to the base directory. Repositories are cloned preserving group/subgroup directory hierarchy.",
	Example: `  grepom clone                        # Clone all repos
  grepom clone web-app                 # Clone a specific repo
  grepom clone --group my-org/frontend # Clone all repos in a group`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{Group: cloneGroup}
		if len(args) > 0 {
			filter.Name = args[0]
		}

		resolver := repo.NewResolver(cfg)
		repos, err := resolver.ResolveAndFilter(context.Background(), filter)
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
			if err := gitpkg.Clone(fullPath, r.SSHURL, r.CloneURL); err != nil {
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
	rootCmd.AddCommand(cloneCmd)
}
