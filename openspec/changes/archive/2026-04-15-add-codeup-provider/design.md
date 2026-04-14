## Context

grepom 当前支持三个 provider：GitLab、GitHub 和 Generic。Provider 通过 `provider.Register()` 在 `init()` 中注册，实现 `Provider` 接口的 `ListRepos` 和 `ListGroups` 方法。

Codeup（阿里云云效）的 API 与现有 provider 有几个关键差异：

1. **API 域名与 clone 域名相同**：都是 `codeup.aliyun.com`，但 API 调用的基础地址是 `devops.aliyun.com`
2. **认证方式不同**：accessToken 通过 URL query parameter 传递，不是 HTTP Header
3. **分页方式不同**：使用 `page`/`perPage` 参数，响应中包含 `total` 字段计算总页数，而非 Link Header
4. **没有按组过滤仓库的 API**：`ListRepositories` 返回整个组织的所有仓库，需要客户端按 `pathWithNamespace` 前缀过滤
5. **需要 organizationId**：所有 API 都需要企业标识（string 类型），且 `ListRepositoryGroups` 还需要 `parentId`（namespaceId，long 类型）
6. **响应结构不同**：所有 Codeup API 响应都包裹在 `{ requestId, success, errorCode, errorMessage, total, result }` 结构中

## Goals / Non-Goals

**Goals:**
- 实现 `codeup` provider 的 `ListRepos` 方法，通过 `ListRepositories` API 全量拉取仓库并按代码组路径过滤
- 实现 `codeup` provider 的 `ListGroups` 方法，通过 `identityGetGroupByPath` + `ListRepositoryGroups` API 查询代码组
- Resource 配置新增可选 `organization_id` 字段
- 在 `config.validate()` 和命令层（sync/list）中支持 `codeup` provider

**Non-Goals:**
- 不支持阿里云 AK/SK 签名认证（仅支持 accessToken）
- 不支持按 `search` 关键字过滤（后续可扩展）
- 不处理 `recursive: false` 的精确层级控制（V1 默认全量，后续优化）
- 不实现 Codeup 私有化部署的自定义域名支持

## Decisions

### D1: API 基础地址映射策略

**决定**：在 codeup provider 内部硬编码 API 基础地址为 `https://devops.aliyun.com`，类似 GitHub 的 `githubAPIURL()` 将 `github.com` 映射为 `api.github.com`。

**理由**：
- 与 GitHub 的处理模式一致，代码模式已被项目采纳
- 用户配置 `url: codeup.aliyun.com` 用于 clone，provider 内部处理 API 地址映射
- Codeup 目前没有已知的私有化部署需求

**备选**：让用户在 `url` 字段配置 API 地址 → 会影响 clone URL 推导，不采用

### D2: 全量拉取 + 客户端过滤

**决定**：`ListRepos` 先调用 `ListRepositories` 全量拉取组织下所有仓库，再按 `group.path` 作为 `pathWithNamespace` 的前缀进行客户端过滤。

**理由**：
- Codeup 的 `ListRepositories` API 没有按代码组过滤的参数
- 这是唯一可行的方案，无需额外 API 调用
- 对于典型组织（几十到几百个仓库），全量拉取性能可接受

**备选**：使用 `search` 参数模糊匹配 → 不可靠，不支持路径前缀精确匹配

### D3: organizationId 存放位置

**决定**：在 Resource 结构体新增 `OrganizationID string` 字段（YAML: `organization_id`）。

**理由**：
- organizationId 是 resource 级别的属性，同一个云效企业下的所有 group 共享
- 语义清晰，不与 `group.path`（代码组路径）混淆
- 可选字段，其他 provider 忽略即可，向后兼容

**备选**：
- 编码进 `url`（如 `codeup.aliyun.com/60de...`）→ 影响 clone URL 推导
- 编码进 `group.path`（如 `60de.../wii/solo`）→ 语义混乱，重复配置

### D4: ListGroups 实现策略

**决定**：通过 `identityGetGroupByPath` API 获取企业根 namespaceId，然后调用 `ListRepositoryGroups` 遍历一级代码组。如果获取根 namespaceId 失败，返回空列表并记录警告。

**理由**：
- `ListRepositoryGroups` 的 `parentId` 是必填的，需要 namespaceId
- `identityGetGroupByPath` 可以通过路径查代码组，可能支持 organizationId 作为根路径
- 根 namespaceId 的获取方式需要实际调试确认，因此需要降级方案

**风险**：如果 `identityGetGroupByPath` 无法通过 organizationId 获取根 namespaceId，`ListGroups` 将不可用

### D5: 认证参数传递

**决定**：accessToken 作为 URL query parameter 传递（`?accessToken=xxx`），不使用 HTTP Header。

**理由**：Codeup API 文档明确要求 `accessToken` 作为查询参数传递，与 GitHub（Bearer Header）和 GitLab（PRIVATE-TOKEN Header）不同。

### D6: 统一响应结构

**决定**：定义 `codeupResponse<T>` 泛型结构体，统一处理所有 Codeup API 的响应解析和错误检测。

**理由**：Codeup 所有 API 响应都包含 `{ requestId, success, errorCode, errorMessage, total, result }` 结构，可以复用同一个解析逻辑。

## Risks / Trade-offs

**[API 基础地址不确定]** → 当前使用 `https://devops.aliyun.com`，如果不对可以快速修改。Provider 内部封装了映射函数，改一处即可。

**[全量拉取性能]** → 对于大型组织（1000+ 仓库），全量拉取会较慢。Mitigation：后续可考虑缓存或增量同步优化。

**[根 namespaceId 获取方式不确定]** → `identityGetGroupByPath` 能否用 organizationId 获取根 namespaceId 需要实际调试。Mitigation：ListGroups 有降级方案（返回空列表），不影响核心的 ListRepos 功能。

**[organization_id 字段影响现有 Resource 结构]** → 新字段是可选的（`omitempty`），不影响 YAML 序列化和现有配置。Mitigation：在 `writeConfig` 中确保空值不出现在输出中。
