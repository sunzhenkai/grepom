## MODIFIED Requirements

### Requirement: token 在实际使用时按需解析
系统 SHALL 在实际需要 token 值时才解析 `${ENV_VAR}` 占位符。仅对当前操作涉及的、已启用的 resource/group/repo 的 token 进行解析。解析 SHALL 通过 `Resource.ResolvedToken()` 方法或 `config.ResolveToken()` 函数进行，不再直接使用 `Resource.Token` 字段作为 API 认证凭据。

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

#### Scenario: sync 命令使用 ResolvedToken 获取 token
- **WHEN** 用户运行 `grepom sync`，resource token 为 `${GITLAB_TOKEN}`，环境变量已设置
- **THEN** sync 命令通过 `res.ResolvedToken()` 获取已解析的 token 值，正常调用 provider API

#### Scenario: pipeline 命令使用 ResolvedToken 获取 token
- **WHEN** 用户运行 `grepom pipeline list web-app`，resource token 为 `${GITLAB_TOKEN}`，环境变量已设置
- **THEN** pipeline 命令通过 `res.ResolvedToken()` 获取已解析的 token 值，正常调用 provider API
