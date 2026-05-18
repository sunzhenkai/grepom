## ADDED Requirements

### Requirement: HTTPS URL 中剥离 userinfo
`extractHost()` 函数在处理 HTTPS/HTTP 格式的 git remote URL 时，SHALL 在提取 host 之前剥离 `user:password@` 部分。

#### Scenario: 包含 oauth2 token 的 GitLab URL
- **WHEN** remote URL 为 `https://oauth2:glpat-xxx@gitlab.company.com/myorg/repo.git`
- **THEN** `extractHost()` SHALL 返回 `gitlab.company.com`

#### Scenario: 包含用户名密码的 URL
- **WHEN** remote URL 为 `https://user:pass@host.example.com/org/repo.git`
- **THEN** `extractHost()` SHALL 返回 `host.example.com`

#### Scenario: 仅含用户名不含密码的 URL
- **WHEN** remote URL 为 `https://token@github.com/user/repo.git`
- **THEN** `extractHost()` SHALL 返回 `github.com`

#### Scenario: 不含 userinfo 的标准 URL
- **WHEN** remote URL 为 `https://gitlab.com/myorg/repo.git`
- **THEN** `extractHost()` SHALL 返回 `gitlab.com`（行为不变，向后兼容）

### Requirement: SSH URL 中剥离 userinfo
`extractHost()` 函数在处理 `ssh://` 格式的 URL 时，SHALL 在提取 host 之前剥离 `user@` 部分。

#### Scenario: ssh:// URL 包含用户名
- **WHEN** remote URL 为 `ssh://git@gitlab.company.com:2222/org/repo.git`
- **THEN** `extractHost()` SHALL 返回 `gitlab.company.com:2222`

#### Scenario: ssh:// URL 不含用户名
- **WHEN** remote URL 为 `ssh://gitlab.company.com/org/repo.git`
- **THEN** `extractHost()` SHALL 返回 `gitlab.company.com`（行为不变）

### Requirement: SCP 风格 URL 不受影响
`extractHost()` 函数在处理 `git@host:path` 格式的 SCP 风格 URL 时，SHALL 保持原有行为不变。

#### Scenario: 标准 git@ SCP 格式
- **WHEN** remote URL 为 `git@github.com:user/repo.git`
- **THEN** `extractHost()` SHALL 返回 `github.com`（行为不变）

#### Scenario: git@ 自托管 SCP 格式
- **WHEN** remote URL 为 `git@gitlab.mycompany.com:org/repo.git`
- **THEN** `extractHost()` SHALL 返回 `gitlab.mycompany.com`（行为不变）
