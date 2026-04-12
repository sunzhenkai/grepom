## Context

grepom 的 clone/pull 命令在并行模式（默认 `--concurrency 4`）下使用 `ProgressRenderer` 显示进度。当前进度只显示 `[3/20] cloning...`，没有仓库信息，用户无法感知当前正在处理哪些仓库。

**当前架构**：
- `git/parallel.go`: `ProgressFunc` 签名为 `func(completed, total int)`，仅在任务完成时触发，不携带仓库名
- `cmd/progress.go`: `ProgressRenderer` 用 `\r` 覆盖单行显示计数，无仓库信息
- `cmd/clone.go` / `cmd/pull.go`: 创建 `ProgressRenderer` 并传入 `CloneAll`/`PullAll` 的回调
- `cmd/interactive.go`: 交互模式中有独立的并行 clone/pull 实现，同样使用 `ProgressRenderer`

**核心瓶颈**：`ProgressFunc` 只传 `(completed, total)`，没有仓库名；且只在完成时触发，无法展示正在进行的任务。

## Goals / Non-Goals

**Goals:**
- 并行模式下实时列出所有正在处理的仓库名
- TTY 环境下使用多行渲染，原地更新（不刷屏）
- 非 TTY 环境下逐行输出完成结果
- pull 命令同步改进
- 顺序模式保持现有行为不变

**Non-Goals:**
- 不引入第三方 TUI 库（继续使用原生 ANSI 转义码）
- 不显示单个仓库的克隆进度（如百分比、传输速度）
- 不修改摘要输出格式（`PrintCloneSummary`/`PrintPullSummary`）
- 不修改 sync/status/list/search 等其他命令

## Decisions

### 决策 1：扩展 `ProgressFunc` 为事件驱动模式

**选择**：新增 `ProgressEvent` 结构体，包含事件类型（Start/Complete）、仓库名、完成计数等。`ProgressFunc` 签名改为 `func(event ProgressEvent)`。

```go
type ProgressEventType int
const (
    ProgressStart    ProgressEventType = iota
    ProgressComplete
)

type ProgressEvent struct {
    Type      ProgressEventType
    RepoName  string
    Completed int
    Total     int
    Err       error  // 仅 Complete 事件，nil 表示成功
}

type ProgressFunc func(ProgressEvent)
```

**理由**：
- Start 事件让渲染器知道哪个仓库刚刚开始处理
- Complete 事件携带仓库名和错误信息
- 单一回调签名覆盖两种事件，简洁统一
- 结构体可扩展，未来可添加更多字段（如耗时）

**替代方案**：
- 两个独立回调 `OnStart`/`OnComplete` → 接口更复杂，且 `CloneAll`/`PullAll` 需要两个参数
- 只改 Complete 事件加仓库名 → 无法展示 in-flight 列表

### 决策 2：TTY 多行渲染使用 ANSI 光标上移

**选择**：在 TTY 环境下，进度显示占多行。第一行是 `[N/M] cloning...`，后续每行显示一个正在处理的仓库。使用 `\033[A`（光标上移）+ `\r` 实现原地更新。

渲染效果示例（`--concurrency 4`，共 20 个仓库）：
```
[3/20] cloning...
  cloning api...
  cloning web-app...
  cloning cli...
```

当 api 完成且 mobile 开始时：
```
[4/20] cloning...
  cloning web-app...
  cloning cli...
  cloning mobile...
```

当所有完成时，清除进度区域，输出摘要。

**理由**：
- 多行列表直观展示所有并行 worker 的状态
- ANSI 转义码广泛支持，无需第三方库
- 行数 = min(concurrency, 剩余任务数)，不会无限增长

**替代方案**：
- 单行紧凑格式 `[3/20] cloning: api, web-app, cli` → 仓库名过长时会截断或换行，可读性差
- 引入 bubbletea/lipgloss 等 TUI 库 → 引入新依赖，过重

### 决策 3：非 TTY 模式逐行输出完成结果

**选择**：非 TTY 模式下，每完成一个仓库输出一行 `✓ repo-name`（成功）或 `✗ repo-name: error`（失败），不使用 `\r` 覆盖。

**理由**：
- 管道/日志场景下每行独立，方便 grep 和分析
- 仓库名信息比纯计数更有用
- 与顺序模式的输出风格一致

### 决策 4：`Done()` 清除进度区域时使用光标上移

**选择**：`Done()` 先将光标上移到进度区域第一行，然后用空行覆盖所有进度行，最后输出摘要。

**理由**：
- 需要清除多行内容（不只是单行 `\r` 覆盖）
- 覆盖后摘要从新行开始，输出整洁

## Risks / Trade-offs

- **[ANSI 兼容性]** 某些非标准终端可能不支持 `\033[A` 光标上移 → 已有 TTY 检测机制，非 TTY 降级为逐行输出；实际风险极低，现代终端均支持
- **[渲染闪烁]** 高频更新多行内容可能在某些终端产生闪烁 → 更新频率由仓库完成速度决定，通常不会太快（秒级），实际体验应可接受
- **[回调签名变更]** `ProgressFunc` 签名变更是 **BREAKING**，但该类型仅在 `git` 包内部和 `cmd` 包使用，不对外暴露，影响可控
