## ADDED Requirements

### Requirement: watch 输出中打印 pipeline Web URL

系统 SHALL 在 watch 模式启动时和 pipeline 结束时各打印一次 pipeline 的 Web URL，方便用户在浏览器中查看 job 详情。

#### Scenario: watch 启动时打印 URL
- **WHEN** watch 开始监控 pipeline #4348，该 pipeline 的 Web URL 为 `https://gitlab.mycompany.com/msa/rt-merger-service/-/pipelines/4348`
- **THEN** 系统 SHALL 在 "Watching pipeline #4348 for rt-merger-service... (Ctrl+C to stop)" 之后打印 URL，格式为 `  👉 <URL>`

#### Scenario: pipeline 结束时打印 URL
- **WHEN** pipeline 到达终态（success/failed/canceled），系统输出 "Pipeline finished: ❌ failed (22s)"
- **THEN** 系统 SHALL 在终态输出之后打印 pipeline Web URL，格式为 `  👉 <URL>`

#### Scenario: pipeline 无 URL
- **WHEN** pipeline 的 Web URL 为空字符串
- **THEN** 系统 SHALL 不打印 URL 行（静默跳过，不输出空行）

#### Scenario: URL 不在状态刷新行中
- **WHEN** watch 正在轮询并使用 `\r` 覆盖状态行
- **THEN** URL SHALL NOT 出现在被 `\r` 覆盖的状态行中，URL 仅出现在独立的、不会被覆盖的输出行中

#### Scenario: Ctrl+C 退出时打印 URL
- **WHEN** 用户按 Ctrl+C 终止 watch，系统输出当前状态后
- **THEN** 系统 SHALL 同样打印 pipeline Web URL
