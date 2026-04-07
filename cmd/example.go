package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const exampleConfig = `# grepom 完整示例配置
# 运行 grepom example 可随时重新生成此文件

# base: 所有仓库的本地根目录（支持 ~/ 展开）
base: ~/projects

# resources: 命名的认证资源，支持 gitlab / github / generic 三种 provider
resources:
  # GitLab 示例（自托管）
  work-gl:
    provider: gitlab
    url: gitlab.mycompany.com      # 纯 host，或带协议前缀如 https://gitlab.mycompany.com
    token: ${GITLAB_TOKEN}         # 支持 ${ENV_VAR} 占位符
    ssh_key: ~/.ssh/id_work        # 可选：SSH 私钥路径
    enabled: true                  # 可选：false 时跳过该 resource 下所有 group/repo

  # GitHub 示例
  github:
    provider: github
    url: github.com
    token: ${GITHUB_TOKEN}

  # Generic 示例：不依赖平台 API，通过显式 URL 管理任意 Git 仓库
  my-git:
    provider: generic
    url: git.internal.com
    token: ${GIT_TOKEN}
    ssh_key: ~/.ssh/id_internal

# groups: 从远程 group/org 自动发现并管理仓库
groups:
  - name: frontend                 # 本地唯一名称
    resource: work-gl              # 引用上方 resources 中的名称
    path: my-org/frontend          # 远程 group/org 路径
    local_path: ./frontend         # 可选：本地存放目录（默认从 path 推导）
    recursive: true                # 可选：递归发现子 group（仅 GitLab）
    ssh_key: ~/.ssh/id_deploy      # 可选：覆盖 resource 的 ssh_key
    token: ${FRONTEND_TOKEN}       # 可选：覆盖 resource 的 token
    enabled: true                  # 可选：false 时跳过整个 group
    exclude_repos:                 # 可选：排除指定仓库名
      - archived-repo
      - legacy-app
    repos: []                      # 由 grepom sync 自动填充，勿手动编辑

  - name: my-org
    resource: github
    path: my-github-org
    recursive: false

# repos: 显式声明的独立仓库（不属于任何 group）
repos:
  - name: dotfiles
    resource: github
    url: https://github.com/me/dotfiles.git
    local_path: ./dotfiles         # 可选：默认为 ./<name>
    ssh_key: ~/.ssh/id_personal    # 可选：覆盖 resource 的 ssh_key

  - name: internal-tool
    resource: my-git
    url: https://git.internal.com/tools/internal-tool.git
`

var outputFile string

var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "输出包含全部功能的示例配置",
	Long:  `输出一份包含所有支持字段和注释的完整示例 .grepom.yml 配置文件。`,
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
	exampleCmd.Flags().StringVarP(&outputFile, "output", "o", "", "写入文件路径（默认输出到 stdout）")
	rootCmd.AddCommand(exampleCmd)
}
