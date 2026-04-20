## Why

当前 `exclude_repos` 机制只在 sync/clone/pull 时跳过匹配的 repos，但已经被克隆到磁盘的 excluded repos 不会被清理，残留目录占用空间且造成混淆。此外，多个 group 之间可能存在同名 repo（如不同 team 各有一个 `api-lib`），用户希望按优先级去重——将低优先级 group 中的同名 repo 排除，但当前只能手动编辑 `exclude_repos`。

## What Changes

- 新增 `grepom prune` 命令：扫描已被 exclude 的 repos，检查本地磁盘克隆状态，安全删除（默认 dry-run，需 `--apply` 才执行删除）。dirty 或 ahead 的 repo 会被跳过，`--force` 可强制删除。
- 新增 `grepom dedup` 命令：指定目标 group（`--group`）和参考 group(s)（`--reference`，可选，默认对比所有其他 group），将目标 group 中与参考 group 同名的 repo 加入 `exclude_repos` 并从 `repos` 列表移除。`--dry-run` 预览不改。

## Capabilities

### New Capabilities
- `prune-command`: 清理已克隆但被 exclude 的 repos，支持安全检查、dry-run 和 force 模式
- `dedup-command`: 跨 group 同名 repo 去重，自动为指定 group 添加 exclude_repos

### Modified Capabilities
<!-- 无需修改现有 capability 的 spec 级别行为 -->

## Impact

- 新增两个 cobra command：`cmd/prune.go`、`cmd/dedup.go`
- `config/config.go`：可能需要新增从 group 的 repos 列表中移除指定 repo 的方法（如 `RemoveGroupRepo`）
- 依赖现有 `repo/resolver.go` 的 `IsExcluded`、`resolveInternal` 等逻辑
- 依赖现有 `git/git.go` 的 `IsCloned`、`GetStatus` 等逻辑
