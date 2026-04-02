## Context

grepom 是一个 Go CLI 工具，用于批量管理 GitLab/GitHub 上的多个仓库。当前 clone 和 pull 命令均为顺序执行（`for` 循环逐个处理），每次仅操作一个仓库。对于管理 50+ 仓库的用户，全量克隆可能需要数十分钟。

现有架构：
- `cmd/clone.go`：顺序遍历 `[]provider.Repo`，逐个调用 `git.Clone()`
- `cmd/pull.go`：顺序遍历，对每个已克隆仓库调用 `git.Pull()`
- `git/git.go`：底层封装 `exec.Command("git", ...)`，`Clone()` 内含 5 级认证优先级链，每级尝试时直接 `fmt.Printf` 输出日志
- `git/GetStatus()`：解析 `git status --porcelain=v2 --branch`，返回 `Status{Branch, Clean, Ahead, Behind, Dirty, ...}`

关键约束：
- `Clone()` 函数当前直接输出日志到 stdout，在并行场景下会导致输出交错
- `Pull()` 当前无分支检查，直接执行 `git pull`，可能在功能分支上意外拉取
- 项目已有 `mattn/go-isatty` 依赖，可检测 TTY 环境

## Goals / Non-Goals

**Goals:**
- clone 和 pull 操作支持并行执行，显著减少大批量仓库的操作时间
- 提供实时进度反馈，用户可直观感知整体进度和每个仓库状态
- pull 命令具备安全检查：仅在默认分支 + clean 时执行
- 支持 `--concurrency` 参数控制并行度
- 单仓库操作（通过名称参数指定）行为不变

**Non-Goals:**
- 不实现断点续传或增量克隆（如 `--depth`、`--filter`）
- 不实现 clone/pull 的速度限制或带宽控制
- 不实现交互式 TUI 进度视图（保持 CLI 风格）
- 不修改 provider API 发现逻辑（sync 命令）
- 不实现全局的并行配置（每次命令行指定即可）

## Decisions

### D1: 自定义进度条实现 vs 第三方库

**选择**：自定义轻量级多行进度显示，不引入第三方进度条库。

**理由**：
- 需求简单：仅需显示 `[完成数/总数]` + 每行仓库名和状态标记
- 不需要百分比、速率估算、spinner 等复杂功能
- 引入 `charmbracelet/bubbles` 等库会增加较多依赖和复杂度
- 自定义实现约 100-150 行代码，完全可控

**替代方案**：
- `charmbracelet/lipgloss` + `bubbles`：功能强大但过重
- `schollz/progressbar`：单行进度条，不适合多仓库场景
- 纯 `fmt.Printf` 计数器：非 TTY 环境下的降级方案

### D2: Worker Pool 模式

**选择**：使用带缓冲 channel 的 worker pool 模式，worker 数量等于 `--concurrency`。

**理由**：
- 标准 Go 并发模式，易于理解和测试
- 通过 channel 传递任务和结果，避免共享状态
- 可精确控制并发度，不会因仓库数量增长而创建过多 goroutine

**实现**：
```
jobs <-chan Repo     (缓冲 N 个任务)
results <-chan Result (缓冲 N 个结果)
N 个 worker goroutine 消费 jobs，产出 results
主 goroutine 收集 results 并更新进度显示
```

### D3: 进度显示策略

**选择**：TTY 环境（终端）使用 `\r` 回车符覆盖更新同一行显示汇总进度；非 TTY 环境（管道、重定向）降级为逐行文本输出。

**理由**：
- `\r` 覆盖更新是最简单且兼容性最好的方案
- 多仓库场景下，并行输出每个仓库的详细日志会交错不可读，因此并行模式下只显示汇总行（如 `[3/20] cloning...`），详细日志在完成后或失败时追加显示
- 非并行模式（`--concurrency 1`）保持原有逐行输出不变

**TTY 检测**：复用 `mattn/go-isatty`（已在依赖中）

### D4: Clone 日志解耦

**选择**：为 `git.Clone()` 添加可选的 `io.Writer` 参数用于日志输出，并行模式下将日志收集到结果中而非直接输出。

**理由**：
- 当前 `Clone()` 直接 `fmt.Printf`，并行调用会导致输出交错
- 通过 `CloneOptions` 增加 `LogWriter io.Writer` 字段，默认为 `os.Stdout`（保持向后兼容）
- 并行模式下传入 `&bytes.Buffer{}`，将每个仓库的克隆日志收集到 `Result` 结构中，完成后统一输出

### D5: 智能 Pull 的默认分支检测

**选择**：通过 `git symbolic-ref refs/remotes/origin/HEAD` 获取远程默认分支名，与 `git status --porcelain=v2 --branch` 中的 `branch.head` 比较。

**理由**：
- `symbolic-ref` 是获取默认分支的标准方式
- 复用已有的 `GetStatus()` 解析逻辑，仅需新增 `DefaultBranch` 字段
- 比 `git remote show origin` 快得多，无需网络请求

### D6: --force 标志

**选择**：pull 命令新增 `--force` 标志，跳过安全检查，对所有已克隆仓库执行 pull（恢复原有行为）。

**理由**：
- 保持向后兼容：不加 `--force` 时启用安全检查
- 用户明确需要强制 pull 时有逃生路径

## Risks / Trade-offs

- **[并行 clone 的目录创建竞争]** → 每个 worker 克隆前调用 `os.MkdirAll`，该操作是幂等的，无竞争问题。但多个 worker 同时 clone 同一目标路径理论上不会发生（repo 列表去重），无需额外加锁。
- **[进度条在 CI/CD 环境下的兼容性]** → 通过 `isatty` 检测自动降级为纯文本输出，确保管道和日志收集工具正常工作。
- **[大量 goroutine 的系统资源]** → `--concurrency` 默认值为 4，用户可手动调高。不设硬上限，信任用户判断。
- **[默认分支检测可能失败]** → 对于未设置 `origin/HEAD` 的仓库（如手动添加的 remote），降级为仅检查 clean 状态，不因检测失败而阻塞 pull。
