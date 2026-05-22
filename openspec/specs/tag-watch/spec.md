## ADDED Requirements

### Requirement: tag 命令 -w/--watch 参数

系统 SHALL 在 `tag` 命令中提供 `-w/--watch` 参数，在 tag 成功创建后自动触发 pipeline watch，监控 CI/CD pipeline 的运行状态。

#### Scenario: 使用 -w 创建 v 版本 tag 并 watch pipeline
- **WHEN** 用户运行 `grepom tag -w`
- **THEN** 系统 SHALL 创建下一个 v 版本 tag，创建成功后自动使用 `resolveCurrentRepoPipeline()` 推断当前仓库信息，并调用 `runWatchLoop()` 监控最新 pipeline

#### Scenario: 使用 -w 创建 t 版本 tag 并 watch pipeline
- **WHEN** 用户运行 `grepom tag -t -w`
- **THEN** 系统 SHALL 创建下一个 t 版本 tag，创建成功后自动 watch 最新 pipeline

#### Scenario: -w 与 -p 组合使用
- **WHEN** 用户运行 `grepom tag -w -p`
- **THEN** 系统 SHALL 创建 tag，推送到所有 remotes，推送成功后自动 watch 最新 pipeline

#### Scenario: -w 与 --dry-run 组合
- **WHEN** 用户运行 `grepom tag -w --dry-run`
- **THEN** 系统 SHALL 仅输出预览信息（如 `[dry-run] Would create tag v0.1.6 locally`），不实际创建 tag，不进入 watch

#### Scenario: tag 创建失败
- **WHEN** 用户运行 `grepom tag -w` 但 tag 创建过程发生错误
- **THEN** 系统 SHALL 输出错误信息并退出，不进入 watch

#### Scenario: tag 创建成功但自动推断 repo 信息失败
- **WHEN** 用户运行 `grepom tag -w`，tag 创建成功，但 `resolveCurrentRepoPipeline()` 三级 fallback 均失败
- **THEN** 系统 SHALL 输出 tag 创建成功的消息，然后输出与 `grepom watch` 相同的详细错误信息（包含诊断和建议），并以非零退出码退出

### Requirement: -w 复用 watch 推断和循环逻辑

系统 SHALL 确保 `tag -w` 的 watch 行为与 `grepom watch` 命令完全一致，复用 `resolveCurrentRepoPipeline()` 和 `runWatchLoop()` 函数。

#### Scenario: 三级 fallback 推断行为一致
- **WHEN** 用户通过 `grepom tag -w` 或 `grepom watch` 在同一目录执行
- **THEN** 两者的 repo 推断逻辑 SHALL 完全一致（Level 1 配置匹配 → Level 2 host 匹配 → Level 3 公共域名 + 环境变量）

#### Scenario: watch 循环行为一致
- **WHEN** 用户通过 `grepom tag -w` 进入 watch
- **THEN** 轮询间隔、状态行格式、终态退出行为、Ctrl+C 处理 SHALL 与 `grepom watch` 完全一致

### Requirement: -w 始终监控最新 pipeline

`tag -w` SHALL 始终监控最新 pipeline，不提供指定 pipeline ID 的选项。

#### Scenario: 不支持 --id
- **WHEN** 用户运行 `grepom tag -w --id 1234`
- **THEN** 系统 SHALL 忽略 `--id` 参数或报错提示不支持，始终监控最新 pipeline

### Requirement: -w 不绑定 -p

`-w` 参数 SHALL 独立于 `-p`（push）参数，不隐含推送行为。

#### Scenario: 仅使用 -w 不使用 -p
- **WHEN** 用户运行 `grepom tag -w`（不带 -p）
- **THEN** 系统 SHALL 创建 tag 本地（可能通过 TTY 提问是否推送），创建后进入 watch

#### Scenario: -w 不隐含推送
- **WHEN** 用户运行 `grepom tag -w`
- **THEN** 系统 SHALL NOT 自动推送 tag，推送行为仍由 `-p` 或 TTY 确认控制