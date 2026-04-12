## Why

当前 clone/pull 命令在并行模式下只显示 `[3/20] cloning...` 这样的进度计数，用户无法知道当前正在处理哪些仓库。并行度越高，信息越不透明，用户面对空白等待不知道进度如何。pull 命令存在完全相同的问题。需要改进进度显示，让用户实时看到所有正在并行处理的仓库名。

## What Changes

- 扩展 `ProgressFunc` 回调，支持任务启动和完成两种事件，携带仓库名信息
- TTY 环境下改为多行渲染：显示 `[N/M] cloning...` + 每个正在处理的仓库各占一行，用 ANSI 转义码实现原地更新
- 非 TTY 环境下改为逐行输出每个仓库的完成结果（`✓ repo-name`）
- pull 命令同步应用相同的进度显示改进
- 交互模式中的 clone/pull 同步生效
- 顺序模式（`--concurrency 1`）保持现有行为不变

## Capabilities

### New Capabilities

（无新增能力）

### Modified Capabilities

- `realtime-progress`: 扩展进度回调支持 start/complete 事件，携带仓库名；改进渲染为多行 TTY 显示
- `parallel-clone`: 更新并行克隆的进度显示规范，要求显示当前正在处理的仓库名

## Impact

- `git/parallel.go`: `ProgressFunc` 签名变更，`CloneAll`/`PullAll` 在任务启动和完成时分别触发回调
- `cmd/progress.go`: `ProgressRenderer` 重构为多行 TTY 渲染，跟踪 in-flight 任务列表
- `cmd/clone.go`: `runParallelClone` 适配新的回调签名和渲染器
- `cmd/pull.go`: `runParallelPull` 适配新的回调签名和渲染器
- `cmd/interactive.go`: 交互模式中 clone/pull 的并行路径同步适配
- 现有测试中 `ProgressFunc` 的调用方式需要更新
