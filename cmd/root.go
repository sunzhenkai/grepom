package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var (
	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "grepom",
	Short: "Git Repository Orchestrator & Manager",
	Long: `A CLI tool for managing multiple git repositories across GitLab groups and GitHub organizations.

Use YAML configuration files to define resources (authentication), groups (remote paths),
and standalone repos. Grepom discovers and manages repositories automatically.`,
	Example: `  grepom -c work.yml list              # List repos from a specific config
  grepom clone --group frontend        # Clone all repos in a group
  grepom status                        # Check status of all cloned repos
  grepom pull web-app                  # Pull updates for a single repo`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".grepom.yml", "path to config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}

func loadConfig() (*config.Config, error) {
	config.SetVerbose(verbose)
	path, err := config.FindConfig(configFile)
	if err != nil {
		return nil, err
	}
	return config.Load(path)
}

func resolvedConfigPath() (string, error) {
	return config.FindConfig(configFile)
}
