## 1. git 包基础改造

- [x] 1.1 在 `CloneOptions` 中新增 `LogWriter io.Writer` 字段，用于控制认证尝试日志的输出目标
- [x] 1.2 修改 `Clone()` 函数：当 `LogWriter` 为 nil 时保持 `fmt.Printf` 行为；非 nil 时将所有日志写入 `LogWriter`（使用 `fmt.Fprintf`）
- [x] 1.3 新增 `GetDefaultBranch(path string) (string, error)` 函数：通过 `git symbolic-ref refs/remotes/origin/HEAD` 获取默认分支名，解析出短分支名返回
- [x] 1.4 为 `GetDefaultBranch` 编写单元测试：覆盖正常获取、origin/HEAD 未设置、非 git 目录等场景

## 2. 并行执行框架

- [x] 2.1 在 `git` 包中新增 `CloneResult` 结构体：包含 `Repo provider.Repo`、`FullPath string`、`Err error`、`Log string`（认证尝试日志文本）
- [x] 2.2 在 `git` 包中新增 `CloneAll(concurrency int, repos []CloneTask) []CloneResult` 函数，实现 worker pool 并行克隆
- [x] 2.3 在 `git` 包中新增 `PullResult` 结构体：包含 `Repo provider.Repo`、`FullPath string`、`Err error`、`Skipped bool`、`SkipReason string`
- [x] 2.4 在 `git` 包中新增 `PullAll(concurrency int, tasks []PullTask) []PullResult` 函数，实现 worker pool 并行 pull
- [x] 2.5 为 `CloneAll` 和 `PullAll` 编写单元测试：覆盖并行执行、部分失败、空列表等场景

## 3. 进度显示

- [x] 3.1 在 `cmd` 包中新增 `isTerminal() bool` 函数，使用 `isatty.IsTerminal()` 检测 stdout 是否为 TTY
- [x] 3.2 实现进度行渲染函数：TTY 模式下使用 `\r` 覆盖更新显示 `[N/M] cloning/pulling...`；非 TTY 模式下逐行输出
- [x] 3.3 实现完成摘要输出函数：统计成功/失败/跳过数量，TTY 模式下用 ✓/✗ 标记各仓库状态

## 4. clone 命令改造

- [x] 4.1 在 `cmd/clone.go` 中新增 `--concurrency` 标志（默认值 4），添加参数校验（必须为正整数）
- [x] 4.2 重构 clone 命令的 `RunE`：过滤已克隆仓库后，当 `concurrency > 1` 时调用 `git.CloneAll`，否则保持原有顺序逻辑
- [x] 4.3 并行模式下：收集 `CloneResult`，使用进度显示函数实时更新，完成后输出摘要（含失败详情）
- [x] 4.4 顺序模式下：保持原有 `fmt.Printf` 输出格式不变
- [x] 4.5 更新 clone 命令的 `Example` 帮助文本，增加 `--concurrency` 用法示例

## 5. pull 命令改造

- [x] 5.1 在 `cmd/pull.go` 中新增 `--concurrency` 标志（默认值 4）和 `--force` 标志（默认 false），添加参数校验
- [x] 5.2 实现安全检查逻辑：对每个已克隆仓库调用 `GetStatus()` + `GetDefaultBranch()`，判断是否在默认分支且 clean
- [x] 5.3 重构 pull 命令的 `RunE`：先执行安全检查过滤，满足条件的仓库当 `concurrency > 1` 时调用 `git.PullAll`，否则顺序执行
- [x] 5.4 `--force` 模式：跳过安全检查，对所有已克隆仓库执行 pull
- [x] 5.5 输出跳过信息（非默认分支、dirty、无法检测默认分支）、进度显示、完成摘要
- [x] 5.6 更新 pull 命令的 `Long` 描述和 `Example` 帮助文本，说明安全检查行为和 `--force` 用法

## 6. 交互模式适配

- [x] 6.1 更新 `cmd/interactive.go` 中的 `interactiveClone()` 函数：调用重构后的 clone 逻辑（支持并行）
- [x] 6.2 更新 `cmd/interactive.go` 中的 `interactivePull()` 函数：调用重构后的 pull 逻辑（支持安全检查和并行）
- [x] 6.3 交互模式下为 clone/pull 增加并发度选择提示（可选，使用 survey 选择器）

## 7. 测试与验证

- [x] 7.1 运行 `go build` 确保编译通过
- [x] 7.2 运行 `go test ./...` 确保现有测试全部通过
- [ ] 7.3 手动测试：`grepom clone --concurrency 1` 顺序克隆，验证输出格式不变
- [ ] 7.4 手动测试：`grepom clone --concurrency 4` 并行克隆，验证进度显示和摘要输出
- [ ] 7.5 手动测试：`grepom pull` 验证安全检查（跳过 dirty 和非默认分支仓库）
- [ ] 7.6 手动测试：`grepom pull --force` 验证跳过安全检查
- [ ] 7.7 手动测试：管道模式 `grepom clone 2>&1 | cat` 验证非 TTY 降级输出
