## 1. 基础设施：loadConfig 返回配置文件路径

- [x] 1.1 修改 `cmd/root.go` 中 `loadConfig()` 函数签名为 `loadConfig() (string, *config.Config, error)`，返回配置文件路径
- [x] 1.2 更新所有调用 `loadConfig()` 的命令（grep 所有 `loadConfig()` 调用点），适配新签名

## 2. 无参数行为：输出配置文件所在目录

- [x] 2.1 修改 `cmd/dir.go` 无参数分支：`filepath.Dir(configPath)` 替代 `cfg.Base`
- [x] 2.2 更新 `cmd/dir_test.go` 中 `TestDirCommand_NoArgs_PrintsBase` 断言，改为校验配置文件所在目录
- [x] 2.3 更新 `TestDirCommand_UpwardConfigSearch` 断言（如果受影响）

## 3. 搜索优先级：先精确后子串

- [x] 3.1 在 `repo/resolver.go` 新增 `ApplyExactFirstSearch(repos []provider.Repo, keyword string, filter Filter) []provider.Repo`，实现两阶段搜索
- [x] 3.2 修改 `cmd/dir.go` 中有参数分支，调用 `ApplyExactFirstSearch` 替代 `ApplySearchFilter`
- [x] 3.3 新增测试：精确匹配恰好一个仓库时直接返回
- [x] 3.4 新增测试：精确匹配多个同名仓库时输出多行
- [x] 3.5 新增测试：无精确匹配退回子串匹配

## 4. 多匹配输出行为：stdout 多行路径

- [x] 4.1 修改 `cmd/dir.go` 中 `default` 分支：多匹配时输出每行一个路径到 stdout，退出码 0
- [x] 4.2 更新 `TestDirCommand_MultipleMatch_ReturnsError` 为新的多匹配行为（不再返回 error）
- [x] 4.3 新增测试：多匹配时 stdout 输出正确行数和路径

## 5. Shell function：运行时检测 fzf

- [x] 5.1 替换 `shellHelperWithFzf` 和 `shellHelperWithoutFzf` 为单一 `shellHelper` 模板，包含运行时 fzf 检测
- [x] 5.2 移除 `fzfAvailable()` 函数
- [x] 5.3 简化 `--shell` 分支：直接输出 `shellHelper`
- [x] 5.4 更新 `TestDirCommand_ShellPrintsFunction` 校验新函数内容（含 `command -v fzf`）
