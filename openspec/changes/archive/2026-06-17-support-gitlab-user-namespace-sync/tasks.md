## 1. GitLab 实体识别与接口路由

- [x] 1.1 在 `provider/gitlab.go` 中新增 path 实体识别逻辑，区分 Group 与 User namespace。
- [x] 1.2 实现 Group 优先、User 回退的仓库列举路由，保留 Group 递归能力。
- [x] 1.3 统一并改进“path 不存在/不可访问”错误语义，避免仅暴露 `Group Not Found`。

## 2. 同步链路兼容与行为校验

- [x] 2.1 确保 User namespace 返回仓库后仍沿用现有 sync 过滤链路（group path、exclude、批次去重）。
- [x] 2.2 补充/更新 `provider/gitlab_test.go`：覆盖 Group 命中、User 命中、双 404、鉴权失败等场景。
- [x] 2.3 补充 sync 相关测试，验证个人命名空间同步成功且不回归现有 Group 行为。

## 3. 文档与验收

- [x] 3.1 更新 README（含多语言版本）中 GitLab `group.path` 使用说明，补充个人命名空间示例。
- [x] 3.2 运行相关测试并记录结果，确认变更满足 specs 中新增场景。
