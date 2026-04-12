### Requirement: Resource 定义认证和连接信息
系统 SHALL 支持在配置文件的 `resources` 字段下定义命名资源。每个资源包含 `provider`（gitlab、github 或 generic）、`url`（host 地址，可选 port，可选 `http://`/`https://` 前缀）和 `token`（认证令牌）。Resource SHALL 支持 `enabled` 布尔字段（默认 `true`）。`url` 字段可填写纯 host（如 `g.wii.pub`）或带协议前缀（如 `http://g.wii.pub`）。无协议前缀时系统以 auto 模式运行（先 HTTPS 后 HTTP）。`resources` 使用 YAML map 格式，key 为资源名称。

#### Scenario: 定义 GitLab 资源（仅 host）
- **WHEN** 配置文件中 `resources` 下存在 `work-gl: {provider: gitlab, url: g.wii.pub, token: ${GITLAB_TOKEN}}`
- **THEN** 系统加载该资源，host 为 `g.wii.pub`，协议为 auto，运行时从环境变量解析 token

#### Scenario: 定义 GitLab 资源（带端口）
- **WHEN** 配置文件中 `resources` 下存在 `work-gl: {provider: gitlab, url: g.wii.pub:8443, token: ${GITLAB_TOKEN}}`
- **THEN** 系统加载该资源，host 为 `g.wii.pub:8443`，协议为 auto

#### Scenario: 定义 GitLab 资源（http 前缀）
- **WHEN** 配置文件中 `resources` 下存在 `work-gl: {provider: gitlab, url: http://g.wii.pub, token: ${GITLAB_TOKEN}}`
- **THEN** 系统加载该资源，host 为 `g.wii.pub`，协议为 http

#### Scenario: 定义 GitLab 资源（https 前缀）
- **WHEN** 配置文件中 `resources` 下存在 `work-gl: {provider: gitlab, url: https://g.wii.pub, token: ${GITLAB_TOKEN}}`
- **THEN** 系统加载该资源，host 为 `g.wii.pub`，协议为 https

#### Scenario: 定义 GitHub 资源
- **WHEN** 配置文件中 `resources` 下存在 `github: {provider: github, url: github.com, token: abc123}`
- **THEN** 系统加载该资源，host 为 `github.com`，token 直接使用明文值

#### Scenario: 定义 generic 资源
- **WHEN** 配置文件中 `resources` 下存在 `my-git: {provider: generic, url: git.internal.com, token: ${GIT_TOKEN}}`
- **THEN** 系统加载该资源，provider 为 generic，host 为 `git.internal.com`

#### Scenario: 资源名称必须唯一
- **WHEN** 配置文件中 `resources` map 有多个同名 key
- **THEN** YAML 解析器自然保证 key 唯一，不会出现重复名称

#### Scenario: resource 的 provider 字段必填
- **WHEN** 配置文件中某 resource 缺少 `provider` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: resource 的 url 字段必填
- **WHEN** 配置文件中某 resource 缺少 `url` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: resource 使用不支持的 provider
- **WHEN** 配置文件中某 resource 的 `provider` 字段值不在 `gitlab`、`github`、`generic` 中
- **THEN** 系统 SHALL 在加载配置时报错，提示不支持的 provider

#### Scenario: 定义禁用的 Resource
- **WHEN** 配置文件中 resource 设置 `enabled: false`
- **THEN** 系统正常加载该 resource 配置，但运行时排除该 resource 下所有 group 和独立 repo

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

### Requirement: 多个 Group/Repo 共享同一 Resource
多个 group 和 repo 可以引用同一个 resource，共享认证信息。

#### Scenario: 两个 group 共享 resource
- **WHEN** group `frontend` 和 group `backend` 都引用 `resource: work-gl`
- **THEN** 两者使用相同的 host 和 token

#### Scenario: group 和独立 repo 共享 resource
- **WHEN** group `frontend` 和独立 repo `special` 都引用 `resource: work-gl`
- **THEN** 两者使用相同的认证信息
