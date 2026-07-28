package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const exampleConfig = `# grepom complete example configuration
# Run grepom example to regenerate this file at any time

# base: local root directory for all repos (supports ~/ expansion)
base: ~/projects

# resources: named auth resources as a YAML map (key = resource name).
# A list of items with "name" fields is also accepted at load time, but map is preferred.
resources:
  # GitLab example (self-hosted)
  work-gl:
    provider: gitlab
    url: gitlab.example.com        # plain host, or with protocol prefix like https://gitlab.example.com
    token: ${GITLAB_TOKEN}         # supports ${ENV_VAR} placeholders
    ssh_key: ~/.ssh/id_work        # recommended: explicit SSH key (especially when SSH agent is unavailable)
    enabled: true                  # optional: skip all groups/repos under this resource when false

  # GitHub example
  github:
    provider: github
    url: github.com
    token: ${GITHUB_TOKEN}

  # Generic example: manages arbitrary Git repos via explicit URLs, no platform API needed
  # Tip: set ssh_key for private SSH hosts; public HTTPS URLs can clone anonymously.
  my-git:
    provider: generic
    url: git.example.com
    token: ${GIT_TOKEN}
    ssh_key: ~/.ssh/id_ed25519     # recommended for SSH; default keys are probed if omitted

# groups: auto-discover and manage repos from remote groups/orgs
groups:
  - name: frontend                 # unique local name
    resource: work-gl              # references a resource name from above
    path: my-org/frontend          # remote group/org path
    local_path: ./frontend         # optional: local directory (defaults to derived from path)
    recursive: true                # optional: recursively discover sub-groups (GitLab only)
    ssh_key: ~/.ssh/id_deploy      # optional: overrides resource ssh_key
    token: ${FRONTEND_TOKEN}       # optional: overrides resource token
    enabled: true                  # optional: skip entire group when false
    exclude_repos:                 # optional: exclude specific repo names
      - archived-repo
      - legacy-app
    repos: []                      # auto-populated by grepom sync, do not edit manually

  - name: my-org
    resource: github
    path: my-github-org
    recursive: false

# virtual_groups: named collections of real groups and optional standalone repos for --vgroup
virtual_groups:
  work:
    groups:
      - frontend
      - my-org
    repos:                         # optional: top-level standalone repo names
      - dotfiles
  personal:
    groups:
      - my-org
    repos:
      - public-demo

# repos: explicitly declared standalone repos (not part of any group)
# url may be a relative path (joined with resource.host) or an absolute https:// / git@ / ssh:// URL
# Absolute URLs are used as-is (not re-concatenated onto resource.url).
repos:
  - name: dotfiles
    resource: github
    url: example/dotfiles.git      # relative path → git@github.com:example/dotfiles.git
    local_path: ./dotfiles         # optional: defaults to ./<name>
    ssh_key: ~/.ssh/id_personal    # optional: overrides resource ssh_key

  - name: public-demo
    resource: my-git
    url: https://git.example.com/tools/public-demo.git  # absolute HTTPS: PreferHTTPS, anonymous OK for public repos

  - name: internal-tool
    resource: my-git
    url: tools/internal-tool.git   # relative path under git.example.com (SSH preferred)

# services: optional local development service definitions
services:
  api:
    cwd: ./backend
    command: make dev
  web:
    cwd: ./frontend
    command:
      - pnpm
      - dev
`

var outputFile string

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Export a complete example configuration",
	Long:  "Export a complete example .grepom.yml configuration file with all supported fields and comments.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if outputFile == "" {
			fmt.Print(exampleConfig)
			return nil
		}
		if _, err := os.Stat(outputFile); err == nil {
			return fmt.Errorf("file already exists: %s", outputFile)
		}
		if err := os.WriteFile(outputFile, []byte(exampleConfig), 0644); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Example config written to %s\n", outputFile)
		return nil
	},
}

func init() {
	exampleCmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file path (defaults to stdout)")
	rootCmd.AddCommand(exampleCmd)
}
