# clone-auth-priority Specification

## Purpose

定义克隆时的认证优先级链，支持 SSH key、token、默认私钥及按 URL 协议选择策略。

## Requirements

### Requirement: 按 URL 协议选择克隆策略
系统 SHALL 根据解析后的 URL 形态选择克隆策略顺序：

- **首选 HTTPS**（绝对 `http://` 或 `https://` URL）：优先尝试 token HTTPS（若有）与**匿名 HTTPS**，再回退 SSH / 默认私钥 / resource token。
- **首选 SSH**（绝对 SSH URL，或相对路径经 resource 推导的 SSH/HTTPS 对）：保持 SSH key → token → resource SSH → default SSH → 默认私钥文件 → resource token 的顺序；相对路径场景不强制把匿名 HTTPS 插到 SSH 之前，以免内网误走 HTTPS。

系统仍 SHALL NOT 触发会阻塞的交互式 git 凭据提示；匿名 HTTPS 仅执行非交互 `git clone`。

#### Scenario: 绝对 HTTPS URL 优先匿名或 token HTTPS
- **WHEN** 解析结果为绝对 HTTPS CloneURL，且未配置 token
- **THEN** 系统在 SSH 之前尝试匿名 HTTPS 克隆

#### Scenario: 相对路径保持 SSH 优先
- **WHEN** repo.url 为相对路径 `org/app.git`，绑定 resource，未配置 group/repo 级 token
- **THEN** 系统仍按 SSH 优先链尝试，不把匿名 HTTPS 插到 SSH 之前

### Requirement: 未配置 ssh_key 时尝试默认私钥文件
当未配置任何 `ssh_key`，且「系统默认 SSH」（不设置 `GIT_SSH_COMMAND`）克隆失败时，系统 SHALL 对存在的常见默认私钥文件依次尝试，使用与显式 `ssh_key` 相同的 `GIT_SSH_COMMAND=ssh -i <path> -o IdentitiesOnly=yes` 方式。探测顺序 SHALL 为：`~/.ssh/id_ed25519`、`~/.ssh/id_rsa`、`~/.ssh/id_ecdsa`、`~/.ssh/id_ed25519_sk`。不存在的文件 SHALL 跳过。

#### Scenario: agent 不可用时默认 ed25519 可用
- **WHEN** 未配置 `ssh_key`，SSH agent 不可用，但 `~/.ssh/id_ed25519` 存在且对该仓库有效
- **THEN** 系统在 default SSH 失败后使用该默认私钥成功克隆

#### Scenario: 默认私钥文件不存在则跳过
- **WHEN** 上述默认私钥路径均不存在
- **THEN** 系统不额外尝试文件密钥，继续后续 token 等策略（若有）

### Requirement: Resource 级别 SSH key 配置
系统 SHALL 支持在 Resource 配置中指定可选的 `ssh_key`（SSH 密钥文件路径），作为 token 之后的二级认证方式。

#### Scenario: Resource 配置 SSH key
- **WHEN** 配置文件中某 resource 设置了 `ssh_key: ~/.ssh/id_work`
- **THEN** 引用该 resource 的所有 repo 在 clone 时，token 认证失败后可使用该 SSH key 进行 SSH 克隆

#### Scenario: Resource 未配置 SSH key
- **WHEN** 某 resource 未配置 `ssh_key`
- **THEN** clone 回退到推导的 SSH URL（系统默认 SSH），并在其失败后尝试默认私钥文件探测（见「未配置 ssh_key 时尝试默认私钥文件」）

#### Scenario: Resource SSH key 路径展开 ~
- **WHEN** resource 的 `ssh_key` 值为 `~/.ssh/deploy`
- **THEN** 系统将 `~` 展开为用户 home 目录

### Requirement: Group/Repo 级别认证覆盖
系统 SHALL 支持 Group、Repo 配置中指定可选的 `ssh_key`（SSH 密钥文件路径）和 `token`（克隆认证 token）字段，覆盖 resource 级别的默认认证。

#### Scenario: Group 配置 SSH key
- **WHEN** 配置文件中某 group 设置了 `ssh_key: ~/.ssh/id_ed25519`
- **THEN** 该 group 下所有 repo clone 时使用指定的 SSH 密钥文件

#### Scenario: Group 配置 token
- **WHEN** 配置文件中某 group 设置了 `token: ${FRONTEND_TOKEN}`
- **THEN** 该 group 下所有 repo clone 时优先使用该 token（支持 `${ENV_VAR}` 语法）

#### Scenario: 独立 Repo 配置 SSH key
- **WHEN** 配置文件中某独立 repo 设置了 `ssh_key: ~/.ssh/id_personal`
- **THEN** 该 repo clone 时使用指定的 SSH 密钥文件

#### Scenario: 独立 Repo 配置 token
- **WHEN** 配置文件中某独立 repo 设置了 `token: ${DOTFILES_TOKEN}`
- **THEN** 该 repo clone 时优先使用该 token

#### Scenario: 不配置认证字段时使用 resource 认证
- **WHEN** group/repo 未配置 `ssh_key` 和 `token`
- **THEN** 系统使用 resource 级别的 token 和 ssh_key

#### Scenario: token 支持 ${ENV_VAR} 语法
- **WHEN** group/repo 的 `token` 字段值为 `${MY_TOKEN}`
- **THEN** 系统运行时从环境变量 `MY_TOKEN` 解析实际 token 值

### Requirement: Token 认证克隆
系统 SHALL 支持使用配置的 token 进行 HTTPS 认证克隆。认证 URL 构建规则：
- GitHub: `https://x-access-token:<token>@<host>/<path>.git`
- GitLab: `https://oauth2:<token>@<host>/<path>.git`

#### Scenario: GitHub 仓库使用 token 克隆
- **WHEN** clone 一个关联 GitHub resource（含 token）的仓库
- **THEN** 系统构建 `https://x-access-token:<token>@github.com/<org>/<repo>.git` URL 并用于 git clone

#### Scenario: GitLab 仓库使用 token 克隆
- **WHEN** clone 一个关联 GitLab resource（含 token）的仓库
- **THEN** 系统构建 `https://oauth2:<token>@gitlab.com/<org>/<repo>.git` URL 并用于 git clone

#### Scenario: Token 为空时跳过 token 认证
- **WHEN** 最终使用的 token（group/repo 级别或 resource 级别）为空字符串
- **THEN** 系统跳过 token 认证方式，直接尝试 SSH 认证

### Requirement: SSH Key 认证克隆
系统 SHALL 支持使用指定的 SSH 密钥文件进行克隆，通过 `GIT_SSH_COMMAND` 环境变量传递密钥路径。

#### Scenario: 使用指定 SSH key 克隆
- **WHEN** clone 一个配置了 `ssh_key: ~/.ssh/deploy_key` 的 repo（group/repo 级别或 resource 级别）
- **THEN** 系统通过设置 `GIT_SSH_COMMAND=ssh -i ~/.ssh/deploy_key -o IdentitiesOnly=yes` 环境变量执行 git clone

#### Scenario: SSH key 路径展开 ~
- **WHEN** 配置的 `ssh_key` 值为 `~/.ssh/id_ed25519`
- **THEN** 系统将 `~` 展开为用户 home 目录后再传递给 SSH 命令

### Requirement: 克隆认证优先级链
系统 SHALL 按「按 URL 协议选择克隆策略」决定的顺序依次尝试克隆，前一种方式成功即停止。相对路径 / 首选 SSH 时的基线顺序为：
1. group/repo 级别 SSH key（SSH + 指定 key）
2. group/repo 级别 token（HTTPS + token URL）
3. resource 级别 SSH key（SSH + 指定 key）
4. 推导的 SSH URL（系统默认 SSH）
5. 默认私钥文件（若存在）
6. resource 级别 token（HTTPS + token URL）

首选 HTTPS 时，匿名 HTTPS 与 token HTTPS SHALL 排在 SSH 尝试之前（有 token 则 token 优先于匿名）。

未配置的级别 SHALL 被跳过，不产生延迟。系统 SHALL NOT 触发交互式凭据提示；匿名 HTTPS 非交互克隆是允许的。

#### Scenario: group/repo 级别 SSH key 最优先
- **WHEN** group 或 repo 配置了 ssh_key，同时也有 token 和 resource 认证，且首选协议为 SSH
- **THEN** 系统优先使用 group/repo 级别的 SSH key 进行 clone

#### Scenario: group/repo 级别 token 作为二级回退
- **WHEN** group 或 repo 配置了 token 但未配置 ssh_key，同时 resource 也有 ssh_key 和 token，且首选协议为 SSH
- **THEN** 系统优先尝试 group/repo 级别的 token 认证；若失败后使用 resource 的 SSH key

#### Scenario: resource SSH key 优先于 resource token
- **WHEN** group/repo 未配置任何认证，resource 同时配置了 ssh_key 和 token，且首选协议为 SSH
- **THEN** 系统先尝试 resource 的 SSH key 认证，再尝试 default SSH 与默认私钥，最后才使用 resource 的 token

#### Scenario: default SSH 优先于 resource token
- **WHEN** group/repo 未配置 ssh_key，resource 配置了 token，且系统有默认 SSH 配置，首选协议为 SSH
- **THEN** 系统在 resource token 之前先尝试 default SSH（系统默认 SSH agent/config）

#### Scenario: group SSH 失败后回退到 default SSH 再到 resource token
- **WHEN** group 配置了 ssh_key 且 clone 失败，resource 配置了 token，首选协议为 SSH
- **THEN** 系统依次尝试 group token → default SSH → 默认私钥 → resource token

#### Scenario: resource SSH key 作为回退
- **WHEN** group/repo 未配置 ssh_key，但 resource 配置了 ssh_key 和 token，且 token 认证失败，首选协议为 SSH
- **THEN** 系统使用 resource 的 SSH key 尝试 SSH clone

#### Scenario: 所有方式均失败
- **WHEN** 所有认证方式均 clone 失败
- **THEN** 系统报告错误（含诊断信息），不保留失败的目录

#### Scenario: 仅推导 SSH 可用
- **WHEN** 无任何 token 和 ssh_key 配置，但有 SSH URL，且首选协议为 SSH
- **THEN** 系统使用推导的 SSH URL 进行 clone（含 default SSH 与默认私钥兜底）

#### Scenario: 无认证信息且 SSH 失败时不再交互
- **WHEN** 无任何 token、ssh_key 配置，SSH 与默认私钥均失败，且不是可匿名的绝对 HTTPS URL
- **THEN** 系统报告所有认证方式失败，不进入交互式凭据提示

### Requirement: Clone 函数接受认证选项
`git.Clone` 函数 SHALL 接受 `CloneOptions` 结构体参数，包含 `Token`、`Provider`、`SSHKey` 和 `LogWriter` 字段。`LogWriter` 用于控制认证尝试日志的输出目标，为 nil 时默认输出到 stdout。

#### Scenario: Clone 函数接收完整认证选项
- **WHEN** 调用 `Clone(path, sshURL, httpURL, CloneOptions{Token: "abc", Provider: "github", SSHKey: "~/.ssh/key"})` 
- **THEN** 函数按优先级链使用认证信息进行 clone，日志输出到 stdout

#### Scenario: Clone 函数使用 LogWriter
- **WHEN** 调用 `Clone(path, sshURL, httpURL, CloneOptions{Token: "abc", Provider: "github", LogWriter: buf})` 
- **THEN** 函数按优先级链使用认证信息进行 clone，所有日志写入 `buf`

#### Scenario: Clone 函数无认证选项
- **WHEN** 调用 `Clone(path, sshURL, httpURL, CloneOptions{})` 时所有认证字段为空
- **THEN** 函数直接使用 SSH → HTTP 回退策略，日志输出到 stdout

### Requirement: Resolver 合并认证信息
`repo.Resolver` 在构建 `provider.Repo` 列表时 SHALL 按以下规则合并各级别认证信息：
1. token：优先使用 group/repo 级别 → 否则使用 resource 级别
2. ssh_key：优先使用 group/repo 级别 → 否则使用 resource 级别
3. provider：始终从 resource 获取

#### Scenario: Group 配置 token 覆盖 resource token
- **WHEN** group 配置了 `token: ${GROUP_TOKEN}`，其 resource 也配置了 `token: ${RES_TOKEN}`
- **THEN** Resolver 生成的 repo 记录使用 group 级别的 token

#### Scenario: GroupRepo 继承 Group 认证
- **WHEN** group 配置了 `ssh_key` 和 `token`，group 下的 repo 未单独配置
- **THEN** 该 group 下所有 GroupRepo 继承 group 的 ssh_key 和 token 设置

#### Scenario: Standalone repo 配置覆盖 resource
- **WHEN** 独立 repo 配置了 `ssh_key: ~/.ssh/special`，其 resource 也配置了 token
- **THEN** Resolver 生成的 repo 记录使用 repo 级别的 ssh_key 和 resource 级别的 token

#### Scenario: Resource ssh_key 作为 fallback
- **WHEN** group/repo 未配置 ssh_key，但 resource 配置了 `ssh_key: ~/.ssh/id_work`
- **THEN** Resolver 生成的 repo 记录携带 resource 级别的 ssh_key

### Requirement: 认证尝试日志
系统 SHALL 在 clone 过程中尝试每种认证方式时输出日志，包含策略标签和实际 URL，让用户了解当前进度和认证回退过程。

并行模式下（`--concurrency > 1`），认证尝试日志 SHALL 输出到 `CloneOptions.LogWriter`（默认 `os.Stdout`），而非直接 `fmt.Printf`。命令层负责将日志收集到结果结构中，完成后统一展示。

#### Scenario: 尝试 token 认证日志（脱敏）
- **WHEN** 系统尝试使用 token 进行 clone（顺序模式）
- **THEN** 系统输出日志 `  [N/M] 尝试 token 认证 (<级别>)... https://<user>:***@<host>/<path>.git`，其中 token 部分用 `***` 替代

#### Scenario: 尝试 SSH key 认证日志
- **WHEN** 系统尝试使用指定 SSH key 进行 clone（顺序模式）
- **THEN** 系统输出日志 `  [N/M] 尝试 SSH key 认证 (<级别>)... git@<host>:<path>.git`

#### Scenario: 尝试默认 SSH 认证日志
- **WHEN** 系统尝试使用推导的 SSH URL 进行 clone（顺序模式）
- **THEN** 系统输出日志 `  [N/M] 尝试 SSH 认证 (默认)... git@<host>:<path>.git`

#### Scenario: 认证失败日志
- **WHEN** 某种认证方式 clone 失败（顺序模式）
- **THEN** 系统输出错误摘要（不含敏感信息），如 `  [N/M] 失败: <错误摘要>`

#### Scenario: 认证成功日志
- **WHEN** 某种认证方式 clone 成功（顺序模式）
- **THEN** 系统输出 `  [N/M] 成功`

#### Scenario: 跳过未配置的级别
- **WHEN** 某个认证级别的 token 或 ssh_key 未配置
- **THEN** 系统跳过该级别，不输出日志

#### Scenario: 并行模式下日志输出到 LogWriter
- **WHEN** 并行克隆（`--concurrency > 1`）且 `CloneOptions.LogWriter` 已设置
- **THEN** 系统将所有认证尝试日志（尝试、失败、成功）写入 `LogWriter`，而非 `os.Stdout`

#### Scenario: LogWriter 为 nil 时保持原有行为
- **WHEN** `CloneOptions.LogWriter` 为 nil（默认值）
- **THEN** 系统直接使用 `fmt.Printf` 输出日志到 stdout，行为与当前实现一致

### Requirement: 敏感信息不在日志中泄露
系统 SHALL 确保在 clone 失败的日志输出中不包含 token 或认证 URL。

#### Scenario: Token 认证 clone 失败时的日志
- **WHEN** token 认证 clone 失败
- **THEN** 系统日志仅显示错误类型和摘要，不显示包含 token 的 URL

#### Scenario: SSH key 认证 clone 失败时的日志
- **WHEN** SSH key 认证 clone 失败
- **THEN** 系统日志仅显示错误摘要，不泄露 SSH key 完整路径中的敏感部分
