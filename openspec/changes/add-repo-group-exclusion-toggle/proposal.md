## Why

当前 grepom 中所有配置的 resource、group 和 repo 都始终处于活跃状态，没有任何机制可以临时禁用或排除某些条目。用户在管理大量 repo 时，经常需要：
- 临时屏蔽某个 group 下的特定 repo（如已废弃、正在迁移、暂不需要同步）
- 整体禁用某个 group 或 resource（如服务器维护期间不需要同步）
- 在 sync 后排除不需要跟踪的 repo，而不是手动从配置中删除

目前只能通过删除配置条目来实现"屏蔽"，操作繁琐且容易丢失配置信息。需要为 resource、group、repo 添加 `enabled` 开关，以及在 group 级别添加 `exclude_repos` 列表来屏蔽特定 repo。

## What Changes

- 为 `Resource` 结构体添加 `enabled` 字段（默认 `true`），禁用后该 resource 下所有 group/repo 均不参与操作
- 为 `Group` 结构体添加 `enabled` 字段（默认 `true`），禁用后该 group 下所有 repo 不参与操作
- 为 `Group` 结构体添加 `exclude_repos` 字段，支持通过 repo 名称或路径模式屏蔽特定 repo
- 为 `Repo` 结构体添加 `enabled` 字段（默认 `true`），禁用后该独立 repo 不参与操作
- 修改 repo 解析逻辑（`repo/resolver.go`），在解析阶段过滤掉被禁用的条目
- 修改 `sync` 命令，在写入配置时保留 `exclude_repos` 列表不被新发现的 repo 覆盖
- 修改 `list`、`clone`、`pull`、`status` 等命令，支持 `--all` 标志来包含被禁用的条目（默认行为为排除禁用项）

## Capabilities

### New Capabilities
- `exclusion-toggle`: resource、group、repo 的 enabled 开关和 group 级别的 exclude_repos 排除机制

### Modified Capabilities
- `sync-command`: sync 时保留 exclude_repos 配置，不被覆盖
- `group-management`: group 结构体新增 enabled 和 exclude_repos 字段
- `resource-management`: resource 结构体新增 enabled 字段

## Impact

- **配置结构**：`config/config.go` 中的 `Resource`、`Group`、`Repo` 结构体将新增字段，YAML 配置格式有新增可选字段（向后兼容）
- **解析逻辑**：`repo/resolver.go` 需要在解析阶段增加过滤逻辑
- **CLI 命令**：`cmd/sync.go`、`cmd/list.go`、`cmd/clone.go`、`cmd/pull.go`、`cmd/status.go` 等命令需要适配禁用过滤
- **测试**：`config/config_test.go`、`repo/resolver_test.go` 需要补充测试用例
- **向后兼容**：新增字段均为可选，默认值为 `true`/空列表，对现有配置完全向后兼容
