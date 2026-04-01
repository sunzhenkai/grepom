### Requirement: Resource 定义认证和连接信息
系统 SHALL 支持在配置文件的 `resources` 字段下定义命名资源。每个资源包含 `provider`（gitlab 或 github）、`url`（API 地址）和 `token`（认证令牌）。`resources` 使用 YAML map 格式，key 为资源名称。

#### Scenario: 定义 GitLab 资源
- **WHEN** 配置文件中 `resources` 下存在 `work-gl: {provider: gitlab, url: https://gitlab.mycompany.com, token: ${GITLAB_TOKEN}}`
- **THEN** 系统加载该资源，名称为 `work-gl`，provider 为 gitlab，运行时从环境变量解析 token

#### Scenario: 定义 GitHub 资源
- **WHEN** 配置文件中 `resources` 下存在 `github: {provider: github, url: https://github.com, token: abc123}`
- **THEN** 系统加载该资源，名称为 `github`，token 直接使用明文值

#### Scenario: 资源名称必须唯一
- **WHEN** 配置文件中 `resources` map 有多个同名 key
- **THEN** YAML 解析器自然保证 key 唯一，不会出现重复名称

#### Scenario: resource 的 provider 字段必填
- **WHEN** 配置文件中某 resource 缺少 `provider` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: resource 的 url 字段必填
- **WHEN** 配置文件中某 resource 缺少 `url` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: resource 的 url 自动补全 https 前缀
- **WHEN** 配置文件中某 resource 的 `url` 为 `gitlab.mycompany.com`（无 scheme 前缀）
- **THEN** 系统自动补全为 `https://gitlab.mycompany.com`

### Requirement: Group 引用 Resource 获取认证
Group 和独立 repo 通过 `resource` 字段引用已定义的资源名称，获取 provider 类型、URL 和 token。

#### Scenario: Group 引用存在的 resource
- **WHEN** 配置文件中某 group 的 `resource` 字段值为 `work-gl`，且 `resources` 下存在该名称
- **THEN** 系统使用 `work-gl` 资源的 provider/url/token 作为该 group 的认证信息

#### Scenario: Group 引用不存在的 resource
- **WHEN** 配置文件中某 group 的 `resource` 字段值为 `nonexistent`，但 `resources` 下不存在该名称
- **THEN** 系统 SHALL 在加载配置时报错，提示引用的资源不存在

#### Scenario: 独立 repo 引用 resource
- **WHEN** 配置文件中某独立 repo 的 `resource` 字段值为 `github`，且 `resources` 下存在该名称
- **THEN** 系统使用该资源的 provider/url/token 作为该 repo 的认证信息

#### Scenario: 独立 repo 引用不存在的 resource
- **WHEN** 配置文件中某独立 repo 的 `resource` 字段值不存在于 `resources` 中
- **THEN** 系统 SHALL 在加载配置时报错

### Requirement: 多个 Group/Repo 共享同一 Resource
多个 group 和 repo 可以引用同一个 resource，共享认证信息。

#### Scenario: 两个 group 共享 resource
- **WHEN** group `frontend` 和 group `backend` 都引用 `resource: work-gl`
- **THEN** 两者使用相同的 GitLab 实例地址和 token

#### Scenario: group 和独立 repo 共享 resource
- **WHEN** group `frontend` 和独立 repo `special` 都引用 `resource: work-gl`
- **THEN** 两者使用相同的认证信息
