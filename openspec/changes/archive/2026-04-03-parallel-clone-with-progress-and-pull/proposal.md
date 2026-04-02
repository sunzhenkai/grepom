## Why

当前 grepom 的 clone 和 pull 操作均为顺序执行，对大量仓库（50+）操作时等待时间过长。用户需要：
1. **并行化克隆**以充分利用网络带宽，大幅缩短整体克隆时间
2. **实时进度条**直观感知操作进度（当前仅有静态文本日志，无法了解整体完成比例）
3. **智能 pull**：仅在仓库处于默认分支且无本地更改时才执行 pull，避免在功能分支上意外拉取导致冲突

## What Changes

- **并行克隆**：clone 命令支持并发克隆多个仓库，通过 `--concurrency` 参数控制并行度（默认 4），使用 worker pool 模式管理 goroutine
- **实时进度条**：clone 和 pull 操作过程中显示多行实时进度，包含仓库名、当前状态（克隆中/完成/失败）、进度计数 `[3/20]`
- **智能 pull**：pull 命令增强为仅对"已克隆 + 在默认分支 + clean"的仓库执行 `git pull`，跳过 dirty、ahead、在非默认分支的仓库
- **pull 命令**：当前 pull 命令直接对每个已克隆仓库执行 `git pull`，无分支检查和 clean 检查。增强后提供 `--force` 标志绕过安全检查
- **单仓库操作**：clone 和 pull 命令均支持指定单个仓库名称进行操作（已支持过滤，但需确保并行模式下的过滤行为一致）

## Capabilities

### New Capabilities
- `parallel-clone`: 并行克隆与进度条——支持并发克隆多个仓库，worker pool 管理 goroutine，实时多行进度显示
- `smart-pull`: 智能 pull 策略——检测默认分支和 clean 状态，安全地批量 pull 更新

### Modified Capabilities
- `cli-commands`: clone 和 pull 命令增加 `--concurrency` 和 `--force` 标志，调整输出格式为进度条模式
- `clone-auth-priority`: 认证尝试日志在并行模式下需适配进度条输出格式（不干扰进度条渲染）

## Impact

- **代码影响**：`cmd/clone.go`、`cmd/pull.go`、`git/git.go`（新增并行执行和进度条逻辑）
- **新增依赖**：可能引入进度条库（如 `charmbracelet/bubbles` 或轻量级自定义实现）
- **交互模式**：`cmd/interactive.go` 中的 clone/pull 逻辑需同步适配并行模式
- **测试影响**：`git/git_test.go` 需新增并行执行和进度条的测试用例
- **向后兼容**：默认行为保持顺序执行（`--concurrency 1`），进度条在非 TTY 环境自动降级为文本输出
