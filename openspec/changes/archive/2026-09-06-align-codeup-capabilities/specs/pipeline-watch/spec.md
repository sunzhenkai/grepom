## ADDED Requirements

### Requirement: watch 支持 Codeup Provider

系统 SHALL 支持对 Codeup 仓库执行 `grepom pipeline watch` 和顶级 `watch` 命令，复用与 pipeline list 相同的云效 Flow 解析逻辑获取 provider、token、organization_id 和仓库路径。

#### Scenario: watch Codeup 仓库最新 pipeline
- **WHEN** 用户对 Codeup 仓库运行 `grepom pipeline watch <repo-name>`
- **THEN** 系统 SHALL 通过云效 Flow 接口获取该仓库关联流水线的最新运行记录，并按既有 watch 语义（5 秒轮询、终态退出、Ctrl+C 终止）监控

#### Scenario: Codeup 仓库无关联流水线
- **WHEN** 组织内没有任何 Flow 流水线的代码源指向该仓库
- **THEN** 系统 SHALL 输出 "No pipelines found for <repo-name>." 并退出
