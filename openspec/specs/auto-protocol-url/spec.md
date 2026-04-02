### Requirement: 从 host:port 推导克隆 URL
系统 SHALL 支持从 resource 的 host:port 地址自动推导出 SSH、HTTPS、HTTP 三种协议的克隆 URL。推导规则：
- SSH URL: `git@<host>:<path>.git`（path 中 `/` 替换为 `:` 后的首个 `/`）
- HTTPS URL: `https://<host>/<path>.git`
- HTTP URL: `http://<host>/<path>.git`

#### Scenario: 从 host 推导 SSH URL
- **WHEN** resource host 为 `g.wii.pub`，repo 路径为 `my-org/my-repo`
- **THEN** 系统推导出 SSH URL `git@g.wii.pub:my-org/my-repo.git`

#### Scenario: 从 host:port 推导 SSH URL
- **WHEN** resource host 为 `g.wii.pub:8022`，repo 路径为 `my-org/my-repo`
- **THEN** 系统推导出 SSH URL `git@g.wii.pub:8022:my-org/my-repo.git`

#### Scenario: 从 host 推导 HTTPS URL
- **WHEN** resource host 为 `g.wii.pub`，repo 路径为 `my-org/my-repo`
- **THEN** 系统推导出 HTTPS URL `https://g.wii.pub/my-org/my-repo.git`

#### Scenario: 从 host:port 推导 HTTPS URL
- **WHEN** resource host 为 `g.wii.pub:8443`，repo 路径为 `my-org/my-repo`
- **THEN** 系统推导出 HTTPS URL `https://g.wii.pub:8443/my-org/my-repo.git`

#### Scenario: 从 host 推导 HTTP URL
- **WHEN** resource host 为 `g.wii.pub`，repo 路径为 `my-org/my-repo`
- **THEN** 系统推导出 HTTP URL `http://g.wii.pub/my-org/my-repo.git`

#### Scenario: 从 host:port 推导 HTTP URL
- **WHEN** resource host 为 `g.wii.pub:8080`，repo 路径为 `my-org/my-repo`
- **THEN** 系统推导出 HTTP URL `http://g.wii.pub:8080/my-org/my-repo.git`

### Requirement: API URL 始终使用 HTTPS
系统 SHALL 为 provider API 调用自动从 host:port 推导 HTTPS URL（`https://<host>/api/...`）。

#### Scenario: 推导 GitLab API URL
- **WHEN** resource host 为 `g.wii.pub`，provider 为 gitlab
- **THEN** 系统 API 调用使用 `https://g.wii.pub/api/v4/...`

#### Scenario: 推导带端口的 API URL
- **WHEN** resource host 为 `g.wii.pub:8443`，provider 为 gitlab
- **THEN** 系统 API 调用使用 `https://g.wii.pub:8443/api/v4/...`

### Requirement: Clone 日志打印实际 URL
系统 SHALL 在 clone 尝试每种认证策略时打印实际使用的 URL，便于用户调试。

#### Scenario: 日志包含实际 URL
- **WHEN** 系统尝试 SSH 认证克隆仓库
- **THEN** 系统输出日志 `  [N/M] 尝试 SSH 认证 (默认)... git@g.wii.pub:org/repo.git`

#### Scenario: token URL 日志脱敏
- **WHEN** 系统尝试 token 认证克隆仓库
- **THEN** 系统输出日志 `  [N/M] 尝试 token 认证 (resource)... https://oauth2:***@g.wii.pub/org/repo.git`（token 部分用 `***` 替代）
