package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
	"github.com/wii/grepom/repo"
)

var (
	listGroup    string
	listResource string
	listType     string
)

var listCmd = &cobra.Command{
	Use:   "list [name]",
	Short: "List resources, groups, or repositories",
	Long: `List resources, groups, or repositories based on the --type flag.

By default (or with --type repos), lists all discovered repositories.
Use --type resources to list configured resources, or --type groups to list configured groups.`,
	Example: `  grepom list                            # List all repos
  grepom list web-app                     # List a specific repo
  grepom list --group frontend            # List repos in a group
  grepom list --resource work-gl          # List repos from a resource
  grepom list --type resources            # List all configured resources
  grepom list --type groups               # List all configured groups`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		switch listType {
		case "resources":
			return runListResources(cfg)
		case "groups":
			return runListGroups(cfg)
		default:
			return runListRepos(cfg, args)
		}
	},
}

func init() {
	listCmd.Flags().StringVar(&listGroup, "group", "", "filter by group name")
	listCmd.Flags().StringVar(&listResource, "resource", "", "filter by resource name")
	listCmd.Flags().StringVar(&listType, "type", "repos", "type to list: repos, resources, groups")
	rootCmd.AddCommand(listCmd)
}

func runListRepos(cfg *config.Config, args []string) error {
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
}

func runListResources(cfg *config.Config) error {
	if len(cfg.Resources) == 0 {
		fmt.Println("No resources found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tPROVIDER\tURL\tSSH_KEY")

	for name, r := range cfg.Resources {
		sshKey := "-"
		if r.SSHKey != "" {
			sshKey = r.SSHKey
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, r.Provider, r.URL, sshKey)
	}
	w.Flush()

	return nil
}

func runListGroups(cfg *config.Config) error {
	if len(cfg.Groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tRESOURCE\tPATH\tLOCAL_PATH\tRECURSIVE\tREPOS")

	for _, g := range cfg.Groups {
		recursive := "no"
		if g.Recursive {
			recursive = "yes"
		}
		localPath := g.LocalPath
		if localPath == "" {
			localPath = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n", g.Name, g.Resource, g.Path, localPath, recursive, len(g.Repos))
	}
	w.Flush()

	return nil
}
