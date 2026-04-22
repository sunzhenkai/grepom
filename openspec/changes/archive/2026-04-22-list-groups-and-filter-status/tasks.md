## 1. 位置参数 groups/resources 快捷方式

- [x] 1.1 修改 `cmd/list.go` 的 `RunE` 函数，在解析 `--type` 后检查位置参数：当位置参数为 `groups` 且 `--type` 为默认值 `repos` 时，自动将 `listType` 设为 `groups`；当位置参数为 `resources` 且 `--type` 为默认值时，自动设为 `resources`
- [x] 1.2 更新 `listCmd` 的 `Use` 和 `Example` 字段，说明 `groups`/`resources` 位置参数的用法
- [x] 1.3 验证：`grepom list groups` 等价于 `grepom list --type groups`，`grepom list resources` 等价于 `grepom list --type resources`
- [x] 1.4 验证：`--type` 标志优先于位置参数关键字（如 `grepom list groups --type repos` 按 repos 列出）

## 2. --no-push 状态筛选标志

- [x] 2.1 在 `cmd/list.go` 中新增 `listNoPush bool` 变量，注册 `--no-push` flag
- [x] 2.2 修改 `runListRepos` 函数：当 `listNoPush` 为 true 时，在获取仓库列表后遍历调用 `git.GetStatus`，过滤掉 ahead == 0 的仓库
- [x] 2.3 确保 `--no-push` 模式下不展示未克隆的仓库
- [x] 2.4 验证：`grepom list --no-push` 仅展示 ahead > 0 的仓库
- [x] 2.5 验证：`grepom list --no-push --group frontend` 在指定 group 内筛选
- [x] 2.6 验证：无匹配时输出 `No repositories found.`

## 3. --no-commit 状态筛选标志

- [x] 3.1 在 `cmd/list.go` 中新增 `listNoCommit bool` 变量，注册 `--no-commit` flag
- [x] 3.2 修改 `runListRepos` 函数：当 `listNoCommit` 为 true 时，在获取仓库列表后遍历调用 `git.GetStatus`，过滤掉 dirty == 0 的仓库
- [x] 3.3 确保 `--no-commit` 模式下不展示未克隆的仓库
- [x] 3.4 验证：`grepom list --no-commit` 仅展示 dirty > 0 的仓库
- [x] 3.5 验证：无匹配时输出 `No repositories found.`

## 4. --no-push 与 --no-commit 组合及 --remote 兼容

- [x] 4.1 修改 `runListRepos`：当 `--no-push` 和 `--no-commit` 同时使用时，保留 ahead > 0 或 dirty > 0 的仓库（并集逻辑）
- [x] 4.2 确保 `--remote` 模式下 `--no-push` 和 `--no-commit` 静默忽略，不报错
- [x] 4.3 优化性能：仅在 `--no-push` 或 `--no-commit` 被指定时才调用 `git.GetStatus`，默认 list 行为不受影响
- [x] 4.4 验证：`grepom list --no-push --no-commit` 展示并集结果
- [x] 4.5 验证：`grepom list --remote --no-push` 正常执行远程查询，忽略 `--no-push`

## 5. 测试与文档

- [x] 5.1 更新 `cmd/list.go` 的 `Example` 字段，添加 `--no-push` 和 `--no-commit` 的使用示例
- [x] 5.2 手动测试所有新增场景，确认向后兼容（现有命令行为不变）
