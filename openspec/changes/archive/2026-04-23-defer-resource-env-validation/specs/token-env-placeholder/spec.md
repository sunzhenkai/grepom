## MODIFIED Requirements

### Requirement: token 支持环境变量占位符
配置文件中 resource 的 `token` 字段 SHALL 支持 `${ENV_VAR}` 占位符语法。系统在实际使用该 token 时（而非配置加载时）从环境变量读取实际值。

#### Scenario: token 使用环境变量占位符
- **WHEN** 配置文件中 resource 的 token 值为 `${GITLAB_TOKEN}`
- **THEN** 系统在加载配置时保留原始占位符字符串，在实际使用该 resource 时从环境变量 `GITLAB_TOKEN` 读取实际 token 值

#### Scenario: 环境变量未设置时按需报错
- **WHEN** 配置文件中 resource 的 token 值为 `${GITLAB_TOKEN}`，但环境变量 `GITLAB_TOKEN` 未设置
- **THEN** 系统 SHALL 在实际使用该 token 时报错，配置加载阶段不报错

#### Scenario: token 为明文值时直接使用
- **WHEN** 配置文件中 resource 的 token 值为 `glpat-xxxxxxxxxxxx`（非占位符格式）
- **THEN** 系统直接使用该值，不做环境变量解析
