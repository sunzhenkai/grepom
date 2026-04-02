## ADDED Requirements

### Requirement: 默认分支检测
pull 命令 SHALL 在执行 pull 前检测仓库是否处于默认分支（远程 HEAD 指向的分支）。

#### Scenario: 仓库在默认分支上
- **WHEN** 某仓库的当前分支为 `main`，且远程 `origin/HEAD` 指向 `origin/main`
- **THEN** 系统判定该仓库在默认分支上，允许执行 pull

#### Scenario: 仓库在非默认分支上
- **WHEN** 某仓库的当前分支为 `feature/login`，远程 `origin/HEAD` 指向 `origin/main`
- **THEN** 系统跳过该仓库的 pull，显示 `skip repo-name (on feature/login, not default branch)`

#### Scenario: 无法检测默认分支
- **WHEN** 某仓库的 `origin/HEAD` 未设置（`git symbolic-ref` 返回错误）
- **THEN** 系统跳过该仓库的 pull，显示 `skip repo-name (cannot detect default branch)`，不执行 pull

### Requirement: Clean 状态检查
pull 命令 SHALL 仅对 clean（无未提交更改）的仓库执行 pull。

#### Scenario: 仓库为 clean 状态
- **WHEN** 某仓库在默认分支上且 `git status` 显示 clean
- **THEN** 系统执行 `git pull`

#### Scenario: 仓库有未提交更改（dirty）
- **WHEN** 某仓库在默认分支上但有未提交的文件更改
- **THEN** 系统跳过该仓库的 pull，显示 `skip repo-name (dirty working tree)`

#### Scenario: 仓库有已暂存更改
- **WHEN** 某仓库在默认分支上且有已 `git add` 但未提交的更改
- **THEN** 系统跳过该仓库的 pull，显示 `skip repo-name (dirty working tree)`

#### Scenario: 仓库有 ahead 提交
- **WHEN** 某仓库在默认分支上，clean，但有 2 个本地未推送的提交（ahead 2）
- **THEN** 系统执行 `git pull`（ahead 不影响安全性，pull 可以 fast-forward 或 merge）

### Requirement: Smart Pull 并行执行
pull 命令 SHALL 支持 `--concurrency` 参数，与 clone 命令一致的并行模型。

#### Scenario: 默认并行 pull
- **WHEN** 用户运行 `grepom pull`（未指定 `--concurrency`）
- **THEN** 系统使用默认并行度 4 并发执行 pull

#### Scenario: 指定并行度 pull
- **WHEN** 用户运行 `grepom pull --concurrency 2`
- **THEN** 系统使用 2 个 worker 并发执行 pull

#### Scenario: 单仓库 pull
- **WHEN** 用户运行 `grepom pull web-app`
- **THEN** 系统仅对 `web-app` 执行 pull（含安全检查）

### Requirement: Pull 进度显示
pull 命令在并行模式下 SHALL 显示实时进度。

#### Scenario: TTY 环境下的 pull 进度
- **WHEN** 在终端环境中并行 pull 15 个仓库，其中 10 个满足条件
- **THEN** 系统显示进度 `[3/10] pulling...`，完成后显示 `[10/10] done`

#### Scenario: 跳过的仓库不计入进度总数
- **WHEN** 15 个仓库中 5 个被跳过（非默认分支或 dirty）
- **THEN** 进度显示基于实际需要 pull 的 10 个仓库，跳过信息在开始前统一展示

### Requirement: Pull 完成摘要
并行 pull 完成后，系统 SHALL 输出操作摘要，包含成功、失败和跳过的统计。

#### Scenario: 全部成功
- **WHEN** 所有满足条件的仓库 pull 成功
- **THEN** 系统输出 `pull complete: 10/10 succeeded, 5 skipped`

#### Scenario: 部分失败
- **WHEN** 10 个仓库中 8 个 pull 成功、2 个失败（如网络错误）
- **THEN** 系统输出 `pull complete: 8/10 succeeded, 2 failed, 5 skipped`

#### Scenario: 全部跳过
- **WHEN** 所有仓库均不满足 pull 条件
- **THEN** 系统输出 `nothing to pull: 15 skipped (not on default branch or dirty)`

### Requirement: --force 标志
pull 命令 SHALL 提供 `--force` 标志，跳过安全检查，对所有已克隆仓库执行 pull。

#### Scenario: 使用 --force pull dirty 仓库
- **WHEN** 用户运行 `grepom pull --force` 且某仓库有未提交更改
- **THEN** 系统对该仓库执行 `git pull`，如果失败则显示错误

#### Scenario: 使用 --force pull 非默认分支仓库
- **WHEN** 用户运行 `grepom pull --force` 且某仓库在功能分支上
- **THEN** 系统对该仓库执行 `git pull`（恢复原有行为）

#### Scenario: --force 与 --concurrency 可组合
- **WHEN** 用户运行 `grepom pull --force --concurrency 8`
- **THEN** 系统使用 8 个 worker 并发对所有已克隆仓库执行 pull，跳过安全检查
