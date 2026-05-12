## 1. 新增实体类型检测方法

- [x] 1.1 在 `provider/github.go` 中新增 `githubUserInfo` 结构体（含 `Type string` 字段）
- [x] 1.2 实现 `getEntityType` 方法：调用 `GET /users/{name}`，解析 `type` 字段，返回 `"Organization"` 或 `"User"`；404 时返回空字符串和 nil error

## 2. 修改 listReposFor 逻辑

- [x] 2.1 重构 `listReposFor` 方法：先调用 `getEntityType` 获取实体类型
- [x] 2.2 根据 type 结果路由：`Organization` → `listOrgRepos`，`User` → `listUserRepos`
- [x] 2.3 保留 404 兜底：当 `getEntityType` 返回 404 时，回退到 `listOrgRepos`

## 3. 测试验证

- [x] 3.1 编写单元测试：模拟 GitHub API 响应，覆盖 Organization、User、404 三种场景
- [x] 3.2 手动验证：`grepom sync -g solo-kingdom` 能发现 24 个仓库（含 dop、unihub 两个私有库）

## 4. 文档更新

- [x] 4.1 同步更新 `README.md` 和 `README_en.md`（如涉及行为描述变化）
