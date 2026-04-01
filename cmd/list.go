package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/repo"
)

var (
	listGroup    string
	listResource string
)

var listCmd = &cobra.Command{
	Use:   "list [name]",
	Short: "List discovered repositories",
	Long:  "List repositories from all configured groups and standalone repos. Optionally filter by name, group, or resource.",
	Example: `  grepom list                            # List all repos
  grepom list web-app                     # List a specific repo
  grepom list --group frontend            # List repos in a group
  grepom list --resource work-gl          # List repos from a resource`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Group:    listGroup,
			Resource: listResource,
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

		w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPATH\tGROUP\tRESOURCE\tCLONED")

		for _, r := range repos {
			cloned := "no"
			fullPath := repo.FullPath(cfg.Base, r)
			if _, err := os.Stat(fullPath); err == nil {
				cloned = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, r.Path, r.GroupName, r.Resource, cloned)
		}
		w.Flush()

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listGroup, "group", "", "filter by group name")
	listCmd.Flags().StringVar(&listResource, "resource", "", "filter by resource name")
	rootCmd.AddCommand(listCmd)
}
