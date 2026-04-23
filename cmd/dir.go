package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/repo"
)

var (
	dirGroup    string
	dirResource string
	dirShell    bool
)

// fzfAvailable 检测系统中是否安装了 fzf
func fzfAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

const shellHelperWithFzf = `gcd() {
  local dir
  if [ $# -eq 0 ]; then
    dir=$(grepom dir)
  else
    dir=$(grepom dir "$@" | fzf --select-1)
  fi || return
  cd "$dir"
}`

const shellHelperWithoutFzf = `gcd() {
  cd "$(grepom dir "$@")"
}`

var dirCmd = &cobra.Command{
	Use:   "dir [name]",
	Short: "Print local directory path for a repo or base directory",
	Long: `Print the local directory path to stdout.

Without arguments, prints the base directory from the config file.
With a name argument, performs case-insensitive substring matching across all
repositories and prints the matching repo's local path.

When exactly one repo matches, the path is printed to stdout (suitable for
cd "$(grepom dir web-app)"). When multiple repos match, a table is printed
to stderr and the command exits with code 1.

Use --shell to print a gcd() shell function for easy directory switching.
Add eval "$(grepom dir --shell)" to your .bashrc/.zshrc.`,
	Example: `  grepom dir                    # Print base directory path
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
			if fzfAvailable() {
				fmt.Println(shellHelperWithFzf)
			} else {
				fmt.Println(shellHelperWithoutFzf)
			}
			return nil
		}

		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		// 无参数：输出 base 目录
		if len(args) == 0 {
			fmt.Println(cfg.Base)
			return nil
		}

		// 有参数：模糊搜索仓库
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
		results := repo.ApplySearchFilter(allRepos, keyword, filter)

		switch len(results) {
		case 0:
			fmt.Fprintf(os.Stderr, "no repo found matching %q\n", keyword)
			return fmt.Errorf("no match")
		case 1:
			// 单个结果：输出路径到 stdout
			fmt.Println(repo.FullPath(cfg.Base, results[0]))
			return nil
		default:
			// 多个结果：列表到 stderr，返回错误
			fmt.Fprintf(os.Stderr, "multiple repos matched %q:\n\n", keyword)
			w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPATH\tGROUP\tRESOURCE")
			for _, r := range results {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, repo.FullPath(cfg.Base, r), r.GroupName, r.Resource)
			}
			w.Flush()
			fmt.Fprintln(os.Stderr, "\nPlease specify a more precise name.")
			return fmt.Errorf("multiple matches")
		}
	},
}

func init() {
	dirCmd.Flags().StringVarP(&dirGroup, "group", "g", "", "filter by group name")
	dirCmd.Flags().StringVarP(&dirResource, "resource", "R", "", "filter by resource name")
	dirCmd.Flags().BoolVar(&dirShell, "shell", false, "print gcd() shell function for easy cd (add to .bashrc/.zshrc)")
	rootCmd.AddCommand(dirCmd)
}
