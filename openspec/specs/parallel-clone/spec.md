## ADDED Requirements

### Requirement: 并行克隆
clone 命令 SHALL 支持通过 `--concurrency` 参数指定并行克隆的仓库数量。当 `--concurrency` 大于 1 时，系统使用 worker pool 模式并发克隆多个仓库。

#### Scenario: 默认并行克隆
- **WHEN** 用户运行 `grepom clone`（未指定 `--concurrency`）
- **THEN** 系统使用默认并行度 4 并发克隆所有未克隆的仓库

#### Scenario: 指定并行度
- **WHEN** 用户运行 `grepom clone --concurrency 8`
- **THEN** 系统使用 8 个 worker 并发克隆仓库

#### Scenario: 顺序克隆（兼容模式）
- **WHEN** 用户运行 `grepom clone --concurrency 1`
- **THEN** 系统顺序逐个克隆仓库，行为与当前实现一致

#### Scenario: 单仓库克隆不受并发影响
- **WHEN** 用户运行 `grepom clone web-app`（指定单个仓库名称）
- **THEN** 系统仅克隆该仓库，并行度参数不影响行为

#### Scenario: 并行度参数校验
- **WHEN** 用户运行 `grepom clone --concurrency 0`
- **THEN** 系统报错提示并发度必须为正整数

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

### Requirement: Clone 日志收集
并行模式下，`git.Clone()` 的认证尝试日志 SHALL 被收集到结果结构中，而非直接输出到 stdout，完成后统一展示。

#### Scenario: 并行模式下克隆日志不交错
- **WHEN** 并行克隆多个仓库，每个仓库的克隆过程涉及多次认证尝试
- **THEN** 各仓库的认证尝试日志互不交错，完成后按仓库分组展示

#### Scenario: 克隆成功的仓库显示摘要
- **WHEN** 某仓库克隆成功
- **THEN** 进度行显示 `✓ repo-name`，详细日志仅在 verbose 模式下展示

#### Scenario: 克隆失败的仓库显示错误
- **WHEN** 某仓库克隆失败
- **THEN** 进度行显示 `✗ repo-name`，并附上错误摘要

### Requirement: 实时进度显示
clone 命令在并行模式下 SHALL 实时显示所有正在并行处理的仓库名，让用户了解当前进度。

#### Scenario: TTY 环境下的多行进度显示
- **WHEN** 在终端（TTY）环境中以 `--concurrency 4` 并行克隆 20 个仓库
- **THEN** 终端显示多行进度区域：第一行 `[N/20] cloning...`，后续每行显示一个正在克隆的仓库。每完成一个仓库或开始处理新仓库时，进度区域原地更新（使用 ANSI 转义码），完成后清除进度区域并显示摘要

#### Scenario: 非 TTY 环境下的逐行输出
- **WHEN** 在管道或重定向环境中并行克隆仓库（如 `grepom clone | tee log.txt`）
- **THEN** 系统逐行输出每个仓库的完成结果（`✓ repo-name` 或 `✗ repo-name: error`），不使用回车覆盖

#### Scenario: 顺序模式下的进度显示不变
- **WHEN** `--concurrency 1` 时克隆仓库
- **THEN** 系统保持原有的逐行输出格式（`cloning xxx...` / `done`），不显示汇总进度行

#### Scenario: 已跳过的仓库不计入总数
- **WHEN** 20 个仓库中有 5 个已克隆被跳过
- **THEN** 进度显示为 `[N/15]`，仅统计实际需要克隆的仓库

#### Scenario: 进度区域清除后摘要独立显示
- **WHEN** TTY 环境下并行克隆完成
- **THEN** 进度区域被清除后有一个换行，`clone complete: ...` 摘要从新行开始

### Requirement: 克隆完成摘要
并行克隆完成后，系统 SHALL 输出操作摘要。

#### Scenario: 全部成功
- **WHEN** 所有仓库克隆成功
- **THEN** 系统输出 `clone complete: 20/20 succeeded`

#### Scenario: 部分失败
- **WHEN** 20 个仓库中 18 个成功、2 个失败
- **THEN** 系统输出 `clone complete: 18/20 succeeded, 2 failed`，并列出失败的仓库名称和错误原因

#### Scenario: 全部跳过
- **WHEN** 所有仓库均已克隆
- **THEN** 系统输出 `all repositories already cloned`
