## ADDED Requirements

### Requirement: dedup 命令检测跨 group 同名 repo

系统 SHALL 提供 `grepom dedup --group <target>` 命令，扫描 target group 的 repos，与参考 group(s) 的 repos 按 name 精确匹配，检测同名冲突。

#### Scenario: 检测到同名冲突
- **WHEN** target group "core-team" 有 repo "api-lib"，reference group "infra-team" 也有 repo "api-lib"
- **THEN** 系统 SHALL 识别 "api-lib" 为同名冲突，并标记为待排除

#### Scenario: 无同名冲突
- **WHEN** target group 的所有 repo names 在 reference group 中均不存在
- **THEN** 系统 SHALL 输出 "no duplicates found"

### Requirement: dedup --group 为必需参数

系统 SHALL 要求 `--group` 参数，指定要去重的目标 group。

#### Scenario: 缺少 --group 参数
- **WHEN** 用户运行 `grepom dedup` 不带 `--group`
- **THEN** 系统 SHALL 返回错误 "required flag(s) \"group\" not set"

### Requirement: dedup --reference 为可选参数

系统 SHALL 支持可选的 `--reference` 参数（逗号分隔），指定参考 group(s)。不指定时 SHALL 使用所有其他 group 作为参考。

#### Scenario: 指定单个 reference group
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team`
- **THEN** 系统 SHALL 只与 infra-team 的 repos 对比

#### Scenario: 指定多个 reference groups（逗号分隔）
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team,legacy-team`
- **THEN** 系统 SHALL 与 infra-team 和 legacy-team 的 repos 取并集后对比

#### Scenario: 不指定 reference（对比所有其他 group）
- **WHEN** 用户运行 `grepom dedup --group core-team`
- **THEN** 系统 SHALL 将除 core-team 外的所有 group 的 repo names 取并集后对比

### Requirement: dedup 只修改 target group

系统 SHALL 只修改 `--group` 指定的 target group 的配置，reference group 不受任何影响。

#### Scenario: reference group 不被修改
- **WHEN** dedup 在 target group 中排除了与 reference group 同名的 repo
- **THEN** reference group 的 exclude_repos 和 repos 列表 SHALL 保持不变

### Requirement: dedup 从 target group 移除同名 repo 并追加 exclude_repos

系统 SHALL 对 target group 中与 reference 同名的 repos 执行两个操作：从 `repos` 列表移除该条目，并将 repo name 追加到 `exclude_repos`（如尚未包含）。

#### Scenario: 移除 repos 条目并追加 exclude_repos
- **WHEN** "api-lib" 在 target group 和 reference group 中都存在
- **THEN** 系统 SHALL 从 target group 的 `repos` 列表中移除 "api-lib" 条目，并将 "api-lib" 追加到 target group 的 `exclude_repos`

#### Scenario: exclude_repos 已包含该 name
- **WHEN** target group 的 `exclude_repos` 已包含 "api-lib"
- **THEN** 系统 SHALL 不重复追加，但仍然从 `repos` 列表中移除该条目

#### Scenario: exclude_repos 的 glob 模式已覆盖该 name
- **WHEN** target group 的 `exclude_repos` 包含 "api-*" 等能匹配 "api-lib" 的 glob 模式
- **THEN** 系统 SHALL 不追加 "api-lib" 到 exclude_repos，但仍然从 `repos` 列表中移除该条目

### Requirement: dedup 默认 dry-run 模式

系统 SHALL 默认以 dry-run 模式运行，只输出计划不修改 config。

#### Scenario: 默认 dry-run 输出计划
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team`
- **THEN** 系统 SHALL 输出去重计划（哪些 repos 将被排除），但不修改 config 文件
- **THEN** 输出末尾 SHALL 提示 "No changes written. Add --apply to execute."

### Requirement: dedup --apply 执行写入

系统 SHALL 支持 `--apply` flag，执行实际的 config 修改。

#### Scenario: --apply 写入 config
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team --apply`
- **THEN** 系统 SHALL 修改 config 文件：更新 target group 的 exclude_repos 和 repos 列表
- **THEN** 完成后 SHALL 提示用户运行 `grepom prune` 清理磁盘上已克隆的 repos

### Requirement: dedup 输出去重计划

系统 SHALL 输出格式化的去重计划，列出每个冲突 repo 的处理结果。

#### Scenario: 计划输出格式
- **WHEN** dedup 检测到 2 个同名冲突
- **THEN** 系统 SHALL 逐行输出每个冲突的处理，如：
  "api-lib → exclude (exists in infra-team)"
  "worker → exclude (exists in infra-team)"
  以及摘要 "2 repos excluded, 3 kept"

### Requirement: dedup 写入使用文件锁

系统 SHALL 在写入 config 时使用 `WithFileLock` 机制，防止并发写入冲突。

#### Scenario: 并发 dedup 安全
- **WHEN** 两个 dedup 操作同时运行
- **THEN** 系统 SHALL 通过文件锁确保 config 写入不冲突
