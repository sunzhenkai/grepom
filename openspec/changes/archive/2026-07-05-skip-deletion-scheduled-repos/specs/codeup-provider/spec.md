## MODIFIED Requirements

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
