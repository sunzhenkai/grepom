## MODIFIED Requirements

### Requirement: API URL 始终使用 HTTPS
系统 SHALL 根据 resource URL 的协议前缀决定 API 调用使用的协议。当 URL 带有 `https://` 前缀时使用 HTTPS；当 URL 带有 `http://` 前缀时使用 HTTP；当 URL 无协议前缀时，`APIURL()` 返回 `https://<host>`，由调用方负责 auto fallback 逻辑。

#### Scenario: URL 带 https 前缀
- **WHEN** resource URL 为 `https://g.wii.pub`，provider 为 gitlab
- **THEN** 系统 API 调用使用 `https://g.wii.pub/api/v4/...`

#### Scenario: URL 带 http 前缀
- **WHEN** resource URL 为 `http://g.wii.pub`，provider 为 gitlab
- **THEN** 系统 API 调用使用 `http://g.wii.pub/api/v4/...`

#### Scenario: URL 无前缀（auto 模式）
- **WHEN** resource URL 为 `g.wii.pub`，provider 为 gitlab
- **THEN** 系统 `APIURL()` 返回 `https://g.wii.pub`，调用方先尝试 HTTPS 失败后自动回退 HTTP

#### Scenario: URL 带端口和 http 前缀
- **WHEN** resource URL 为 `http://g.wii.pub:8080`，provider 为 gitlab
- **THEN** 系统 API 调用使用 `http://g.wii.pub:8080/api/v4/...`

#### Scenario: URL 带端口无前缀（auto 模式）
- **WHEN** resource URL 为 `g.wii.pub:8080`，provider 为 gitlab
- **THEN** 系统 `APIURL()` 返回 `https://g.wii.pub:8080`，调用方先尝试 HTTPS 失败后自动回退 HTTP

## ADDED Requirements

### Requirement: auto 模式下 API 连接失败自动回退 HTTP
当 resource URL 无协议前缀（auto 模式）时，系统 SHALL 先使用 HTTPS 发起 API 请求；若连接失败（TCP 连接错误），系统 SHALL 自动使用 HTTP 重试。若 HTTP 也失败，系统 SHALL 报告错误。

#### Scenario: HTTPS 可用时使用 HTTPS
- **WHEN** resource URL 无前缀，GitLab 服务器 HTTPS 端口可达
- **THEN** 系统使用 HTTPS 成功完成 API 请求，不尝试 HTTP

#### Scenario: HTTPS 不可用时自动回退 HTTP
- **WHEN** resource URL 无前缀，GitLab 服务器 HTTPS 端口不可达但 HTTP 端口可达
- **THEN** 系统先尝试 HTTPS 连接失败，然后自动使用 HTTP 成功完成 API 请求

#### Scenario: HTTPS 和 HTTP 均不可用
- **WHEN** resource URL 无前缀，GitLab 服务器 HTTPS 和 HTTP 端口均不可达
- **THEN** 系统报告 HTTPS 连接错误（不报告 HTTP 重试错误，避免混淆）

#### Scenario: HTTP 错误码不触发 fallback
- **WHEN** resource URL 无前缀，HTTPS 连接成功但返回 401 未授权
- **THEN** 系统直接报告 401 错误，不回退到 HTTP

#### Scenario: auto fallback 时打印提示
- **WHEN** resource URL 无前缀，系统从 HTTPS 回退到 HTTP 成功
- **THEN** 系统输出警告信息，提示用户该 resource 使用了 HTTP 协议，建议在 URL 中显式指定 `http://` 前缀

### Requirement: 显式协议前缀不触发 fallback
当 resource URL 带有 `https://` 或 `http://` 前缀时，系统 SHALL 仅使用指定协议，不进行任何 fallback。

#### Scenario: URL 带 https 前缀且连接失败
- **WHEN** resource URL 为 `https://g.wii.pub`，GitLab 服务器 HTTPS 端口不可达
- **THEN** 系统直接报告连接错误，不尝试 HTTP

#### Scenario: URL 带 http 前缀且连接成功
- **WHEN** resource URL 为 `http://g.wii.pub`，GitLab 服务器 HTTP 端口可达
- **THEN** 系统使用 HTTP 成功完成 API 请求
