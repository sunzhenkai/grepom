package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	statusGroup    string
	statusResource string
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show git status for repositories",
	Long:  "Show git status (branch, clean/dirty, ahead/behind) for cloned repositories.",
	Example: `  grepom status           # Status of all repos
  grepom status web-app    # Status of a specific repo
  grepom status --group frontend  # Status of repos in a group`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Group:    statusGroup,
			Resource: statusResource,
		}
		if len(args) > 0 {
			filter.Name = args[0]
		}

		resolver := repo.NewResolver(cfg)
		repos, err := resolver.ResolveAndFilter(filter)
		if err != nil {
			return err
		}

		if len(repos) == 0 {
			fmt.Println("No repositories found.")
			return nil
		}

		for _, r := range repos {
			fullPath := repo.FullPath(cfg.Base, r)

			st := gitpkg.GetStatus(fullPath)

			if !st.Cloned {
				fmt.Printf("%s: not cloned\n", r.Path)
				continue
			}

			if st.NotARepo {
				fmt.Printf("%s: not a git repository\n", r.Path)
				continue
			}

			parts := []string{st.Branch}

			if st.Clean {
				parts = append(parts, "clean")
			} else {
				parts = append(parts, fmt.Sprintf("dirty (%d files)", st.Dirty))
			}

			if st.Ahead > 0 {
				parts = append(parts, fmt.Sprintf("ahead %d", st.Ahead))
			}
			if st.Behind > 0 {
				parts = append(parts, fmt.Sprintf("behind %d", st.Behind))
			}

			fmt.Printf("%s: %s\n", r.Path, strings.Join(parts, ", "))
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().StringVar(&statusGroup, "group", "", "filter by group name")
	statusCmd.Flags().StringVar(&statusResource, "resource", "", "filter by resource name")
	rootCmd.AddCommand(statusCmd)
}
