package cmd

import (
	"os"
	"sort"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCompletionCmd())
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for bash, zsh, or fish.

Load completions in your current shell session:

  # bash
  source <(grepom completion bash)

  # zsh
  source <(grepom completion zsh)

  # fish
  grepom completion fish | source`,
		Example: `  eval "$(grepom completion bash)"
  eval "$(grepom completion zsh)"
  grepom completion fish | source`,
		ValidArgs: []string{"bash", "zsh", "fish"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			}
			return nil
		},
	}
}

func completeSvcNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	mgr, err := resolveServiceManager()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	names := make(map[string]struct{})
	for name := range mgr.Services {
		names[name] = struct{}{}
	}
	entries, err := mgr.List()
	if err == nil {
		for _, e := range entries {
			names[e.Record.Name] = struct{}{}
		}
	}

	result := make([]string, 0, len(names))
	for name := range names {
		result = append(result, name)
	}
	sort.Strings(result)
	return result, cobra.ShellCompDirectiveNoFileComp
}
