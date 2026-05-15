## Why

grepom 目前可以管理多仓库的 clone、pull、push、sync 等操作，但缺少一个高频开发动作——提交 Merge Request / Pull Request。开发者通常需要切到浏览器手动操作，中断命令行工作流。作为一个 Git 仓库编排工具，grepom 应该支持在 CLI 中直接创建 MR/PR，打通从代码编写到代码审查的最后一公里。

## What Changes

- 新增 `mr` 命令（`pr` 作为别名），用于创建 Merge Request / Pull Request
- 支持 GitHub 和 GitLab 两个 provider，通过 REST API 创建 PR/MR
- Codeup 暂不支持，执行时给出友好提示并引导到浏览器
- 无参数时自动检测当前 git 仓库的分支信息和 remote，智能推断 from/to 分支
- 支持 `--from` 和 `--to` 参数手动指定源分支和目标分支
- 支持 `--draft` 创建草稿 MR/PR
- 支持 `--web` 在浏览器中打开创建页面
- 支持自动从最新 commit message 提取 title/body，也可通过 `--title` / `--body` 手动指定
- 当分支有未推送的 commit 时，交互式提示用户是否先 push
- 新增 `mergerequest/` 包，采用与 `cicd/` 包一致的独立注册表模式
- 在 `git/` 包中新增分支检测和 remote 解析等辅助函数

## Capabilities

### New Capabilities
- `mr-pr-command`: MR/PR 创建命令，包括分支检测、provider 识别、token 获取、API 调用和 `--web` 浏览器打开
- `merge-request-provider`: MR/PR provider 接口层，包括 GitHub PR 和 GitLab MR 的 API 实现，以及 Codeup 的不支持提示
- `git-branch-detect`: Git 仓库分支检测和 remote URL 解析能力（当前分支、默认分支、remote URL、未推送 commit 检测）

### Modified Capabilities
（无现有 spec 需要修改）

## Impact

- **新增包**: `mergerequest/`（mergerequest.go、github.go、gitlab.go 及测试）
- **修改包**: `git/git.go`（新增 GetCurrentBranch、GetRemoteURL、HasUnpushedCommits、GetHeadCommitMessage 函数）
- **修改包**: `cmd/`（新增 mr.go，在 root.go 中注册 `mr` 和 `pr` 两个命令别名）
- **依赖**: 复用已有的 `github.com/AlecAivazis/survey/v2`（交互提示）和标准 `net/http`（API 调用）
- **文档**: 需同步更新 README.md 和 README_en.md，在命令列表中新增 `mr` 命令说明
