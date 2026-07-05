## Why

并行 `clone` / `pull` 时，TTY 进度区域出现严重错位、行重叠、完成计数倒退（如 `86 → 81 → 92`）的乱序输出。根因是 `ProgressRenderer` 没有任何并发保护：`ProgressStart` 事件由 N 个 worker goroutine 并发触发，`ProgressComplete` 由结果收集 goroutine 触发，二者同时读写共享状态（`inflight` 切片、`rendered` 行数、`completed`）并并发向 stdout 写 ANSI 转义序列，导致数据竞争与渲染交织。这是日常高频操作（拉取 629 个仓库）的核心体验缺陷。

## What Changes

- 为 `ProgressRenderer` 增加互斥锁，串行化 `Handle` 与 `Done` 对共享状态的读写及 stdout 写入。
- 修复 `renderTTY` 在 in-flight 行数减少时未清除残留旧行的问题：当新渲染行数小于上次渲染行数时，用空行覆盖多余的旧行。
- 统一 `ProgressStart`/`ProgressComplete` 的渲染路径，确保计数与 in-flight 列表在任何并发交错下保持自洽。
- 同步更新 README（含多语言版本）中并行进度展示的说明，明确其并发安全行为。

## Capabilities

### New Capabilities
<!-- 本次不新增独立能力，属于对现有能力的缺陷修复。 -->

### Modified Capabilities
- `realtime-progress`: 明确 `ProgressRenderer` 在并发事件回调下的线程安全契约，并修复 in-flight 行数减少时残留旧行未清除的渲染缺陷。

## Impact

- 受影响代码：`cmd/progress.go`（核心修复）、`cmd/progress_test.go`（新增并发测试，如不存在则新建）。
- 受影响文档：`openspec/specs/realtime-progress/spec.md`（补充并发安全与残留行清除场景）、`README.md` / `README_en.md`（进度行为说明）。
- 外部影响：无 breaking change；CLI 输出更稳定，无新增参数或配置字段。
