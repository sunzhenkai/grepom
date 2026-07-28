## ADDED Requirements

### Requirement: 绑定 resource 时 repo.url 分类解析
当独立 repo 绑定了 resource 时，系统 SHALL 按 `repo.url` 形态解析克隆地址，而不是无条件用 `resource.url` 二次拼接完整 URL：

1. **相对路径**（无 `http(s)://`、`git@`、`ssh://` 前缀）：使用 `resource.url` 作为 host，分别生成 HTTPS 与 SSH 克隆 URL（与现有 `HTTPSURL` / `SSHURL` 一致）。
2. **绝对 HTTP(S) URL**：CloneURL 使用该 URL 原样（可补全 `.git` 后缀若缺失）；不得把整个绝对 URL 当作 path 再拼到 `resource.url` 后。
3. **绝对 SSH URL**（`git@host:path` 或 `ssh://...`）：SSHURL 使用该 URL 原样交给 `git`；不得生成 `git@<resource.url>:<原始绝对URL>` 这类无效地址。

resource 仍用于提供 `token`、`ssh_key`、`provider` 等认证元数据。

#### Scenario: 相对路径与 resource host 拼接
- **WHEN** resource `url` 为 `git.example.com`，repo `url` 为 `tools/internal-tool.git`
- **THEN** SSHURL 为 `git@git.example.com:tools/internal-tool.git`，CloneURL 为 `https://git.example.com/tools/internal-tool.git`

#### Scenario: 完整 HTTPS URL 不再二次拼接
- **WHEN** repo `url` 为 `https://git.example.com/tools/internal-tool.git`，且绑定了任意 resource
- **THEN** CloneURL 为该 HTTPS URL 本身，不会变成 `git@<resource.host>:https://...` 或错误 path

#### Scenario: 完整 ssh:// URL 原样用于 SSH
- **WHEN** repo `url` 为 `ssh://git@git.example.com/tools/internal-tool.git`
- **THEN** 系统使用该 URL（或 git 等价形式）进行 SSH 克隆，不会将其拼进 `git@<resource.host>:ssh://...`

#### Scenario: 完整 git@ URL 原样用于 SSH
- **WHEN** repo `url` 为 `git@git.example.com:tools/internal-tool.git`
- **THEN** SSHURL 为该字符串本身

### Requirement: 完整 HTTPS URL 支持匿名克隆
当 repo 解析结果的首选协议为 HTTPS（`repo.url` 为绝对 `http://` 或 `https://`）时，系统 SHALL 在认证链中包含无 token 的匿名 HTTPS 克隆尝试，使公开仓库在无 SSH key、无 token 时仍可成功克隆。

#### Scenario: 公开 HTTPS 仓库无认证可克隆
- **WHEN** 独立 repo 的 `url` 为公开可匿名访问的 `https://github.com/example/public-repo.git`，未配置 token 与 ssh_key
- **THEN** 系统通过匿名 HTTPS `git clone` 成功克隆该仓库

#### Scenario: 私有 HTTPS 匿名失败后回退
- **WHEN** repo `url` 为私有 HTTPS 地址，匿名克隆失败，但配置了可用的 token 或 ssh_key
- **THEN** 系统继续按认证优先级尝试后续策略，不因匿名失败而立即放弃

## MODIFIED Requirements

### Requirement: generic provider 支持纯 Git URL 仓库管理
系统 SHALL 提供 `generic` provider 类型，允许用户通过显式声明 Git URL 管理不依赖任何平台 API 的仓库。`generic` provider 不支持通过 API 自动发现仓库，仓库必须在配置文件中显式声明。绑定 resource 时的 URL 解析 MUST 遵循「绑定 resource 时 repo.url 分类解析」。

#### Scenario: 使用 generic provider 克隆仓库
- **WHEN** 配置文件中存在 `provider: generic` 的 resource，且该 resource 下有显式声明的 repo
- **THEN** 系统按 repo.url 分类结果克隆，无需调用任何平台 API

#### Scenario: generic provider 的 sync 命令静默跳过
- **WHEN** 用户运行 `grepom sync`，且存在 `provider: generic` 的 resource
- **THEN** 系统跳过该 resource 的 API 发现步骤，不报错，不修改该 resource 下已声明的 repos

#### Scenario: generic provider 不支持 list --remote
- **WHEN** 用户运行 `grepom list --remote --type groups`，且存在 `provider: generic` 的 resource
- **THEN** 系统跳过 generic resource，仅查询支持 API 的 provider
