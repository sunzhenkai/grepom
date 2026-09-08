## Why

`grepom tag -pw` 在 push 完成后仅固定等待 1 秒，然后盲取 pipeline 列表第一条当作"最新 pipeline"。但 GitHub Actions 的 workflow run 是异步创建的，通常滞后数秒才出现在 API 中；在竞态窗口内取到的是**上一个 tag 的 run**（往往已终态），watch 随即显示"已完成"并立即退出——用户监控的实际是旧版本的 pipeline。根因分析已通过确定性复现测试证实（见 `cmd/zz_debug_tag_watch_race_test.go`，修复后删除）。

## What Changes

- **SHA 绑定 watch**：`tag -w`（含 `-pw` 组合）在 push 成功后，解析新 tag 指向的 commit SHA，轮询等待"该 SHA 对应的 pipeline"出现后再进入 watch；不再依赖固定 1 秒等待和盲取 `pipelines[0]`。
- **ListPipelines 支持 SHA 过滤**：`ListPipelinesParams` 新增 SHA 字段；GitHub provider 传递 `head_sha` 查询参数，GitLab provider 传递 `sha` 参数；Codeup（云效 Flow）无法按 SHA 过滤，保持现状（best-effort，不因 SHA 参数报错）。
- **等待超时与反馈**：SHA 匹配的 pipeline 在超时窗口内未出现时，给出明确的警告/错误信息（而非静默监控旧 pipeline），超时窗口与轮询间隔在设计中定义。
- **行为不变部分**：`grepom watch`、`grepom pipeline watch <repo>`（不带 --id）保持"监控最新 pipeline"语义不变；`pipeline watch --id` 不变；tag 的创建、推送、确认交互均不变。
- **清理**：删除一次性调试测试 `cmd/zz_debug_tag_watch_race_test.go`，其场景转化为正式回归测试。
- **文档同步**：更新 `README.md` 与 `README_en.md` 中 `tag -w` 的行为说明（监控新 tag 的 pipeline，而非任意最新 pipeline）。

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `tag-watch`：`tag -w` 的监控目标从"最新 pipeline"改为"刚创建/推送的 tag 所指向 commit 的 pipeline"——push 成功后按 SHA 轮询等待目标 pipeline 出现（超时给出明确错误），取代固定 1 秒等待 + 盲取第一条的行为；`grepom watch` 与 `tag -w` 的循环渲染/终态/Ctrl+C 行为保持一致，但监控目标的确定方式分叉（watch=最新，tag -w=SHA 绑定）。

## Impact

- **代码**：`cmd/tag.go`（push 后获取 tag SHA、传入 watch）、`cmd/pipeline.go`（`runWatchLoop` 支持按 SHA 等待目标 pipeline）、`cicd/cicd.go`（`ListPipelinesParams` 增加字段）、`cicd/github.go`、`cicd/gitlab.go`（SHA 过滤参数）、`git/tag.go`（新增解析 tag commit SHA 的辅助函数）。
- **测试**：`cmd/tag_watch_test.go`、`cicd/github_test.go`、`cicd/gitlab_test.go` 新增/调整用例；删除 `cmd/zz_debug_tag_watch_race_test.go`。
- **API 兼容**：`ListPipelinesParams` 为内部结构体，新增字段不构成破坏性变更；GitHub `head_sha`、GitLab `sha` 均为官方支持的过滤参数。
- **用户可见行为**：`tag -pw` 不再显示旧版本 pipeline；CI 未触发（如 workflow 未监听 tag push 事件）时会在超时后得到明确错误而非错误地报告旧 pipeline 成功。
