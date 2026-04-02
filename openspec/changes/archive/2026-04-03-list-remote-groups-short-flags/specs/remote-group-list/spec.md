## ADDED Requirements

### Requirement: Provider ListGroups 接口
Provider 接口 SHALL 提供 `ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error)` 方法，用于通过 provider API 查询可用的 groups/orgs 列表。

`ListGroupsParams` SHALL 包含 `ServerURL` 和 `Token` 字段。

`RemoteGroup` SHALL 包含 `Name`（group/org 名称）、`Path`（完整路径）和 `Provider`（provider 类型标识）字段。

#### Scenario: GitLab provider 列出 groups
- **WHEN** 调用 GitLab provider 的 `ListGroups` 方法
- **THEN** 系统通过 `GET /api/v4/groups?per_page=100` 分页查询所有可见 groups，返回包含 Name、Path、Provider 的 RemoteGroup 列表

#### Scenario: GitHub provider 列出 orgs
- **WHEN** 调用 GitHub provider 的 `ListGroups` 方法
- **THEN** 系统通过 `GET /user/orgs` 分页查询用户所属的所有 orgs，返回包含 Name、Path、Provider 的 RemoteGroup 列表

#### Scenario: API 认证失败
- **WHEN** 调用 `ListGroups` 时 token 无效或过期
- **THEN** 系统返回包含 provider 信息的错误（如 "gitlab: authentication failed"）

#### Scenario: API 速率限制
- **WHEN** 调用 `ListGroups` 时触发 provider 的速率限制
- **THEN** 系统返回包含重试信息的错误（复用现有的速率限制检测逻辑）

### Requirement: list --remote --type groups 命令
系统 SHALL 支持 `grepom list --remote --type groups` 命令，通过 provider API 实时查询远程 groups/orgs 列表。

系统 SHALL 遍历配置中所有 resources，对每个 resource 调用对应 provider 的 `ListGroups` 方法，汇总结果以表格形式输出。

输出 SHALL 包含 NAME、RESOURCE、PATH 三列。

#### Scenario: 远程列出所有 resources 的 groups
- **WHEN** 用户运行 `grepom list --remote --type groups`
- **THEN** 系统遍历所有已配置的 resources，通过各自 provider API 查询 groups，以表格输出 NAME、RESOURCE、PATH

#### Scenario: 远程列出特定 resource 的 groups
- **WHEN** 用户运行 `grepom list --remote --type groups --resource work-gl`
- **THEN** 系统仅查询 resource `work-gl` 对应 provider 的 groups

#### Scenario: 远程列出特定 group 的信息
- **WHEN** 用户运行 `grepom list --remote --type groups --group frontend`
- **THEN** 系统在远程查询结果中按名称过滤，仅显示名称匹配 `frontend` 的 groups

#### Scenario: 无配置 resources
- **WHEN** 用户运行 `grepom list --remote --type groups` 且配置中无 resources
- **THEN** 系统输出 `No resources found.`

#### Scenario: 某个 resource 查询失败
- **WHEN** 某个 resource 的 provider API 查询失败（如网络错误、认证失败）
- **THEN** 系统向 stderr 输出错误信息，继续查询其他 resources，最终输出成功查询到的 groups

#### Scenario: 远程查询无结果
- **WHEN** 所有 resource 查询完成但未找到任何 groups
- **THEN** 系统输出 `No remote groups found.`
