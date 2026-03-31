package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var (
	initBase string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new config file",
	Long:  "Create a .grepom.yml config file in the current directory. Use --base to specify the repository root directory.",
	Example: `  grepom init                                    # Create .grepom.yml with default base
  grepom init --base ~/work/repos                   # Specify a custom base directory
  grepom init --provider gitlab --url https://gitlab.com --group my-org
  grepom init --provider github --url https://github.com --org my-org`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unknown argument %q. Did you mean `grepom clone`?", args[0])
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolvedConfigPath()
		if err != nil {
			path = configFile
		}

		if err := config.InitConfig(path, initBase); err != nil {
			return err
		}

		fmt.Printf("Created config file: %s\n", path)

		// If provider flags are set, add a source
		if initProvider != "" {
			source := config.Source{
				Provider: initProvider,
				URL:      initURL,
				Token:    initToken,
			}

			for _, g := range initGroup {
				source.Groups = append(source.Groups, config.GroupSource{
					Path:      g,
					Recursive: initRecursive,
				})
			}

			for _, o := range initOrg {
				source.Orgs = append(source.Orgs, config.OrgSource{Name: o})
			}

			if len(source.Groups) == 0 && len(source.Orgs) == 0 {
				return nil
			}

			if err := config.AddSource(path, source); err != nil {
				return err
			}

			fmt.Printf("Added %s source\n", initProvider)
		}

		return nil
	},
}

var (
	initProvider  string
	initURL       string
	initToken     string
	initGroup     []string
	initOrg       []string
	initRecursive bool
)

func init() {
	initCmd.Flags().StringVar(&initBase, "base", "", "root directory for cloned repos (default: ~/projects)")
	initCmd.Flags().StringVar(&initProvider, "provider", "", "provider type (gitlab or github)")
	initCmd.Flags().StringVar(&initURL, "url", "", "API base URL")
	initCmd.Flags().StringVar(&initToken, "token", "", "API token (or use ${ENV_VAR} syntax)")
	initCmd.Flags().StringArrayVar(&initGroup, "group", nil, "group path to fetch (GitLab)")
	initCmd.Flags().StringArrayVar(&initOrg, "org", nil, "organization name (GitHub)")
	initCmd.Flags().BoolVar(&initRecursive, "recursive", false, "recursively fetch subgroups")
	rootCmd.AddCommand(initCmd)
}
