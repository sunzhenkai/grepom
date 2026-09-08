## Context

根因（已由复现测试证实）：`cmd/tag.go:121-122` 在 push 后固定 `time.Sleep(1s)`，随后 `runWatchLoop(target, 0)`（`cmd/pipeline.go:199-209`）盲取 `ListPipelines(Limit:1)` 的第一条。GitHub Actions 异步创建 run（数秒滞后），竞态窗口内第一条是上一个 tag 的已终态 run，watch 立即退出。`cicd/github.go` 的列表接口未按 ref/SHA 过滤，`ListPipelinesParams` 也没有承载 SHA 的字段。

约束：
- GitLab 在 push 处理期间基本同步建 pipeline，1 秒对 GitLab 历史上够用；GitHub 明显不够。
- Codeup（云效 Flow）的运行记录无法按 commit SHA 过滤，`ListPipelines` 是"按仓库匹配 Flow 后取运行记录"的模型。
- `grepom watch` / `pipeline watch <repo>`（不带 --id）没有可绑定的 SHA，保持"监控最新"语义。

## Goals / Non-Goals

**Goals**
- `tag -w` 的监控目标绑定新 tag 指向的 commit SHA：轮询等待匹配 pipeline 出现，杜绝监控到旧 pipeline。
- `ListPipelines` 具备按 SHA 过滤的通道（GitHub/GitLab 真过滤，Codeup 忽略）。
- 目标超时未出现时给出可诊断的错误，不再静默错报。
- 清理 `[DEBUG-a4f2]` 一次性测试，转为正式回归测试。

**Non-Goals**
- 不改变 `grepom watch` / `pipeline watch` 的"最新"语义。
- 不改变 tag 的创建、版本计算、推送交互。
- 不引入对 provider 的事件/webhook 订阅等重型机制。

## Decisions

1. **SHA 从 git 侧获取，而非从 provider 反查**：在 `git/tag.go` 新增 `TagCommitSHA(path, tag)`（`git rev-parse <tag>^{commit}`，自动解引用 annotated tag）。tag 刚由本进程创建，本地对象必然存在且与 push 内容一致；反查 provider 反而受异步延迟影响。
   - 备选：从 watch 结果的 SHA 匹配 tag 名——依赖 provider 字段完整性，放弃。

2. **`WatchTarget` 增加 `WatchSHA string`，`runWatchLoop` 内等待**：`targetID==0 && WatchSHA!=""` 时进入"等待目标出现"阶段；否则维持现有"取最新"路径。等待阶段直接复用 `ListPipelines`（带 SHA 过滤）+ 本地比对 `Pipeline.SHA` 前缀，双保险兼容不过滤的 provider（如 Codeup）。
   - 备选：在 `tag.go` 里先等再调 `runWatchLoop`——等待逻辑与轮询/信号处理交织，分散两处不如内聚在 watch 入口。

3. **SHA 过滤参数放 provider 层**：`ListPipelinesParams.SHA`（完整 40 位 SHA）。GitHub 拼 `&head_sha=<sha>`；GitLab 拼 `&sha=<sha>`（均官方支持）；Codeup 忽略该字段。空值时行为与现状完全一致。

4. **等待窗口 60s、轮询间隔 2s**（等待阶段专用，独立于进入 watch 后的 5s 轮询）。GitHub 官方文档对 run 创建无时延承诺，社区经验为 1~10s；60s 覆盖慢队列且不至于让用户久等。超时错误信息包含：新 tag 名、SHA 前缀、已等待时长、可能原因（未推送 / workflow 未监听 tag）。
   - 备选：无限等待直到 Ctrl+C——与"退出即可见"的工具性格不符，放弃。

5. **多 pipeline 匹配同一 SHA 时取列表第一条**（provider 按创建时间倒序返回，即最新创建的）。GitHub tag push 触发多个 workflow 时行为确定，且未来可在列表输出中体现。

6. **等待阶段支持 Ctrl+C**：等待轮询与 watch 循环共用 `signal.NotifyContext` 的 ctx，用户可随时中断。

## Risks / Trade-offs

- [provider 列表延迟超过 60s（极少见）] → 超时报错给出明确提示，用户可稍后手动 `grepom watch`；不会误报旧 pipeline。
- [Codeup 无法 SHA 过滤，靠本地比对] → Flow 运行记录若不回传 SHA，`Pipeline.SHA` 为空则永不匹配，60s 后超时报错——但 Codeup 的 tag 流水线创建近实时，风险低；错误信息可解释。
- [annotated tag 与 commit SHA 混淆] → `rev-parse <tag>^{commit}` 统一解引用为 commit SHA，provider（GitHub `head_sha`、GitLab `sha`）均以 commit SHA 为准。
- [等待阶段增加 `tag -pw` 总时长最多 60s] → 正常情况 GitHub 数秒内命中，只有异常（未触发 CI）才付满窗口，且换来正确性。

## Migration Plan

纯 CLI 行为修正，无数据迁移。按 tasks 顺序实现即可；回滚即 revert 提交。`grepom watch`、`pipeline watch` 不受影响，风险面限定在 `tag -w` 路径。

## Open Questions

（无）
