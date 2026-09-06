## MODIFIED Requirements

### Requirement: MergeRequestProvider 接口
系统 SHALL 在 `mergerequest/` 包中定义 `MergeRequestProvider` 接口，包含 `CreateMergeRequest` 和 `BuildWebURL` 两个方法。包 SHALL 使用独立注册表（与 `cicd/` 包模式一致）。

#### Scenario: 接口注册
- **WHEN** 系统启动时
- **THEN** `github`、`gitlab` 和 `codeup` provider SHALL 通过 `init()` 注册到注册表中

#### Scenario: 获取不存在的 provider
- **WHEN** 调用 `mergerequest.Get("generic")`
- **THEN** SHALL 返回错误 "unsupported merge request provider: generic"

## ADDED Requirements

### Requirement: Codeup MR 创建
系统 SHALL 通过 Codeup OAPI v1 的创建合并请求端点创建 Merge Request。API 基础地址 SHALL 按 `provider` 包相同的规则映射（`codeup.aliyun.com` → `https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/{orgId}`），认证 SHALL 使用 `x-yunxiao-token` 请求头。创建前 SHALL 先将仓库路径（`pathWithNamespace`）解析为 Codeup 仓库 ID。

#### Scenario: 正常创建 Codeup MR
- **WHEN** 调用 `CreateMergeRequest` 且 provider 为 codeup，参数含 organizationID
- **THEN** SHALL 先通过仓库列表接口将 `RepoPath` 解析为仓库 ID
- **THEN** SHALL 向 `{apiBase}/repositories/{repoId}/changeRequests` 发送 POST 请求
- **THEN** 请求体 SHALL 包含 `sourceBranch`、`sourceProjectId`（int，同库 MR 时为仓库 ID）、`targetBranch`、`targetProjectId`（int）、`title` 字段，`description` 非空时包含
- **THEN** 请求 SHALL 使用 `x-yunxiao-token: {token}` 认证

#### Scenario: Codeup MR 创建成功响应
- **WHEN** API 返回成功状态码
- **THEN** SHALL 解析响应并返回 `MergeRequest` 结构（含 Number、Title、URL 等）

#### Scenario: Codeup MR 创建失败
- **WHEN** API 返回非成功状态码（如分支无差异）
- **THEN** SHALL 返回包含状态码和错误信息的 error

#### Scenario: 仓库路径无法解析为仓库 ID
- **WHEN** 仓库列表接口中找不到与 `RepoPath` 精确匹配 `pathWithNamespace` 的仓库
- **THEN** SHALL 返回包含该路径的错误信息

#### Scenario: 缺少 organizationID
- **WHEN** 调用 `CreateMergeRequest` 且 provider 为 codeup，但 organizationID 为空
- **THEN** SHALL 返回错误，提示 codeup 需要配置 organization_id

### Requirement: 已有打开中 Codeup MR 幂等返回
系统 SHALL 在创建 Codeup MR 前查询是否已存在相同源分支的打开中合并请求：向 `{apiBase}/changeRequests?projectIds={repoId}&state=opened&perPage=100` 发送 GET 请求（ListChangeRequests 接口不支持按源分支服务端过滤），并在客户端按 `sourceBranch` 过滤；若命中，SHALL 直接返回该 MR 并置 `AlreadyExists` 为 true，与 GitLab/GitHub provider 行为一致。

#### Scenario: 已存在打开中 MR
- **WHEN** 目标仓库已存在 sourceBranch 相同且状态为打开的合并请求
- **THEN** SHALL 返回该已有 MR，`AlreadyExists` 为 true，不再调用创建接口

#### Scenario: 不存在打开中 MR
- **WHEN** 目标仓库不存在匹配的打开中合并请求
- **THEN** SHALL 调用创建接口创建新 MR

### Requirement: Codeup MR Web URL
系统 SHALL 构建 Codeup MR 浏览器创建页面 URL：`{server}/{path}/merge_requests/new?source={from}&target={to}`。

#### Scenario: Codeup Web URL 构建
- **WHEN** 调用 `BuildWebURL`，serverURL="https://codeup.aliyun.com"，repoPath="wii/solo/grepom"，source="feature-x"，target="master"
- **THEN** SHALL 返回 `https://codeup.aliyun.com/wii/solo/grepom/merge_requests/new?source=feature-x&target=master`

### Requirement: Codeup 响应映射
Codeup API 返回的合并请求数据 SHALL 映射为 `MergeRequest` 结构：`Number` 映射自合并请求的本地编号（`localId`），`URL` 映射自 `webUrl`（为空时兜底 `detailUrl`），`State` 映射自 `state` 字段（为空时兜底 `status`），状态映射规则为 `UNDER_DEV`/`UNDER_REVIEW`/`TO_BE_MERGED` → `open`，`MERGED` → `merged`，`CLOSED` → `closed`。

#### Scenario: Codeup 响应映射
- **WHEN** Codeup API 返回合并请求数据
- **THEN** `Number` SHALL 映射自 `localId` 字段，`URL` 映射自 `webUrl` 字段
