## MODIFIED Requirements

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
