package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	gitpkg "github.com/wii/grepom/git"
	"github.com/wii/grepom/scanner"
)

var (
	pushForce       bool
	pushGitleaksCfg string
)

var pushCmd = &cobra.Command{
	Use:   "push [--] [git-push-args...]",
	Short: "Scan for secrets before pushing",
	Long: `Automatically scan the current directory for sensitive information before executing git push.

If secrets are found, push is rejected by default. Use -f/--force to force push (with warning).

This command does not depend on grepom config files and can be used in any git repository.
Use -- to pass remaining arguments through to git push.`,
	Example: `  grepom push                        # Scan and push (if no secrets found)
  grepom push -f                     # Force push even if secrets found (prints warning)
  grepom push -- origin main         # Pass arguments through to git push
  grepom push --gitleaks-config ./rules.toml  # Use custom scan rules`,
	RunE: runPush,
}

func init() {
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "force push even when secrets are found (still prints warning)")
	pushCmd.Flags().StringVar(&pushGitleaksCfg, "gitleaks-config", "", "path to custom gitleaks.toml config file")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	// 检测当前目录是否为 git 仓库
	if !gitpkg.IsCloned(".") {
		return fmt.Errorf("current directory is not a git repository")
	}

	// 收集透传给 git push 的参数（跳过 cobra 已解析的标志）
	gitArgs := extractGitPushArgs(cmd, args)

	// 创建 scanner 并扫描当前目录
	s := scanner.NewScanner(scanner.Options{
		GitleaksConfigPath: pushGitleaksCfg,
	})

	ctx := context.Background()
	findings, err := s.ScanDir(ctx, ".")
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// 有发现项时处理
	if len(findings) > 0 {
		// 设置 repo 名称为当前目录
		dir, _ := os.Getwd()
		for i := range findings {
			findings[i].Repo = dir
		}

		// 输出扫描结果
		if err := outputTable(os.Stderr, findings); err != nil {
			return err
		}

		if pushForce {
			fmt.Fprintln(os.Stderr, "\n⚠ Secrets detected but push forced")
		} else {
			fmt.Fprintln(os.Stderr, "\n❌ Push rejected: secrets detected. Use -f to force push.")
			return fmt.Errorf("push aborted: %d secret(s) detected", len(findings))
		}
	}

	// 执行 git push
	return gitpkg.Push(".", gitArgs...)
}

// extractGitPushArgs 从命令参数中提取需要透传给 git push 的参数。
// cobra 的 flags 已经被解析，剩余的 args 直接传递给 git push。
func extractGitPushArgs(cmd *cobra.Command, args []string) []string {
	// cobra 在解析 flags 后，剩余的位置参数就是需要透传的参数
	return args
}
