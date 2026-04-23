package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/repo"
)

var (
	dirGroup    string
	dirResource string
	dirShell    bool
)

const shellHelper = `gcd() {
  local dir
  if [ $# -eq 0 ]; then
    dir=$(grepom dir)
  else
    if command -v fzf >/dev/null 2>&1; then
      dir=$(grepom dir "$@" | fzf --select-1)
    else
      dir=$(grepom dir "$@" | head -n 1)
    fi
  fi || return
  cd "$dir"
}`

var dirCmd = &cobra.Command{
	Use:   "dir [name]",
	Short: "Print local directory path for a repo or config directory",
	Long: `Print the local directory path to stdout.

Without arguments, prints the directory containing the config file (.grepom.yml).
With a name argument, performs exact-first then substring case-insensitive matching
across all repositories and prints matching repo local path(s).

When exactly one repo matches, the path is printed to stdout (suitable for
cd "$(grepom dir web-app)"). When multiple repos match, all paths are printed
(one per line) to stdout.

Use --shell to print a gcd() shell function for easy directory switching.
Add eval "$(grepom dir --shell)" to your .bashrc/.zshrc.`,
	Example: `  grepom dir                    # Print config file directory path
  grepom dir web-app            # Print path of repo matching "web-app"
  grepom dir web --group fe     # Search "web" in group "fe"
  cd "$(grepom dir web-app)"    # Jump to the repo directory
  grepom dir --shell            # Print gcd() shell function
  eval "$(grepom dir --shell)"  # Enable gcd command in current shell`,
	Args: cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// --shell：输出 shell function 代码
		if dirShell {
			fmt.Println(shellHelper)
			return nil
		}

		configPath, cfg, err := loadConfig()
		if err != nil {
			return err
		}

		// 无参数：输出配置文件所在目录
		if len(args) == 0 {
			fmt.Println(filepath.Dir(configPath))
			return nil
		}

		// 有参数：搜索仓库（精确优先，后子串）
		keyword := args[0]

		resolver := repo.NewResolver(cfg)
		allRepos, err := resolver.Resolve()
		if err != nil {
			return err
		}

		filter := repo.Filter{
			Group:    dirGroup,
			Resource: dirResource,
		}
		results := repo.ApplyExactFirstSearch(allRepos, keyword, filter)

		switch len(results) {
		case 0:
			fmt.Fprintf(os.Stderr, "no repo found matching %q\n", keyword)
			return fmt.Errorf("no match")
		default:
			// 单个或多个结果：输出路径到 stdout（每行一个）
			for _, r := range results {
				fmt.Println(repo.FullPath(cfg.Base, r))
			}
			return nil
		}
	},
}

func init() {
	dirCmd.Flags().StringVarP(&dirGroup, "group", "g", "", "filter by group name")
	dirCmd.Flags().StringVarP(&dirResource, "resource", "R", "", "filter by resource name")
	dirCmd.Flags().BoolVar(&dirShell, "shell", false, "print gcd() shell function for easy cd (add to .bashrc/.zshrc)")
	rootCmd.AddCommand(dirCmd)
}
