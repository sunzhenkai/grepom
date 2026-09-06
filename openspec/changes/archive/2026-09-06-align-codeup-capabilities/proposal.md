## Why

grepom 已支持 Codeup 仓库发现与克隆，但相比 GitLab/GitHub 存在两块能力缺口：

1. `grepom mr` 对 Codeup 仓库只能输出"暂不支持通过 API 创建 Merge Request"的提示并回退到浏览器手动创建（`cmd/mr.go` 中的特判分支）。实际上 Codeup OAPI v1 已提供完整的合并请求接口族（CreateMergeRequest / ListMergeRequests 等），该限制是 grepom 的实现缺口而非平台缺口。
2. `grepom pipeline list/watch` 的 `cicd` 注册表只有 gitlab/github 两个 provider，Codeup 仓库查询流水线直接报 "unsupported cicd provider: codeup"。Codeup 本身无内置 CI，但同一组织下的云效 Flow 流水线可通过 OAPI v1 Flow 接口查询。

补齐这两块后，Codeup 在日常开发闭环（推送 → 建 MR → 看流水线）上与 GitLab/GitHub 体验对齐。

## What Changes

- **新增 `mergerequest` 包的 Codeup provider**：通过 OAPI v1 创建合并请求，支持标题/正文/草稿语义、已有打开中 MR 时返回已有 MR（与 gitlab/github 行为一致的幂等语义）、浏览器创建页 Web URL 构建。
- **修改 `grepom mr` 命令**：移除 Codeup "暂不支持" 的特判分支，Codeup 仓库走与 gitlab/github 相同的创建流程；provider 识别时额外解析 Codeup resource 的 `organization_id`。
- **新增 `cicd` 包的 Codeup provider**：基于云效 Flow OAPI v1 查询流水线运行记录，实现 `ListPipelines` / `GetPipeline`，使 `pipeline list`、`pipeline watch`、`watch` 命令对 Codeup 仓库可用。Flow 流水线是组织级资源，需要"按仓库地址匹配流水线"的映射策略（见 design.md）。
- **同步 README（中文 + English）**：功能特性与 provider 支持矩阵更新为 Codeup 支持 MR 创建与 pipeline 查询。

## Capabilities

### New Capabilities

（无新能力路径；两块改动都落在既有能力的扩展上。）

### Modified Capabilities

- `merge-request-provider`：新增 Codeup 合并请求创建与 Web URL 构建需求；修改"获取不存在的 provider"场景——`mergerequest.Get("codeup")` 不再返回 unsupported 错误。
- `mr-pr-command`：移除"Codeup 不支持提示"需求，Codeup 仓库执行 `grepom mr` 走标准 MR 创建流程；Token 获取策略补充 Codeup resource 需要 `organization_id` 的要求。
- `pipeline-list`：新增对 Codeup provider（云效 Flow）的支持需求，包括按仓库匹配 Flow 流水线的行为与无匹配流水线时的错误提示。
- `pipeline-watch`：补充 watch 流程对 Codeup provider 可用的要求（复用 pipeline-list 的解析与轮询语义）。

## Impact

- **代码**：
  - 新增 `mergerequest/codeup.go`、`cicd/codeup.go`（含单元测试）
  - 修改 `cmd/mr.go`（移除 Codeup 特判、识别逻辑透传 organization_id）、`cmd/pipeline.go`（Codeup resource 解析透传 organization_id）
  - `mergerequest.MergeRequestProvider` / `cicd.PipelineProvider` 接口本身不变，仅新增实现
- **API 依赖**：Codeup OAPI v1（`https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/{orgId}` 与 `.../oapi/v1/flow/organizations/{orgId}`），认证沿用 `x-yunxiao-token`，无新增第三方依赖
- **配置**：沿用现有 codeup resource（`provider: codeup` + `organization_id`），无配置格式变更、无 BREAKING
- **文档**：README.md / README_en.md 功能列表与多 provider 说明
- **行为变化（用户可见）**：Codeup 仓库 `grepom mr` 从"提示手动创建"变为"直接 API 创建"；`pipeline list/watch` 从报错变为可用
