## MODIFIED Requirements

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
