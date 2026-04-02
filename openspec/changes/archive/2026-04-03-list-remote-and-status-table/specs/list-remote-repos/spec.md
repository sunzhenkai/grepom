## ADDED Requirements

### Requirement: list --remote 远程仓库查询
`list` 命令 SHALL 支持 `--remote` 标志，当使用 `--remote` 时（默认 `--type repos`），系统通过 provider API 实时查询远程仓库列表并以表格形式展示。

输出列 SHALL 包含：`NAME`、`PATH`、`GROUP`、`RESOURCE`、`CLONE_URL`。

#### Scenario: 查询所有远程仓库
- **WHEN** 用户运行 `grepom list --remote`
- **THEN** 系统遍历配置中所有 groups，通过各 group 关联的 provider API 查询远程仓库，以表格形式展示所有发现的仓库

#### Scenario: 按 group 过滤远程仓库
- **WHEN** 用户运行 `grepom list --remote --group frontend`
- **THEN** 系统仅查询 group `frontend` 关联的远程仓库并展示

#### Scenario: 按 resource 过滤远程仓库
- **WHEN** 用户运行 `grepom list --remote --resource work-gl`
- **THEN** 系统仅查询引用 resource `work-gl` 的 groups 所关联的远程仓库并展示

#### Scenario: 远程查询无结果
- **WHEN** 用户运行 `grepom list --remote` 且远程无仓库
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
