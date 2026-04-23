package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var (
	initBase     string
	initProvider string
	initURL      string
	initToken    string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a configuration file",
	Long:  "Create a new .grepom.yml configuration file in the current directory or at a specified path.",
	Example: `  grepom init                          # Create config with default base ~/projects
  grepom init --base ~/work/repos      # Specify base directory
  grepom init --provider gitlab --url https://gitlab.com --token ${GITLAB_TOKEN}`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolvedConfigPath()
		if err != nil {
			path = defaultConfigPath()
		}

		if err := config.InitConfig(path, initBase); err != nil {
			return err
		}

		// Optionally add first resource
		if initProvider != "" && initURL != "" && initToken != "" {
			name := initProvider
			if initProvider == "gitlab" {
				name = "gitlab"
			}
			res := config.Resource{
				Provider: initProvider,
				URL:      initURL,
				Token:    initToken,
			}
			if err := config.AddResource(path, name, res); err != nil {
				return fmt.Errorf("add resource: %w", err)
			}
		}

		fmt.Printf("Created config file: %s\n", path)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&initBase, "base", "b", "", "base directory for cloned repos (default: ~/projects)")
	initCmd.Flags().StringVarP(&initProvider, "provider", "p", "", "provider type for initial resource (gitlab or github)")
	initCmd.Flags().StringVarP(&initURL, "url", "u", "", "API base URL for initial resource")
	initCmd.Flags().StringVarP(&initToken, "token", "k", "", "API token for initial resource (supports ${ENV_VAR} syntax)")
	rootCmd.AddCommand(initCmd)
}
