## 1. XDG 状态目录

- [x] 1.1 在 `service/scope.go` 新增 `StateHomeFunc` 和 `defaultStateHome()`，优先 `$XDG_STATE_HOME`，fallback `~/.local/state`
- [x] 1.2 将 `StateDir()` 改为 `StateHome/grepom/services/<scopeID>/`，移除对 `UserConfigDirFunc` 的依赖
- [x] 1.3 更新 `service/scope_test.go` 和 `service/manager_test.go` 中的目录断言
- [x] 1.4 确认 registry/logs 写入前目录自动创建（`MkdirAll`）

## 2. 紧凑列表与 verbose 模式

- [x] 2.1 为 `printSvcTable` 增加 `verbose bool` 参数：默认 4 列，verbose 6 列
- [x] 2.2 实现 PATH 显示的 `~` home 前缀缩短（仅展示层，不改 registry）
- [x] 2.3 为 `svc list` 添加 `-v`/`--verbose` 标志并接入 `printSvcTable`
- [x] 2.4 `svc status <name>` 保持完整元数据输出（verbose 模式）
- [x] 2.5 无参数 `svc status` 行为与默认 `svc list` 一致（紧凑 4 列）
- [x] 2.6 更新 `cmd/svc_test.go` 覆盖紧凑/verbose 两种输出

## 3. Shell 补全

- [x] 3.1 新增 `cmd/completion.go`，注册 `grepom completion bash|zsh|fish` 子命令
- [x] 3.2 实现 `completeSvcNames`：合并 registry 服务名与 `.grepom.yml` 配置名，去重排序
- [x] 3.3 在 `registerSvcSubcommands` 中为 `logs`、`kill`、`dir`、`status`、`run` 挂载 `ValidArgsFunction`
- [x] 3.4 补全失败时静默返回空列表（`ShellCompDirectiveNoFileComp`）
- [x] 3.5 添加补全相关单元测试（至少验证 `completeSvcNames` 逻辑）

## 4. 文档

- [x] 4.1 更新 `README.md`：XDG 状态目录说明、旧路径不迁移提示、`list -v` 用法、shell 补全安装示例
- [x] 4.2 更新 `README_en.md` 保持与中文文档一致
- [x] 4.3 更新 `svc list` 命令 help 示例，说明 `-v` 标志

## 5. 验证

- [x] 5.1 运行 `go test ./service/... ./cmd/...` 确保全部通过
- [x] 5.2 手动验证：新路径下 `svc run` → `svc list`（4 列）→ `svc list -v`（6 列）→ `svc status api`（完整）
- [x] 5.3 手动验证：`eval "$(grepom completion zsh)"` 后 `grepom svc kill <TAB>` 可补全服务名
