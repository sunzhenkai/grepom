## Why

`list --type resources` 目前只能展示本地配置文件中已有的 resource 元数据（provider、url、ssh_key），无法直接查看远程 provider 上实际存在的仓库信息。用户若想了解某个 resource 对应的远程有哪些仓库，需要先 `sync` 再 `list`，流程割裂。

同时，`status` 命令的头部摘要当前以单行文本形式输出（如 `168 repos:, 168 clean`），当 repo 数量多、状态类型多时，信息不够直观，不易快速扫读。

## What Changes

- **`list` 命令新增 `--remote` 标志**：当使用 `--type repos --remote` 时，系统通过 provider API 实时查询远程仓库列表并展示，而非仅读取本地配置。支持 `--group` 和 `--resource` 过滤。
- **`status` 命令头部摘要改为表格展示**：将当前的单行文本摘要（`N repos: N clean, N dirty, ...`）改为表格形式，每行一个状态类型及对应数量，提升可读性。

## Capabilities

### New Capabilities
- `list-remote-repos`: `list` 命令支持通过 `--remote` 标志实时查询远程 provider 仓库并展示

### Modified Capabilities
- `cli-commands`: `status` 命令头部摘要输出格式从单行文本改为表格

## Impact

- **`cmd/list.go`**：新增 `--remote` 标志及 `runListRemoteRepos` 函数，调用 provider API 查询远程仓库
- **`cmd/status.go`**：修改头部摘要输出逻辑，使用 tabwriter 输出表格格式
- **`openspec/specs/cli-commands/spec.md`**：更新 status command 的概要行描述为表格格式
- **`openspec/specs/resource-list/spec.md`**：无变更（`list --type resources` 行为不变，`--remote` 仅作用于 `--type repos`）
- **依赖**：复用现有 `provider.Provider.ListRepos` 接口，无需新增 provider API 调用逻辑
