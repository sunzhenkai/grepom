## 1. 命令行参数

- [x] 1.1 在 `cmd/tag.go` 中新增 `tagWatch` bool 变量，并在 `init()` 中注册 `-w/--watch` flag
- [x] 1.2 更新 `tagCmd` 的 `Short`、`Long` 和 `Example` 文档字符串，说明 `-w` 参数的用途和用法

## 2. 核心逻辑实现

- [x] 2.1 修改 `runTag` 函数：在 `runVTag` / `runTTag` 成功返回后，检查 `tagWatch` 标志，若为 true 则调用 `resolveCurrentRepoPipeline()` 获取 `WatchTarget`，再调用 `runWatchLoop(target, 0, cmd)` 监控最新 pipeline
- [x] 2.2 确保 `--dry-run` 模式下不触发 watch（`runVTag` / `runTTag` 在 dry-run 时直接返回 nil，需在 watch 调用前检查 `tagDryRun`）
- [x] 2.3 处理 `resolveCurrentRepoPipeline()` 失败的情况：tag 创建成功但 repo 推断失败时，输出成功创建的消息，然后显示详细错误信息并以非零退出码退出
- [x] 2.4 使用 `cmd`（tag 命令的 `cobra.Command`）作为 `runWatchLoop()` 的参数，确保 signal handling 正常工作

## 3. 测试

- [x] 3.1 编写 `cmd/tag_watch_test.go` 单元测试：验证 `-w` flag 注册正确
- [x] 3.2 编写集成测试场景：`tag -w` 在 dry-run 模式不进入 watch
- [x] 3.3 编写集成测试场景：`tag -w` 在 tag 创建失败时不进入 watch

## 4. 文档更新

- [x] 4.1 更新 README（中文版）中 tag 命令的使用说明，添加 `-w/--watch` 参数描述
- [x] 4.2 更新 README（英文版）中 tag 命令的使用说明，添加 `-w/--watch` 参数描述