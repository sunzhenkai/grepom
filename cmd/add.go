package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/config"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a resource, group, or repository to the config file",
	Long:  "Append a new resource, group, or repo entry to the configuration file.",
	Example: `  grepom add resource --name work-gl --provider gitlab --url https://gitlab.com --token ${GL_TOKEN}
  grepom add group --name frontend --resource work-gl --path my-org/frontend --local-path ./frontend --recursive
  grepom add repo --name dotfiles --resource github --url https://github.com/me/dotfiles.git`,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

// --- add resource ---
var (
	addResName     string
	addResProvider string
	addResURL      string
	addResToken    string
	addResSSHKey   string
)

var addResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Add an authentication resource to the config file",
	Long:  "Add a GitLab or GitHub API resource with connection and authentication details.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addResProvider == "" {
			return fmt.Errorf("--provider is required (gitlab or github)")
		}
		if addResProvider != "gitlab" && addResProvider != "github" {
			return fmt.Errorf("unsupported provider: %s (use gitlab or github)", addResProvider)
		}
		if addResURL == "" {
			return fmt.Errorf("--url is required")
		}
		if addResName == "" {
			return fmt.Errorf("--name is required")
		}

		path, err := resolvedConfigPath()
		if err != nil {
			path = configFile
		}

		res := config.Resource{
			Provider: addResProvider,
			URL:      addResURL,
			Token:    addResToken,
			SSHKey:   addResSSHKey,
		}

		if err := config.AddResource(path, addResName, res); err != nil {
			return err
		}

		fmt.Printf("Added resource %s to %s\n", addResName, path)
		return nil
	},
}

func init() {
	addResourceCmd.Flags().StringVarP(&addResName, "name", "n", "", "resource name (required)")
	addResourceCmd.Flags().StringVarP(&addResProvider, "provider", "p", "", "provider type (gitlab or github)")
	addResourceCmd.Flags().StringVarP(&addResURL, "url", "u", "", "API base URL")
	addResourceCmd.Flags().StringVarP(&addResToken, "token", "k", "", "API token (supports ${ENV_VAR} syntax)")
	addResourceCmd.Flags().StringVarP(&addResSSHKey, "ssh-key", "s", "", "SSH private key path for clone (supports ~/")
	addCmd.AddCommand(addResourceCmd)
}

// --- add group ---
var (
	addGroupName     string
	addGroupResource string
	addGroupPath     string
	addGroupLocal    string
	addGroupRecurse  bool
	addGroupSSHKey   string
	addGroupToken    string
)

var addGroupCmd = &cobra.Command{
	Use:   "group",
	Short: "Add a group to the config file",
	Long:  "Add a remote group (GitLab group or GitHub org) whose repos will be managed.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addGroupName == "" {
			return fmt.Errorf("--name is required")
		}
		if addGroupResource == "" {
			return fmt.Errorf("--resource is required")
		}
		if addGroupPath == "" {
			return fmt.Errorf("--path is required")
		}

		path, err := resolvedConfigPath()
		if err != nil {
			path = configFile
		}

		// Pre-add validation: check resource exists and name uniqueness
		cfg, err := loadExistingConfig(path)
		if err == nil {
			if _, ok := cfg.Resources[addGroupResource]; !ok {
				return fmt.Errorf("resource %q not found", addGroupResource)
			}
			for _, g := range cfg.Groups {
				if g.Name == addGroupName {
					return fmt.Errorf("group %q already exists", addGroupName)
				}
			}
		}

		group := config.Group{
			Name:      addGroupName,
			Resource:  addGroupResource,
			Path:      addGroupPath,
			LocalPath: addGroupLocal,
			Recursive: addGroupRecurse,
			SSHKey:    addGroupSSHKey,
			Token:     addGroupToken,
		}

		if err := config.AddGroup(path, group); err != nil {
			return err
		}

		fmt.Printf("Added group %s to %s\n", addGroupName, path)
		return nil
	},
}

func init() {
	addGroupCmd.Flags().StringVarP(&addGroupName, "name", "n", "", "group name (required)")
	addGroupCmd.Flags().StringVarP(&addGroupResource, "resource", "R", "", "resource name to use for authentication")
	addGroupCmd.Flags().StringVarP(&addGroupPath, "path", "p", "", "remote group path (e.g. my-org/frontend)")
	addGroupCmd.Flags().StringVarP(&addGroupLocal, "local-path", "l", "", "local directory path relative to base (default: ./<name>)")
	addGroupCmd.Flags().BoolVarP(&addGroupRecurse, "recursive", "r", false, "recursively discover subgroups (GitLab only)")
	addGroupCmd.Flags().StringVarP(&addGroupSSHKey, "ssh-key", "s", "", "SSH private key path for clone (overrides resource)")
	addGroupCmd.Flags().StringVarP(&addGroupToken, "token", "k", "", "token for clone (overrides resource, supports ${ENV_VAR})")
	addCmd.AddCommand(addGroupCmd)
}

// --- add repo ---
var (
	addRepoName      string
	addRepoResource  string
	addRepoURL       string
	addRepoLocalPath string
	addRepoGroup     string
	addRepoPath      string
	addRepoSSHKey    string
	addRepoToken     string
)

var addRepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Add a repository to the config file",
	Long:  "Add a standalone repo or a repo to an existing group.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addRepoName == "" {
			return fmt.Errorf("--name is required")
		}
		if addRepoURL == "" {
			return fmt.Errorf("--url is required")
		}

		path, err := resolvedConfigPath()
		if err != nil {
			path = configFile
		}

		if addRepoGroup != "" {
			// Add to group: validate group exists
			cfg, err := loadExistingConfig(path)
			if err == nil {
				found := false
				for _, g := range cfg.Groups {
					if g.Name == addRepoGroup {
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("group %q not found", addRepoGroup)
				}
			}

			repo := config.GroupRepo{
				Name: addRepoName,
				URL:  addRepoURL,
				Path: addRepoPath,
			}
			if repo.Path == "" {
				repo.Path = addRepoName
			}

			if err := config.AddGroupRepo(path, addRepoGroup, repo); err != nil {
				return err
			}
			fmt.Printf("Added repo %s to group %s in %s\n", addRepoName, addRepoGroup, path)
		} else {
			// Standalone repo: validate resource and name uniqueness
			if addRepoResource == "" {
				return fmt.Errorf("--resource is required")
			}

			cfg, err := loadExistingConfig(path)
			if err == nil {
				if _, ok := cfg.Resources[addRepoResource]; !ok {
					return fmt.Errorf("resource %q not found", addRepoResource)
				}
				for _, r := range cfg.Repos {
					if r.Name == addRepoName {
						return fmt.Errorf("repo %q already exists", addRepoName)
					}
				}
			}

			repo := config.Repo{
				Name:      addRepoName,
				Resource:  addRepoResource,
				URL:       addRepoURL,
				LocalPath: addRepoLocalPath,
				SSHKey:    addRepoSSHKey,
				Token:     addRepoToken,
			}

			if err := config.AddRepo(path, repo); err != nil {
				return err
			}
			fmt.Printf("Added repo %s to %s\n", addRepoName, path)
		}

		return nil
	},
}

func init() {
	addRepoCmd.Flags().StringVarP(&addRepoName, "name", "n", "", "repository name")
	addRepoCmd.Flags().StringVarP(&addRepoResource, "resource", "R", "", "resource name for authentication")
	addRepoCmd.Flags().StringVarP(&addRepoURL, "url", "u", "", "clone URL")
	addRepoCmd.Flags().StringVarP(&addRepoLocalPath, "local-path", "l", "", "local path relative to base (default: ./<name>)")
	addRepoCmd.Flags().StringVarP(&addRepoGroup, "group", "g", "", "group name to add this repo to (omit for standalone)")
	addRepoCmd.Flags().StringVarP(&addRepoPath, "path", "p", "", "remote path for group repo (e.g. my-org/frontend/repo-name)")
	addRepoCmd.Flags().StringVarP(&addRepoSSHKey, "ssh-key", "s", "", "SSH private key path for clone (overrides resource)")
	addRepoCmd.Flags().StringVarP(&addRepoToken, "token", "k", "", "token for clone (overrides resource, supports ${ENV_VAR})")
	addCmd.AddCommand(addRepoCmd)
}

// loadExistingConfig loads the config file for validation purposes.
// Returns nil (no error) if file doesn't exist, allowing add commands to proceed.
// This avoids triggering env var resolution errors during pre-add validation.
func loadExistingConfig(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
