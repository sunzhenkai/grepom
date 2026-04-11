## 1. 导出 isExcluded 函数

- [x] 1.1 将 `repo/resolver.go` 中的 `isExcluded()` 函数导出为 `IsExcluded()`，签名改为 `func IsExcluded(excludeRepos []string, repoName string) bool`
- [x] 1.2 更新 `resolver.go` 内部所有调用处，从 `isExcluded` 改为 `IsExcluded`
- [x] 1.3 运行现有测试确保无回归

## 2. sync 命令跳过被排除仓库

- [x] 2.1 在 `cmd/sync.go` 中，远程发现仓库转换为 `GroupRepo` 列表后，增加 `repo.IsExcluded(g.ExcludeRepos, r.Name)` 过滤，跳过匹配的仓库
- [x] 2.2 在 verbose 模式下输出每个 group 跳过的被排除仓库数量
- [x] 2.3 在同步摘要中区分"发现的仓库数"和"被排除跳过的仓库数"

## 3. list --remote 过滤被排除仓库

- [x] 3.1 在 `cmd/list.go` 的 `runListRemoteRepos` 中，遍历 groups 时增加 `enabled` 检查，跳过禁用的 group 和 resource（除非 `listAll` 为 true）
- [x] 3.2 在远程仓库展示前，增加 `exclude_repos` 过滤逻辑（除非 `listAll` 为 true）
- [x] 3.3 当 `listAll` 为 true 时，为被排除的仓库标注 `[excluded]`，为禁用 group 下的仓库标注 `[disabled]`

## 4. 测试

- [x] 4.1 为 `repo.IsExcluded()` 编写或更新单元测试
- [x] 4.2 为 sync 跳过 exclude_repos 仓库编写测试
- [x] 4.3 为 list --remote 过滤 exclude_repos 编写测试
- [x] 4.4 为 list --remote --all 包含被排除仓库编写测试
- [x] 4.5 运行完整测试套件确保无回归
