### Requirement: 配置加载时不解析 token 环境变量
`config.Load()` SHALL 仅保存 token 字段的原始值（可能是 `${ENV_VAR}` 占位符或明文），不在加载阶段解析环境变量。加载 SHALL 不因环境变量未设置而失败。

#### Scenario: 环境变量未设置时配置加载成功
- **WHEN** 配置文件中 resource 的 token 为 `${GITLAB_TOKEN}`，但环境变量 `GITLAB_TOKEN` 未设置
- **THEN** `config.Load()` 成功返回 Config 对象，token 字段保留 `${GITLAB_TOKEN}` 原始字符串

#### Scenario: 多个 resource 部分环境变量未设置
- **WHEN** 配置文件中有两个 resource，token 分别为 `${GITHUB_TOKEN}`（已设置）和 `${GITLAB_TOKEN}`（未设置）
- **THEN** `config.Load()` 成功返回，两个 token 均保留原始占位符字符串

#### Scenario: 禁用的 resource 环境变量未设置
- **WHEN** 配置文件中某 resource 设置 `enabled: false`，token 为 `${UNSET_VAR}`
- **THEN** `config.Load()` 成功返回，不解析该 token

### Requirement: token 在实际使用时按需解析
系统 SHALL 在实际需要 token 值时（如 Resolver 解析 repo 列表、clone、sync 等操作）才解析 `${ENV_VAR}` 占位符。仅对当前操作涉及的、已启用的 resource/group/repo 的 token 进行解析。

#### Scenario: Resolver 仅解析启用的 resource token
- **WHEN** 配置中有两个 resource：`github`（enabled, token `${GITHUB_TOKEN}` 已设置）和 `gitlab`（disabled, token `${GITLAB_TOKEN}` 未设置）
- **AND** 用户执行 `grepom sync` 操作
- **THEN** 系统仅解析 `github` 的 token，不解析 `gitlab` 的 token，操作正常执行

#### Scenario: 使用未设置的 token 时报错包含上下文
- **WHEN** 用户执行需要 resource `work-gl` 的操作，其 token `${GITLAB_TOKEN}` 未设置
- **THEN** 系统报错，错误信息包含 resource 名称 `work-gl` 和环境变量名 `GITLAB_TOKEN`

#### Scenario: Group 的 token 覆盖在解析时生效
- **WHEN** Group `frontend` 引用 resource `work-gl`（token `${GL_TOKEN}`），且 Group 自身设置 `token: ${FRONTEND_TOKEN}`
- **AND** `FRONTEND_TOKEN` 已设置但 `GL_TOKEN` 未设置
- **THEN** 系统使用 Group 的 `FRONTEND_TOKEN`，不解析 resource 的 `GL_TOKEN`，操作正常执行

### Requirement: add resource 命令仍立即验证 token
`grepom add resource` 命令 SHALL 在添加 resource 时立即验证 token 的环境变量是否存在，如果未设置则报错。这是用户主动操作，需要立即反馈。

#### Scenario: add resource 时环境变量未设置
- **WHEN** 用户执行 `grepom add resource --name work-gl --token '${GITLAB_TOKEN}'`，`GITLAB_TOKEN` 未设置
- **THEN** 命令报错，提示环境变量 `GITLAB_TOKEN` 未设置

#### Scenario: add resource 时环境变量已设置
- **WHEN** 用户执行 `grepom add resource --name work-gl --token '${GITLAB_TOKEN}'`，`GITLAB_TOKEN` 已设置
- **THEN** resource 成功添加到配置文件
