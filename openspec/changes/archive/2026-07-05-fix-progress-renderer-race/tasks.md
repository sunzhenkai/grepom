## 1. 并发安全：为 ProgressRenderer 加互斥锁

- [x] 1.1 在 `cmd/progress.go` 的 `ProgressRenderer` 结构体增加 `mu sync.Mutex` 字段。
- [x] 1.2 在 `Handle` 方法入口加锁（`p.mu.Lock(); defer p.mu.Unlock()`），覆盖 `inflight` 切片增删、`completed` 赋值及 `renderTTY` 调用。
- [x] 1.3 在 `Done` 方法入口加锁，覆盖光标复位与逐行清除的整段 stdout 写入。
- [x] 1.4 确认 `p.completed` 仅在 `ProgressComplete` 分支内更新，`ProgressStart` 分支不修改计数（保证单调不减）。

## 2. 渲染修复：清除行数缩减时的残留旧行

- [x] 2.1 重构 `renderTTY`：在重绘前记录 `prev := p.rendered`，重绘 `lines = 1 + len(inflight)` 行。
- [x] 2.2 当 `lines < prev` 时，对多余的 `prev - lines` 行继续输出 `\n` + 空行（pad 至 `maxWidth`）以覆盖旧内容。
- [x] 2.3 重绘结束后将光标移回进度区域第一行起点（确保下次重绘基准一致），再更新 `p.rendered = lines`。
- [x] 2.4 复核 `Done` 的清除逻辑与新 `rendered` 语义保持一致（按当前 `rendered` 行数清除并换行）。

## 3. 测试：固化并发安全与残留行清除行为

- [x] 3.1 新增 `cmd/progress_test.go`（若不存在），构造 `ProgressRenderer` 并以多 goroutine 并发触发 Start/Complete 事件，验证无 panic、无数据竞争（`go test -race`）。
- [x] 3.2 新增测试：TTY 模式下捕获缓冲 stdout，断言 in-flight 行数减少后旧仓库名被空行覆盖（残留行清除场景）。
- [x] 3.3 新增测试：非 TTY 模式下并发触发多个 Complete 事件，断言每行输出完整不交错。
- [x] 3.4 新增测试：高并发（如 50 worker × 200 任务）下，捕获的 `[N/M]` 计数序列单调不减。
- [x] 3.5 运行 `go test ./cmd/... -race` 与 `go test ./git/...` 全量通过，确认无回归。

## 4. 文档与验收

- [x] 4.1 更新 `README.md` 与 `README_en.md` 中并行进度展示相关说明，补充并发安全/稳定渲染的行为描述。
- [x] 4.2 运行 `openspec validate fix-progress-renderer-race --strict` 通过校验。
- [x] 4.3 人工/脚本验证：对一批仓库执行 `grepom pull -j 8`，确认进度区域无错位、计数不回退。
