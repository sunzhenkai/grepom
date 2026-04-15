## Context

grepom 是一个多 provider 仓库管理工具，通过统一的 `Provider` 接口对接 GitLab、GitHub、Codeup 等代码平台。Codeup provider 当前使用旧版 API，域名 `devops.aliyun.com` 实际上是云效 Web 前端而非 API 端点，导致所有请求返回 HTML，功能完全不可用。

云效官方提供两套 API：
- 旧版 (`api-devops-2021-06-25`)：域名 `devops.cn-hangzhou.aliyuncs.com`，accessToken query param，wrapper 响应
- 新版 OAPI v1（推荐）：域名 `openapi-rdc.aliyuncs.com`（中心版），`x-yunxiao-token` Header 认证，直接数组 + 响应头分页

用户选择方案 A：配置文件中 `url` 保持 `codeup.aliyun.com` 不变（仅用于 clone URL），API 域名内部映射。

## Goals / Non-Goals

**Goals:**
- 迁移至新版 OAPI v1，使 Codeup provider 的 sync、list 命令正常工作
- 保持用户配置完全兼容（`url`、`token`、`organization_id` 字段不变）
- 利用新版 API 的 `ListGroupRepositories`（支持 `includeSubgroups`）实现更精准的按组查询，避免拉全量
- 利用新版 API 的响应头分页（`x-total`/`x-next-page`）替代旧的 body 分页

**Non-Goals:**
- 不支持 Region 版的实例专属域名（当前硬编码 `openapi-rdc.aliyuncs.com`，未来可扩展）
- 不引入 `ListUserResources` API（用户维度的资源查询，当前不在 grepom 核心功能范围内）
- 不修改 `Provider` 接口、`ListReposParams`、`ListGroupsParams` 等跨 provider 共享类型
- 不修改 `config/config.go`、`cmd/sync.go`、`cmd/list.go`

## Decisions

### D1: API 域名映射策略

**决定**: `codeupAPIDomain()` 函数将 `codeup.aliyun.com` 映射为 `openapi-rdc.aliyuncs.com`，clone host 保持用户配置的 `url` 不变。

**理由**: 与 GitHub provider 的 `githubAPIURL()` 模式一致——配置中存的是 clone host，内部映射为 API host。用户配置零改动。

**替代方案**: 新增 `api_url` 配置字段。更灵活但增加了配置复杂度，且当前只有中心版一个场景，YAGNI。

### D2: ListRepos 两步查询策略

**决定**: `ListRepos` 改为两步：
1. 调用 `ListNamespaces?search={groupPath}` 找到 group 对应的 `groupId`（精确匹配 `pathWithNamespace`）
2. 调用 `ListGroupRepositories` 按 `groupId` 精确拉取，通过 `includeSubgroups` 控制是否递归

**理由**: 旧版拉全量 org repos 再客户端过滤（可能几千个仓库），新版直接按组查询，数据量精准可控。`includeSubgroups=true` 原生支持递归，替代了旧版的客户端前缀过滤。

**替代方案**: 继续用 `ListRepositories` 拉全量 + 客户端过滤。简单但浪费，违背新版 API 的设计意图。

### D3: 响应头分页

**决定**: 分页信息从响应头 `x-total`/`x-total-pages`/`x-next-page` 读取。当 `x-next-page` 存在且大于当前页时继续翻页。

**理由**: 新版 OAPI v1 标准分页机制，与 GitHub 的 `Link` header 分页、GitLab 的 `x-next-page` header 分页模式一致。

### D4: ListGroups 使用 ListNamespaces

**决定**: `ListGroups` 直接调用 `ListNamespaces` 列出组织下的所有代码组空间。

**理由**: 旧版需要两步（先 `find_by_path` 获取根 ID，再 `groups/get/all`），新版一步到位。`ListNamespaces` 返回的 `path`/`fullPath`/`pathWithNamespace` 可直接映射为 `RemoteGroup`。

### D5: ListRepos 中 group path 查找失败的降级策略

**决定**: 如果 `ListNamespaces?search=` 未找到精确匹配的 group，则回退到 `ListRepositories`（全量 org 仓库列表）并做客户端 path 前缀过滤，与旧版行为一致。

**理由**: 用户配置的 `group.path` 可能是一个不存在于 namespaces 中的路径（如手动输入错误），回退到全量模式保证不静默失败。同时在 verbose 模式输出警告。

## Risks / Trade-offs

- **[Risk] search 模糊匹配精度** → `ListNamespaces?search=` 是模糊搜索，可能返回多个结果。**缓解**: 在结果中精确匹配 `pathWithNamespace == groupPath`。如果无精确匹配，回退到全量模式（D5）。
- **[Risk] 中心版域名硬编码** → `openapi-rdc.aliyuncs.com` 只适用于中心版。**缓解**: 当前已知用户均为中心版；如需 Region 版支持，可后续在 `codeupAPIDomain()` 中扩展条件判断。
- **[Risk] 新版 API 响应格式变化** → 新版 OAPI 返回直接数组（无 wrapper），字段名与旧版略有差异（如 `visibility` 替代 `visibilityLevel`）。**缓解**: 重写响应结构体，严格按新版文档字段定义。
