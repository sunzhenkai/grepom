## MODIFIED Requirements

### Requirement: Provider ListGroups 接口
Provider 接口 SHALL 提供 `ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error)` 方法，用于通过 provider API 查询可用的 groups/orgs 列表。

`ListGroupsParams` SHALL 包含 `ServerURL`、`Token` 和 `OrganizationID` 字段。`OrganizationID` 为可选字段，仅 Codeup provider 使用。

`RemoteGroup` SHALL 包含 `Name`（group/org 名称）、`Path`（完整路径）和 `Provider`（provider 类型标识）字段。

#### Scenario: GitLab provider 列出 groups
- **WHEN** 调用 GitLab provider 的 `ListGroups` 方法
- **THEN** 系统通过 `GET /api/v4/groups?per_page=100` 分页查询所有可见 groups，返回包含 Name、Path、Provider 的 RemoteGroup 列表

#### Scenario: GitHub provider 列出 orgs
- **WHEN** 调用 GitHub provider 的 `ListGroups` 方法
- **THEN** 系统通过 `GET /user/orgs` 分页查询用户所属的所有 orgs，返回包含 Name、Path、Provider 的 RemoteGroup 列表

#### Scenario: Codeup provider 列出代码组
- **WHEN** 调用 Codeup provider 的 `ListGroups` 方法，OrganizationID 为 `60de7a6852743a5162b5f957`
- **THEN** 系统通过 `identityGetGroupByPath` 获取根 namespaceId，再通过 `ListRepositoryGroups` 列出一级代码组，返回包含 Name、Path、Provider 的 RemoteGroup 列表

#### Scenario: API 认证失败
- **WHEN** 调用 `ListGroups` 时 token 无效或过期
- **THEN** 系统返回包含 provider 信息的错误（如 "gitlab: authentication failed" 或 "codeup: authentication failed"）

#### Scenario: API 速率限制
- **WHEN** 调用 `ListGroups` 时触发 provider 的速率限制
- **THEN** 系统返回包含重试信息的错误（复用现有的速率限制检测逻辑）
