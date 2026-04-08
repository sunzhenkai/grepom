## MODIFIED Requirements

### Requirement: add resource 命令的 provider 验证
`add resource` 命令 SHALL 验证 `--provider` 参数值为合法的 provider 类型（gitlab、github 或 generic）。不合法时 SHALL 报错并列出所有支持的 provider。

#### Scenario: 使用 generic provider 添加资源
- **WHEN** 用户运行 `grepom add resource --name my-git --provider generic --url git.internal.com --token ${GIT_TOKEN}`
- **THEN** 系统正常添加资源，provider 为 `generic`

#### Scenario: 使用不支持的 provider 添加资源
- **WHEN** 用户运行 `grepom add resource --name test --provider unknown --url example.com`
- **THEN** 系统报错提示不支持的 provider，并列出 `gitlab`、`github`、`generic`

#### Scenario: provider 参数缺失
- **WHEN** 用户运行 `grepom add resource --name test --url example.com`（未指定 `--provider`）
- **THEN** 系统报错提示 `--provider` 是必填参数
