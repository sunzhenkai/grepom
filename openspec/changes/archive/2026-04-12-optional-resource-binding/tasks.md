## 1. 配置验证层改造

- [x] 1.1 修改 `config/config.go` 的 `validate()` 函数：移除 group 的 resource 必填检查，改为可选验证（如果指定了 resource 则必须存在于 Resources map 中）
- [x] 1.2 修改 `config/config.go` 的 `validate()` 函数：移除 standalone repo 的 resource 必填检查，添加"无 resource 时 url 必填"的新验证规则
- [x] 1.3 修改 `config/config.go` 的 `validate()` 函数：有 resource 的 group 仍要求 path 必填，无 resource 的 group path 可选
- [x] 1.4 为新的验证规则编写单元测试（`config/config_test.go`）

## 2. Resolver 适配

- [x] 2.1 修改 `repo/resolver.go` 的 `ResolveGroups()`：处理 group 未绑定 resource 的情况，对无 resource 的 group 中的 repo 直接使用其 url 字段
- [x] 2.2 修改 `repo/resolver.go` 的 `ResolveRepos()`：处理 standalone repo 未绑定 resource 的情况，直接使用其 url 字段
- [x] 2.3 为 resolver 的可选 resource 场景编写单元测试

## 3. 命令层适配

- [x] 3.1 修改 `cmd/sync.go`：检测无 resource 的 group，跳过远程发现并输出提示信息
- [x] 3.2 修改 `cmd/list.go`：`list --remote` 跳过无 resource 的 group 并输出提示
- [x] 3.3 修改 `cmd/add.go`：`add group` 和 `add repo` 的 `--resource` 参数改为可选
- [x] 3.4 修改 `cmd/clone.go`：处理无 resource 的 repo，使用系统默认认证执行 clone（无需修改，resolver 已处理）
- [x] 3.5 修改 `cmd/pull.go`：处理无 resource 的 repo，使用系统默认认证执行 pull（无需修改，resolver 已处理）

## 4. 集成验证

- [x] 4.1 验证已有配置（绑定 resource）的行为不受影响
- [x] 4.2 验证无 resource 的 group 的手动管理流程正常工作
- [x] 4.3 验证 sync/sync --group/sync --resource 在无 resource 场景下的提示输出
