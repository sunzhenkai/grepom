package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/repo"
)

var (
	listSource string
	listGroup  string
)

var listCmd = &cobra.Command{
	Use:   "list [name]",
	Short: "List discovered repositories",
	Long:  "List repositories from all configured sources. Optionally filter by name, group, or provider.",
	Example: `  grepom list                        # List all repos
  grepom list web-app                 # List a specific repo
  grepom list --source gitlab         # Filter by provider
  grepom list --group my-org/frontend # Filter by group`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Provider: listSource,
			Group:    listGroup,
		}
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

		w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPATH\tPROVIDER\tCLONED")

		for _, r := range repos {
			cloned := "no"
			fullPath := repo.FullPath(cfg.Base, r)
			if _, err := os.Stat(fullPath); err == nil {
				cloned = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, r.Path, r.Provider, cloned)
		}
		w.Flush()

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listSource, "source", "", "filter by provider (gitlab, github)")
	listCmd.Flags().StringVar(&listGroup, "group", "", "filter by group path")
	rootCmd.AddCommand(listCmd)
}
