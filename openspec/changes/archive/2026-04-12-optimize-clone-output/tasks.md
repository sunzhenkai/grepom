## 1. 扩展进度回调 API

- [x] 1.1 在 `git/parallel.go` 中定义 `ProgressEventType`（`ProgressStart`/`ProgressComplete`）和 `ProgressEvent` 结构体（Type、RepoName、Completed、Total、Err）
- [x] 1.2 将 `ProgressFunc` 签名从 `func(completed, total int)` 改为 `func(ProgressEvent)`
- [x] 1.3 修改 `CloneAll`：在 worker 开始处理每个任务前触发 `ProgressStart` 事件，完成后触发 `ProgressComplete` 事件
- [x] 1.4 修改 `PullAll`：同样在任务开始和完成时分别触发 Start/Complete 事件

## 2. 重构 ProgressRenderer 为多行渲染

- [x] 2.1 重构 `cmd/progress.go` 中的 `ProgressRenderer`：新增 in-flight 任务跟踪（`[]inflightTask`），记录当前正在处理的仓库名
- [x] 2.2 新增 `Handle(event ProgressEvent)` 方法：Start 事件将仓库名加入 in-flight 列表并渲染，Complete 事件从列表移除并渲染
- [x] 2.3 实现 TTY 多行渲染：第一行 `[N/M] action...`，后续每行 `  action repo-name...`，使用 `\033[A` 光标上移实现原地更新
- [x] 2.4 实现非 TTY 逐行输出：Complete 事件时输出 `✓ repo-name`（成功）或 `✗ repo-name: error`（失败）
- [x] 2.5 修改 `Done()` 方法：TTY 模式下使用光标上移清除多行进度区域，非 TTY 模式下无需操作

## 3. 更新命令层适配新 API

- [x] 3.1 修改 `cmd/clone.go` 的 `runParallelClone`：使用新的 `Handle` 方法替代 `Update`，移除非 TTY 模式下的独立 stderr 错误循环（已由渲染器处理）
- [x] 3.2 修改 `cmd/pull.go` 的 `runParallelPull`：同上适配
- [x] 3.3 修改 `cmd/interactive.go` 的 `interactiveClone` 并行分支：适配新的 `ProgressRenderer` API
- [x] 3.4 修改 `cmd/interactive.go` 的 `interactivePull` 并行分支：适配新的 `ProgressRenderer` API

## 4. 更新测试

- [x] 4.1 更新 `git/git_test.go` 中使用 `ProgressFunc` 的测试用例，适配新签名
- [x] 4.2 运行全部测试确认无回归

## 5. 验证

- [ ] 5.1 手动测试：TTY 环境下并行 clone，确认多行进度显示正确（显示正在处理的仓库名）
- [ ] 5.2 手动测试：非 TTY 环境下并行 clone，确认逐行输出 ✓/✗ 结果
- [ ] 5.3 手动测试：并行 pull，确认行为与 clone 一致
- [ ] 5.4 手动测试：顺序模式（`--concurrency 1`）输出保持不变
- [ ] 5.5 手动测试：交互模式中的 clone/pull 并行路径
