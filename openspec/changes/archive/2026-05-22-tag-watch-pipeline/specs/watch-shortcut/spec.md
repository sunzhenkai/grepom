## MODIFIED Requirements

### Requirement: watch 命令复用 pipeline watch 的 watch 循环逻辑

`grepom watch`、`grepom pipeline watch` 和 `grepom tag -w` SHALL 共享同一 watch 轮询循环实现，不重复实现轮询、状态渲染、Ctrl+C 处理等逻辑。

#### Scenario: watch 循环行为一致
- **WHEN** 用户通过 `grepom watch web-app`、`grepom pipeline watch web-app` 或 `grepom tag -w` 监控同一个 pipeline
- **THEN** 三者的轮询间隔、状态行格式、终态退出行为 SHALL 完全一致

#### Scenario: Ctrl+C 行为一致
- **WHEN** 用户在 `grepom watch` 或 `grepom tag -w` 运行过程中按 Ctrl+C
- **THEN** 系统 SHALL 与 `pipeline watch` 相同的方式优雅退出