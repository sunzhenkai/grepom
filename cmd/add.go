package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a source or repository to the config file",
	Long:  "Append a new API source or explicit repository entry to the configuration file.",
	Example: `  grepom add source --provider gitlab --url https://gitlab.com --group my-org/frontend --recursive
  grepom add source --provider github --url https://github.com --org my-org
  grepom add repo --name special --url https://gitlab.com/other/special.git --path ./special`,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// add source subcommand
var (
	addProvider  string
	addURL       string
	addToken     string
	addGroup     []string
	addOrg       []string
	addRecursive bool
)

var addSourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Add an API source to the config file",
	Long:  "Add a GitLab or GitHub API source to fetch repositories from.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addProvider == "" {
			return fmt.Errorf("--provider is required (gitlab or github)")
		}
		if addProvider != "gitlab" && addProvider != "github" {
			return fmt.Errorf("unsupported provider: %s (use gitlab or github)", addProvider)
		}
		if addURL == "" {
			return fmt.Errorf("--url is required")
		}

		path, err := resolvedConfigPath()
		if err != nil {
			path = configFile
		}

		source := config.Source{
			Provider: addProvider,
			URL:      addURL,
			Token:    addToken,
		}

		for _, g := range addGroup {
			source.Groups = append(source.Groups, config.GroupSource{
				Path:      g,
				Recursive: addRecursive,
			})
		}

		for _, o := range addOrg {
			source.Orgs = append(source.Orgs, config.OrgSource{Name: o})
		}

		if len(source.Groups) == 0 && len(source.Orgs) == 0 {
			return fmt.Errorf("at least one --group or --org is required")
		}

		if err := config.AddSource(path, source); err != nil {
			return err
		}

		fmt.Printf("Added %s source to %s\n", addProvider, path)
		return nil
	},
}

func init() {
	addSourceCmd.Flags().StringVar(&addProvider, "provider", "", "provider type (gitlab or github)")
	addSourceCmd.Flags().StringVar(&addURL, "url", "", "API base URL")
	addSourceCmd.Flags().StringVar(&addToken, "token", "", "API token (or use ${ENV_VAR} syntax)")
	addSourceCmd.Flags().StringArrayVar(&addGroup, "group", nil, "group path to fetch (GitLab)")
	addSourceCmd.Flags().StringArrayVar(&addOrg, "org", nil, "organization name (GitHub)")
	addSourceCmd.Flags().BoolVar(&addRecursive, "recursive", false, "recursively fetch subgroups")
	addCmd.AddCommand(addSourceCmd)
}

// add repo subcommand
var (
	addRepoName string
	addRepoURL  string
	addRepoPath string
)

var addRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Add an explicit repository to the config file",
	Long:  "Add a repository entry with a custom clone URL and path.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addRepoName == "" {
			return fmt.Errorf("--name is required")
		}
		if addRepoURL == "" {
			return fmt.Errorf("--url is required")
		}
		if addRepoPath == "" {
			addRepoPath = "./" + addRepoName
		}

		path, err := resolvedConfigPath()
		if err != nil {
			path = configFile
		}

		entry := config.RepoEntry{
			Name: addRepoName,
			URL:  addRepoURL,
			Path: addRepoPath,
		}

		if err := config.AddRepo(path, entry); err != nil {
			return err
		}

		fmt.Printf("Added repo %s to %s\n", addRepoName, path)
		return nil
	},
}

func init() {
	addRepoCmd.Flags().StringVar(&addRepoName, "name", "", "repository name")
	addRepoCmd.Flags().StringVar(&addRepoURL, "url", "", "clone URL")
	addRepoCmd.Flags().StringVar(&addRepoPath, "path", "", "relative path from base (default: ./<name>)")
	addCmd.AddCommand(addRepoCmd)
}
