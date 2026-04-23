## MODIFIED Requirements

### Requirement: token 支持环境变量占位符
配置文件中 resource 的 `token` 字段 SHALL 支持 `${ENV_VAR}` 占位符语法。系统在实际使用该 token 时（而非配置加载时）从环境变量读取实际值。`ResolveToken()` SHALL 在解析前自动清理 token 值首尾成对的单引号或双引号（`'...'` 或 `"..."`），以兼容 YAML 和 CLI 参数中引号包裹的场景。

#### Scenario: token 使用环境变量占位符
- **WHEN** 配置文件中 resource 的 token 值为 `${GITLAB_TOKEN}`
- **THEN** 系统在加载配置时保留原始占位符字符串，在实际使用该 resource 时从环境变量 `GITLAB_TOKEN` 读取实际 token 值

#### Scenario: 环境变量未设置时按需报错
- **WHEN** 配置文件中 resource 的 token 值为 `${GITLAB_TOKEN}`，但环境变量 `GITLAB_TOKEN` 未设置
- **THEN** 系统 SHALL 在实际使用该 token 时报错，配置加载阶段不报错

#### Scenario: token 为明文值时直接使用
- **WHEN** 配置文件中 resource 的 token 值为 `glpat-xxxxxxxxxxxx`（非占位符格式）
- **THEN** 系统直接使用该值，不做环境变量解析

#### Scenario: token 被单引号包裹
- **WHEN** token 值为 `'${GITLAB_TOKEN}'`（YAML 单引号包裹或 CLI 传入）
- **THEN** `ResolveToken()` 自动去除首尾单引号后解析为环境变量 `GITLAB_TOKEN` 的值

#### Scenario: token 被双引号包裹
- **WHEN** token 值为 `"${GITLAB_TOKEN}"`（双引号包裹）
- **THEN** `ResolveToken()` 自动去除首尾双引号后解析为环境变量 `GITLAB_TOKEN` 的值

#### Scenario: token 引号不对称时不处理
- **WHEN** token 值为 `'${GITLAB_TOKEN}"`（首尾引号不同）
- **THEN** `ResolveToken()` 不去除引号，按原值处理

#### Scenario: 明文 token 被引号包裹
- **WHEN** token 值为 `"glpat-xxxxx"`（双引号包裹的明文 token）
- **THEN** `ResolveToken()` 去除双引号后返回 `glpat-xxxxx`

### Requirement: token 占位符在配置回写时保持不变
当系统回写配置文件时 SHALL 保留 token 字段的原始占位符字符串，不将展开后的明文值写入磁盘。引号清理仅在运行时 `ResolveToken()` 内部进行，不影响存储在 `rawTokens` map 中的原始值。

#### Scenario: sync 更新配置后 token 占位符不变
- **WHEN** 配置文件中 token 为 `${GITLAB_TOKEN}`，sync 执行后回写配置文件
- **THEN** 配置文件中 token 仍为 `${GITLAB_TOKEN}`，而非环境变量的实际值

#### Scenario: add resource 时保留 token 占位符
- **WHEN** 用户运行 `grepom add resource --name work-gl --token '${GITLAB_TOKEN}' --provider gitlab ...`
- **THEN** 配置文件中 token 字段写入 `${GITLAB_TOKEN}`
