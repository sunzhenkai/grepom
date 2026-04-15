## 1. 修改 IsExcluded 函数

- [x] 1.1 修改 `IsExcluded` 函数签名，新增 `remotePath string` 参数
- [x] 1.2 在 `IsExcluded` 中实现匹配分支：无通配符走 `pattern == repoName` 精确匹配，含通配符走 `filepath.Match(pattern, remotePath)` glob 匹配
- [x] 1.3 添加 `hasWildcard` 辅助函数（`strings.ContainsAny(pattern, "*?[")`）

## 2. 更新调用点

- [x] 2.1 更新 `resolver.go` 中 `resolveInternal` 两处 `IsExcluded` 调用，传入 `gr.Path`（远端路径）
- [x] 2.2 更新 `cmd/list.go` 和 `cmd/sync.go` 中所有 `IsExcluded` 调用，传入远端路径

## 3. 测试

- [x] 3.1 为 `IsExcluded` 补充单元测试：覆盖精确匹配、glob 前缀匹配、跨层级匹配、混合模式、不匹配等场景
