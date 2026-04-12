## MODIFIED Requirements

### Requirement: Group 内 repo 路径自动推导
Group 内 repo 的本地完整路径 SHALL 通过公式自动推导：`base + group.local_path + trimPrefix(repo.path, group.path)`。Group 内 repo 不存储 `local_path` 字段。无 resource 的 group，其 repo 的 url 字段 SHALL 由用户手动填写完整 clone URL。

#### Scenario: 直接子项目
- **WHEN** group path 为 `my-org/frontend`，local_path 为 `./frontend`，repo path 为 `my-org/frontend/shared-utils`
- **THEN** 本地路径推导为 `<base>/frontend/shared-utils`

#### Scenario: 多级 subgroup 下的项目
- **WHEN** group path 为 `my-org/frontend`，local_path 为 `./frontend`，repo path 为 `my-org/frontend/ui/design-system`
- **THEN** 本地路径推导为 `<base>/frontend/ui/design-system`

#### Scenario: 三级 subgroup 下的项目
- **WHEN** group path 为 `my-org/frontend`，local_path 为 `./frontend`，repo path 为 `my-org/frontend/ui/components/button-lib`
- **THEN** 本地路径推导为 `<base>/frontend/ui/components/button-lib`

#### Scenario: repo path 恰好等于 group path
- **WHEN** group path 为 `my-org/frontend`，repo path 也为 `my-org/frontend`（项目名和 group 同名）
- **THEN** 本地路径推导为 `<base>/frontend/`（trimPrefix 后为空，路径为 group local_path 本身）

#### Scenario: 无 resource 的 group 内 repo 使用手动 url
- **WHEN** group 未绑定 resource，其 repo 的 url 为 `git@github.com:org/repo.git`
- **THEN** clone/pull 时直接使用该 url，不从 resource 推导
