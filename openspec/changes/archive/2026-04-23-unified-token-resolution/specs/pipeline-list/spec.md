## MODIFIED Requirements

### Requirement: pipeline list Token 获取
系统 SHALL 复用现有 token 优先级链路获取 API 访问令牌。`resolvePipelineInput` 函数 SHALL 通过 `Resource.ResolvedToken()` 获取已解析的 token，不再直接读取 `Resource.Token` 字段。

#### Scenario: Token 解析
- **WHEN** 执行 pipeline list
- **THEN** 系统 SHALL 通过 `repo.Resolver.ResolveAndFilter` 查找 repo，从其 Resource 获取 ServerURL 和 Provider 类型，并通过 `res.ResolvedToken()` 获取已解析的 token

#### Scenario: pipeline token 环境变量未设置时报错
- **WHEN** 用户运行 `grepom pipeline list web-app`，resource token 为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 未设置
- **THEN** 系统通过 `res.ResolvedToken()` 获得错误，输出包含 resource 名称和环境变量名的错误信息
