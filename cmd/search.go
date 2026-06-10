package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/git"
	"github.com/wii/grepom/repo"
)

var (
	searchGroup    string
	searchVGroup   string
	searchResource string
)

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search repositories by name",
	Long:  "Search all repositories (groups and standalone) by name using case-insensitive substring matching. Supports --group and --resource filters to narrow the search scope.",
	Example: `  grepom search web                      # Search all repos with "web" in name
  grepom search api --group backend        # Search "api" in group "backend"
  grepom search ui --resource work-gl      # Search "ui" in repos from resource "work-gl"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := args[0]

		_, cfg, err := loadConfig()
		if err != nil {
			return err
		}

		resolver := repo.NewResolver(cfg)
		allRepos, err := resolver.Resolve()
		if err != nil {
			return err
		}

		filter, err := buildRepoFilter(cfg, searchGroup, searchVGroup, searchResource, false)
		if err != nil {
			return err
		}
		results := repo.ApplySearchFilter(allRepos, keyword, filter)

		if len(results) == 0 {
			fmt.Println("No matching repos found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPATH\tGROUP\tRESOURCE\tCLONED")

		for _, r := range results {
			cloned := "no"
			fullPath := repo.FullPath(cfg.Base, r)
			if git.IsCloned(fullPath) {
				cloned = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.Path, r.GroupName, r.Resource, cloned)
		}
		w.Flush()

		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchGroup, "group", "g", "", "filter by group name")
	searchCmd.Flags().StringVarP(&searchVGroup, "vgroup", "V", "", "filter by virtual group name")
	searchCmd.Flags().StringVarP(&searchResource, "resource", "R", "", "filter by resource name")
	rootCmd.AddCommand(searchCmd)
}
