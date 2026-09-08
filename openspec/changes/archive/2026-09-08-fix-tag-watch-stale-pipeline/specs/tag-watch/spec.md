## MODIFIED Requirements

### Requirement: tag 命令 -w/--watch 参数

系统 SHALL 在 `tag` 命令中提供 `-w/--watch` 参数，在 tag 成功创建后自动触发 pipeline watch，监控**新 tag 所指向 commit** 的 CI/CD pipeline 运行状态。

#### Scenario: 使用 -w 创建 v 版本 tag 并 watch pipeline
- **WHEN** 用户运行 `grepom tag -w`
- **THEN** 系统 SHALL 创建下一个 v 版本 tag，创建成功后自动使用 `resolveCurrentRepoPipeline()` 推断当前仓库信息，并监控新 tag 所指向 commit 的 pipeline

#### Scenario: 使用 -w 创建 t 版本 tag 并 watch pipeline
- **WHEN** 用户运行 `grepom tag -t -w`
- **THEN** 系统 SHALL 创建下一个 t 版本 tag，创建成功后自动监控新 tag 所指向 commit 的 pipeline

#### Scenario: -w 与 -p 组合使用
- **WHEN** 用户运行 `grepom tag -w -p`
- **THEN** 系统 SHALL 创建 tag，推送到所有 remotes，推送成功后自动监控新 tag 所指向 commit 的 pipeline

#### Scenario: -w 与 --dry-run 组合
- **WHEN** 用户运行 `grepom tag -w --dry-run`
- **THEN** 系统 SHALL 仅输出预览信息（如 `[dry-run] Would create tag v0.1.6 locally`），不实际创建 tag，不进入 watch

#### Scenario: tag 创建失败
- **WHEN** 用户运行 `grepom tag -w` 但 tag 创建过程发生错误
- **THEN** 系统 SHALL 输出错误信息并退出，不进入 watch

#### Scenario: tag 创建成功但自动推断 repo 信息失败
- **WHEN** 用户运行 `grepom tag -w`，tag 创建成功，但 `resolveCurrentRepoPipeline()` 三级 fallback 均失败
- **THEN** 系统 SHALL 输出 tag 创建成功的消息，然后输出与 `grepom watch` 相同的详细错误信息（包含诊断和建议），并以非零退出码退出

## ADDED Requirements

### Requirement: -w 监控新 tag 对应的 pipeline

`tag -w` SHALL 以新 tag 指向的 commit SHA 为监控目标：push（或本地创建）完成后，系统 SHALL 按 SHA 轮询查询目标 pipeline，直到出现匹配的 pipeline 或超过等待窗口。系统 SHALL NOT 在目标 pipeline 尚未出现时退而监控其他不相关 pipeline。

#### Scenario: 新 pipeline 未立即出现时持续等待
- **WHEN** 用户运行 `grepom tag -pw`，push 成功后 provider 的 pipeline 列表尚未出现该 tag 指向 commit 的 pipeline（异步创建延迟）
- **THEN** 系统 SHALL 按 SHA 轮询持续等待，直到匹配的 pipeline 出现后进入 watch，监控的 pipeline 的 SHA 与新 tag 指向的 commit 一致

#### Scenario: 等待窗口内目标 pipeline 出现
- **WHEN** 匹配 SHA 的 pipeline 在等待窗口内出现
- **THEN** 系统 SHALL 立即进入 watch 循环，其轮询间隔、状态行格式、终态退出行为、Ctrl+C 处理与 `grepom watch` 完全一致

#### Scenario: 超时未出现
- **WHEN** 等待窗口耗尽仍未出现匹配 SHA 的 pipeline（如 workflow 未监听 tag push 事件、未推送导致远端无触发）
- **THEN** 系统 SHALL 输出明确的错误信息（说明未找到新 tag 对应的 pipeline，并提示可能原因），且 SHALL NOT 监控或报告其他 pipeline 的状态

#### Scenario: 未推送即 watch
- **WHEN** 用户运行 `grepom tag -w`（不带 -p）且未确认推送，随后目标 pipeline 在等待窗口内未出现
- **THEN** 系统 SHALL 按超时场景输出明确错误信息，错误信息中 SHALL 包含"可能未推送"的提示

#### Scenario: 同一 SHA 存在多个 pipeline
- **WHEN** 新 tag 的 push 触发了多个 pipeline（如多个 workflow 文件监听 tag push 事件）
- **THEN** 系统 SHALL 选取其中最新创建的一个进入 watch，且 SHALL NOT 监控其他 SHA 的 pipeline

#### Scenario: 不支持 --id
- **WHEN** 用户运行 `grepom tag -w --id 1234`
- **THEN** 系统 SHALL 忽略 `--id` 参数或报错提示不支持，始终监控新 tag 对应的 pipeline

## REMOVED Requirements

### Requirement: -w 始终监控最新 pipeline

**Reason**: "取列表第一条当作最新"正是本缺陷的根因——provider 侧 pipeline 异步创建期间，列表第一条是上一个 tag 的已终态 run，导致 `tag -pw` 显示旧版本 pipeline 并立即退出。监控目标必须绑定到新 tag 的 commit SHA，而非任意"当前最新"。

**Migration**: `tag -w` 改为按新 tag 指向的 commit SHA 绑定监控目标（见 ADDED Requirement "-w 监控新 tag 对应的 pipeline"）；`grepom watch` 与 `grepom pipeline watch <repo>`（不带 --id）仍保持"监控最新 pipeline"语义，行为不变。
