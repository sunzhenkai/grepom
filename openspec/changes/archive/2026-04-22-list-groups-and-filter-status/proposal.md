## Why

用户在使用 `grepom list` 时缺少两种常用操作：一是需要用 `--type groups` 才能列出所有 groups，命令较长不够直观；二是在管理多个仓库时，经常需要快速找出"有未推送提交"或"有未提交更改"的仓库，当前只能先运行 `grepom status` 再人工筛选。这两个痛点降低了日常使用效率。

## What Changes

- `grepom list` 的位置参数支持 `groups` 关键字，使其等价于 `--type groups`，提供更简洁的命令写法（如 `grepom list groups`）
- `grepom list` 新增 `--no-push` 标志，仅展示有未推送提交（ahead > 0）的仓库
- `grepom list` 新增 `--no-commit` 标志，仅展示有未提交更改（dirty > 0）的仓库
- `--no-push` 和 `--no-commit` 可以组合使用，展示满足任一条件的仓库（并集）
- 这两个标志仅对已克隆的仓库有效，未克隆的仓库在这些模式下不展示

## Capabilities

### New Capabilities
- `list-status-filter`: `grepom list` 命令支持按 git 状态筛选仓库，通过 `--no-push`（未推送）和 `--no-commit`（未提交）标志过滤

### Modified Capabilities
- `cli-commands`: `list` 命令的位置参数新增 `groups` 关键字支持，等价于 `--type groups`
- `group-list`: `list groups` 作为位置参数时，行为与 `list --type groups` 完全一致

## Impact

- **cmd/list.go**: 修改 `RunE` 函数，新增 `groups` 位置参数解析逻辑和状态筛选标志处理
- **cmd/list.go**: `runListRepos` 函数需要集成 `git.GetStatus` 进行状态过滤
- **repo/resolver.go**: `Filter` 结构体可能需要扩展以支持状态过滤（或在 cmd 层处理）
- **git/git.go**: 无需修改，现有 `GetStatus` 已提供所需信息
- 无 breaking changes，所有现有命令行为保持不变
