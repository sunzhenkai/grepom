## 1. 重构 pipeline watch 核心逻辑

- [x] 1.1 在 `cmd/pipeline.go` 中定义 `WatchTarget` 结构体（Provider、ServerURL、RepoPath、Token、RepoName 字段）
- [x] 1.2 从 `runPipelineWatch` 中提取 watch 轮询循环逻辑到独立函数 `runWatchLoop(target WatchTarget, pipelineID int, ctx context.Context) error`，包含：获取最新/指定 pipeline、启动信息打印、轮询、状态渲染、终态退出、Ctrl+C 处理
- [x] 1.3 改造 `runPipelineWatch` 使用 `resolvePipelineInput` 构造 `WatchTarget`，再调用 `runWatchLoop`
- [x] 1.4 验证 `grepom pipeline watch <repo-name>` 行为与重构前一致

## 2. watch 输出增加 pipeline URL 打印

- [x] 2.1 在 `runWatchLoop` 中，watch 启动信息之后打印 pipeline URL（格式：`  👉 <URL>`），当 URL 为空时静默跳过
- [x] 2.2 在 pipeline 终态退出时打印 pipeline URL
- [x] 2.3 在 Ctrl+C 退出时打印 pipeline URL
- [x] 2.4 验证 `pipeline watch` 的 URL 打印输出（开始 + 结束 + Ctrl+C 三个时机）

## 3. 实现 watch 自动推断逻辑

- [x] 3.1 在 `cmd/watch.go` 中实现 `resolveCurrentRepoPipeline()` 函数：前置检查（git repo、remote origin）→ Level 1 配置精确匹配 → Level 2 host 匹配 → Level 3 公共域名，返回 `WatchTarget` 或详细错误
- [x] 3.2 实现 Level 1：获取 remote URL → `ExtractRemotePath` → 遍历 `resolver.Resolve()` 的所有 repo 比对 CloneURL/SSHURL 的 remotePath
- [x] 3.3 实现 Level 2：遍历 config Resources 比对 host（使用 `extractHost` + `parseResourceURL`），匹配后用 resource 的 provider + token + remote URL 的 remotePath 构造 WatchTarget
- [x] 3.4 实现 Level 3：host 为 github.com / gitlab.com 时，从环境变量获取 token
- [x] 3.5 实现详细错误信息生成函数，覆盖四种场景：非 git 仓库、无 remote origin、配置中无匹配（含 host/remotePath/remoteURL 诊断信息）、已知域名但无 token

## 4. 注册 watch 顶级命令

- [x] 4.1 在 `cmd/watch.go` 中定义 `watchCmd` cobra 命令：`Use: "watch [repo-name]"`、`Args: cobra.MaximumNArgs(1)`、`--id` flag
- [x] 4.2 实现 `runWatch`：有 repo-name 参数时走 `resolvePipelineInput` 构造 WatchTarget；无参数时走 `resolveCurrentRepoPipeline()`
- [x] 4.3 在 `cmd/watch.go` 的 `init()` 中将 `watchCmd` 注册到 `rootCmd`
- [x] 4.4 验证 `grepom watch`（无参数自动推断）、`grepom watch <repo-name>`、`grepom watch --id <id>`、`grepom watch <repo-name> --id <id>` 四种用法

## 5. 更新文档

- [x] 5.1 更新 README.md：在命令列表中添加 `watch` 命令说明和示例
- [x] 5.2 更新 README_en.md：同步添加 `watch` 命令的英文说明和示例
