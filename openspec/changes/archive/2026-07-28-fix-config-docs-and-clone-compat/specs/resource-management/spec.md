## ADDED Requirements

### Requirement: resources list 格式兼容加载
系统 SHALL 在加载配置时同时接受 `resources` 的 YAML map 与 list 两种写法。map 的 key 为资源名；list 中每一项 MUST 包含非空 `name` 字段，系统将该 `name` 作为 map key 归一为内部 `map[string]Resource`。正式文档与 `grepom example` SHALL 以 map 格式为准。

#### Scenario: map 格式正常加载
- **WHEN** 配置文件使用 `resources: { my-gitlab: { provider: gitlab, url: gitlab.example.com, token: ${GITLAB_TOKEN} } }`
- **THEN** 系统加载成功，可通过名称 `my-gitlab` 查找该 resource

#### Scenario: list 格式兼容加载
- **WHEN** 配置文件使用 `resources: [{ name: my-gitlab, provider: gitlab, url: gitlab.example.com, token: ${GITLAB_TOKEN} }]`
- **THEN** 系统加载成功，内部归一为名为 `my-gitlab` 的 resource，行为与 map 格式一致

#### Scenario: list 项缺少 name
- **WHEN** `resources` 为 list，且某项缺少 `name` 或 `name` 为空
- **THEN** 系统 SHALL 在加载配置时报错，提示 list 项必须包含 name

#### Scenario: list 中 name 重复
- **WHEN** `resources` 为 list，且两个项的 `name` 相同
- **THEN** 系统 SHALL 在加载配置时报错，提示资源名称重复

#### Scenario: resources 既非 map 也非 list
- **WHEN** `resources` 节点类型不是 mapping 或 sequence
- **THEN** 系统 SHALL 在加载配置时报错

## MODIFIED Requirements

### Requirement: Resource 定义认证和连接信息
系统 SHALL 支持在配置文件的 `resources` 字段下定义命名资源。每个资源包含 `provider`（gitlab、github、generic 或 codeup）、`url`（host 地址，可选 port，可选 `http://`/`https://` 前缀）和 `token`（认证令牌，支持 `${ENV_VAR}` 占位符）。Resource SHALL 支持 `enabled` 布尔字段（默认 `true`）。`url` 字段可填写纯 host（如 `g.wii.pub`）或带协议前缀（如 `http://g.wii.pub`）。无协议前缀时系统以 auto 模式运行（先 HTTPS 后 HTTP）。`resources` 的正式格式为 YAML map（key 为资源名称）；系统同时兼容带 `name` 字段的 list 写法（见 resources list 格式兼容加载）。

Resource SHALL 支持可选的 `organization_id` 字段（字符串），用于 Codeup 等需要企业标识的 provider。`organization_id` 对 GitLab、GitHub、Generic provider 无影响。

Resource 的 token 环境变量 SHALL 在实际使用该 resource 时解析，而非在配置加载时。配置加载时仅校验 provider、url 等结构字段，不校验 token 对应的环境变量。

#### Scenario: 定义 GitLab 资源（仅 host）
- **WHEN** 配置文件中 `resources` 下存在 `work-gl: {provider: gitlab, url: g.wii.pub, token: ${GITLAB_TOKEN}}`
- **THEN** 系统加载该资源，host 为 `g.wii.pub`，协议为 auto，token 保留占位符直到实际使用时解析

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

#### Scenario: 定义 Codeup 资源
- **WHEN** 配置文件中 `resources` 下存在 `my-codeup: {provider: codeup, url: codeup.aliyun.com, token: ${CODEUP_TOKEN}, organization_id: "60de7a6852743a5162b5f957"}`
- **THEN** 系统加载该资源，provider 为 codeup，host 为 `codeup.aliyun.com`，organization_id 为 `"60de7a6852743a5162b5f957"`

#### Scenario: Codeup 资源缺少 organization_id
- **WHEN** 配置文件中 `resources` 下存在 `my-codeup: {provider: codeup, url: codeup.aliyun.com, token: xxx}`（无 organization_id）
- **THEN** 系统 SHALL 在加载配置时报错，提示 Codeup 资源必须提供 organization_id

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
- **WHEN** 配置文件中某 resource 的 `provider` 字段值不在 `gitlab`、`github`、`generic`、`codeup` 中
- **THEN** 系统 SHALL 在加载配置时报错，提示不支持的 provider

#### Scenario: 定义禁用的 Resource
- **WHEN** 配置文件中 resource 设置 `enabled: false`
- **THEN** 系统正常加载该 resource 配置，运行时排除该 resource 下所有 group 和独立 repo，且不解析该 resource 的 token 环境变量

#### Scenario: organization_id 对非 Codeup provider 无影响
- **WHEN** GitLab resource 配置了 `organization_id` 字段
- **THEN** 系统正常加载，`organization_id` 字段被忽略，不影响 GitLab 功能
