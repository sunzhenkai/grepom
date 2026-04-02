## MODIFIED Requirements

### Requirement: 克隆认证优先级链
系统 SHALL 按以下优先级依次尝试克隆，前一种方式成功即停止：
1. group/repo 级别 SSH key（SSH + 指定 key）
2. group/repo 级别 token（HTTPS + token URL）
3. resource 级别 SSH key（SSH + 指定 key）
4. 推导的 SSH URL（系统默认 SSH）
5. resource 级别 token（HTTPS + token URL）
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
- **THEN** 系统先尝试 resource 的 SSH key 认证，再尝试 default SSH，最后才使用 resource 的 token

#### Scenario: default SSH 优先于 resource token
- **WHEN** group/repo 未配置 ssh_key，resource 配置了 token，且系统有默认 SSH 配置
- **THEN** 系统在 resource token 之前先尝试 default SSH（系统默认 SSH agent/config）

#### Scenario: group SSH 失败后回退到 default SSH 再到 resource token
- **WHEN** group 配置了 ssh_key 且 clone 失败，resource 配置了 token
- **THEN** 系统依次尝试 group token → default SSH → resource token

#### Scenario: resource SSH key 作为回退
- **WHEN** group/repo 未配置 ssh_key，但 resource 配置了 ssh_key 和 token，且 token 认证失败
- **THEN** 系统使用 resource 的 SSH key 尝试 SSH clone

#### Scenario: 所有方式均失败
- **WHEN** 所有认证方式均 clone 失败
- **THEN** 系统报告错误，不保留失败的目录

#### Scenario: 仅推导 SSH 可用
- **WHEN** 无任何 token 和 ssh_key 配置，但有 SSH URL
- **THEN** 系统使用推导的 SSH URL 进行 clone
