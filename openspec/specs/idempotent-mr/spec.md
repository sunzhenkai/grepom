## ADDED Requirements

### Requirement: GitLab MR 创建的幂等性

当 GitLab API 返回 409 Conflict（表示同源分支已有打开的 MR）时，系统 SHALL 通过 GitLab 搜索 API 查找该已有 MR 并返回其信息，而非报错退出。

#### Scenario: GitLab 同源分支已有打开的 MR
- **WHEN** 调用 GitLab 创建 MR API 返回 409 Conflict
- **THEN** 系统 SHALL 调用 `GET /api/v4/projects/:id/merge_requests?source_branch=<branch>&state=opened` 搜索已有 MR
- **THEN** 若搜索成功且找到 MR，系统 SHALL 返回该 MR 对象，且 `AlreadyExists` 字段为 `true`
- **THEN** 若搜索成功但未找到 MR，系统 SHALL 返回原始 409 错误

#### Scenario: GitLab 搜索 API 失败时的回退
- **WHEN** 调用 GitLab 创建 MR API 返回 409 Conflict，且后续搜索 API 调用失败
- **THEN** 系统 SHALL 返回原始的 409 错误信息，而非搜索失败错误

### Requirement: GitHub PR 创建的幂等性

当 GitHub API 返回 422 Unprocessable Entity（表示同源分支已有打开的 PR）时，系统 SHALL 通过 GitHub 搜索 API 查找该已有 PR 并返回其信息，而非报错退出。

#### Scenario: GitHub 同源分支已有打开的 PR
- **WHEN** 调用 GitHub 创建 PR API 返回 422 Unprocessable Entity
- **THEN** 系统 SHALL 调用 `GET /repos/:owner/:repo/pulls?head=<owner>:<branch>&state=open` 搜索已有 PR
- **THEN** 若搜索成功且找到 PR，系统 SHALL 返回该 PR 对象，且 `AlreadyExists` 字段为 `true`
- **THEN** 若搜索成功但未找到 PR，系统 SHALL 返回原始 422 错误

#### Scenario: GitHub 搜索 API 失败时的回退
- **WHEN** 调用 GitHub 创建 PR API 返回 422 Unprocessable Entity，且后续搜索 API 调用失败
- **THEN** 系统 SHALL 返回原始的 422 错误信息，而非搜索失败错误

### Requirement: MergeRequest 结构体新增 AlreadyExists 字段

`MergeRequest` 结构体 SHALL 新增 `AlreadyExists bool` 字段，用于标识返回的 MR/PR 是新建的还是已存在的。

#### Scenario: 新创建的 MR
- **WHEN** MR/PR 创建成功（GitLab 201 / GitHub 201）
- **THEN** 返回的 `MergeRequest` 对象的 `AlreadyExists` 字段 SHALL 为 `false`

#### Scenario: 已存在的 MR
- **WHEN** 通过搜索 API 找到已有的 MR/PR
- **THEN** 返回的 `MergeRequest` 对象的 `AlreadyExists` 字段 SHALL 为 `true`

### Requirement: MR/PR 已存在时的输出提示

`cmd/mr.go` SHALL 根据 `AlreadyExists` 字段输出不同的提示信息，让用户明确知道结果是新建还是已存在。

#### Scenario: 新创建的 MR/PR 输出
- **WHEN** `CreateMergeRequest` 返回 `AlreadyExists` 为 `false`
- **THEN** 系统 SHALL 输出 `✅ MR #N: <title>` 格式的信息及 URL

#### Scenario: 已存在的 MR/PR 输出
- **WHEN** `CreateMergeRequest` 返回 `AlreadyExists` 为 `true`
- **THEN** 系统 SHALL 输出 `ℹ️ MR #N already exists: <title>` 格式的信息及 URL，明确标识为已存在