## MODIFIED Requirements

### Requirement: Group 引用 Resource 获取认证
Group 和独立 repo 通过 `resource` 字段引用已定义的资源名称，获取 provider 类型、host 和 token。`resource` 字段为可选。未指定 resource 的 group 或 repo 视为手动管理模式，使用系统默认认证。

#### Scenario: Group 引用存在的 resource
- **WHEN** 配置文件中某 group 的 `resource` 字段值为 `work-gl`，且 `resources` 下存在该名称
- **THEN** 系统使用 `work-gl` 资源的 provider/host/token 作为该 group 的认证信息

#### Scenario: Group 不指定 resource
- **WHEN** 配置文件中某 group 的 `resource` 字段为空或未设置
- **THEN** 系统正常加载该 group，该 group 不具备远程 API 能力，repos 由用户手动维护

#### Scenario: Group 引用不存在的 resource
- **WHEN** 配置文件中某 group 的 `resource` 字段值为 `nonexistent`，但 `resources` 下不存在该名称
- **THEN** 系统 SHALL 在加载配置时报错，提示引用的资源不存在

#### Scenario: 独立 repo 引用 resource
- **WHEN** 配置文件中某独立 repo 的 `resource` 字段值为 `github`，且 `resources` 下存在该名称
- **THEN** 系统使用该资源的 provider/host/token 作为该 repo 的认证信息

#### Scenario: 独立 repo 不指定 resource 但指定 url
- **WHEN** 配置文件中某独立 repo 未指定 `resource`，但指定了 `url: git@github.com:user/repo.git`
- **THEN** 系统正常加载该 repo，clone/pull 使用 url 字段的值

#### Scenario: 独立 repo 不指定 resource 且不指定 url
- **WHEN** 配置文件中某独立 repo 未指定 `resource` 且未指定 `url`
- **THEN** 系统 SHALL 在加载配置时报错，提示必须提供 resource 或 url

#### Scenario: 独立 repo 引用不存在的 resource
- **WHEN** 配置文件中某独立 repo 的 `resource` 字段值不存在于 `resources` 中
- **THEN** 系统 SHALL 在加载配置时报错
