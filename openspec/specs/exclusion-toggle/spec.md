### Requirement: Resource enabled 开关
Resource 结构体 SHALL 支持 `enabled` 布尔字段（默认 `true`）。当 `enabled` 为 `false` 时，该 resource 下引用它的所有 group 和独立 repo SHALL 被排除在正常操作之外。

#### Scenario: Resource 默认启用
- **WHEN** 配置文件中 resource 未设置 `enabled` 字段或 `enabled: true`
- **THEN** 该 resource 正常参与所有操作

#### Scenario: Resource 禁用后排除所有关联条目
- **WHEN** 配置文件中 resource `work-gl` 设置 `enabled: false`，group `frontend` 和独立 repo `special` 均引用 `work-gl`
- **THEN** `grepom clone`、`grepom pull`、`grepom status`、`grepom list` 均不包含 `frontend` group 下的 repo 和独立 repo `special`

#### Scenario: Resource 禁用不影响配置文件结构
- **WHEN** 配置文件中 resource 设置 `enabled: false`
- **THEN** 配置正常加载，不报错，该 resource 及其关联条目仅在运行时被排除

#### Scenario: 使用 --all 包含被禁用的 Resource
- **WHEN** resource `work-gl` 设置 `enabled: false`，用户运行 `grepom list --all`
- **THEN** 列表中包含该 resource 下的所有 repo，并标注 disabled 状态

### Requirement: Group enabled 开关
Group 结构体 SHALL 支持 `enabled` 布尔字段（默认 `true`）。当 `enabled` 为 `false` 时，该 group 下的所有 repo SHALL 被排除在正常操作之外。

#### Scenario: Group 默认启用
- **WHEN** 配置文件中 group 未设置 `enabled` 字段或 `enabled: true`
- **THEN** 该 group 正常参与所有操作

#### Scenario: Group 禁用后排除其下所有 repo
- **WHEN** 配置文件中 group `frontend` 设置 `enabled: false`，该 group 下有 3 个 repo
- **THEN** `grepom clone`、`grepom pull`、`grepom status`、`grepom list` 均不包含这 3 个 repo

#### Scenario: Group 禁用不影响其他 group
- **WHEN** group `frontend` 设置 `enabled: false`，group `backend` 保持启用
- **THEN** `grepom clone` 仅克隆 `backend` 下的 repo，`frontend` 下的 repo 被排除

#### Scenario: Resource 禁用时 Group enabled 无效
- **WHEN** resource `work-gl` 设置 `enabled: false`，引用它的 group `frontend` 设置 `enabled: true`
- **THEN** group `frontend` 仍被排除（resource 级别优先）

### Requirement: 独立 Repo enabled 开关
独立 Repo 结构体 SHALL 支持 `enabled` 布尔字段（默认 `true`）。当 `enabled` 为 `false` 时，该 repo SHALL 被排除在正常操作之外。

#### Scenario: 独立 Repo 默认启用
- **WHEN** 配置文件中独立 repo 未设置 `enabled` 字段或 `enabled: true`
- **THEN** 该 repo 正常参与所有操作

#### Scenario: 独立 Repo 禁用后被排除
- **WHEN** 配置文件中独立 repo `dotfiles` 设置 `enabled: false`
- **THEN** `grepom clone`、`grepom pull`、`grepom status`、`grepom list` 均不包含该 repo

#### Scenario: Resource 禁用时独立 Repo enabled 无效
- **WHEN** resource `work-gl` 设置 `enabled: false`，引用它的独立 repo `dotfiles` 设置 `enabled: true`
- **THEN** 独立 repo `dotfiles` 仍被排除（resource 级别优先）

### Requirement: Group exclude_repos 排除列表
Group 结构体 SHALL 支持 `exclude_repos` 字段（字符串数组，默认为空）。每个条目的匹配规则如下：
- 不含通配符（`*`、`?`、`[`）的条目：通过 repo 的 `name` 字段精确匹配（原有行为不变）
- 含通配符的条目：通过 repo 的**远端路径**（remote path，即 `GroupRepo.Path`，如 `my-org/frontend/web-app`）进行 `filepath.Match` glob 匹配

被排除的 repo 不参与正常操作。

#### Scenario: exclude_repos 精确匹配 repo name（向后兼容）
- **WHEN** group `frontend` 的 `exclude_repos` 包含 `"deprecated-app"`（无通配符），该 group 下有 repo name 为 `deprecated-app`
- **THEN** `deprecated-app` 被排除

#### Scenario: exclude_repos glob 匹配远端路径前缀
- **WHEN** group `my-group` 的 `exclude_repos` 包含 `"my-org/frontend/*"`，该 group 下有远端路径为 `my-org/frontend/web-app` 的 repo
- **THEN** 该 repo 被排除

#### Scenario: exclude_repos glob 匹配跨层级模式
- **WHEN** group `my-group` 的 `exclude_repos` 包含 `"*/legacy-*"`，该 group 下有远端路径为 `my-org/backend/legacy-api` 和 `my-org/frontend/web-app` 的 repo
- **THEN** 远端路径匹配 `*/legacy-*` 的 `legacy-api` 被排除，`web-app` 不受影响

#### Scenario: exclude_repos 混合使用精确匹配和 glob 匹配
- **WHEN** group `my-group` 的 `exclude_repos` 为 `["deprecated-app", "my-org/frontend/*"]`
- **THEN** name 为 `deprecated-app` 的 repo 被排除（精确匹配），远端路径匹配 `my-org/frontend/*` 的所有 repo 也被排除

#### Scenario: exclude_repos glob 不匹配的 repo 不受影响
- **WHEN** group `my-group` 的 `exclude_repos` 包含 `"my-org/backend/*"`，该 group 下有远端路径为 `my-org/frontend/web-app` 的 repo
- **THEN** `web-app` 正常参与所有操作

#### Scenario: exclude_repos 为空时无排除效果
- **WHEN** group `frontend` 未设置 `exclude_repos` 字段或 `exclude_repos: []`
- **THEN** 该 group 下所有 repo 正常参与操作

#### Scenario: 使用 --all 包含被排除的 repo
- **WHEN** group `frontend` 的 `exclude_repos` 包含 `"my-org/frontend/*"`，用户运行 `grepom list --all`
- **THEN** 列表中包含被 glob 排除的 repo，并标注 excluded 状态

### Requirement: 排除优先级
排除逻辑 SHALL 按以下优先级逐层过滤：Resource enabled → Group enabled → Group exclude_repos → Repo enabled。上层禁用时下层开关无效。

#### Scenario: Resource 禁用优先于 Group exclude_repos
- **WHEN** resource `work-gl` 设置 `enabled: false`，引用它的 group `frontend` 的 `exclude_repos` 为空
- **THEN** group `frontend` 下所有 repo 被排除，无论 exclude_repos 设置如何

#### Scenario: Group enabled 优先于 exclude_repos
- **WHEN** group `frontend` 设置 `enabled: false`，其 `exclude_repos` 为空
- **THEN** 该 group 下所有 repo 被排除，enabled: true 的 repo 也被排除

### Requirement: 排除过滤在 Resolver 层统一处理
所有排除逻辑 SHALL 在 `repo/resolver.go` 的 `Resolve()` 或 `ApplyFilter()` 中实现。`Filter` 结构体 SHALL 新增 `IncludeDisabled` 字段，设为 `true` 时包含被排除的条目。

#### Scenario: Resolve 默认排除被禁用条目
- **WHEN** 调用 `Resolve()` 且 Filter 的 `IncludeDisabled` 为 `false`（默认）
- **THEN** 返回的 repo 列表不包含被禁用或被排除的条目

#### Scenario: ResolveWithAll 包含被禁用条目
- **WHEN** 调用 `Resolve()` 且 Filter 的 `IncludeDisabled` 为 `true`
- **THEN** 返回的 repo 列表包含所有条目，包括被禁用和被排除的条目
