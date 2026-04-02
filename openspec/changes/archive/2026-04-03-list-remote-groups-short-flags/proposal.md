## Why

当前 `grepom list --remote` 仅支持 `--type repos`（默认），无法远程查询 groups/orgs。当用户需要查看 provider 上有哪些可用的 group 时，只能手动登录 GitLab/GitHub 查找，再回来配置，体验不佳。同时，所有子命令的 flag 均无短别名，频繁使用时输入冗长（如 `grepom list --type groups --group frontend --resource work-gl`）。

## What Changes

1. **支持 `list --remote --type groups`**：通过 provider API 实时查询远程 groups/orgs 列表，基于配置的 resources 进行查询，输出 NAME、RESOURCE、PATH 列。
2. **为常用 flag 添加短别名**：为各子命令的常用 flag 注册短标志，减少输入长度。

## Capabilities

### New Capabilities
- `remote-group-list`: 通过 provider API 远程查询 groups/orgs 列表的能力（Provider 接口新增 ListGroups 方法，GitLab/GitHub 各自实现）

### Modified Capabilities
- `cli-commands`: 为常用 flag 添加短别名（如 `-g`/`-R`/`-t`/`-r` 等），更新 list 命令的 `--remote` 支持 `--type groups`
- `group-list`: 新增 `--remote` 远程列出 groups 的场景

## Impact

- **Provider 接口**：新增 `ListGroups` 方法，`GitLabProvider` 和 `GitHubProvider` 需要实现
- **list 命令**：移除 `--remote` 仅限 `--type repos` 的硬编码限制，新增 `runListRemoteGroups` 函数
- **所有子命令的 flag 注册**：从 `StringVar`/`BoolVar` 改为 `StringVarP`/`BoolVarP`，增加短别名
- **依赖**：无新增外部依赖，GitLab API `GET /api/v4/groups` 和 GitHub API `GET /user/orgs` 均为已有端点
