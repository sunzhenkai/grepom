## Context

`grepom` 的并行 `clone` / `pull` 通过 `git.CloneAll` / `git.PullAll`（`git/parallel.go`）以 worker pool 执行。进度回调 `ProgressFunc` 在两处被调用：

- **`ProgressStart`**：由 N 个 worker goroutine 在取到任务时各自触发（`parallel.go` CloneAll L69-74、PullAll L169-174）。
- **`ProgressComplete`**：由唯一的结果收集 goroutine 在 `for r := range results` 循环中触发（L113-121 / L203-211）。

`cmd/progress.go` 的 `ProgressRenderer.Handle` 直接读写共享字段 `inflight`（切片 append/删除）、`completed`、`rendered`，并在 `renderTTY` 中向 stdout 输出 ANSI 光标控制序列（`\033[<n>A`、`\r`）。整个过程**无任何锁**。

由此产生两类缺陷：

1. **数据竞争 + 输出交织**：多个 worker 并发调用 `Handle(ProgressStart)` → 并发 append `inflight`、并发写 stdout，表现为进度行互相覆盖、`[N/M]` 计数倒退（如 `86 → 81 → 92`）、仓库名错位堆叠。`go test -race` 会直接告警。
2. **残留旧行未清除**：`renderTTY` 仅重写 `1 + len(inflight)` 行。当 in-flight 数量减少（任务完成多于新开始）时，`rendered` 缩小，但下方旧行未被空行覆盖，终端残留过期的仓库名。

目标是在不改变对外 CLI 行为、不新增依赖的前提下，让进度渲染在任何并发交错下保持正确与稳定。

## Goals / Non-Goals

**Goals:**
- 消除 `ProgressRenderer` 的数据竞争，保证 `Handle` / `Done` 在并发回调下线程安全。
- 修复 in-flight 行数减少时残留旧行的渲染缺陷，确保进度区域始终自洽。
- 保持非 TTY 模式逐行输出语义与 TTY 模式多行渲染语义不变。
- 通过可并发压测的单元测试固化行为，防止回归。

**Non-Goals:**
- 不重构 `git/parallel.go` 的事件分发模型（保持 Start 由 worker 触发、Complete 由收集器触发的现状）。
- 不引入第三方进度条库（保持零新增依赖）。
- 不改变 `ProgressEvent` 结构或 `ProgressFunc` 签名。
- 不调整 `--concurrency` 默认值或 CLI 参数。

## Decisions

### 1) 用 `sync.Mutex` 串行化渲染器状态
- 决策：为 `ProgressRenderer` 增加 `mu sync.Mutex`，在 `Handle` 与 `Done` 入口加锁，保护 `inflight` / `completed` / `rendered` 的读写以及整段 stdout 写入（光标移动 + 文本输出必须作为一个临界区，否则 ANSI 序列仍会交织）。
- 理由：锁粒度只需覆盖「状态变更 + 一次完整重绘」，渲染本身是 O(in-flight) 的轻量操作（通常 ≤ 并行度，几十行内），互斥不会成为瓶颈；git clone/pull 才是耗时主体，进度重绘频率远低于临界区成本。
- 备选方案：
  - 方案 A：在 `parallel.go` 用一个事件 channel 把所有 Start/Complete 事件汇聚到单 goroutine 再回调。改动面更大，且需修改 `CloneAll`/`PullAll` 内部结构，回归风险高。
  - 方案 B：用 `sync/atomic` 保护计数、用 `sync.Map` 管理 in-flight。但 `renderTTY` 的多行 ANSI 输出天然是非原子的，仍需一把锁来保证「光标上移 → 写各行」不被打断，故不如直接用一把互斥锁简单。

### 2) 重绘时按「上一次渲染行数」清除多余旧行
- 决策：`renderTTY` 先记住 `prev := p.rendered`，按当前 `lines` 重绘后，若 `lines < prev`，则继续用 `\n` + 空行（pad 到 `maxWidth`）覆盖剩余的 `prev - lines` 行，最后把光标移回第一行起点，再更新 `p.rendered = lines`。
- 理由：终端是覆盖式画布，行数缩减时必须显式擦除，否则历史内容会残留并和新内容叠加。
- 备选方案：
  - 方案 A：每次重绘固定画 `maxConcurrent` 行（预留空位）。会引入「最大并行度」概念且产生空行噪音。
  - 方案 B：用 `\033[J`（清屏至底部）替代逐行空格。可行但部分终端/管道行为不一致；逐行 pad 到 `maxWidth` 已是现有策略，保持一致更稳。

### 3) 锁内统一渲染，Start/Complete 共用同一路径
- 决策：`ProgressStart` 与 `ProgressComplete` 都在锁内调用 `renderTTY`（TTY 模式），不区分事件类型走不同输出分支。`p.completed` 仅由 Complete 事件更新，Start 事件不修改计数（保证计数单调不减）。
- 理由：消除「Start 用旧 completed 渲染导致计数回退」的可见缺陷——虽然加锁后已不会交织，但统一路径让语义更清晰，且计数只在 Complete 推进。
- 备选方案：
  - 方案 A：Start 事件不触发重绘，仅更新 in-flight 列表。会导致新任务开始时屏幕不及时刷新，体验下降。

## Risks / Trade-offs

- [风险] 锁粒度若覆盖到慢操作会拖慢进度刷新  
  → 仅锁「状态 + 重绘」，不锁 git 操作；重绘只做内存拼接与少量 IO，耗时可控。

- [风险] 并发压测在 CI 上可能因时序不稳定而 flaky  
  → 测试用 `sync.WaitGroup` + 高并发（如 50 worker × 200 任务）配合 `isTTY=false` 固定路径断言，并通过 `-race` 固化；TTY 分支用缓冲 stdout 捕获 ANSI 序列做确定性断言。

- [风险] 残留行清除依赖 `maxWidth=120` 的 pad 假设，超长仓库名可能截断残留  
  → 现状已是既有约束；本次不改变宽度策略，仅保证「已写过的行」被完整覆盖。超长名截断属另一议题，不在本次范围。

- [权衡] 不引入事件汇聚 channel，保留了 Start 由 worker 触发的设计  
  → 减少改动面与回归风险，互斥锁足以解决可见缺陷。
