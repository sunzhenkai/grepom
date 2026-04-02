## MODIFIED Requirements

### Requirement: list command
系统 SHALL 提供 `grepom list` 命令，支持通过 `--type` 标志切换列出目标。`--type` 支持三个值：`repos`（默认）、`resources`、`groups`。

当 `--type` 为 `repos`（默认）时，行为与原有 list 命令一致：列出所有已发现的仓库，支持位置参数 `[name]` 按名称过滤、`--group` 按分组过滤、`--resource` 按 resource 过滤。

当 `--type` 为 `resources` 或 `groups` 时，位置参数和过滤标志不生效。

`--remote` 标志 SHALL 支持 `--type groups`，通过 provider API 查询远程 groups/orgs 列表。`--remote` 不支持 `--type resources`。

list 命令的 flag SHALL 支持短别名：`-t`（`--type`）、`-r`（`--remote`）、`-g`（`--group`）、`-R`（`--resource`）。

#### Scenario: List all repos
- **WHEN** user runs `grepom list`
- **THEN** the system displays all repos from all groups and independent repos, with name, path, provider, and clone status

#### Scenario: List single repo
- **WHEN** user runs `grepom list web-app`
- **THEN** the system displays info for only `web-app`

#### Scenario: List by group
- **WHEN** user runs `grepom list --group frontend`
- **THEN** the system displays repos only from group `frontend`

#### Scenario: List by group using short flag
- **WHEN** user runs `grepom list -g frontend`
- **THEN** the system displays repos only from group `frontend`，行为与 `--group` 完全一致

#### Scenario: List by resource
- **WHEN** user runs `grepom list --resource work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`

#### Scenario: List by resource using short flag
- **WHEN** user runs `grepom list -R work-gl`
- **THEN** the system displays repos from all groups and independent repos that reference resource `work-gl`，行为与 `--resource` 完全一致

#### Scenario: --type repos 等同默认行为
- **WHEN** user runs `grepom list --type repos`
- **THEN** 系统行为与 `grepom list` 完全一致，列出所有 repos

#### Scenario: --type repos 使用短别名
- **WHEN** user runs `grepom list -t repos`
- **THEN** 系统行为与 `grepom list --type repos` 完全一致

#### Scenario: --type 列出 resources
- **WHEN** user runs `grepom list --type resources`
- **THEN** 系统列出所有已配置的 resources，输出包含名称、provider、url 和 ssh_key

#### Scenario: --type 列出 groups
- **WHEN** user runs `grepom list --type groups`
- **THEN** 系统列出所有已配置的 groups，输出包含名称、关联 resource、路径、recursive 和 repo 数量

#### Scenario: --remote --type groups 远程列出 groups
- **WHEN** user runs `grepom list --remote --type groups`
- **THEN** 系统通过 provider API 查询远程 groups/orgs 列表并输出

#### Scenario: --remote --type groups 使用短别名
- **WHEN** user runs `grepom list -r -t groups`
- **THEN** 系统通过 provider API 查询远程 groups/orgs 列表并输出，行为与长 flag 版本完全一致

#### Scenario: --remote 不支持 --type resources
- **WHEN** user runs `grepom list --remote --type resources`
- **THEN** 系统输出错误信息 "--remote is not supported with --type resources"

#### Scenario: 使用混合短别名
- **WHEN** user runs `grepom list -r -t groups -R work-gl`
- **THEN** 系统仅查询 resource `work-gl` 的远程 groups，行为与 `grepom list --remote --type groups --resource work-gl` 完全一致
