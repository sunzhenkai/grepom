## Why

并行克隆/拉取操作没有可感知的实时进度反馈。当前实现中 `CloneAll()` / `PullAll()` 阻塞直到所有任务完成后才调用一次 `Update(N)`，导致进度行仅在操作结束时闪现 `[N/N] cloning...` 随即被 `Done()` 清除——用户在整个克隆过程中看不到任何进度变化。

此外，TTY 检测使用 `go-isatty` 的 `IsTerminal()` 依赖 `syscall.IoctlGetTermios`，在 Arch Linux + zsh 环境下，当使用 kitty、alacritty 等终端模拟器，或 zsh 配置了某些插件（如 zsh-autosuggestions、powerlevel10k）时，stdout 的 fd 状态可能异常，导致 TTY 被误判为 non-TTY，进而退化为逐行输出模式（每行打印一次），失去了 `\r` 覆盖的进度效果。

## What Changes

- **将 `CloneAll` / `PullAll` 改为回调驱动**：接受 `OnResult func(result)` 回调，每完成一个仓库立即触发进度更新，实现真正的逐个递增进度显示
- **增强 TTY 检测健壮性**：在 `go-isatty` 基础上增加多层级回退检测（`TERM` 环境变量检查、`isatty` 命令验证、`/proc/self/fd/0` inode 检查），确保在 zsh + 各类终端模拟器下正确识别 TTY
- **优化进度行清除逻辑**：`Done()` 方法在清除进度行后确保输出换行，避免后续摘要输出被进度行残留覆盖
- **统一 clone 和 pull 的进度行为**：pull 操作存在相同的问题，一并修复

## Capabilities

### New Capabilities

- `realtime-progress`: 回调驱动的实时进度更新机制，替代当前的批量阻塞模式
- `robust-tty-detection`: 多策略回退的 TTY 检测，提升跨 shell 和终端模拟器兼容性

### Modified Capabilities

- `parallel-clone`: 修改 `CloneAll` 接口签名，增加回调参数以支持实时进度；修改进度显示为逐个仓库递增更新
