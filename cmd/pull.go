package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var pullCmd = &cobra.Command{
	Use:   "pull [name]",
	Short: "Pull updates for repositories",
	Long:  "Run git pull on cloned repositories. Skips repositories that have not been cloned yet.",
	Example: `  grepom pull           # Pull all cloned repos
  grepom pull web-app    # Pull a specific repo`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{}
		if len(args) > 0 {
			filter.Name = args[0]
		}

		resolver := repo.NewResolver(cfg)
		repos, err := resolver.ResolveAndFilter(context.Background(), filter)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories found.")
			return nil
		}

		for _, r := range repos {
			fullPath := repo.FullPath(cfg.Base, r)

			if !gitpkg.IsCloned(fullPath) {
				fmt.Printf("skip %s (not cloned)\n", r.Path)
				continue
			}

			fmt.Printf("pulling %s...\n", r.Path)
			if err := gitpkg.Pull(fullPath); err != nil {
				fmt.Fprintf(os.Stderr, "error pulling %s: %v\n", r.Path, err)
				continue
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
