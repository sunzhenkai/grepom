## 1. 基础设施：SHA 获取与参数通道

- [x] 1.1 `git/tag.go` 新增 `TagCommitSHA(path, tag string) (string, error)`：`git rev-parse <tag>^{commit}` 解引用 annotated tag，返回完整 commit SHA；补充单元测试
- [x] 1.2 `cicd/cicd.go` 的 `ListPipelinesParams` 增加 `SHA string` 字段（完整 commit SHA，空值表示不过滤），更新字段注释

## 2. Provider 层 SHA 过滤

- [x] 2.1 `cicd/github.go`：`ListPipelines` 在 `params.SHA != ""` 时追加 `&head_sha=<sha>` 查询参数；`cicd/github_test.go` 增加断言（带 SHA 请求的 URL 含参数、空 SHA 时 URL 与现状一致）
- [x] 2.2 `cicd/gitlab.go`：`ListPipelines` 在 `params.SHA != ""` 时追加 `&sha=<sha>` 查询参数；`cicd/gitlab_test.go` 增加对应断言
- [x] 2.3 确认 `cicd/codeup.go` 忽略 `SHA` 字段且不报错（如无现有用例覆盖，补一条"SHA 字段不影响 Codeup 查询"的测试）

## 3. watch 入口：按 SHA 等待目标 pipeline

- [x] 3.1 `cmd/pipeline.go`：`WatchTarget` 增加 `WatchSHA string`；`runWatchLoop` 在 `targetID==0 && WatchSHA!=""` 时进入等待阶段——以 2s 间隔轮询 `ListPipelines`（携带 SHA 过滤）并本地比对 `Pipeline.SHA` 前缀，命中最新一条后进入现有 watch 循环；将 `signal.NotifyContext` 提前覆盖等待阶段以支持 Ctrl+C
- [x] 3.2 等待超时（60s）时输出明确错误：新 tag 名、SHA 前 7 位、已等待时长、可能原因（未推送 / workflow 未监听 tag push）；不进入 watch、不监控任何其他 pipeline
- [x] 3.3 用一个 fake provider（旧 run 已终态、新 run 延迟 N 秒出现）编写正式回归测试，覆盖规格场景：竞态窗口内不误取旧 pipeline、超时路径报错、Ctrl+C 中断等待路径；该测试先于实现编写并确认失败（红→绿）

## 4. tag 命令接线

- [x] 4.1 `cmd/tag.go`：`runVTag`/`runTTag` 成功路径返回新 tag 名（或通过共享变量/out 参数传出）；push 后路径解析 `TagCommitSHA` 填入 `WatchTarget.WatchSHA`
- [x] 4.2 移除"Waiting 1s for GitLab to create pipeline..."固定等待，替换为等待阶段的进度提示（如 `Waiting for pipeline of <tag> (<sha7>)...`）
- [x] 4.3 未推送（用户拒绝 TTY 确认或无 TTY 跳过推送）时仍设置 `WatchSHA`：依赖超时路径的"可能未推送"提示文案，验证错误信息包含该提示

## 5. 清理与验证

- [x] 5.1 删除 `cmd/zz_debug_tag_watch_race_test.go`（`[DEBUG-a4f2]`），其场景已由 3.3 的正式回归测试覆盖；`grep -rn "DEBUG-a4f2"` 确认无残留
- [x] 5.2 全量验证：`go build ./...`、`go vet ./...`、`go test ./...` 通过
- [x] 5.3 行为核对：`grepom watch`、`grepom pipeline watch <repo>`（不带 --id）与 `--id` 路径输出与改动前一致（对照现有测试与手动 smoke）

## 6. 文档同步

- [x] 6.1 更新 `README.md`：`tag -w`/`tag -pw` 行为说明改为"监控新 tag 对应的 pipeline（按 commit SHA 绑定），最多等待 60 秒，超时给出明确错误"
- [x] 6.2 更新 `README_en.md`：同步 6.1 的英文说明
