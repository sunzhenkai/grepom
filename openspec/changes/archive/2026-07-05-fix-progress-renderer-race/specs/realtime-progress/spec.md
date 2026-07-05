## ADDED Requirements

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
