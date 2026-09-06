## ADDED Requirements

### Requirement: pipeline list 支持 Codeup Provider（云效 Flow）
系统 SHALL 通过云效 Flow OAPI v1 获取 Codeup 仓库关联的流水线运行记录。API 基础地址 SHALL 映射为 `https://openapi-rdc.aliyuncs.com/oapi/v1/flow/organizations/{orgId}`，认证 SHALL 使用 `x-yunxiao-token` 请求头。由于 Flow 流水线是组织级资源，系统 SHALL 先列出组织流水线，按流水线代码源地址与目标仓库匹配，再查询该流水线的运行记录。

#### Scenario: Codeup API 调用
- **WHEN** repo 的 Provider 为 "codeup"，且 resource 配置了 organization_id
- **THEN** 系统 SHALL 列出该组织的 Flow 流水线，匹配代码源地址指向目标仓库的流水线，查询其最近运行记录并返回

#### Scenario: 无关联流水线
- **WHEN** 组织内没有任何 Flow 流水线的代码源指向目标仓库
- **THEN** 系统 SHALL 输出 "No pipelines found for <repo-name>."

#### Scenario: 缺少 organization_id
- **WHEN** repo 的 Provider 为 "codeup"，但 resource 未配置 organization_id
- **THEN** 系统 SHALL 报错并提示在 config 的 codeup resource 中配置 `organization_id`

#### Scenario: 多条流水线匹配同一仓库
- **WHEN** 组织内有多条 Flow 流水线的代码源指向目标仓库
- **THEN** 系统 SHALL 合并各流水线的运行记录，按开始时间倒序返回前 N 条
