## ADDED Requirements

### Requirement: prune 命令扫描并清理已克隆的 excluded repos

系统 SHALL 提供 `grepom prune` 命令，扫描所有 `DisabledReason="excluded"` 的 repos，检查本地磁盘克隆状态，输出清理计划。

#### Scenario: 默认 dry-run 模式展示计划
- **WHEN** 用户运行 `grepom prune`
- **THEN** 系统 SHALL 列出所有已克隆的 excluded repos，标注每个的状态（clean/dirty/ahead/not cloned），并输出将要删除的 repos 列表，但不实际删除任何文件

#### Scenario: --apply 模式执行删除
- **WHEN** 用户运行 `grepom prune --apply`
- **THEN** 系统 SHALL 对所有通过安全检查的 excluded repos 执行 `rm -rf` 删除，并输出删除结果摘要

### Requirement: prune 安全检查跳过 dirty repos

系统 SHALL 在执行删除前检查 repo 的 git status。当 repo 存在未提交的改动（dirty）时 SHALL 跳过删除并输出警告。

#### Scenario: dirty repo 被跳过
- **WHEN** 一个 excluded repo 已克隆且 `git status` 显示有未提交的改动
- **THEN** 系统 SHALL 跳过该 repo 的删除，输出 "skipped: <name> (dirty, N files changed)"，继续处理其他 repos

### Requirement: prune 安全检查跳过 ahead repos

系统 SHALL 在执行删除前检查 repo 是否有本地领先的 commit。当 repo 本地领先远程时 SHALL 跳过删除并输出警告。

#### Scenario: ahead repo 被跳过
- **WHEN** 一个 excluded repo 已克隆且 `git status` 显示 ahead > 0
- **THEN** 系统 SHALL 跳过该 repo 的删除，输出 "skipped: <name> (ahead N commits)"，继续处理其他 repos

### Requirement: prune --force 跳过安全检查

系统 SHALL 支持 `--force` flag，跳过 dirty/ahead 安全检查，对所有已克隆的 excluded repos 执行删除。

#### Scenario: --force 删除 dirty repo
- **WHEN** 用户运行 `grepom prune --apply --force`
- **THEN** 系统 SHALL 不检查 dirty/ahead 状态，直接删除所有已克隆的 excluded repos

### Requirement: prune 按 group 和 resource 过滤

系统 SHALL 支持 `--group` 和 `--resource` flag，只处理匹配的 excluded repos。

#### Scenario: 按 group 过滤
- **WHEN** 用户运行 `grepom prune --group frontend`
- **THEN** 系统 SHALL 只扫描 frontend group 中的 excluded repos

#### Scenario: 按 resource 过滤
- **WHEN** 用户运行 `grepom prune --resource work-gl`
- **THEN** 系统 SHALL 只扫描使用 work-gl resource 的 excluded repos

### Requirement: prune 输出结果摘要

系统 SHALL 在执行完成后输出摘要，包含：已删除数量、跳过数量及原因、未克隆数量。

#### Scenario: 摘要输出
- **WHEN** prune 完成（dry-run 或 --apply）
- **THEN** 系统 SHALL 输出格式化的摘要，如 "prune: 3 deleted, 1 skipped (dirty), 2 not cloned"

### Requirement: prune 提示已克隆的 excluded repos

系统 SHALL 在完成时检测是否有 excluded repos 已被克隆，如有则提示用户运行 prune。

#### Scenario: 无 excluded repos 已克隆
- **WHEN** 所有 excluded repos 均未克隆到磁盘
- **THEN** 系统 SHALL 输出 "no excluded repos to prune"
