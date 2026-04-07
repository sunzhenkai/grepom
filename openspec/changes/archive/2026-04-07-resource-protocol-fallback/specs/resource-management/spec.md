## MODIFIED Requirements

### Requirement: Resource 定义认证和连接信息
系统 SHALL 支持在配置文件的 `resources` 字段下定义命名资源。每个资源包含 `provider`（gitlab 或 github）、`url`（host 地址，可选 port，可选 `http://`/`https://` 前缀）和 `token`（认证令牌）。Resource SHALL 支持 `enabled` 布尔字段（默认 `true`）。`url` 字段可填写纯 host（如 `g.wii.pub`）或带协议前缀（如 `http://g.wii.pub`）。无协议前缀时系统以 auto 模式运行（先 HTTPS 后 HTTP）。`resources` 使用 YAML map 格式，key 为资源名称。

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

#### Scenario: 资源名称必须唯一
- **WHEN** 配置文件中 `resources` map 有多个同名 key
- **THEN** YAML 解析器自然保证 key 唯一，不会出现重复名称

#### Scenario: resource 的 provider 字段必填
- **WHEN** 配置文件中某 resource 缺少 `provider` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: resource 的 url 字段必填
- **WHEN** 配置文件中某 resource 缺少 `url` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: resource 的 url 保留协议前缀信息
- **WHEN** 配置文件中某 resource 的 `url` 为 `https://g.wii.pub`（含协议前缀）
- **THEN** 系统解析为 host `g.wii.pub` 和协议 `https`

#### Scenario: resource 的 url 解析 http 前缀
- **WHEN** 配置文件中某 resource 的 `url` 为 `http://g.wii.pub:8080`（含 http 前缀和端口）
- **THEN** 系统解析为 host `g.wii.pub:8080` 和协议 `http`

#### Scenario: resource 的 url 无前缀为 auto
- **WHEN** 配置文件中某 resource 的 `url` 为 `g.wii.pub`（无协议前缀）
- **THEN** 系统解析为 host `g.wii.pub`，协议为空（auto）

#### Scenario: 定义禁用的 Resource
- **WHEN** 配置文件中 resource 设置 `enabled: false`
- **THEN** 系统正常加载该 resource 配置，但运行时排除该 resource 下所有 group 和独立 repo
