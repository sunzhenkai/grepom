## 1. TTY 检测增强

- [x] 1.1 在 `cmd/progress.go` 中重写 `isStdoutTerminal()` 函数，实现多层级检测：go-isatty → TERM 环境变量 → /proc/self/fd/1 inode 检查（Linux）
- [x] 1.2 在 verbose 模式下为 `isStdoutTerminal()` 各层检测添加日志输出，便于排查问题

## 2. CloneAll / PullAll 回调机制

- [x] 2.1 在 `git/parallel.go` 中定义 `ProgressFunc` 类型：`type ProgressFunc func(completed, total int)`
- [x] 2.2 修改 `CloneAll` 函数签名，增加 `onProgress ProgressFunc` 参数，在结果收集循环中每收到一个结果调用回调（nil 时跳过）
- [x] 2.3 修改 `PullAll` 函数签名，增加 `onProgress ProgressFunc` 参数，在结果收集循环中每收到一个结果调用回调（nil 时跳过）

## 3. 进度显示实时化

- [x] 3.1 修改 `cmd/clone.go` 中 `runParallelClone()`，将 `progress.Update` 放入 `CloneAll` 的回调中，实现逐个仓库递增进度
- [x] 3.2 修改 `cmd/pull.go` 中 `runParallelPull()`，将 `progress.Update` 放入 `PullAll` 的回调中，实现逐个仓库递增进度
- [x] 3.3 修改 `ProgressRenderer.Done()`，在清除进度行后输出 `\n`，确保后续摘要从新行开始

## 4. 验证与测试

- [x] 4.1 编译项目确认无编译错误
- [ ] 4.2 在 TTY 环境下手动测试 `grepom clone`，确认进度行逐个递增显示
- [ ] 4.3 在非 TTY 环境（管道）下手动测试 `grepom clone | cat`，确认逐行输出行为正常
- [ ] 4.4 在 Arch Linux + zsh 环境下测试，确认 TTY 检测正确且进度条可见
