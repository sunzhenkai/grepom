## MODIFIED Requirements

### Requirement: list groups 命令
系统 SHALL 提供 `grepom list --type groups` 命令，列出所有已配置的真实 group 和虚拟分组。输出以表格形式显示 TYPE、NAME、RESOURCE、PATH、LOCAL_PATH、RECURSIVE、REPOS、GROUPS 列。真实 group 行的 TYPE 为 `group`，GROUPS 列显示 `-`；虚拟分组行的 TYPE 为 `vgroup`，RESOURCE、PATH、LOCAL_PATH、RECURSIVE 列显示 `-`，GROUPS 列显示成员真实 group 名称列表，REPOS 列显示成员真实 groups 的 repo 总数。

系统 SHALL 支持位置参数 `groups` 作为 `--type groups` 的快捷方式，即 `grepom list groups` SHALL 与 `grepom list --type groups` 行为完全一致。

系统 SHALL 支持 `grepom list --remote --type groups` 命令，通过 provider API 实时查询远程 groups/orgs 列表，输出以表格形式显示 NAME、RESOURCE、PATH 列。远程 groups 列表 SHALL 只表示 provider 返回的远程 group/org，不展示本地虚拟分组。

#### Scenario: 列出所有分组
- **WHEN** 用户运行 `grepom list --type groups`
- **THEN** 系统以表格形式输出所有已配置的真实 group 和虚拟分组，每行通过 TYPE 区分 `group` 或 `vgroup`

#### Scenario: 位置参数 groups 列出所有分组
- **WHEN** 用户运行 `grepom list groups`
- **THEN** 系统行为与 `grepom list --type groups` 完全一致，以表格形式输出所有已配置的真实 group 和虚拟分组

#### Scenario: 配置中无 group
- **WHEN** 用户运行 `grepom list --type groups` 且配置文件中 `groups` 和 `virtual_groups` 均为空
- **THEN** 系统输出 `No groups found.`

#### Scenario: group 下有多个 repo
- **WHEN** 配置中某真实 group 的 `repos` 列表包含 5 个条目
- **THEN** 该真实 group 行的 REPOS 列显示 `5`

#### Scenario: group 下无 repo
- **WHEN** 配置中某真实 group 的 `repos` 列表为空或省略
- **THEN** 该真实 group 行的 REPOS 列显示 `0`

#### Scenario: group 的 recursive 字段
- **WHEN** 配置中某真实 group 的 `recursive` 为 `true`
- **THEN** 该真实 group 行的 RECURSIVE 列显示 `yes`

#### Scenario: group 的 recursive 默认值
- **WHEN** 配置中某真实 group 未设置 `recursive` 字段
- **THEN** 该真实 group 行的 RECURSIVE 列显示 `no`

#### Scenario: --type groups 忽略过滤标志
- **WHEN** 用户运行 `grepom list --type groups --group frontend --vgroup work --resource work-gl`
- **THEN** 系统忽略 `--group`、`--vgroup` 和 `--resource` 标志，仍列出所有真实 group 和虚拟分组

#### Scenario: 列出虚拟分组
- **WHEN** 配置中虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`，且二者 repo 总数为 7
- **THEN** `grepom list groups` 输出一行 TYPE 为 `vgroup`、NAME 为 `work`、REPOS 为 `7`、GROUPS 包含 `frontend,backend` 的记录

#### Scenario: 远程列出所有 resources 的 groups
- **WHEN** 用户运行 `grepom list --remote --type groups`
- **THEN** 系统遍历所有已配置的 resources，通过各自 provider API 查询远程 groups，以表格输出 NAME、RESOURCE、PATH，不展示本地虚拟分组

#### Scenario: 远程列出特定 resource 的 groups
- **WHEN** 用户运行 `grepom list --remote --type groups --resource work-gl`
- **THEN** 系统仅查询 resource `work-gl` 对应 provider 的 groups

#### Scenario: 远程列出 groups 无结果
- **WHEN** 用户运行 `grepom list --remote --type groups` 且所有 resource 查询无结果
- **THEN** 系统输出 `No remote groups found.`
