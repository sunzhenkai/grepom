### Requirement: Codeup provider 注册与初始化
系统 SHALL 提供 `codeup` provider，在 `init()` 函数中通过 `provider.Register("codeup", ...)` 注册。Codeup provider SHALL 实现 `Provider` 接口的 `ListRepos` 和 `ListGroups` 方法。

#### Scenario: codeup provider 注册
- **WHEN** 程序启动时
- **THEN** `codeup` provider 已注册到 provider 注册表，可通过 `provider.Get("codeup")` 获取实例

#### Scenario: codeup provider 实现 Provider 接口
- **WHEN** 通过 `provider.Get("codeup")` 获取 provider 实例
- **THEN** 该实例实现了 `ListRepos(ctx, params)` 和 `ListGroups(ctx, params)` 方法

### Requirement: Codeup API 基础地址映射
Codeup provider SHALL 将用户配置的 `resource.url`（如 `codeup.aliyun.com`）映射为 API 基础地址 `https://devops.aliyun.com`，用于所有 API 调用。用户配置的 `resource.url` 仅用于 clone URL 推导。

#### Scenario: API 地址映射
- **WHEN** resource.url 为 `codeup.aliyun.com`
- **THEN** Codeup provider 使用 `https://devops.aliyun.com` 作为 API 基础地址

#### Scenario: clone URL 使用原始 url
- **WHEN** resource.url 为 `codeup.aliyun.com`，仓库 pathWithNamespace 为 `wii/solo/grepom`
- **THEN** HTTPS clone URL 为 `https://codeup.aliyun.com/wii/solo/grepom.git`，SSH clone URL 为 `git@codeup.aliyun.com:wii/solo/grepom.git`

### Requirement: Codeup ListRepos 实现
Codeup provider 的 `ListRepos` 方法 SHALL 通过 `GET /repository/list` API 全量拉取 organizationId 对应组织下的所有代码库，然后按 `group.path` 作为 `pathWithNamespace` 的前缀进行客户端过滤。

分页 SHALL 使用 `page`（从 1 开始）和 `perPage`（最大 100）参数，通过响应中的 `total` 字段计算总页数，循环拉取所有页面。

认证 SHALL 通过 URL query parameter `accessToken` 传递 token。

每个返回的仓库 SHALL 映射为 `provider.Repo`，其中：
- `Name` = `result.Name`
- `Path` = `result.PathWithNamespace`
- `CloneURL` = 通过 `resource.url` + `PathWithNamespace` 推导的 HTTPS URL
- `SSHURL` = 通过 `resource.url` + `PathWithNamespace` 推导的 SSH URL
- `Provider` = `"codeup"`

#### Scenario: 列出代码组下的所有仓库
- **WHEN** group.path 为 `wii/solo`，organizationId 为 `60de7a6852743a5162b5f957`
- **THEN** provider 调用 `GET /repository/list?organizationId=60de7a6852743a5162b5f957&accessToken=xxx&page=1&perPage=100`，获取全量仓库后过滤 `pathWithNamespace` 以 `wii/solo/` 开头的仓库

#### Scenario: 多页分页拉取
- **WHEN** 组织有 250 个仓库，perPage=100
- **THEN** provider 依次请求 page=1（返回100条，total=250）、page=2（返回100条）、page=3（返回50条），合并所有结果

#### Scenario: 无匹配仓库
- **WHEN** group.path 为 `nonexistent/path`，全量仓库中没有任何仓库的 pathWithNamespace 以该路径开头
- **THEN** provider 返回空列表，不报错

#### Scenario: group.path 为空时不进行过滤
- **WHEN** group.path 为空字符串
- **THEN** provider 返回全量仓库，不进行前缀过滤

#### Scenario: API 认证失败
- **WHEN** accessToken 无效或过期
- **THEN** provider 返回包含 "codeup: authentication failed" 的错误

#### Scenario: API 返回错误
- **WHEN** API 响应 `success: false`
- **THEN** provider 返回包含 errorCode 和 errorMessage 的错误信息

### Requirement: Codeup 统一响应结构
所有 Codeup API 响应 SHALL 通过统一的响应结构解析，包含 `requestId`、`success`、`errorCode`、`errorMessage`、`total`、`result` 字段。当 `success` 为 `false` 时 SHALL 返回错误。

#### Scenario: 成功响应解析
- **WHEN** API 返回 `{ success: true, total: 5, result: [...] }`
- **THEN** provider 正确解析 result 数组和 total 值

#### Scenario: 失败响应处理
- **WHEN** API 返回 `{ success: false, errorCode: "SYSTEM_UNKNOWN_ERROR", errorMessage: "some error" }`
- **THEN** provider 返回错误，包含 errorCode 和 errorMessage 信息

### Requirement: Codeup ListGroups 实现
Codeup provider 的 `ListGroups` 方法 SHALL 通过 `identityGetGroupByPath` API 获取企业根 namespaceId，然后通过 `ListRepositoryGroups` API 列出一级代码组。

如果获取根 namespaceId 失败，SHALL 返回空列表并在 verbose 模式下输出警告信息。

#### Scenario: 成功列出代码组
- **WHEN** 调用 ListGroups，organizationId 为 `60de7a6852743a5162b5f957`
- **THEN** provider 先通过 `identityGetGroupByPath` 获取根 namespaceId，再调用 `ListRepositoryGroups` 获取一级代码组列表，返回 `RemoteGroup` 数组

#### Scenario: 获取根 namespaceId 失败
- **WHEN** `identityGetGroupByPath` API 调用失败或返回异常
- **THEN** provider 返回空列表，在 verbose 模式下输出警告

#### Scenario: 代码组映射为 RemoteGroup
- **WHEN** `ListRepositoryGroups` 返回代码组 `{ path: "my-group", pathWithNamespace: "org/my-group" }`
- **THEN** 映射为 `RemoteGroup{ Name: "my-group", Path: "org/my-group", Provider: "codeup" }`
