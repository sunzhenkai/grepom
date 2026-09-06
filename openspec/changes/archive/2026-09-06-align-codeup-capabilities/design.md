## Context

现状约束（动机见 proposal.md）：

- `provider/codeup.go` 已实现 OAPI v1 的 API 地址映射（`codeup.aliyun.com` → `https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/{orgId}`）、`x-yunxiao-token` 认证、响应头分页（`x-total`/`x-next-page`）等基础设施，但均为 `provider` 包私有。新增的 `mergerequest/codeup.go` 与 `cicd/codeup.go` 需要复用同一套规则。
- `cmd/mr.go` 的 `detectProvider` 目前只返回 `(providerName, serverURL, token)`，codeup 走知名域名分支时 token 为空、无 organization_id；cmd 层拿不到 resource 的 `organization_id`。
- `cicd.PipelineProvider` 接口的 `ListPipelinesParams`/`GetPipelineParams` 只有 `ServerURL/Token/RepoPath`，同样缺少 organization_id 通道。
- Codeup 无内置 CI；云效 Flow 流水线是组织级资源，与仓库是多对多关系（一条流水线绑定一个代码源，一个仓库可被多条流水线引用），模型与 GitLab/GitHub "仓库→流水线" 的一对多不同。

## Goals / Non-Goals

**Goals:**
- Codeup 仓库 `grepom mr` 全流程可用（创建、幂等返回已有 MR、Web URL），体验与 gitlab/github 一致
- `pipeline list` / `pipeline watch` / 顶级 `watch` 对 Codeup 仓库可用
- 复用现有注册表模式与 OAPI v1 基础设施，不改动 provider 接口签名之外的无关注册项

**Non-Goals:**
- 不实现 Codeup MR 的评审、合并、关闭等操作（只做创建与查询已有）
- 不实现 Flow 流水线的创建/触发/取消（只读查询与监控）
- 不处理 codeup.aliyun.com 以外的私有部署 Codeup（OAPI v1 地址映射不支持，保持与 `provider/codeup.go` 现状一致）
- 不引入 Codeup webhook、部署密钥等其他对齐项（平台 API 具备但本期不做）

## Decisions

### D1: 抽取 Codeup OAPI v1 公共基础设施到新包 `codeupapi`

将 `provider/codeup.go` 中的 API 地址映射、分页解析、认证 GET/POST 请求封装抽取为独立内部包（如 `internal/codeupapi` 或 `provider` 包导出），供 `provider`、`mergerequest`、`cicd` 三个包共用。

- **备选**：三个包各自复制一份 → 三份重复代码，地址映射规则漂移风险高，否决。
- **备选**：只导出 `provider` 包的方法 → `provider` 包职责膨胀且产生包间反向依赖（mergerequest → provider 仅为了 HTTP 工具），否决。

### D2: 参数结构增加 `OrganizationID` 字段

`mergerequest.CreateMergeRequestParams` 与 `cicd.ListPipelinesParams`/`GetPipelineParams` 增加 `OrganizationID string` 字段（沿用 `provider.ListReposParams` 的既有先例）。gitlab/github provider 忽略该字段，无接口破坏。

cmd 层：`detectProvider`（mr.go）与 `resolvePipelineInput`（pipeline.go）在命中 codeup resource 时把 `res.OrganizationID` 透传下去。`detectProvider` 返回值扩展（加返回 organizationID，或改为返回一个小的 struct）。

### D3: Codeup MR 创建流程：路径 → 仓库 ID → 查重 → 创建

端点以实测网关与官方文档（CreateChangeRequest / ListChangeRequests）为准，路径段为 `changeRequests`（而非早期假设的 `mergeRequests`），且挂在组织前缀下（`{apiBase}` 已含 `/organizations/{orgId}`）。

1. **路径解析**：`GET {apiBase}/repositories?page=N&perPage=100` 分页遍历，找到 `pathWithNamespace == RepoPath` 的条目取 `id`。找不到则报错。
2. **幂等查重**：`GET {apiBase}/changeRequests?projectIds={id}&state=opened&perPage=100`——ListChangeRequests 不支持按源分支服务端过滤，故客户端按 `sourceBranch` 过滤，命中则直接返回（`AlreadyExists: true`），与 gitlab provider 的 `findOpenMR` 语义对齐。
3. **创建**：`POST {apiBase}/repositories/{id}/changeRequests`，body 含 `sourceBranch`/`sourceProjectId`(int)/`targetBranch`/`targetProjectId`(int)/`title`/`description`（同库 MR 时 source/target projectId 均为仓库 ID）。
4. **草稿语义**：Codeup 无 draft 字段，`--draft` 时与 GitLab 一致在 title 前加 "Draft: " 前缀（Web URL 无法携带 draft，忽略）。
5. **响应映射**：`localId` → `Number`，`webUrl`（兜底 `detailUrl`）→ `URL`，`state`（兜底 `status`）按 `UNDER_DEV`/`UNDER_REVIEW`/`TO_BE_MERGED` → open、`MERGED` → merged、`CLOSED` → closed 映射。

- **备选**：跳过查重依赖 API 409 冲突报错 → 报错信息不友好，且与 gitlab/github 的幂等体验不一致，否决。

### D4: Flow 流水线按"代码源地址"匹配仓库

`ListPipelines` 实现：

1. `GET /oapi/v1/flow/organizations/{orgId}/pipelines` 分页列出组织流水线。
2. 对每条流水线读取代码源配置，匹配其仓库地址（URL 归一化后与 `RepoPath` 比较）。
3. 对匹配的流水线调用 `GET .../pipelines/{pipelineId}/runs` 取运行记录，合并后按开始时间倒序截断到 `Limit`。
4. `GetPipeline` 按 runId 查询单次运行详情。

状态映射：Flow 运行状态（RUNNING/SUCCESS/FAIL/CANCEL 等）映射到现有 `PipelineStatus` 五值。

- **备选**：要求用户在 config 中手工绑定 repo → pipelineId → 配置格式变更、维护成本高，否决；自动匹配覆盖绝大多数场景。
- **已知代价**：组织内流水线数量大时列表全量拉取较慢 → 接受（list/watch 是低频操作；可加 verbose 日志）。

### D5: 环境变量兜底 `GREPOM_CODEUP_TOKEN`

`detectProvider`/`resolveToken` 的 codeup 知名域名分支补 `GREPOM_CODEUP_TOKEN` 环境变量兜底，与 github/gitlab 的 `GREPOM_GITHUB_TOKEN`/`GREPOM_GITLAB_TOKEN` 模式对齐。但 organization_id 无环境变量兜底——知名域名命中但 config 无 codeup resource 时直接报错提示配置 `organization_id`（见 mr-pr-command delta spec）。

- **备选**：也提供 `GREPOM_CODEUP_ORG_ID` → organization_id 是企业级标识，放 config 更合理；且 mr 之外 sync 等流程本就强制要求 config 中有 organization_id，保持一致，否决。

## Risks / Trade-offs

- [OAPI v1 MR/Flow 接口字段与文档存在出入（Codeup 文档质量参差，grepom 开发中已遇到过分页行为与文档不符的情况）] → 实现时以真实 API 响应为准编写解析逻辑，单元测试用录制式 mock；每个接口先手工 curl 验证再编码。
- [Flow 流水线列表在大组织下分页拉取慢，list/watch 首次延迟高] → verbose 模式输出拉取进度；后续可加本地缓存（本期不做）。
- [一条仓库匹配多条流水线时输出语义与 GitLab 的"单流水线"模型有差异] → spec 中明确"合并按时间倒序"的确定行为；输出表格不显示流水线名，保持现有格式。
- [删除中仓库（deletion_scheduled）在 MR 创建路径解析时被误命中] → 路径精确匹配 `pathWithNamespace`，删除中仓库路径已带后缀不会命中，天然规避。

## Migration Plan

纯新增能力 + 行为升级，无数据迁移：

1. 抽取 `codeupapi` 公共包（`provider/codeup.go` 改为薄封装，行为不变，现有测试保底）
2. 实现 `mergerequest/codeup.go` + 改造 `cmd/mr.go`
3. 实现 `cicd/codeup.go` + 改造 `cmd/pipeline.go` 参数透传
4. README 双语更新
5. 发布说明中注明行为变化：Codeup 仓库 `mr` 从"提示手动创建"变为"API 直接创建"

回滚策略：无持久状态，回退版本即恢复旧行为。

## Open Questions

- Flow 流水线代码源字段的确切名称与结构（`codeSource` / `sources` 等）需在实现时以真实响应确认，不影响 spec 中"按代码源地址匹配"的行为定义。
