## 1. 基础设施：API 域名映射、认证、响应解析

- [x] 1.1 重写 `codeupAPIURL()` → `codeupAPIBaseURL()`，将 `codeup.aliyun.com` 映射为 `openapi-rdc.aliyuncs.com`（返回完整 API base URL），保留 clone host 推导逻辑
- [x] 1.2 实现 `get()` / `getWithPagination()` 通用请求方法：使用 `x-yunxiao-token` Header 认证，返回值包含响应头分页信息
- [x] 1.3 实现响应头分页解析：从 `x-total`、`x-next-page`、`x-page`、`x-per-page` 提取分页状态，`x-next-page` 为空或 0 时终止翻页
- [x] 1.4 删除旧版 `codeupResponse` wrapper 结构体及 `getWithTotal()` 方法

## 2. ListGroups：迁移至 ListNamespaces

- [x] 2.1 定义新版 `codeupNamespace` 响应结构体：`id`、`path`、`fullPath`、`pathWithNamespace`、`name`、`parentId`、`kind`、`visibility`
- [x] 2.2 重写 `ListGroups()`：调用 `GET .../namespaces?perPage=100&page={page}`，分页拉取所有 namespace，映射为 `RemoteGroup{ Name: path, Path: pathWithNamespace, Provider: "codeup" }`
- [x] 2.3 删除旧版 `codeupGroup`、`codeupGroupDetail` 结构体，删除 `getRootNamespaceID()` 和 `listTopLevelGroups()` 方法

## 3. ListRepos：迁移至 ListGroupRepositories（两步查询）

- [x] 3.1 定义新版 `codeupRepo` 响应结构体：`id`、`name`、`path`、`pathWithNamespace`、`nameWithNamespace`、`webUrl`、`visibility`、`archived`、`namespaceId`（按新版 OAPI v1 文档字段）
- [x] 3.2 实现 `resolveGroupID()`：调用 `ListNamespaces?search={groupPath}`，精确匹配 `pathWithNamespace == groupPath`，返回 groupId
- [x] 3.3 重写 `listGroupRepos()` 主路径：使用 `resolveGroupID()` 获取 groupId，调用 `GET .../groups/{groupId}/repositories?includeSubgroups={recursive}` 按组拉取
- [x] 3.4 实现回退路径：当 `resolveGroupID()` 无精确匹配时，回退到 `GET .../repositories?perPage=100&page={page}` 全量拉取 + 客户端 `pathWithNamespace` 前缀过滤，verbose 输出警告
- [x] 3.5 clone URL 推导：使用配置的 cloneHost（`resource.url`）构建 `https://{cloneHost}/{pathWithNamespace}.git` 和 `git@{cloneHost}:{pathWithNamespace}.git`

## 4. 测试重写

- [x] 4.1 重写 `TestCodeupAPIURL` → `TestCodeupAPIBaseURL`，验证域名映射 `codeup.aliyun.com` → `openapi-rdc.aliyuncs.com`
- [x] 4.2 重写 `TestCodeupProvider_ListRepos_PathFilter`：mock server 返回新版格式（直接数组 + 响应头分页 + x-yunxiao-token Header），验证两步查询路径
- [x] 4.3 重写 `TestCodeupProvider_ListRepos_Pagination` 和 `TestCodeupProvider_ListRepos_MultiPage`：使用 `x-next-page`/`x-total` 响应头模拟分页
- [x] 4.4 重写 `TestCodeupProvider_ListRepos_AuthFailure` 和 `TestCodeupProvider_ListRepos_APIError`：适配新版响应格式
- [x] 4.5 重写 `TestCodeupProvider_ListGroups`：mock server 模拟 `ListNamespaces` 响应，验证 namespace → RemoteGroup 映射
- [x] 4.6 新增 `TestCodeupProvider_ListRepos_FallbackToFullList`：验证 group path 未匹配 namespace 时的全量回退路径
- [x] 4.7 新增 `TestCodeupProvider_ListRepos_Recursive`：验证 `includeSubgroups=true` 传递
- [x] 4.8 运行全部测试确认通过：`go test ./provider/...`
