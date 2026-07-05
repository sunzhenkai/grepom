## 1. 基础：检测函数与参数结构

- [x] 1.1 在 `provider/` 包新增 `isDeletionScheduled(name, pathWithNamespace string) bool` 辅助函数：当 `name` 或 `pathWithNamespace` 包含子串 `deletion_scheduled` 时返回 `true`
- [x] 1.2 为 `isDeletionScheduled` 添加单元测试（覆盖 name 含标记、仅 path 含标记、正常库三种情形）
- [x] 1.3 在 `provider.ListReposParams` 新增 `IncludeDeleted bool` 字段

## 2. 发现层：Codeup provider 过滤

- [x] 2.1 在 `provider/codeup.go` 的 `listGroupReposByID` 中，当 `params.IncludeDeleted` 为 false 时，使用 `isDeletionScheduled` 剔除删除中代码库
- [x] 2.2 在 `listAllReposFull` 中应用相同过滤逻辑（覆盖 group 与 fallback 两条路径）
- [x] 2.3 让 `ListRepos` 把 `params.IncludeDeleted` 透传到两个子方法（调整签名或读取 params 字段）
- [x] 2.4 在 `provider/codeup_test.go` 新增测试：默认剔除删除中库、`IncludeDeleted=true` 保留、组被删除时其下库被剔除

## 3. 兜底层：resolver 运行时拦截

- [x] 3.1 在 `repo/resolver.go` 的 `resolveInternal` 中，对 group repo 与独立 repo 调用 `isDeletionScheduled`，命中时设置 `DisabledReason = "deletion_scheduled"`
- [x] 3.2 确认 `Resolve()` 默认剔除、`ResolveAndFilter(IncludeDisabled=true)` 保留的逻辑天然覆盖新 reason（无需额外改动则验证，否则补齐）
- [x] 3.3 在 `repo/resolver_test.go` 新增测试：配置含删除中库时默认跳过、`IncludeDisabled=true` 时保留并标注 reason

## 4. CLI 标志与提示

- [x] 4.1 在 `cmd/sync.go` 新增 `--include-deleted` 布尔标志，透传到 `ListReposParams.IncludeDeleted`
- [x] 4.2 在 `cmd/list.go` 新增 `--include-deleted` 布尔标志，透传到 `ListReposParams.IncludeDeleted`
- [x] 4.3 在 `cmd/sync.go` verbose 模式下输出被跳过的删除中代码库数量（如 `skipped N deletion_scheduled repos`）
- [x] 4.4 为 `--include-deleted` 标志添加命令测试（`cmd/sync_test.go`、`cmd/list_test.go`）

## 5. 文档更新

- [x] 5.1 更新 `README.md`：说明默认跳过删除中代码库的行为与 `--include-deleted` 标志
- [x] 5.2 更新 `README_en.md`（与中文版对应）

## 6. 验证

- [x] 6.1 运行 `go build ./...` 与 `go test ./...` 全量通过
- [x] 6.2 手动验证：对含删除中库的 Codeup group 执行 `grepom sync -v`，确认无 `all authentication methods failed` 且输出跳过计数（单元测试已通过 Codeup mock 覆盖该路径；真实 Codeup 端到端验证留给用户）
