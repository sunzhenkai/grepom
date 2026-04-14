## Why

当前 grepom 支持 GitLab 和 GitHub 两个远程仓库 provider，但国内大量团队使用阿里云云效（Codeup）进行代码托管。需要新增 Codeup provider，让这些团队也能使用 grepom 进行仓库批量管理和同步。

## What Changes

- 新增 `codeup` provider 实现，通过阿里云云效 API 发现和列出代码库
- 支持三个 Codeup API：`ListRepositories`（列出代码库）、`ListRepositoryGroups`（列出代码组）、`identityGetGroupByPath`（按路径查询代码组）
- Resource 配置新增 `organization_id` 字段，用于存放云效企业标识
- Codeup 使用全量拉取 + 客户端路径前缀过滤的模式（与 GitLab 按组精确拉取不同）
- Clone URL 通过 `resource.url` + `pathWithNamespace` 推导，复用现有的 `HTTPSURL()` 和 `deriveSSHURL()` 逻辑
- 认证方式为 accessToken（个人访问令牌），通过 URL query parameter 传递
- 分页使用 page/perPage 参数，通过响应中的 `total` 字段计算总页数

## Capabilities

### New Capabilities
- `codeup-provider`: 阿里云云效（Codeup）代码托管平台的 provider 实现，支持通过 API 发现代码库和代码组

### Modified Capabilities
- `resource-management`: Resource 结构新增 `organization_id` 字段（可选），用于 Codeup 等需要企业标识的 provider；provider 验证列表新增 `codeup`
- `remote-group-list`: Codeup provider 需实现 `ListGroups` 方法，通过 `ListRepositoryGroups` API 查询代码组
- `list-remote-repos`: cmd/sync.go 和 cmd/list.go 的 provider 分支逻辑需支持 `codeup` 类型
- `sync-command`: cmd/sync.go 的 provider 分支逻辑需支持 `codeup` 类型，Codeup 使用 Groups 查询模式

## Impact

- **新增文件**: `provider/codeup.go`（~250 行）、`provider/codeup_test.go`
- **修改文件**:
  - `config/config.go` — Resource 结构体新增 `OrganizationID` 字段，`validate()` 新增 `codeup` 验证
  - `cmd/sync.go` — provider 分支判断新增 `codeup` 处理
  - `cmd/list.go` — provider 分支判断新增 `codeup` 处理
- **API 依赖**: 阿里云云效 API（`devops.aliyun.com`），使用 accessToken 认证
- **向后兼容**: `organization_id` 为可选字段，不影响现有 GitLab/GitHub/Generic 配置
