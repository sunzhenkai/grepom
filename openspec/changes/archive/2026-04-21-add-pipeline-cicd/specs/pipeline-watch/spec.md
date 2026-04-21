## ADDED Requirements

### Requirement: pipeline watch 命令实时监控 pipeline 状态

系统 SHALL 提供 `grepom pipeline watch <repo-name>` 命令，实时监控指定仓库最新一条 pipeline 的运行状态。

#### Scenario: watch 最新 pipeline
- **WHEN** 用户运行 `grepom pipeline watch web-app`
- **THEN** 系统 SHALL 获取该仓库最新的一条 pipeline，开始轮询监控（每 5 秒查询一次状态），在终端实时刷新状态行

#### Scenario: watch 指定 pipeline
- **WHEN** 用户运行 `grepom pipeline watch web-app --id 1234`
- **THEN** 系统 SHALL 监控 ID 为 1234 的 pipeline

#### Scenario: repo 不存在
- **WHEN** 用户指定的 repo name 在 config 中找不到
- **THEN** 系统 SHALL 输出错误信息并退出

#### Scenario: 无 pipeline 记录
- **WHEN** 指定仓库没有任何 pipeline 记录
- **THEN** 系统 SHALL 输出 "No pipelines found for <repo-name>." 并退出

### Requirement: watch 自然结束

系统 SHALL 在 pipeline 到达终态时自动停止并退出。

#### Scenario: pipeline 成功完成
- **WHEN** watch 监控的 pipeline 状态变为 `success`
- **THEN** 系统 SHALL 输出最终状态 "Pipeline finished: ✅ success (2m 34s)"，然后退出

#### Scenario: pipeline 失败
- **WHEN** watch 监控的 pipeline 状态变为 `failed`
- **THEN** 系统 SHALL 输出最终状态 "Pipeline finished: ❌ failed (5m 12s)"，然后退出

#### Scenario: pipeline 被取消
- **WHEN** watch 监控的 pipeline 状态变为 `canceled`
- **THEN** 系统 SHALL 输出最终状态 "Pipeline finished: 🚫 canceled"，然后退出

#### Scenario: 终态定义
- 终态 SHALL 包括：`success`、`failed`、`canceled`
- 非终态：`running`、`pending`

### Requirement: watch 支持 Ctrl+C 提前终止

系统 SHALL 捕获 SIGINT 信号，优雅退出 watch 模式。

#### Scenario: 用户按 Ctrl+C
- **WHEN** watch 正在运行且用户按下 Ctrl+C
- **THEN** 系统 SHALL 停止轮询，输出当前已知的 pipeline 状态，然后退出（退出码 0）

### Requirement: watch 输出格式

系统 SHALL 在 watch 过程中实时刷新状态行。

#### Scenario: 轮询输出
- **WHEN** watch 正在运行
- **THEN** 系统 SHALL 在同一行覆盖输出当前状态，格式为：
  ```
  🔄 running   #1234  main  abc1234  (1m 23s)
  ```
  使用 `\r` 回到行首覆盖，不产生多行输出。每次轮询刷新一次。

#### Scenario: watch 启动信息
- **WHEN** watch 开始
- **THEN** 系统 SHALL 先输出一行启动信息：
  ```
  Watching pipeline #1234 for web-app... (Ctrl+C to stop)
  ```
  然后开始轮询状态行。

### Requirement: watch 轮询间隔

系统 SHALL 每 5 秒调用一次 API 获取 pipeline 状态。

#### Scenario: 轮询间隔
- **WHEN** watch 正在运行
- **THEN** 系统 SHALL 每 5 秒调用 `GetPipeline` API，刷新状态行，间隔期间响应 Ctrl+C

### Requirement: watch 复用 Token 链路

系统 SHALL 复用与 pipeline list 相同的 token 获取和 repo 解析逻辑。

#### Scenario: Token 和 Resource 解析
- **WHEN** 执行 pipeline watch
- **THEN** 系统 SHALL 使用与 `pipeline list` 相同的流程获取 token、ServerURL、Provider 类型和远程路径
