## Why

GitHub sync 在发现 organization 仓库时，错误地使用了 `/users/{org}/repos` 端点。该端点对 Organization 虽然返回 200，但**只返回公开仓库**，私有仓库被静默丢弃。这导致用户新建的私有仓库永远无法通过 `grepom sync` 被发现和同步。

根因：`listReposFor` 的回退逻辑仅在 `/users/` 返回 404 时才切换到 `/orgs/`，但 GitHub 对 `/users/{org}/repos` 返回 200，回退路径永远不会触发。

## What Changes

- GitHub provider 新增实体类型检测：在列举仓库前，先通过 `GET /users/{name}` 获取 `type` 字段，判断目标是 `Organization` 还是 `User`
- 根据 type 智能选择 API 端点：Organization 走 `/orgs/{name}/repos`，User 走 `/users/{name}/repos`
- 保留 404 回退作为兜底，确保未知边界情况仍能工作
- 私有仓库将被正确发现和同步

## Capabilities

### New Capabilities

- `github-entity-detection`: GitHub provider 在列举仓库前检测目标实体类型（User/Organization），智能选择正确的 API 端点

### Modified Capabilities

（无已有 spec 需要修改）

## Impact

- **代码影响**：`provider/github.go` 的 `listReposFor` 方法和新增辅助方法
- **API 调用**：每次 sync 多一次轻量 `GET /users/{name}` 请求（只返回用户元信息，不含仓库列表）
- **行为变更**：Organization 类型的 group 现在能正确发现私有仓库，之前被静默忽略
- **兼容性**：完全向后兼容，对已有公开仓库的同步无影响
- **文档**：需同步更新 README
