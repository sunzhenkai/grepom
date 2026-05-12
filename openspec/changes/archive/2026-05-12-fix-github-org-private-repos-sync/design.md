## Context

grepom 的 GitHub provider 在列举仓库时，对 `solo-kingdom` 这类 Organization 使用了 `/users/{name}/repos` 端点。该端点对 Org 虽然返回 200，但只返回公开仓库。正确的端点 `/orgs/{name}/repos` 需要通过 404 回退才能触发，而 200 响应阻止了回退。

当前代码流程（`provider/github.go`）：

```
listReposFor(name)
  ├── listUserRepos(name)   → GET /users/{name}/repos
  │   └── 返回 200? → 直接返回 (仅 public repos)
  └── listOrgRepos(name)    → GET /orgs/{name}/repos (永远不走)
```

实际验证结果（`solo-kingdom` org）：
- `/users/solo-kingdom/repos` → 22 repos（仅公开）
- `/orgs/solo-kingdom/repos` + token → 24 repos（含 2 个私有）

## Goals / Non-Goals

**Goals:**
- Organization 类型的 GitHub 实体能正确发现所有仓库（含私有）
- 保持对个人用户（User 类型）的兼容
- 最小化 API 调用开销

**Non-Goals:**
- 不修改其他 provider（GitLab、Codeup）的行为
- 不改变 config 文件格式或用户配置方式
- 不增加缓存层（本次不做）

## Decisions

### Decision 1: 实体类型预检

**选择**：在列举仓库前，先调用 `GET /users/{name}` 获取 `type` 字段，根据类型选择端点。

```
listReposFor(name)
  ├── getEntityType(name)       → GET /users/{name}
  │   ├── type == "Organization" → listOrgRepos(name)   → /orgs/{name}/repos
  │   └── type == "User"         → listUserRepos(name)   → /users/{name}/repos
  └── 404 → listOrgRepos(name)  (兜底)
```

**替代方案**：
- **方案 B: 反转优先级**（先 `/orgs/` 后 `/users/`）—— 对个人用户会多一次无效 API 调用，且语义上 `/orgs/` 对个人用户也会 404，不优雅
- **方案 C: 两个都查、合并去重**—— 每次都发两次完整仓库列表请求，开销翻倍，不必要

**理由**：方案 A 只多一次轻量 `GET /users/{name}` 调用（返回约 500 字节的用户元信息），然后精准选择正确的端点，一次到位。

### Decision 2: 新增 githubUserInfo 辅助结构

**选择**：新增一个轻量结构体用于解析 `/users/{name}` 响应，只关心 `type` 字段。

```go
type githubUserInfo struct {
    Type string `json:"type"` // "User" or "Organization"
}
```

### Decision 3: 保留 404 兜底

如果 `getEntityType` 返回 404（理论上 GitHub 不存在用户和组织同名的情况，但防御性编程），直接尝试 `/orgs/{name}/repos`。

## Risks / Trade-offs

- **[额外 API 调用]** 每次 sync 每个 group 多一次 `GET /users/{name}` 调用 → 该请求极轻量，对 rate limit 影响可忽略
- **[GitHub API 一致性]** 依赖 GitHub `/users/{name}` 返回的 `type` 字段准确 → 这是 GitHub 公开 API 的稳定字段，文档明确保证
- **[缓存未做]** 多次 sync 同一 group 会重复调用类型检测 → 后续可加缓存优化，当前不引入复杂度
