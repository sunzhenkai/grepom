## 1. 抽取 Codeup OAPI v1 公共包

- [x] 1.1 新建公共包（`internal/codeupapi` 或 `provider` 包导出）：迁移 `provider/codeup.go` 中的 API 地址映射（`codeupAPIBaseURL`）、clone host 提取、分页头解析（`parsePagination`）、带认证的 GET 请求（含 401 识别与非 200 错误包装）
- [x] 1.2 在公共包中补充带认证的 POST 请求方法（JSON body、`x-yunxiao-token` 头、错误处理与 GET 一致）
- [x] 1.3 `provider/codeup.go` 改为调用公共包的薄封装，保持 `ListRepos`/`ListGroups` 行为不变，现有 `provider/codeup_test.go` 全部通过
- [x] 1.4 迁移公共包相关单元测试（地址映射、分页解析），新增 POST 方法测试

## 2. mergerequest Codeup provider

- [x] 2.1 新建 `mergerequest/codeup.go`：`init()` 注册 `codeup`；定义 Codeup MR 请求/响应结构（`localId`、`webUrl`、`status` 等字段）
- [x] 2.2 实现仓库路径解析：`GET .../repositories` 分页查找 `pathWithNamespace == RepoPath` 的仓库 ID，找不到返回含路径的错误
- [x] 2.3 实现幂等查重：`GET .../repositories/{id}/mergeRequests?status=opened&sourceBranch={from}`，命中返回 `AlreadyExists: true` 的 MR
- [x] 2.4 实现创建：`POST .../repositories/{id}/mergeRequests`（`sourceBranch`/`targetBranch`/`title`/`description`），`Draft` 为 true 时 title 加 "Draft: " 前缀；响应映射 `localId`→`Number`、`webUrl`→`URL`、`status`→`State`
- [x] 2.5 实现 `BuildWebURL`：`{server}/{path}/merge_requests/new?source={from}&target={to}`
- [x] 2.6 `CreateMergeRequestParams` 增加 `OrganizationID string` 字段；organizationID 为空时返回提示配置 `organization_id` 的错误
- [x] 2.7 编写 `mergerequest/codeup_test.go`：mock HTTP 服务覆盖路径解析失败、查重命中、正常创建、草稿前缀、Web URL 构建、缺 organizationID 报错

## 3. cmd/mr.go 接入 Codeup

- [x] 3.1 `detectProvider` 扩展返回值（增加 organizationID 或改为 struct），命中 codeup resource 时透传 `res.OrganizationID`；知名域名分支补 `GREPOM_CODEUP_TOKEN` 环境变量兜底
- [x] 3.2 移除 `runMR` 中 Codeup "暂不支持" 特判分支（含 `buildCodeupWebURL`，其逻辑并入 `mergerequest/codeup.go` 的 `BuildWebURL`）
- [x] 3.3 Codeup 无 organization_id 时报错并提示在 config 的 codeup resource 中配置
- [x] 3.4 更新 `cmd/mr_test.go`：移除"Codeup 提示"断言，新增 codeup 走标准创建流程、缺 organization_id 报错的用例

## 4. cicd Codeup provider（云效 Flow）

- [x] 4.1 `cicd.ListPipelinesParams`/`GetPipelineParams` 增加 `OrganizationID string` 字段
- [x] 4.2 新建 `cicd/codeup.go`：`init()` 注册 `codeup`；Flow API 基础地址映射为 `.../oapi/v1/flow/organizations/{orgId}`；定义 Flow 流水线与运行记录结构（先 curl 真实 API 确认字段名，见 design.md Open Questions）
- [x] 4.3 实现流水线匹配：分页列出组织流水线，读取代码源配置，URL 归一化后与 `RepoPath` 匹配；无匹配返回空列表（由 cmd 层输出 "No pipelines found"）
- [x] 4.4 实现 `ListPipelines`：查询匹配流水线的运行记录，合并按开始时间倒序，截断到 `Limit`；Flow 状态映射到 `PipelineStatus` 五值
- [x] 4.5 实现 `GetPipeline`：按运行 ID 查询单次运行详情
- [x] 4.6 organizationID 为空时返回提示配置 `organization_id` 的错误
- [x] 4.7 编写 `cicd/codeup_test.go`：mock 覆盖流水线匹配、多流水线合并排序、无匹配、状态映射、缺 organizationID

## 5. cmd/pipeline.go 参数透传

- [x] 5.1 `resolvePipelineInput` 命中 codeup resource 时透传 `OrganizationID`（返回值扩展或并入 WatchTarget 链路）
- [x] 5.2 验证 `pipeline list`、`pipeline watch`、顶级 `watch` 三条路径对 codeup repo 均可用（含缺 organization_id 的报错提示）

## 6. 文档与收尾

- [x] 6.1 README.md / README_en.md：功能特性中 MR/PR 创建与 pipeline 查询的支持范围补充 Codeup；provider 说明更新
- [x] 6.2 `make build` + 全量 `go test ./...` 通过
- [ ] 6.3 真实 Codeup 组织冒烟验证（可选，需有效 token）：`grepom mr` 创建 MR、`grepom pipeline list <codeup-repo`（阻塞：当前 $CODEUP_TOKEN 缺少合并请求读写与 Flow 权限，网关返回 Forbidden；路由存在性已用 Forbidden vs NotFound 差异确认，需换有权限 token 后补验）
