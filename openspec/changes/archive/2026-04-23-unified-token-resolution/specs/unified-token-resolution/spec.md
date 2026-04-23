## ADDED Requirements

### Requirement: Resource.ResolvedToken() 统一 token 解析入口
`Resource` 类型 SHALL 提供 `ResolvedToken() (string, error)` 方法，封装 token 的环境变量解析逻辑。该方法 SHALL 内部调用 `ResolveToken(r.Token)` 完成占位符解析。所有需要获取 token 实际值的代码 SHALL 通过此方法或直接调用 `ResolveToken()` 获取，不再直接读取 `Resource.Token` 字段作为 API 认证凭据。

#### Scenario: ResolvedToken 解析环境变量占位符
- **WHEN** Resource 的 Token 字段为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 已设置为 `glpat-xxx`
- **THEN** `res.ResolvedToken()` 返回 `("glpat-xxx", nil)`

#### Scenario: ResolvedToken 处理明文 token
- **WHEN** Resource 的 Token 字段为 `glpat-xxxxxxxxxxxx`（非占位符格式）
- **THEN** `res.ResolvedToken()` 返回 `("glpat-xxxxxxxxxxxx", nil)`

#### Scenario: ResolvedToken 环境变量未设置时报错
- **WHEN** Resource 的 Token 字段为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 未设置
- **THEN** `res.ResolvedToken()` 返回 `("", error)`，错误信息包含环境变量名 `GITLAB_TOKEN`

#### Scenario: ResolvedToken 处理空 token
- **WHEN** Resource 的 Token 字段为空字符串
- **THEN** `res.ResolvedToken()` 返回 `("", nil)`
