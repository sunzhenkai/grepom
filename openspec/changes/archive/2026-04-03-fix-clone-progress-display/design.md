## Context

grepom 是一个 Go CLI 工具，使用 cobra 框架管理多个 git 仓库。当前并行克隆和拉取使用 worker pool 模式（`git/parallel.go`），但进度显示存在两个问题：

1. **进度不实时**：`CloneAll()` 和 `PullAll()` 阻塞直到全部完成，返回后只调用一次 `Update(len(results))`，导致进度行闪现后立即被清除，用户无法感知任何进度变化
2. **TTY 检测不可靠**：仅依赖 `go-isatty.IsTerminal(os.Stdout.Fd())`，在 Arch Linux + zsh 环境下（kitty/alacritty 终端、powerlevel10k 等插件）可能误判为 non-TTY

当前代码路径：
- `cmd/clone.go` → `runParallelClone()` → `gitpkg.CloneAll()` → 阻塞 → `progress.Update(N)` → `progress.Done()`
- `cmd/pull.go` → `runParallelPull()` → `gitpkg.PullAll()` → 阻塞 → `progress.Update(N)` → `progress.Done()`

## Goals / Non-Goals

**Goals:**
- 每完成一个仓库克隆/拉取就立即更新进度行，让用户看到 `[3/20] cloning...` 的逐步递增
- 增强终端检测的鲁棒性，确保 zsh + 各类终端模拟器（kitty、alacritty、wezterm 等）下正确识别 TTY
- 保持 `CloneAll` / `PullAll` 的结果排序保证和错误隔离特性
- 保持 non-TTY 模式的逐行输出行为不变

**Non-Goals:**
- 不引入第三方进度条库（lipgloss/bubbles、schollz/progressbar 等），保持轻量自定义实现
- 不添加百分比、速率估算、ETA 等复杂进度信息
- 不修改认证链、并发模型等核心逻辑
- 不处理 interactive 模式下的进度显示（interactive 模式有自己的输出逻辑）

## Decisions

### 决策 1：使用回调函数实现实时进度

**选择**：为 `CloneAll` / `PullAll` 添加 `OnProgress func(completed, total int)` 回调参数

**替代方案**：
- **channel 模式**：通过 channel 流式发送结果 → 需要调用方管理 channel 消费，增加复杂度，且结果收集和进度更新耦合
- **Observer 模式**：定义 ProgressObserver 接口 → 过度设计，当前场景只需要一个回调

**理由**：回调是最简单直接的方案。在 `CloneAll` 的结果收集循环中，每收到一个结果就调用回调。调用方（`cmd/clone.go`）在回调中调用 `progress.Update(completed)` 即可。

**接口变更**：
```go
// 之前
func CloneAll(concurrency int, tasks []CloneTask) []CloneResult

// 之后
type ProgressFunc func(completed, total int)

func CloneAll(concurrency int, tasks []CloneTask, onProgress ProgressFunc) []CloneResult
```

### 决策 2：多策略 TTY 检测回退

**选择**：在 `go-isatty` 基础上增加环境变量 + inode 检查作为回退

**检测策略优先级**：
1. `go-isatty.IsTerminal()` — 基于 `ioctl` 的标准检测（最可靠）
2. `TERM` 环境变量检查 — 如果 `TERM` 非空且不等于 `dumb`/`unknown`，大概率是终端
3. `/proc/self/fd/1` inode 检查 — Linux 特有，通过 stat 检查 fd 1 指向的是否是 tty 设备（major=5）

**替代方案**：
- 使用 `golang.org/x/term` 的 `IsTerminal()` — 内部实现与 `go-isatty` 类似，无额外收益
- 执行外部 `tty` 命令 — 有 fork 开销，且在受限环境下可能不可用
- 仅依赖 `TERM` 环境变量 — 在 tmux/screen 等场景下 `TERM` 可能为 `screen-256color`，不够精确

**理由**：`go-isatty` 在绝大多数场景下工作正常。Arch Linux + zsh 下的问题可能是特定终端模拟器或插件对 stdout fd 的处理异常（如管道重定向、prompt hook 等）。增加回退检测可以在 `go-isatty` 失败时提供第二、第三层保障，且检测逻辑全部内聚在 `isStdoutTerminal()` 函数中，不引入外部依赖。

### 决策 3：进度行清除后确保换行

**选择**：在 `Done()` 清除进度行后，显式输出 `\n`

**理由**：当前 `Done()` 仅用空格覆盖进度行并 `\r` 回到行首。如果后续 `PrintCloneSummary` 输出不以换行开头，摘要会和进度行残留在同一行。Arch Linux + zsh 下某些终端模拟器（如 kitty 的 GPU 渲染管线）对 `\r` + 空格覆盖的渲染可能有细微差异，显式换行更安全。

## Risks / Trade-offs

- **[回调引入耦合]** `CloneAll` / `PullAll` 接口变更属于破坏性变更 → 缓解：这两个函数仅在 `cmd/` 包内调用，且调用点集中，影响范围可控。同时保留向后兼容：`onProgress` 为 nil 时不调用回调，行为与之前一致
- **[TTY 回退误判]** `TERM` 环境变量回退可能在某些 CI 环境（如设置了 `TERM=xterm-256color` 的 Docker 容器）中误判 → 缓解：`TERM` 检查仅作为第二层回退，且优先级低于 `go-isatty`；非 TTY 场景下进度退化为逐行输出，不会导致功能异常
- **[并发安全性]** 进度回调在 goroutine 中触发 → 缓解：回调内仅调用 `fmt.Fprintf(os.Stdout, ...)` 和更新计数器，Go 的 `os.Stdout.Write` 是并发安全的（内部有 mutex），且 `ProgressRenderer.completed` 由单一 goroutine（结果收集循环）更新，无需额外同步
