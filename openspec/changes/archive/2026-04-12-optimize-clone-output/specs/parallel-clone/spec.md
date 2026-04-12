## MODIFIED Requirements

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
