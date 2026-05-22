## ADDED Requirements

### Requirement: 组内 URL 去重检测

系统 SHALL 检测指定 group 内 URL 规范化后相同的重复仓库条目。规范化使用 `NormalizeRepoURL` 函数。

#### Scenario: 检测到组内 URL 重复
- **WHEN** group "core-team" 中有两个 repo 条目的 URL 规范化后均为 `gitlab.com/my-org/api-lib`
- **THEN** 系统 SHALL 识别这两个条目为重复，保留第一个，标记后续条目为待删除

#### Scenario: 组内无重复
- **WHEN** group 中所有 repo 的 URL 规范化后均不相同
- **THEN** 系统 SHALL 报告该组无重复

### Requirement: 组内去重只删除多余条目不加 exclude_repos

系统 SHALL 对组内重复的 repo 只从 `repos` 列表中删除多余条目（保留第一个出现的），不将 repo name 追加到 `exclude_repos`。

#### Scenario: 删除重复条目不加 exclude_repos
- **WHEN** group "core-team" 中有两个 URL 指向同一仓库的条目
- **THEN** 系统 SHALL 从 repos 列表删除第二个条目，不修改 exclude_repos

#### Scenario: 多于两个重复条目
- **WHEN** group 中有三个 repo 条目的 URL 规范化后相同
- **THEN** 系统 SHALL 保留第一个，删除其余两个

### Requirement: 组内去重默认检查所有 group

当未指定 `--group` 时，系统 SHALL 对所有 group 执行组内 URL 去重检测。

#### Scenario: 无 --group 参数检查所有组
- **WHEN** 用户运行 `grepom dedup` 且配置中有 3 个 group
- **THEN** 系统 SHALL 对所有 3 个 group 分别执行组内去重检测

#### Scenario: 指定 --group 只检查指定组
- **WHEN** 用户运行 `grepom dedup --group core-team`
- **THEN** 系统 SHALL 只对 core-team 执行组内去重检测

### Requirement: 组内去重输出格式

系统 SHALL 输出每个 group 的组内去重结果，列出被删除的重复条目信息。

#### Scenario: 有重复的输出
- **WHEN** group "core-team" 中检测到 2 个重复条目
- **THEN** 系统 SHALL 在 "Intra-group dedup" 部分输出该组名、被删除条目的 name 和 URL，以及保留和删除的计数

#### Scenario: 无重复的输出
- **WHEN** group 中无重复
- **THEN** 系统 SHALL 输出该组 "no duplicates"

### Requirement: 组内去重写入使用文件锁

系统 SHALL 在 --apply 模式下写入 config 时使用 `WithFileLock` 机制，防止并发写入冲突。

#### Scenario: 并发 dedup 安全
- **WHEN** 两个 dedup 操作同时运行且都使用 --apply
- **THEN** 系统 SHALL 通过文件锁确保 config 写入不冲突