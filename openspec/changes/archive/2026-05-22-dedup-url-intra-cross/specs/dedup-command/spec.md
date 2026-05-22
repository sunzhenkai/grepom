## MODIFIED Requirements

### Requirement: dedup 命令检测跨 group 同名 repo

系统 SHALL 提供 `grepom dedup [--group <target>] [--reference <refs>]` 命令。当同时指定 `--group` 和 `--reference` 时，扫描 target group 的 repos，与参考 group(s) 的 repos 按 name 精确匹配，检测同名冲突。

#### Scenario: 检测到同名冲突
- **WHEN** 指定了 `--group core-team --reference infra-team`，且 core-team 有 repo "api-lib"，infra-team 也有 repo "api-lib"
- **THEN** 系统 SHALL 识别 "api-lib" 为同名冲突，并标记为待排除

#### Scenario: 无同名冲突
- **WHEN** 指定了 `--group` 和 `--reference`，且 target group 的所有 repo names 在 reference group 中均不存在
- **THEN** 系统 SHALL 输出 "no duplicates found"

### Requirement: dedup --group 为可选参数

系统 SHALL 将 `--group` 设为可选参数。不指定时 SHALL 对所有 group 执行组内去重和跨组 URL 警告。指定 `--group` 时 SHALL 只处理指定 group。

#### Scenario: 不指定 --group
- **WHEN** 用户运行 `grepom dedup` 不带 `--group`
- **THEN** 系统 SHALL 对所有 group 执行组内 URL 去重和跨组 URL 警告

#### Scenario: 指定 --group
- **WHEN** 用户运行 `grepom dedup --group core-team`
- **THEN** 系统 SHALL 只对 core-team 执行组内去重，只检查 core-team 与其他 group 的跨组 URL 重复

### Requirement: dedup --reference 为可选参数

系统 SHALL 支持可选的 `--reference` 参数（逗号分隔），指定参考 group(s)。不指定时不触发按 name 跨组排除逻辑。指定时与 `--group` 配合触发原有跨组 name 去重。

#### Scenario: 不指定 --reference（不触发跨组 name 去重）
- **WHEN** 用户运行 `grepom dedup --group core-team` 不带 `--reference`
- **THEN** 系统 SHALL 不执行按 name 的跨组排除逻辑

#### Scenario: 指定单个 reference group
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team`
- **THEN** 系统 SHALL 执行组内去重和跨组警告后，额外执行与 infra-team 的按 name 跨组排除

#### Scenario: 指定多个 reference groups（逗号分隔）
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team,legacy-team`
- **THEN** 系统 SHALL 与 infra-team 和 legacy-team 的 repos 取并集后对比（跨组 name 去重部分）

### Requirement: dedup 命令执行流程

系统 SHALL 按以下顺序执行 dedup 命令：

1. **Step 1**: 组内 URL 去重 — 始终执行，对指定或所有 group 检测并清理组内 URL 重复
2. **Step 2**: 跨组 URL 警告 — 始终执行，检测不同 group 之间的 URL 重复并输出警告
3. **Step 3**: 跨组 name 去重 — 仅当同时指定 `--group` + `--reference` 时执行，行为与原逻辑相同

#### Scenario: 无参数运行
- **WHEN** 用户运行 `grepom dedup`
- **THEN** 系统 SHALL 执行 Step 1（所有 group 组内去重）和 Step 2（跨组 URL 警告），不执行 Step 3

#### Scenario: 仅指定 --group
- **WHEN** 用户运行 `grepom dedup --group core-team`
- **THEN** 系统 SHALL 执行 Step 1（core-team 组内去重）和 Step 2（core-team 与其他组跨组警告），不执行 Step 3

#### Scenario: 同时指定 --group 和 --reference
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team`
- **THEN** 系统 SHALL 依次执行 Step 1、Step 2、Step 3

### Requirement: dedup 从 target group 移除同名 repo 并追加 exclude_repos

系统 SHALL 对 Step 3 中 target group 与 reference 同名的 repos 执行两个操作：从 `repos` 列表移除该条目，并将 repo name 追加到 `exclude_repos`（如尚未包含）。此行为与原逻辑完全一致。

#### Scenario: 移除 repos 条目并追加 exclude_repos
- **WHEN** "api-lib" 在 target group 和 reference group 中都存在，且 Step 3 被触发
- **THEN** 系统 SHALL 从 target group 的 `repos` 列表中移除 "api-lib" 条目，并将 "api-lib" 追加到 target group 的 `exclude_repos`

#### Scenario: exclude_repos 已包含该 name
- **WHEN** target group 的 `exclude_repos` 已包含 "api-lib"，且 Step 3 被触发
- **THEN** 系统 SHALL 不重复追加，但仍然从 `repos` 列表中移除该条目

#### Scenario: exclude_repos 的 glob 模式已覆盖该 name
- **WHEN** target group 的 `exclude_repos` 包含 "api-*" 等能匹配 "api-lib" 的 glob 模式，且 Step 3 被触发
- **THEN** 系统 SHALL 不追加 "api-lib" 到 exclude_repos，但仍然从 `repos` 列表中移除该条目

### Requirement: dedup 默认 dry-run 模式

系统 SHALL 默认以 dry-run 模式运行，只输出计划不修改 config。

#### Scenario: 默认 dry-run 输出计划
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team`
- **THEN** 系统 SHALL 输出去重计划（组内重复、跨组警告、跨组排除），但不修改 config 文件
- **THEN** 输出末尾 SHALL 提示 "No changes written. Add --apply to execute."

### Requirement: dedup --apply 执行写入

系统 SHALL 支持 `--apply` flag，执行实际的 config 修改。

#### Scenario: --apply 写入 config
- **WHEN** 用户运行 `grepom dedup --group core-team --reference infra-team --apply`
- **THEN** 系统 SHALL 修改 config 文件：执行组内去重（删除重复 repos 条目）、执行跨组 name 去重（更新 target group 的 exclude_repos 和 repos 列表）

### Requirement: dedup 输出去重计划

系统 SHALL 输出格式化的去重结果，分三个部分。

#### Scenario: 完整输出格式
- **WHEN** dedup 检测到组内重复、跨组 URL 重复、跨组 name 冲突
- **THEN** 系统 SHALL 分三部分输出：
  - "Intra-group dedup"：每组列出删除的重复条目和计数
  - "Cross-group URL warnings"：每个跨组重复 URL 列出涉及的 group
  - "Cross-group name dedup"（仅 Step 3 触发时）：每个同名冲突列出处理结果
