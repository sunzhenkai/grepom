## ADDED Requirements

### Requirement: MergeRequestProvider 接口
系统 SHALL 在 `mergerequest/` 包中定义 `MergeRequestProvider` 接口，包含 `CreateMergeRequest` 和 `BuildWebURL` 两个方法。包 SHALL 使用独立注册表（与 `cicd/` 包模式一致）。

#### Scenario: 接口注册
- **WHEN** 系统启动时
- **THEN** `github` 和 `gitlab` provider SHALL 通过 `init()` 注册到注册表中

#### Scenario: 获取不存在的 provider
- **WHEN** 调用 `mergerequest.Get("codeup")`
- **THEN** SHALL 返回错误 "unsupported merge request provider: codeup"

### Requirement: GitHub PR 创建
系统 SHALL 通过 GitHub REST API 的 `POST /repos/{owner}/{repo}/pulls` 端点创建 Pull Request。

#### Scenario: 正常创建 GitHub PR
- **WHEN** 调用 `CreateMergeRequest` 且 provider 为 github
- **THEN** SHALL 向 `https://api.github.com/repos/{owner}/{repo}/pulls`（或 GHE 对应地址）发送 POST 请求
- **THEN** 请求体 SHALL 包含 `title`、`body`、`head`、`base`、`draft` 字段
- **THEN** 请求 SHALL 使用 `Authorization: Bearer {token}` 认证

#### Scenario: GitHub PR 创建成功响应
- **WHEN** API 返回 201 Created
- **THEN** SHALL 解析响应并返回 `MergeRequest` 结构（含 Number、Title、URL 等）

#### Scenario: GitHub PR 创建失败
- **WHEN** API 返回非 201 状态码（如 422 分支无差异）
- **THEN** SHALL 返回包含状态码和错误信息的 error

### Requirement: GitHub PR Web URL
系统 SHALL 构建 GitHub PR 浏览器创建页面 URL：`https://github.com/{owner}/{repo}/compare/{base}...{head}?expand=1`。`--draft` 时追加 `&draft=1`。

#### Scenario: GitHub Web URL 构建
- **WHEN** 调用 `BuildWebURL`，repoPath="myorg/repo"，source="feature-x"，target="main"
- **THEN** SHALL 返回 `https://github.com/myorg/repo/compare/main...feature-x?expand=1`

#### Scenario: GitHub Web URL with draft
- **WHEN** draft=true
- **THEN** SHALL 返回 URL 追加 `&draft=1`

### Requirement: GitLab MR 创建
系统 SHALL 通过 GitLab REST API 的 `POST /projects/{id}/merge_requests` 端点创建 Merge Request。项目路径 SHALL 使用 URL 编码。

#### Scenario: 正常创建 GitLab MR
- **WHEN** 调用 `CreateMergeRequest` 且 provider 为 gitlab
- **THEN** SHALL 向 `{server}/api/v4/projects/{url-encoded-path}/merge_requests` 发送 POST 请求
- **THEN** 请求体 SHALL 包含 `source_branch`、`target_branch`、`title`、`description` 字段
- **THEN** `draft` 为 true 时 SHALL 将 title 前缀加上 "Draft: "
- **THEN** 请求 SHALL 使用 `PRIVATE-TOKEN: {token}` 认证

#### Scenario: GitLab MR 创建成功响应
- **WHEN** API 返回 201 Created
- **THEN** SHALL 解析响应并返回 `MergeRequest` 结构（含 IID、Title、WebURL 等）

#### Scenario: GitLab MR 创建失败
- **WHEN** API 返回非 201 状态码
- **THEN** SHALL 返回包含状态码和错误信息的 error

### Requirement: GitLab MR Web URL
系统 SHALL 构建 GitLab MR 浏览器创建页面 URL：`{server}/{path}/-/merge_requests/new?merge_request[source_branch]={from}&merge_request[target_branch]={to}`。

#### Scenario: GitLab Web URL 构建
- **WHEN** 调用 `BuildWebURL`，serverURL="https://gitlab.com"，repoPath="myorg/repo"，source="feature-x"，target="main"
- **THEN** SHALL 返回 `https://gitlab.com/myorg/repo/-/merge_requests/new?merge_request[source_branch]=feature-x&merge_request[target_branch]=main`

### Requirement: MergeRequest 数据结构
系统 SHALL 定义 `MergeRequest` 结构体，包含以下字段：ID（int）、Number（int，GitHub 为 number，GitLab 为 iid）、Title（string）、Description（string）、URL（string）、State（string）、SourceBranch（string）、TargetBranch（string）、Draft（bool）。

#### Scenario: GitHub 响应映射
- **WHEN** GitHub API 返回 PR 数据
- **THEN** `Number` SHALL 映射自 `number` 字段，`URL` 映射自 `html_url`

#### Scenario: GitLab 响应映射
- **WHEN** GitLab API 返回 MR 数据
- **THEN** `Number` SHALL 映射自 `iid` 字段，`URL` 映射自 `web_url`

### Requirement: GitHub Enterprise Server 兼容
系统 SHALL 正确处理 GitHub Enterprise Server 的 API URL 格式。对于非 `github.com` 域名，SHALL 使用 `{server}/api/v3/` 作为 API 前缀；对于 `github.com`，SHALL 使用 `https://api.github.com`。

#### Scenario: GHE API URL
- **WHEN** serverURL 为 `https://ghe.example.com`
- **THEN** API URL SHALL 为 `https://ghe.example.com/api/v3`

#### Scenario: GitHub.com API URL
- **WHEN** serverURL 为 `https://github.com`
- **THEN** API URL SHALL 为 `https://api.github.com`
