### Requirement: generic provider 支持纯 Git URL 仓库管理
系统 SHALL 提供 `generic` provider 类型，允许用户通过显式声明 Git URL 管理不依赖任何平台 API 的仓库。`generic` provider 不支持通过 API 自动发现仓库，仓库必须在配置文件中显式声明。

#### Scenario: 使用 generic provider 克隆仓库
- **WHEN** 配置文件中存在 `provider: generic` 的 resource，且该 resource 下有显式声明的 repo
- **THEN** 系统使用该 repo 的 URL 直接克隆，无需调用任何平台 API

#### Scenario: generic provider 的 sync 命令静默跳过
- **WHEN** 用户运行 `grepom sync`，且存在 `provider: generic` 的 resource
- **THEN** 系统跳过该 resource 的 API 发现步骤，不报错，不修改该 resource 下已声明的 repos

#### Scenario: generic provider 不支持 list --remote
- **WHEN** 用户运行 `grepom list --remote --type groups`，且存在 `provider: generic` 的 resource
- **THEN** 系统跳过 generic resource，仅查询支持 API 的 provider

### Requirement: generic provider 的认证支持
`generic` provider SHALL 支持与其他 provider 相同的认证优先级链（SSH key 优先，token 次之）。

#### Scenario: 使用 SSH key 克隆 generic 仓库
- **WHEN** generic resource 或 repo 配置了 `ssh_key`
- **THEN** 系统优先使用 SSH key 进行克隆

#### Scenario: 使用 token 克隆 generic 仓库
- **WHEN** generic resource 配置了 `token`，且未配置 SSH key
- **THEN** 系统使用 token 进行 HTTPS 认证克隆，token 用户名为 `token`
