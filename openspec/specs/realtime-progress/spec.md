## ADDED Requirements

### Requirement: 实时进度回调
`CloneAll` 和 `PullAll` SHALL 接受可选的 `ProgressFunc` 回调参数。当回调非 nil 时，每完成一个仓库操作 SHALL 立即调用回调，传入当前已完成数和总任务数。

#### Scenario: 克隆时逐个更新进度
- **WHEN** 并行克隆 20 个仓库，`onProgress` 回调非 nil
- **THEN** 每完成一个仓库克隆后，系统立即调用 `onProgress(completed, 20)`，`completed` 从 1 递增到 20

#### Scenario: 拉取时逐个更新进度
- **WHEN** 并行拉取 15 个仓库，`onProgress` 回调非 nil
- **THEN** 每完成一个仓库拉取后，系统立即调用 `onProgress(completed, 15)`，`completed` 从 1 递增到 15

#### Scenario: 回调为 nil 时不影响行为
- **WHEN** `onProgress` 为 nil 时调用 `CloneAll`
- **THEN** 系统正常执行克隆，不调用任何回调，行为与修改前一致

#### Scenario: 进度回调在结果收集时触发
- **WHEN** worker 完成 N 个仓库后，主 goroutine 从 results channel 收集到结果
- **THEN** 系统在收集到每个结果后立即触发回调，不等待所有任务完成

### Requirement: 进度显示逐个递增
`runParallelClone` 和 `runParallelPull` SHALL 在进度回调中调用 `ProgressRenderer.Update()`，使进度行实时递增显示。

#### Scenario: TTY 环境下进度行实时覆盖更新
- **WHEN** 在 TTY 环境下并行克隆 20 个仓库
- **THEN** 终端上依次显示 `[1/20] cloning...`、`[2/20] cloning...`、...、`[20/20] cloning...`，每完成一个仓库进度行更新一次，使用 `\r` 覆盖

#### Scenario: 非 TTY 环境下逐行输出
- **WHEN** 在非 TTY 环境下并行克隆 20 个仓库
- **THEN** 系统逐行输出 `[1/20] cloning...`、`[2/20] cloning...` 等，每个完成事件占一行

### Requirement: 进度行清除后换行
`ProgressRenderer.Done()` 在清除进度行后 SHALL 输出一个换行符，确保后续摘要输出从新行开始。

#### Scenario: 清除进度行后摘要独立成行
- **WHEN** TTY 环境下并行克隆完成，`Done()` 被调用
- **THEN** 进度行被清除后输出一个空行，`clone complete: ...` 摘要从新行开始显示
