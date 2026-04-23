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

### Requirement: token 占位符在配置回写时保持不变
当系统回写配置文件时 SHALL 保留 token 字段的原始占位符字符串，不将展开后的明文值写入磁盘。

#### Scenario: sync 更新配置后 token 占位符不变
- **WHEN** 配置文件中 token 为 `${GITLAB_TOKEN}`，sync 执行后回写配置文件
- **THEN** 配置文件中 token 仍为 `${GITLAB_TOKEN}`，而非环境变量的实际值

#### Scenario: add resource 时保留 token 占位符
- **WHEN** 用户运行 `grepom add resource --name work-gl --token '${GITLAB_TOKEN}' --provider gitlab ...`
- **THEN** 配置文件中 token 字段写入 `${GITLAB_TOKEN}`

### Requirement: 仅 token 字段支持环境变量占位符
系统 SHALL 仅对 Resource.Token 字段进行环境变量占位符解析，不对配置文件其他字段做全局环境变量展开。

#### Scenario: base 路径中的波浪线仍正常工作
- **WHEN** 配置文件中 base 为 `~/projects`
- **THEN** 系统正常展开波浪线为用户 home 目录

#### Scenario: 其他字段不做环境变量展开
- **WHEN** 配置文件中 URL 字段包含 `$` 字符
- **THEN** 系统不对 URL 做环境变量展开
