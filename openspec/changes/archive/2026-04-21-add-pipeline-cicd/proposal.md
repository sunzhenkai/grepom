## Why

grepom 目前是一个仓库管理工具（发现、克隆、拉取、状态检查），但用户在本地开发时经常需要快速查看某个仓库的 CI/CD 运行状态——尤其是推送代码后想看 pipeline 是否通过。当前只能打开浏览器手动查看，效率低下。需要在 grepom 中直接查看和监控 pipeline 状态。

## What Changes

- 新增 `cicd` 包：定义 `PipelineProvider` 接口及 GitLab、GitHub 两个实现，通过各 Provider 的 CI/CD REST API 查询 pipeline/runs 数据
- 新增 `grepom pipeline list <repo>` 命令：列出某个仓库最近 N 条 pipeline 的状态（默认 5 条，支持 `-n` 参数调整）
- 新增 `grepom pipeline watch <repo>` 命令：watch 某个仓库的最新 pipeline（或指定 `--id`），轮询刷新状态，pipeline 到达终态时自动退出，也支持 Ctrl+C 提前终止
- 导出 `repo.ExtractRemotePath`：从 CloneURL 反推远程仓库路径，供 cicd 包使用
- Token 复用现有的 repo → group → resource 优先级链路

初始范围仅支持 GitLab 和 GitHub，不支持 Codeup。

## Capabilities

### New Capabilities
- `pipeline-list`: 列出指定仓库最近的 CI/CD pipeline 状态
- `pipeline-watch`: 实时监控指定 pipeline 的运行状态，终态自动退出

### Modified Capabilities
- 无 spec 级别修改，仅导出 `repo.ExtractRemotePath` 函数

## Impact

- 新增 `cicd/` 包（cicd.go、gitlab.go、github.go）
- 新增 `cmd/pipeline.go`（pipeline list + pipeline watch 子命令）
- `repo/resolver.go`：导出 `extractRepoPath` 为 `ExtractRemotePath`
- 依赖现有 `provider/github.go` 的 `githubAPIURL` URL 转换逻辑（cicd 包内复制或导出）
