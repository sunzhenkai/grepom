## Why

创建 tag 后，开发者通常需要手动执行 `grepom watch` 或 `grepom pipeline watch` 来监控 CI/CD pipeline 的运行状态。这是一个高频的连续操作：打 tag → 推送 → 等待 pipeline 结果。将 watch 集成到 tag 命令中可以减少操作步骤，实现一键打标并监控。

## What Changes

- `tag` 命令新增 `-w/--watch` 参数，在 tag 创建并推送后自动触发 pipeline watch
- `-w` 仅在成功创建 tag 后生效；如果 tag 创建失败或被 `--dry-run` 跳过，则不进入 watch
- `-w` 复用现有 `resolveCurrentRepoPipeline()` 三级 fallback 推断 repo 信息，与 `grepom watch` 无参数时的行为一致
- watch 过程中按 Ctrl+C 的退出行为与 `grepom watch` 完全一致

## Capabilities

### New Capabilities
- `tag-watch`: tag 命令的 `-w/--watch` 参数，在 tag 创建和推送成功后自动监控 CI/CD pipeline 状态

### Modified Capabilities
- `watch-shortcut`: 补充说明 tag 命令的 `-w` 参数也复用 `resolveCurrentRepoPipeline()` 和 `runWatchLoop()` 逻辑

## Impact

- **代码变更**: `cmd/tag.go` — 新增 `-w/--watch` flag，修改 `runTag` 函数在成功路径末尾添加 watch 逻辑
- **共享逻辑**: 复用 `cmd/watch.go` 中的 `resolveCurrentRepoPipeline()` 和 `cmd/pipeline.go` 中的 `runWatchLoop()`
- **文档变更**: README 多语言文档需同步更新，添加 `-w` 参数说明