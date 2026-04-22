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
	Short: "推送前自动扫描敏感信息",
	Long: `在执行 git push 前自动扫描当前目录的敏感信息。
如果发现敏感信息，默认拒绝推送。使用 -f/--force 可以强制推送，但会打印警告。

该命令不依赖 grepom 配置文件，可在任何 git 仓库中使用。
使用 -- 将后续参数透传给 git push。`,
	Example: `  grepom push                        # 扫描后推送（无敏感信息时）
  grepom push -f                     # 发现敏感信息仍强制推送（打印警告）
  grepom push -- origin main         # 透传参数给 git push
  grepom push --gitleaks-config ./rules.toml  # 使用自定义扫描规则`,
	RunE: runPush,
}

func init() {
	pushCmd.Flags().BoolVarP(&pushForce, "force", "f", false, "发现敏感信息时强制推送（仍打印警告）")
	pushCmd.Flags().StringVar(&pushGitleaksCfg, "gitleaks-config", "", "自定义 gitleaks.toml 配置文件路径")
	rootCmd.AddCommand(pushCmd)
}

func runPush(cmd *cobra.Command, args []string) error {
	// 检测当前目录是否为 git 仓库
	if !gitpkg.IsCloned(".") {
		return fmt.Errorf("当前目录不是 git 仓库")
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
		return fmt.Errorf("扫描失败: %w", err)
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
