## Why

当前 `exclude_repos` 仅作为运行时过滤器在 `repo/resolver.go` 中生效，覆盖 clone、pull、status、list（本地）、search 等通过 Resolver 的命令。但 `sync` 和 `list --remote` 直接遍历 config 中的 groups，**不经过 Resolver**，导致被排除的仓库在同步发现和远程列表查询中仍然出现。用户配置 `exclude_repos` 的意图是全面排除这些仓库，当前实现存在不一致。

## What Changes

- `sync` 命令在发现远程仓库时，跳过匹配 `exclude_repos` 的仓库，不再将其写入配置
- `list --remote` 命令在展示远程仓库时，过滤掉匹配 `exclude_repos` 的仓库，与 `list`（本地）行为一致
- `list --remote` 增加 `--all` 支持，允许用户查看包含被排除仓库的完整远程列表

## Capabilities

### New Capabilities

（无新增 capability）

### Modified Capabilities

- `sync-command`: sync 发现仓库时跳过 `exclude_repos` 中匹配的仓库，不再写入配置
- `list-remote-repos`: `list --remote` 过滤被排除的仓库，新增 `--all` 标志可查看完整列表

## Impact

- `cmd/sync.go`：在发现新仓库后写入配置前，增加 `exclude_repos` 过滤逻辑
- `cmd/list.go`（`runListRemoteRepos`）：在展示远程仓库时增加 `exclude_repos` 过滤，新增 `--all` flag
- `repo/resolver.go`：可能需要导出 `isExcluded` 辅助函数供 sync/list-remote 复用
- 现有测试需要更新以覆盖新行为
- **不破坏现有配置格式**，`exclude_repos` 字段不变
