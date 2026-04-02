## MODIFIED Requirements

### Requirement: 实时进度显示
clone 命令在并行模式下 SHALL 显示实时进度，每完成一个仓库克隆就更新进度行，让用户了解整体克隆进度。

#### Scenario: TTY 环境下的进度显示
- **WHEN** 在终端（TTY）环境中并行克隆 20 个仓库
- **THEN** 系统在单行显示实时更新的进度 `[3/20] cloning...`，每完成一个仓库更新一次，完成后切换为清除进度行并显示摘要

#### Scenario: 非 TTY 环境下的进度显示
- **WHEN** 在管道或重定向环境中并行克隆仓库（如 `grepom clone | tee log.txt`）
- **THEN** 系统逐行输出每个仓库完成时的进度（`[1/20] cloning...`、`[2/20] cloning...` 等），不使用回车覆盖

#### Scenario: 顺序模式下的进度显示
- **WHEN** `--concurrency 1` 时克隆仓库
- **THEN** 系统保持原有的逐行输出格式（`cloning xxx...` / `done`），不显示汇总进度行

#### Scenario: 已跳过的仓库计入总数
- **WHEN** 20 个仓库中有 5 个已克隆被跳过
- **THEN** 进度显示为 `[15/20]`，仅统计实际需要克隆的仓库

#### Scenario: 进度行清除后摘要独立显示
- **WHEN** TTY 环境下并行克隆完成
- **THEN** 进度行被清除后有一个换行，`clone complete: ...` 摘要从新行开始

### Requirement: Worker Pool 并发模型
并行克隆 SHALL 使用 worker pool 模式：主 goroutine 分发任务，N 个 worker goroutine 消费任务并返回结果。结果收集时 SHALL 支持可选的进度回调。

#### Scenario: 仓库数量小于并行度
- **WHEN** 有 2 个仓库需要克隆，`--concurrency` 为 4
- **THEN** 系统仅创建 2 个任务，2 个 worker 各处理一个任务，每完成一个触发进度回调

#### Scenario: 仓库数量大于并行度
- **WHEN** 有 20 个仓库需要克隆，`--concurrency` 为 4
- **THEN** 系统创建 4 个 worker，分批处理 20 个任务，每完成一个触发进度回调

#### Scenario: 单个仓库克隆失败不影响其他仓库
- **WHEN** 并行克隆中某个仓库的所有认证方式均失败
- **THEN** 系统记录该仓库的失败结果，触发进度回调，其他 worker 继续处理剩余仓库

#### Scenario: 无进度回调时行为不变
- **WHEN** `CloneAll` 的 `onProgress` 参数为 nil
- **THEN** 系统正常执行克隆，不触发回调，返回结果与之前完全一致
