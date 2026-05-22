## ADDED Requirements

### Requirement: 跨组 URL 重复检测

系统 SHALL 检测不同 group 之间 URL 规范化后相同的仓库，输出警告信息。规范化使用 `NormalizeRepoURL` 函数。

#### Scenario: 检测到跨组 URL 重复
- **WHEN** group "infra-team" 有 repo "api-lib" URL 为 `https://gitlab.com/my-org/api-lib.git`，group "core-team" 也有 repo URL 规范化后为 `gitlab.com/my-org/api-lib`
- **THEN** 系统 SHALL 输出警告，列出该 URL 出现在哪些 group 中

#### Scenario: 无跨组重复
- **WHEN** 所有 group 之间的 URL 规范化后均不相同
- **THEN** 系统 SHALL 不输出跨组警告

### Requirement: 跨组 URL 警告只打印不删除

系统 SHALL 对跨组 URL 重复只输出警告信息，不删除任何 repo 条目，不修改 exclude_repos。

#### Scenario: 跨组重复不修改配置
- **WHEN** 不同 group 中存在 URL 规范化后相同的 repo
- **THEN** 系统 SHALL 不从任何 group 中删除 repo，不修改 exclude_repos

### Requirement: 跨组 URL 警告不影响退出码

系统 SHALL 在存在跨组 URL 重复时仍返回退出码 0。

#### Scenario: 跨组重复退出码为 0
- **WHEN** 检测到跨组 URL 重复并输出警告
- **THEN** 命令退出码 SHALL 为 0

### Requirement: 跨组警告默认检查所有 group 之间

当未指定 `--group` 时，系统 SHALL 检查所有 group 之间的 URL 重复。指定 `--group` 时 SHALL 只检查该 group 与其他 group 之间。

#### Scenario: 无 --group 检查所有组之间
- **WHEN** 用户运行 `grepom dedup` 且配置中有 3 个 group
- **THEN** 系统 SHALL 检查所有 3 个 group 两两之间的 URL 重复

#### Scenario: 指定 --group 只检查该组与其他组之间
- **WHEN** 用户运行 `grepom dedup --group core-team`
- **THEN** 系统 SHALL 只检查 core-team 与其他 group 之间的 URL 重复

### Requirement: 跨组 URL 警告输出格式

系统 SHALL 输出跨组警告信息，列出重复的 URL 和涉及的 group 名称。

#### Scenario: 警告输出格式
- **WHEN** URL `gitlab.com/my-org/api-lib` 出现在 infra-team 和 core-team 两个 group 中
- **THEN** 系统 SHALL 在 "Cross-group URL warnings" 部分输出警告，格式包含 ⚠️ 标记、规范化后的 URL（或原始 URL）、涉及的 group 名称列表
