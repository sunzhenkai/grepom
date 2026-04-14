## 1. 配置层扩展

- [x] 1.1 在 `config/config.go` 的 `Resource` 结构体新增 `OrganizationID string` 字段（YAML tag: `organization_id,omitempty`）
- [x] 1.2 在 `config/config.go` 的 `validate()` 方法中，将 `codeup` 加入 `validProviders` map
- [x] 1.3 在 `validate()` 中新增校验：当 provider 为 `codeup` 时，`organization_id` 为必填字段，缺少时返回明确错误信息
- [x] 1.4 在 `provider/provider.go` 的 `ListGroupsParams` 结构体新增 `OrganizationID string` 字段

## 2. Codeup Provider 核心 API 实现

- [x] 2.1 创建 `provider/codeup.go`，实现 `init()` 注册 `codeup` provider
- [x] 2.2 定义 Codeup 统一响应结构体 `codeupResponse[T]`（包含 requestId, success, errorCode, errorMessage, total, result 字段）
- [x] 2.3 定义 Codeup 仓库结构体 `codeupRepo`（映射 API 返回的 Result 对象字段：Id, Name, Path, PathWithNamespace 等）
- [x] 2.4 定义 Codeup 代码组结构体 `codeupGroup`（映射 ListRepositoryGroups 返回的字段：id, path, name, pathWithNamespace 等）
- [x] 2.5 实现 `codeupAPIURL()` 函数，将 `codeup.aliyun.com` 映射为 `https://devops.aliyun.com`
- [x] 2.6 实现通用 HTTP GET 方法 `getCodeup()`，处理 accessToken query parameter 传递、响应解析、错误检测（success==false 时返回错误）

## 3. ListRepos 实现

- [x] 3.1 实现 `ListRepos()` 方法：遍历 `params.Groups`，对每个 group 调用 `GET /repository/list` 全量拉取仓库
- [x] 3.2 实现分页逻辑：使用 page/perPage 参数，通过响应 `total` 计算总页数，循环拉取所有页面
- [x] 3.3 实现客户端路径前缀过滤：当 `group.Path` 非空时，仅保留 `pathWithNamespace` 以 `group.Path + "/"` 开头的仓库
- [x] 3.4 实现仓库映射：将 codeupRepo 映射为 `provider.Repo`，推导 HTTPS 和 SSH clone URL

## 4. ListGroups 实现

- [x] 4.1 实现 `findGroupByPath()` 内部方法：调用 `GET /api/4/groups/find_by_path`，通过代码组路径获取 namespaceId
- [x] 4.2 实现 `ListGroups()` 方法：先通过 `findGroupByPath` 获取根 namespaceId，再调用 `GET /repository/groups/get/all` 列出一级代码组
- [x] 4.3 实现 ListGroups 降级逻辑：当获取根 namespaceId 失败时返回空列表并在 verbose 模式输出警告

## 5. 命令层集成

- [x] 5.1 修改 `cmd/sync.go`：在 provider 分支判断中新增 `codeup` case，使 Codeup 使用 Groups 查询模式（与 GitLab 一致），并传递 `OrganizationID` 到 `ListReposParams`
- [x] 5.2 修改 `cmd/list.go`：在 `runListRemoteRepos()` 和 `runListRemoteGroups()` 中新增 `codeup` 分支处理，传递 `OrganizationID`

## 6. 测试

- [x] 6.1 为 `codeupAPIURL()` 编写单元测试
- [x] 6.2 为 `ListRepos` 编写测试：使用 httptest 模拟 API 响应，验证全量拉取 + 路径前缀过滤逻辑
- [x] 6.3 为 `ListRepos` 编写测试：验证多页分页场景
- [x] 6.4 为 `ListRepos` 编写测试：验证认证失败和 API 错误场景
- [x] 6.5 为 `ListGroups` 编写测试：使用 httptest 模拟代码组查询
- [x] 6.6 验证 `config.Load()` 能正确加载包含 codeup resource 的配置文件
- [x] 6.7 运行全量测试确保不破坏现有功能：`go test ./...`
