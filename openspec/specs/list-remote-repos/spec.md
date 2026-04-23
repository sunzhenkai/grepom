### Requirement: list --remote 远程仓库查询
`list` 命令 SHALL 支持 `--remote` 标志，当使用 `--remote` 时（默认 `--type repos`），系统通过 provider API 实时查询远程仓库列表并以表格形式展示。`list --remote` SHALL 通过 `Resource.ResolvedToken()` 获取已解析的 token 用于 provider API 认证，不再直接读取 `Resource.Token` 字段。

`list --remote` SHALL 默认过滤掉匹配 group `exclude_repos` 列表的仓库，不展示被排除的仓库。当同时使用 `--remote --all` 时，SHALL 展示所有仓库包括被排除的仓库，并为被排除的仓库标注 `[excluded]`。`list --remote` SHALL 跳过 `enabled: false` 的 group 和 resource（除非使用 `--all`）。

对于 Codeup provider，`list --remote` SHALL 使用与 GitLab 相同的 Groups 查询模式（非 GitHub 的 Orgs 模式）。

输出列 SHALL 包含：`NAME`、`PATH`、`GROUP`、`RESOURCE`、`CLONE_URL`。

#### Scenario: 查询所有远程仓库
- **WHEN** 用户运行 `grepom list --remote`
- **THEN** 系统遍历配置中所有启用的 groups，通过 `res.ResolvedToken()` 获取已解析 token，通过各 group 关联的 provider API 查询远程仓库，过滤掉匹配 `exclude_repos` 的仓库后以表格形式展示

#### Scenario: list --remote token 环境变量未设置时报错
- **WHEN** 用户运行 `grepom list --remote`，某 resource token 为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 未设置
- **THEN** 系统向 stderr 输出包含 resource 名称和环境变量名的错误信息，跳过该 group，继续查询其他 groups

#### Scenario: 按 group 过滤远程仓库
- **WHEN** 用户运行 `grepom list --remote --group frontend`
- **THEN** 系统仅查询 group `frontend` 关联的远程仓库并展示

#### Scenario: 按 resource 过滤远程仓库
- **WHEN** 用户运行 `grepom list --remote --resource work-gl`
- **THEN** 系统仅查询引用 resource `work-gl` 的 groups 所关联的远程仓库并展示

#### Scenario: 远程查询无结果
- **WHEN** 用户运行 `grepom list --remote` 且远程无仓库（或所有仓库均被排除）
- **THEN** 系统输出 `No remote repositories found.`

#### Scenario: --remote 与 --type resources 不兼容
- **WHEN** 用户运行 `grepom list --remote --type resources`
- **THEN** 系统报错提示 `--remote` 仅支持 `--type repos`（默认）

#### Scenario: --remote 与 --type groups 不兼容
- **WHEN** 用户运行 `grepom list --remote --type groups`
- **THEN** 系统报错提示 `--remote` 仅支持 `--type repos`（默认）

#### Scenario: 远程查询 API 错误
- **WHEN** 用户运行 `grepom list --remote` 且某个 group 的 provider API 调用失败
- **THEN** 系统向 stderr 输出错误信息，继续查询其他 groups，最终展示成功查询到的仓库

#### Scenario: 远程查询遇到 rate limit
- **WHEN** 用户运行 `grepom list --remote` 且 provider 返回 rate limit 错误
- **THEN** 系统向 stderr 输出 rate limit 提示信息，跳过该 group 继续查询其他 groups

#### Scenario: list --remote 过滤被排除的仓库
- **WHEN** 用户运行 `grepom list --remote`，group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`
- **THEN** 展示结果中不包含 `deprecated-app`

#### Scenario: list --remote --all 包含被排除的仓库
- **WHEN** 用户运行 `grepom list --remote --all`，group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`
- **THEN** 展示结果中包含 `deprecated-app`，NAME 列显示为 `deprecated-app [excluded]`

#### Scenario: list --remote 跳过禁用的 group
- **WHEN** 用户运行 `grepom list --remote`，group `frontend` 设置 `enabled: false`
- **THEN** 系统跳过 `frontend`，不查询其远程仓库

#### Scenario: list --remote --all 包含禁用 group 的仓库
- **WHEN** 用户运行 `grepom list --remote --all`，group `frontend` 设置 `enabled: false`
- **THEN** 系统查询 `frontend` 的远程仓库并展示，NAME 列标注 `[disabled]`

#### Scenario: Codeup provider 使用 Groups 模式查询
- **WHEN** 用户运行 `grepom list --remote`，group 的 resource provider 为 `codeup`
- **THEN** 系统使用 Groups 查询模式（与 GitLab 一致），通过 `ListRepos` 全量拉取并按代码组路径过滤
