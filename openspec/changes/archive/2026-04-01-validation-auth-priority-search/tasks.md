## 1. 认证优先级调整（SSH 优先）

- [x] 1.1 修改 `git/git.go` 的 `Clone` 函数中策略构建顺序： 将 SSH key 策略移到 token 策略之前（group/repo token → resource SSH key → resource token → 默认 SSH → 裸 HTTP）
- [x] 1.2 删除 `authLevel` 函数，改用 `buildAuthStrategies` + source tracking field实现正确的 6 级认证链

- [x] 1.3 更新 `git/git_test.go` 中现有 clone 认证优先级相关测试用例， 验证 SSH 优先行为
- [x] 1.4 更新 `repo/resolver_test.go` 中认证合并相关测试， 新增 source tracking 测试

 新增 `provider.Repo` 的 `HasGroupToken`/`HasGroupSSHKey` source tracking 字段

- [x] 2.1-2.5 修改 `cmd/add.go` 的 `addGroupCmd` 和 `addRepoCmd` 添加前置校验: resource 存在性、 group/repo 名称唯一性)
- [x] 3.1 在 `repo/resolver.go` 中新增 `ApplySearchFilter` 函数: 对 repo 列表按名称子串进行大小写不敏感匹配)
- [x] 3.2-3.4 新建 `cmd/search.go` : 实现 `grepom search <keyword>` 命令, 支持 `--group` 和 `--resource` 过滤器
- [x] 3.5 在 `repo/resolver_test.go` 中增加 `ApplySearchFilter` 的单元测试

- [x] 4.1 运行 `go build` 确保编译通过
 ✓
- [x] 4.2 运行 `go test ./...` 确保所有测试通过
 ✓
- [x] 4.3 手动验证: 添加 group 引用不存在的 resource 时正确报错
 ✓
- [x] 4.4 手动验证: search 命令按子串搜索仓库并正确显示结果