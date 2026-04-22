## MODIFIED Requirements

### Requirement: list groups 命令
系统 SHALL 提供 `grepom list --type groups` 命令，列出所有已配置的分组。输出以表格形式显示 NAME、RESOURCE、PATH、LOCAL_PATH、RECURSIVE、REPOS 列。

系统 SHALL 支持位置参数 `groups` 作为 `--type groups` 的快捷方式，即 `grepom list groups` SHALL 与 `grepom list --type groups` 行为完全一致。

系统 SHALL 支持 `grepom list --remote --type groups` 命令，通过 provider API 实时查询远程 groups/orgs 列表，输出以表格形式显示 NAME、RESOURCE、PATH 列。

#### Scenario: 列出所有分组
- **WHEN** 用户运行 `grepom list --type groups`
- **THEN** 系统以表格形式输出所有已配置的 group，每行包含名称、关联 resource、远端路径、本地路径、是否递归（yes/no）和 repo 数量

#### Scenario: 位置参数 groups 列出所有分组
- **WHEN** 用户运行 `grepom list groups`
- **THEN** 系统行为与 `grepom list --type groups` 完全一致，以表格形式输出所有已配置的 group

#### Scenario: 配置中无 group
- **WHEN** 用户运行 `grepom list --type groups` 且配置文件中 `groups` 为空
- **THEN** 系统输出 `No groups found.`

#### Scenario: group 下有多个 repo
- **WHEN** 配置中某 group 的 `repos` 列表包含 5 个条目
- **THEN** 表格中 REPOS 列显示 `5`

#### Scenario: group 下无 repo
- **WHEN** 配置中某 group 的 `repos` 列表为空或省略
- **THEN** 表格中 REPOS 列显示 `0`

#### Scenario: group 的 recursive 字段
- **WHEN** 配置中某 group 的 `recursive` 为 `true`
- **THEN** 表格中 RECURSIVE 列显示 `yes`

#### Scenario: group 的 recursive 默认值
- **WHEN** 配置中某 group 未设置 `recursive` 字段
- **THEN** 表格中 RECURSIVE 列显示 `no`

#### Scenario: --type groups 忽略过滤标志
- **WHEN** 用户运行 `grepom list --type groups --group frontend --resource work-gl`
- **THEN** 系统忽略 `--group` 和 `--resource` 标志，仍列出所有 groups

#### Scenario: 远程列出所有 resources 的 groups
- **WHEN** 用户运行 `grepom list --remote --type groups`
- **THEN** 系统遍历所有已配置的 resources，通过各自 provider API 查询 groups，以表格输出 NAME、RESOURCE、PATH

#### Scenario: 远程列出特定 resource 的 groups
- **WHEN** 用户运行 `grepom list --remote --type groups --resource work-gl`
- **THEN** 系统仅查询 resource `work-gl` 对应 provider 的 groups

#### Scenario: 远程列出 groups 无结果
- **WHEN** 用户运行 `grepom list --remote --type groups` 且所有 resource 查询无结果
- **THEN** 系统输出 `No remote groups found.`
