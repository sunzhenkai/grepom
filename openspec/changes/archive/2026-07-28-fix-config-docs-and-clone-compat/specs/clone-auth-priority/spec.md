## ADDED Requirements

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

## MODIFIED Requirements

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
