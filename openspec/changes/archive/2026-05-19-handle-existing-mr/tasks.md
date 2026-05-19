## 1. 数据模型

- [x] 1.1 在 `MergeRequest` 结构体中新增 `AlreadyExists bool` 字段（mergerequest/mergerequest.go）

## 2. GitLab Provider 幂等性处理

- [x] 2.1 在 `gitlab.go` 中实现 `findOpenMR` 方法：通过 `GET /api/v4/projects/:id/merge_requests?source_branch=xxx&state=opened` 搜索已有 MR，返回 `(*MergeRequest, error)`
- [x] 2.2 修改 `GitLabMRProvider.CreateMergeRequest`：在收到 409 状态码时调用 `findOpenMR`，找到则设置 `AlreadyExists=true` 返回；搜索失败则回退到原始 409 错误
- [x] 2.3 在 `gitlab_test.go` 中添加测试：模拟 409 响应 + 搜索 API 返回已有 MR，验证返回正确的 `MergeRequest`（含 `AlreadyExists=true`）
- [x] 2.4 在 `gitlab_test.go` 中添加测试：模拟 409 响应 + 搜索 API 无结果，验证返回原始 409 错误
- [x] 2.5 在 `gitlab_test.go` 中添加测试：模拟 409 响应 + 搜索 API 也失败，验证返回原始 409 错误

## 3. GitHub Provider 幂等性处理

- [x] 3.1 在 `github.go` 中实现 `findOpenPR` 方法：通过 `GET /repos/:owner/:repo/pulls?head=owner:branch&state=open` 搜索已有 PR，从 `RepoPath` 中提取 owner 拼接 head 参数，返回 `(*MergeRequest, error)`
- [x] 3.2 修改 `GitHubMRProvider.CreateMergeRequest`：在收到 422 状态码时调用 `findOpenPR`，找到则设置 `AlreadyExists=true` 返回；搜索失败则回退到原始 422 错误
- [x] 3.3 在 `github_test.go` 中添加测试：模拟 422 响应 + 搜索 API 返回已有 PR，验证返回正确的 `MergeRequest`（含 `AlreadyExists=true`）
- [x] 3.4 在 `github_test.go` 中添加测试：模拟 422 响应 + 搜索 API 无结果，验证返回原始 422 错误
- [x] 3.5 在 `github_test.go` 中添加测试：模拟 422 响应 + 搜索 API 也失败，验证返回原始 422 错误

## 4. 命令行输出

- [x] 4.1 修改 `cmd/mr.go` 的 `runMR` 函数：根据 `mr.AlreadyExists` 区分输出格式，新建显示 `✅ MR #N: <title>`，已存在显示 `ℹ️ MR #N already exists: <title>` 及 URL

## 5. 文档同步

- [x] 5.1 更新 README.md 和 README_en.md，补充 `mr` 命令支持幂等创建的说明（已有 MR 时自动返回地址而非报错）