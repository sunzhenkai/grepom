## REMOVED Requirements

### Requirement: Codeup 不支持提示
**Reason**: Codeup OAPI v1 已提供完整的合并请求接口，`mergerequest` 包新增 codeup provider 后，Codeup 仓库可直接通过 API 创建 MR，不再局限于浏览器手动创建。
**Migration**: 无用户迁移成本。Codeup 仓库执行 `grepom mr` 将走与 GitLab/GitHub 相同的标准创建流程；`--web` 模式仍可通过浏览器创建页 URL 打开。

## MODIFIED Requirements

### Requirement: Token 获取策略
系统 SHALL 按以下优先级获取 API token：
1. config 中匹配的 resource token（支持 `${ENV_VAR}` 占位符）
2. 环境变量 `GREPOM_GITHUB_TOKEN`（GitHub）、`GREPOM_GITLAB_TOKEN`（GitLab）或 `GREPOM_CODEUP_TOKEN`（Codeup）
3. 无法获取 → 报错

当 provider 为 codeup 时，系统 SHALL 同时获取匹配 resource 的 `organization_id`；若无法获得 organization_id（如仅通过知名域名识别而无 config resource），SHALL 报错并提示在 config 中为 codeup resource 配置 `organization_id`。

#### Scenario: 从 config 获取 token
- **WHEN** remote URL 匹配 config 中的 resource，且该 resource 有 token 配置
- **THEN** 系统 SHALL 使用该 resource 的 token

#### Scenario: 从环境变量获取 token
- **WHEN** remote URL 未匹配 config 中的 resource，但设置了 `GREPOM_GITHUB_TOKEN` 环境变量
- **THEN** 系统 SHALL 使用环境变量中的 token

#### Scenario: Codeup 从 config 获取 organization_id
- **WHEN** remote URL 匹配 config 中 provider 为 codeup 的 resource
- **THEN** 系统 SHALL 使用该 resource 的 token 和 `organization_id`

#### Scenario: Codeup 缺少 organization_id
- **WHEN** 识别出 provider 为 codeup，但无法从 config resource 获得 `organization_id`
- **THEN** 系统 SHALL 报错并提示在 config 的 codeup resource 中配置 `organization_id`

#### Scenario: 无法获取 token
- **WHEN** remote URL 未匹配 config，也没有对应的环境变量
- **THEN** 系统 SHALL 报错并提示设置环境变量或在 config 中添加 resource
