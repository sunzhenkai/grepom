### Requirement: Codeup provider 注册与初始化
系统 SHALL 提供 `codeup` provider，在 `init()` 函数中通过 `provider.Register("codeup", ...)` 注册。Codeup provider SHALL 实现 `Provider` 接口的 `ListRepos` 和 `ListGroups` 方法。

#### Scenario: codeup provider 注册
- **WHEN** 程序启动时
- **THEN** `codeup` provider 已注册到 provider 注册表，可通过 `provider.Get("codeup")` 获取实例

#### Scenario: codeup provider 实现 Provider 接口
- **WHEN** 通过 `provider.Get("codeup")` 获取 provider 实例
- **THEN** 该实例实现了 `ListRepos(ctx, params)` 和 `ListGroups(ctx, params)` 方法

### Requirement: Codeup API 基础地址映射
Codeup provider SHALL 将用户配置的 `resource.url`（如 `codeup.aliyun.com`）映射为 OAPI v1 API 基础地址 `https://openapi-rdc.aliyuncs.com`，用于所有 API 调用。用户配置的 `resource.url` 仅用于 clone URL 推导。

#### Scenario: API 地址映射
- **WHEN** resource.url 为 `codeup.aliyun.com`
- **THEN** Codeup provider 使用 `https://openapi-rdc.aliyuncs.com` 作为 API 基础地址

#### Scenario: clone URL 使用原始 url
- **WHEN** resource.url 为 `codeup.aliyun.com`，仓库 pathWithNamespace 为 `wii/solo/grepom`
- **THEN** HTTPS clone URL 为 `https://codeup.aliyun.com/wii/solo/grepom.git`，SSH clone URL 为 `git@codeup.aliyun.com:wii/solo/grepom.git`

### Requirement: Codeup 认证方式
Codeup provider 的所有 API 请求 SHALL 通过 HTTP 请求头 `x-yunxiao-token` 传递个人访问令牌进行认证，不使用 URL query parameter 传递 token。

#### Scenario: 请求头认证
- **WHEN** 发起任意 Codeup API 请求
- **THEN** 请求头中包含 `x-yunxiao-token: {token}` 字段

#### Scenario: API 认证失败
- **WHEN** token 无效或过期
- **THEN** provider 返回包含 "codeup: authentication failed" 的错误

### Requirement: Codeup 响应解析与分页
Codeup provider SHALL 直接解析新版 OAPI v1 的响应体（JSON 数组，无 wrapper），分页信息从响应头读取。

响应头分页字段：
- `x-total`: 总记录数
- `x-total-pages`: 总页数
- `x-page`: 当前页码
- `x-next-page`: 下一页页码（不存在或为 0 表示无下一页）
- `x-per-page`: 每页大小

#### Scenario: 单页响应
- **WHEN** API 返回状态码 200，响应体为 JSON 数组，响应头 `x-total` 为 15，`x-per-page` 为 100
- **THEN** provider 正确解析数组内容，无需翻页

#### Scenario: 多页响应
- **WHEN** API 返回 `x-total` 为 250，`x-per-page` 为 100，`x-next-page` 为 2
- **THEN** provider 依次请求 page=1、page=2、page=3（当 `x-next-page` 不存在时停止），合并所有结果

#### Scenario: API 返回错误状态码
- **WHEN** API 返回非 200 状态码（如 403、500）
- **THEN** provider 返回包含状态码和响应体的错误信息

### Requirement: Codeup ListRepos 实现
Codeup provider 的 `ListRepos` 方法 SHALL 通过两步查询实现：

**Step 1**: 调用 `GET /oapi/v1/codeup/organizations/{orgId}/namespaces?search={groupPath}&perPage=100`，在返回结果中精确匹配 `pathWithNamespace == groupPath` 的条目以获取 `groupId`。

**Step 2**: 调用 `GET /oapi/v1/codeup/organizations/{orgId}/groups/{groupId}/repositories?page={page}&perPage=100&includeSubgroups={recursive}`，拉取该组下的代码库。

如果 Step 1 未找到精确匹配的 namespace，SHALL 回退到 `GET /oapi/v1/codeup/organizations/{orgId}/repositories?perPage=100&page={page}` 全量拉取组织代码库，然后按 `group.path` 作为 `pathWithNamespace` 的前缀进行客户端过滤，并在 verbose 模式下输出警告。

每个返回的仓库 SHALL 映射为 `provider.Repo`，其中：
- `Name` = `name`
- `Path` = `pathWithNamespace`
- `CloneURL` = `"https://" + cloneHost + "/" + pathWithNamespace + ".git"`
- `SSHURL` = `"git@" + cloneHost + ":" + pathWithNamespace + ".git"`
- `Provider` = `"codeup"`

**删除中代码库过滤**：当 `ListReposParams.IncludeDeleted` 为 `false`（默认）时，`ListRepos` SHALL 通过 `isDeletionScheduled(name, pathWithNamespace)` 检测并剔除处于"计划删除"状态的代码库（`name` 或 `pathWithNamespace` 含 `deletion_scheduled` 标记）。`IncludeDeleted` 为 `true` 时 SHALL 保留这些代码库。

#### Scenario: 按代码组精确查询仓库
- **WHEN** group.path 为 `wii/solo`，organizationId 为 `646c6887be6b046b8f87bb30`，ListNamespaces 返回匹配的 groupId 为 12345
- **THEN** provider 调用 `GET .../groups/12345/repositories?perPage=100&page=1&includeSubgroups=false`，返回该组下的仓库

#### Scenario: 递归查询子组仓库
- **WHEN** group.path 为 `wii`，recursive 为 true
- **THEN** provider 调用 `ListGroupRepositories` 时 `includeSubgroups=true`，获取该组及所有子组下的仓库

#### Scenario: group path 未匹配 namespace 时回退
- **WHEN** group.path 为 `some/path`，ListNamespaces search 无精确匹配结果
- **THEN** provider 回退到 ListRepositories 全量拉取，客户端按 `some/path/` 前缀过滤，verbose 模式输出警告

#### Scenario: 无匹配仓库
- **WHEN** group.path 对应的组下没有任何仓库
- **THEN** provider 返回空列表，不报错

#### Scenario: group.path 为空时不进行过滤
- **WHEN** group.path 为空字符串
- **THEN** provider 使用 ListRepositories 全量拉取，不进行前缀过滤

#### Scenario: 多页分页拉取
- **WHEN** 组下有 250 个仓库，perPage=100
- **THEN** provider 依次请求 page=1、page=2、page=3（通过 x-next-page 判断），合并所有结果

#### Scenario: 默认剔除删除中代码库
- **WHEN** ListGroupRepositories 返回 3 个代码库，其中 1 个 `name` 含 `deletion_scheduled`，`ListReposParams.IncludeDeleted` 未设置（默认 false）
- **THEN** `ListRepos` 返回 2 个正常代码库，删除中的代码库被剔除

#### Scenario: 组被删除时其下所有代码库被剔除
- **WHEN** ListGroupRepositories 返回的代码库 `pathWithNamespace` 形如 `dsp-services-deletion_scheduled-452/repo-a`，`IncludeDeleted` 为 false
- **THEN** 该代码库被剔除（pathWithNamespace 含 `deletion_scheduled`）

#### Scenario: IncludeDeleted 为 true 时保留删除中代码库
- **WHEN** `ListReposParams.IncludeDeleted` 为 `true`，发现结果含删除中代码库
- **THEN** `ListRepos` 返回结果包含这些删除中代码库

### Requirement: Codeup ListGroups 实现
Codeup provider 的 `ListGroups` 方法 SHALL 调用 `GET /oapi/v1/codeup/organizations/{orgId}/namespaces?page={page}&perPage=100` 列出组织下的所有代码组空间，将结果映射为 `RemoteGroup` 数组。

每个 namespace 条目 SHALL 映射为：
- `Name` = `path`
- `Path` = `pathWithNamespace`（如不存在则使用 `fullPath`）
- `Provider` = `"codeup"`

#### Scenario: 成功列出代码组
- **WHEN** 调用 ListGroups，organizationId 为 `646c6887be6b046b8f87bb30`
- **THEN** provider 调用 `GET .../namespaces?perPage=100&page=1`，返回所有代码组空间

#### Scenario: 代码组映射为 RemoteGroup
- **WHEN** ListNamespaces 返回 `{ path: "my-group", pathWithNamespace: "org/my-group" }`
- **THEN** 映射为 `RemoteGroup{ Name: "my-group", Path: "org/my-group", Provider: "codeup" }`

#### Scenario: 多页分页
- **WHEN** 组织有 150 个代码组空间，perPage=100
- **THEN** provider 请求 page=1 和 page=2，合并结果

#### Scenario: ListNamespaces 调用失败
- **WHEN** API 调用返回非 200 状态码
- **THEN** provider 返回空列表，在 verbose 模式下输出警告
