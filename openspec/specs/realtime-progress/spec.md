## ADDED Requirements

### Requirement: 实时进度回调
`CloneAll` 和 `PullAll` SHALL 接受 `ProgressFunc` 回调参数。回调签名 SHALL 为 `func(ProgressEvent)`，其中 `ProgressEvent` 包含事件类型（Start/Complete）、仓库名、已完成数、总数和错误信息。每开始处理一个仓库时 SHALL 触发 Start 事件，每完成一个仓库时 SHALL 触发 Complete 事件。

#### Scenario: 克隆时触发 Start 和 Complete 事件
- **WHEN** 并行克隆 20 个仓库，`onProgress` 回调非 nil
- **THEN** 系统在每个 worker 开始处理仓库时触发 `ProgressEvent{Type: ProgressStart, RepoName: "xxx", Completed: N, Total: 20}`，在每个仓库克隆完成时触发 `ProgressEvent{Type: ProgressComplete, RepoName: "xxx", Completed: N, Total: 20, Err: nil}`

#### Scenario: 拉取时触发 Start 和 Complete 事件
- **WHEN** 并行拉取 15 个仓库，`onProgress` 回调非 nil
- **THEN** 系统在每个 worker 开始处理仓库时触发 Start 事件，在每个仓库拉取完成时触发 Complete 事件

#### Scenario: 回调为 nil 时不影响行为
- **WHEN** `onProgress` 为 nil 时调用 `CloneAll`
- **THEN** 系统正常执行克隆，不触发任何回调，行为与修改前一致

#### Scenario: 克隆失败时 Complete 事件携带错误
- **WHEN** 某仓库克隆失败
- **THEN** 触发的 Complete 事件中 `Err` 字段非 nil，包含错误信息

### Requirement: 进度显示列出正在处理的仓库
`runParallelClone` 和 `runParallelPull` SHALL 在进度回调中更新 `ProgressRenderer`，使进度区域实时展示所有正在并行处理的仓库名。

#### Scenario: TTY 环境下多行显示正在克隆的仓库
- **WHEN** 在 TTY 环境下以 `--concurrency 4` 并行克隆 20 个仓库
- **THEN** 终端显示进度区域：第一行为 `[N/20] cloning...`，后续每行显示一个正在处理的仓库 `  cloning repo-name...`，最多显示 4 行（等于并行度）。当某个仓库完成后，其行被替换为新开始的仓库。使用 ANSI 光标上移实现原地更新。

#### Scenario: TTY 环境下多行显示正在拉取的仓库
- **WHEN** 在 TTY 环境下以 `--concurrency 4` 并行拉取 10 个仓库
- **THEN** 终端显示进度区域：第一行为 `[N/10] pulling...`，后续每行显示一个正在处理的仓库 `  pulling repo-name...`

#### Scenario: 非 TTY 环境下逐行输出完成结果
- **WHEN** 在非 TTY 环境下并行克隆 20 个仓库
- **THEN** 系统逐行输出每个仓库的完成结果：成功时输出 `✓ repo-name`，失败时输出 `✗ repo-name: error message`

#### Scenario: 非 TTY 环境下逐行输出拉取完成结果
- **WHEN** 在非 TTY 环境下并行拉取 10 个仓库
- **THEN** 系统逐行输出每个仓库的完成结果：成功时输出 `✓ repo-name`，失败时输出 `✗ repo-name: error message`

#### Scenario: 任务数少于并行度
- **WHEN** 有 2 个仓库需要克隆，`--concurrency` 为 4
- **THEN** 进度区域只显示 2 个正在处理的仓库行，不显示空行

#### Scenario: 接近完成时行数减少
- **WHEN** 并行克隆 20 个仓库，已完成 19 个，仅剩 1 个未完成
- **THEN** 进度区域仅显示 1 个正在处理的仓库行

### Requirement: 进度区域清除后换行
`ProgressRenderer.Done()` SHALL 清除进度区域的所有行（使用 ANSI 光标上移 + 空行覆盖），然后输出换行符，确保后续摘要输出从新行开始。

#### Scenario: 清除多行进度区域后摘要独立成行
- **WHEN** TTY 环境下并行克隆完成，`Done()` 被调用
- **THEN** 进度区域所有行被清除，`clone complete: ...` 摘要从新行开始显示

### Requirement: 进度渲染器并发安全
`ProgressRenderer` SHALL 在 `Handle` 与 `Done` 的全部状态读写与 stdout 输出上加互斥保护，使得当 `ProgressStart` 由多个 worker goroutine 并发触发、`ProgressComplete` 由结果收集 goroutine 触发时，渲染区域不会出现行交织、ANSI 转义序列错位或共享字段数据竞争。`completed` 计数 SHALL 仅由 `ProgressComplete` 事件推进，在单次批量操作内单调不减。

#### Scenario: 高并发下进度计数单调不减
- **WHEN** 以 `--concurrency 8` 并行拉取 200 个仓库，TTY 环境下 `ProgressStart` 与 `ProgressComplete` 事件并发到达
- **THEN** 渲染出的 `[N/200] pulling...` 摘要行中 `N` 在时间序列上单调不减，不出现 `86 → 81 → 92` 这类回退

#### Scenario: 并发事件不产生交织输出
- **WHEN** 多个 worker goroutine 几乎同时触发 `ProgressStart`，且同时有 `ProgressComplete` 到达
- **THEN** 每次重绘输出的 ANSI 光标控制序列（`\033[<n>A`、`\r`）与文本内容作为完整临界区写入，不与其他重绘的字节序列交错，`go test -race` 下无数据竞争告警

#### Scenario: 非 TTY 模式并发安全
- **WHEN** 非 TTY 环境下并发触发多个 `ProgressComplete` 事件
- **THEN** 每行完成结果（`✓ repo-name` 或 `✗ repo-name: error`）作为完整一行输出，不与其他行的字符交错

### Requirement: 进度区域行数缩减时清除残留旧行
当 in-flight 任务数减少导致本次需渲染的行数小于上一次渲染行数时，`renderTTY` SHALL 用空行（pad 至 `maxWidth`）覆盖多余的旧行，使终端不再显示已完成的仓库名或过期内容。

#### Scenario: 任务陆续完成时无残留仓库名
- **WHEN** TTY 环境下并行拉取 20 个仓库，已完成 18 个、仅剩 2 个 in-flight（上一次渲染了 5 行）
- **THEN** 当次重绘后进度区域仅包含 1 行摘要 + 2 行 in-flight 仓库，原本第 4、5 行的旧仓库名被空行覆盖，不残留可见内容

#### Scenario: 行数在增减间反复变化
- **WHEN** 并行度 4，in-flight 数量在 1↔4 之间多次反复增减直至全部完成
- **THEN** 任意时刻终端显示的 in-flight 行数与实际未完成任务数一致，旧内容不残留、不与新内容叠加
