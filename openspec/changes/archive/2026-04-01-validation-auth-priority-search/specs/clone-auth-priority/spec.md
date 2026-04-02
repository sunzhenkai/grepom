## MODIFIED Requirements

### Requirement: 克隆认证优先级链
系统 SHALL 按以下优先级依次尝试克隆，前一种方式成功即停止：
1. group/repo 级别 SSH key（SSH + 指定 key）
2. group/repo 级别 token（HTTPS + token URL）
3. resource 级别 SSH key（SSH + 指定 key）
4. resource 级别 token（HTTPS + token URL）
5. 推导的 SSH URL（系统默认 SSH）
6. 裸 HTTP URL

未配置的级别 SHALL 被跳过，不产生延迟。

#### Scenario: group/repo 级别 SSH key 最优先
- **WHEN** group 或 repo 配置了 ssh_key，同时也有 token 和 resource 认证
- **THEN** 系统优先使用 group/repo 级别的 SSH key 进行 clone

#### Scenario: group/repo 级别 token 作为二级回退
- **WHEN** group 或 repo 配置了 token 但未配置 ssh_key，同时 resource 也有 ssh_key 和 token
- **THEN** 系统优先尝试 group/repo 级别的 token 认证；若失败后使用 resource 的 SSH key

#### Scenario: resource SSH key 优先于 resource token
- **WHEN** group/repo 未配置任何认证，resource 同时配置了 ssh_key 和 token
- **THEN** 系统先尝试 resource 的 SSH key 认证，若失败后使用 resource 的 token

#### Scenario: resource SSH key 作为回退
- **WHEN** group/repo 未配置 ssh_key，但 resource 配置了 ssh_key 和 token，且 token 认证失败
- **THEN** 系统使用 resource 的 SSH key 尝试 SSH clone

#### Scenario: 所有方式均失败
- **WHEN** 所有认证方式均 clone 失败
- **THEN** 系统报告错误，不保留失败的目录

#### Scenario: 仅推导 SSH 可用
- **WHEN** 无任何 token 和 ssh_key 配置，但有 SSH URL
- **THEN** 系统使用推导的 SSH URL 进行 clone

#### Scenario: 仅 HTTP 可用
- **WHEN** 无任何 SSH 和 token 配置
- **THEN** 系统使用裸 HTTP URL 进行 clone

### Requirement: Token 认证克隆
系统 SHALL 支持使用配置的 token 进行 HTTPS 认证克隆。认证 URL 构建规则：
- GitHub: `https://x-access-token:<token>@<host>/<path>.git`
- GitLab: `https://oauth2:<token>@<host>/<path>.git`

Token 认证在 SSH key 认证失败后使用。

#### Scenario: GitHub 仓库使用 token 克隆
- **WHEN** clone 一个关联 GitHub resource（含 token）的仓库，且 SSH key 认证失败
- **THEN** 系统构建 `https://x-access-token:<token>@github.com/<org>/<repo>.git` URL 并用于 git clone

#### Scenario: GitLab 仓库使用 token 克隆
- **WHEN** clone 一个关联 GitLab resource（含 token）的仓库，且 SSH key 认证失败
- **THEN** 系统构建 `https://oauth2:<token>@gitlab.com/<org>/<repo>.git` URL 并用于 git clone

#### Scenario: Token 为空时跳过 token 认证
- **WHEN** 最终使用的 token（group/repo 级别或 resource 级别）为空字符串
- **THEN** 系统跳过 token 认证方式，直接尝试下一种认证方式

### Requirement: SSH Key 认证克隆
系统 SHALL 支持使用指定的 SSH 密钥文件进行克隆，通过 `GIT_SSH_COMMAND` 环境变量传递密钥路径。SSH key 认证优先于 token 认证。

#### Scenario: 使用指定 SSH key 克隆
- **WHEN** clone 一个配置了 `ssh_key: ~/.ssh/deploy_key` 的 repo（group/repo 级别或 resource 级别）
- **THEN** 系统通过设置 `GIT_SSH_COMMAND=ssh -i ~/.ssh/deploy_key -o IdentitiesOnly=yes` 环境变量执行 git clone

#### Scenario: SSH key 路径展开 ~
- **WHEN** 配置的 `ssh_key` 值为 `~/.ssh/id_ed25519`
- **THEN** 系统将 `~` 展开为用户 home 目录后再传递给 SSH 命令

### Requirement: 认证尝试日志
系统 SHALL 在 clone 过程中尝试每种认证方式时输出日志，让用户了解当前进度和认证回退过程。

#### Scenario: 尝试 SSH key 认证日志
- **WHEN** 系统尝试使用指定 SSH key 进行 clone
- **THEN** 系统输出日志 `  [N/M] 尝试 SSH key 认证 (<级别>)...`

#### Scenario: 尝试 token 认证日志
- **WHEN** 系统尝试使用 token 进行 clone
- **THEN** 系统输出日志 `  [N/M] 尝试 token 认证 (<级别>)...`，其中级别为 "group/repo" 或 "resource"

#### Scenario: 尝试默认 SSH 认证日志
- **WHEN** 系统尝试使用推导的 SSH URL 进行 clone
- **THEN** 系统输出日志 `  [N/M] 尝试 SSH 认证 (默认)...`

#### Scenario: 尝试 HTTP 克隆日志
- **WHEN** 系统尝试使用裸 HTTP URL 进行 clone
- **THEN** 系统输出日志 `  [N/M] 尝试 HTTP 克隆...`

#### Scenario: 认证失败日志
- **WHEN** 某种认证方式 clone 失败
- **THEN** 系统输出错误摘要（不含敏感信息），如 `  [N/M] 失败: <错误摘要>`

#### Scenario: 认证成功日志
- **WHEN** 某种认证方式 clone 成功
- **THEN** 系统输出 `  [N/M] 成功`

#### Scenario: 跳过未配置的级别
- **WHEN** 某个认证级别的 SSH key 或 token 未配置
- **THEN** 系统跳过该级别，不输出日志
