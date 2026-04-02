## Context

grepom 是一个多仓库管理 CLI 工具，支持 GitLab 和 GitHub 两个 provider。当前 `list --remote` 通过 provider API 查询仓库列表，但仅限于 repos 类型。groups 类型的远程查询被硬编码禁止。

现有 flag 注册模式：全局 flag（`-c`/`-v`）使用 `VarP` 注册短别名，所有子命令本地 flag 使用 `Var` 无短别名。

Provider 接口当前仅有 `ListRepos` 方法，GitLab 通过 `GET /api/v4/groups/{id}/projects` 获取项目，GitHub 通过 `GET /orgs/{org}/repos` 获取仓库。

## Goals / Non-Goals

**Goals:**
- 支持 `list --remote --type groups` 查询 provider API 获取可用的 groups/orgs 列表
- 为所有子命令的常用 flag 添加短别名，减少日常使用输入量
- 保持与现有行为的向后兼容

**Non-Goals:**
- 不支持 `list --remote --type resources`（resources 是纯本地配置概念，无远程 API 对应）
- 不添加子命令的短别名（如 `ls` → `list`），仅添加 flag 的短别名
- 不修改 provider API 的分页逻辑或错误处理模式

## Decisions

### 1. Provider 接口扩展：新增 ListGroups 方法

**选择**：在 `Provider` 接口中新增 `ListGroups(ctx, params) ([]RemoteGroup, error)` 方法。

**替代方案**：
- A) 复用 `ListReposParams`：不适合，groups 查询不需要 orgs/groups 路径参数
- B) 新建 `ListGroupsParams` 结构体：**选择此方案**，语义清晰

**理由**：groups 查询的参数和返回类型与 repos 完全不同。GitLab 查询 groups 使用 `GET /api/v4/groups`（支持搜索和分页），GitHub 使用 `GET /user/orgs`。返回的是 group/org 信息而非 repo 信息。

```go
type RemoteGroup struct {
    Name     string // group/org 名称
    Path     string // 完整路径（GitLab: full_path, GitHub: login）
    Provider string // "gitlab" 或 "github"
}

type ListGroupsParams struct {
    ServerURL string
    Token     string
}
```

### 2. 远程 groups 查询的数据源

**选择**：遍历配置中所有 resources，对每个 resource 调用对应的 provider 查询 groups。

**理由**：与 `runListRemoteRepos` 的模式一致——基于本地配置中的 resources 来确定查询哪些 provider。这样用户只需配置 resource，即可远程浏览可用的 groups/orgs。

**输出列**：`NAME`、`RESOURCE`、`PATH`。不显示 `LOCAL_PATH`、`RECURSIVE`、`REPOS` 等本地配置字段，因为这些是本地配置概念。

### 3. Flag 短别名分配

**选择**：为各子命令的常用 flag 添加短别名，遵循以下原则：
- 同一命令内不冲突
- 常用 flag 优先分配直觉性强的单字母
- 不与全局 flag（`-c`、`-v`）冲突

**短别名映射表**：

| Flag | 短别名 | 适用命令 |
|------|--------|----------|
| `--group` | `-g` | list, clone, status, pull, search, sync |
| `--resource` | `-R` | list, clone, status, pull, search, sync |
| `--type` | `-t` | list |
| `--remote` | `-r` | list |
| `--force` | `-f` | pull |
| `--concurrency` | `-n` | clone, pull |
| `--name` | `-n` | add resource, add group, add repo |
| `--provider` | `-p` | add resource, init |
| `--url` | `-u` | add resource, init, add repo |
| `--token` | `-k` | add resource, add group, add repo, init |
| `--ssh-key` | `-s` | add resource, add group, add repo |
| `--path` | `-p` | add group, add repo |
| `--local-path` | `-l` | add group, add repo, init |
| `--recursive` | `-r` | add group |
| `--base` | `-b` | init |

**注意**：`--resource` 使用 `-R`（大写）以与 `--recursive` 的 `-r`（小写）区分。`--concurrency` 在 clone/pull 中使用 `-n`，与 add 子命令的 `--name` 使用 `-n` 不冲突（不同命令上下文）。

### 4. `--remote` 限制的调整

**选择**：将硬编码的 `listRemote && listType != "repos"` 检查改为仅排除 `resources`，允许 `groups`。

```go
if listRemote && listType == "resources" {
    return fmt.Errorf("--remote is not supported with --type resources")
}
```

**理由**：`resources` 是纯本地配置（provider 凭据），无远程 API 可查。`groups` 则有对应的远程 API。

## Risks / Trade-offs

- **[GitHub API 限制]** `GET /user/orgs` 仅返回用户所属的 orgs，无法列出所有公开 org → 在输出中说明此限制
- **[GitLab groups 数量]** 自托管 GitLab 实例可能有大量 groups，API 默认分页 20 条/页 → 复用现有分页逻辑，遍历所有页
- **[短别名记忆负担]** 短别名增多可能需要用户记忆 → 通过 `--help` 展示所有短别名，保持长 flag 依然可用
- **[接口变更]** `Provider` 接口新增方法会影响所有实现 → 当前仅有 GitLab 和 GitHub 两个实现，影响可控
