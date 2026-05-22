## Context

grepom 的 `tag` 命令用于在 git 仓库中创建版本标签（v 版本或 t 版本），支持推送、dry-run 和 annotated tag 等选项。`watch` 和 `pipeline watch` 命令支持实时监控 CI/CD pipeline 状态。

当前痛点：开发者执行 `grepom tag -p` 推送标签后，需手动再执行 `grepom watch` 或 `grepom pipeline watch <repo>` 来查看 pipeline 是否通过。这是两个连续的高频操作，可以合并为一步。

现有架构：
- `cmd/tag.go` — tag 命令入口，`runTag` → `runVTag` / `runTTag` → `handlePush`
- `cmd/watch.go` — 顶级 watch 命令，通过 `resolveCurrentRepoPipeline()` 推断 repo，调用 `runWatchLoop()`
- `cmd/pipeline.go` — `runWatchLoop()` 共享 watch 循环逻辑

## Goals / Non-Goals

**Goals:**
- 在 `tag` 命令中新增 `-w/--watch` 参数，tag 创建并推送成功后自动触发 pipeline watch
- 复用现有 `resolveCurrentRepoPipeline()` 和 `runWatchLoop()` 逻辑，不作重复实现
- `-w` 的行为与 `grepom watch` 一致（三级 fallback、Ctrl+C 优雅退出）

**Non-Goals:**
- 不改变 `watch` 或 `pipeline watch` 命令本身的行为
- 不支持在 `-w` 模式下指定 `--id`（tag 后总是 watch 最新 pipeline）
- 不改变 tag 的版本计算逻辑

## Decisions

### 决策 1：`-w` 仅在 tag 创建成功后触发

**选择**: 只有 tag 成功创建（非 `--dry-run`）后才进入 watch；如果创建失败或 dry-run 跳过，则不进入 watch。

**理由**: 如果 tag 没有实际创建，就不会触发 pipeline，watch 没有对象可监控。dry-run 模式属于预览，不应产生后续副作用。

**备选方案**: 无论成功失败都进入 watch — 被排除，因为失败的 tag 不会触发 pipeline。

### 决策 2：`-w` 不要求 `-p`（推送），但无推送时 watch 可能监控不到新 pipeline

**选择**: `-w` 独立于 `-p`，不做强制绑定。用户可以 `tag -w` 不推送，watch 会监控当前最新 pipeline。

**理由**: 某些场景下 tag 可能已被 push 到远端（通过其他方式），用户只想 watch 已有的 pipeline。强制绑定 `-p` 会限制灵活性。

**备选方案**: `-w` 隐含 `-p` — 被排除，因为用户可能已经通过其他方式推送了 tag。

### 决策 3：复用 `resolveCurrentRepoPipeline()` 推断 repo

**选择**: 使用与 `grepom watch`（无参数）相同的三级 fallback 推断逻辑。

**理由**: tag 命令在本地 git 仓库中运行，自然有当前目录的 git 上下文。复用现有逻辑减少代码重复。

**备选方案**: 要求用户在 `-w` 后指定 repo name — 被排除，因为 tag 命令已经在仓库中运行，无需额外指定。

### 决策 4：使用 `cmd.RootCmd` 的 `cobra.Command` 作为 watch 上下文

**选择**: 构造一个 `cobra.Command` 上下文给 `runWatchLoop()`，复用其 context 和 flag 机制。

**理由**: `runWatchLoop()` 需要 `cobra.Command` 来获取 context（用于 signal handling）。在 tag 命令中，可以复用 tag 命令自身的 `cmd` 参数。

## Risks / Trade-offs

- **[风险] 自动推断失败** → 缓解：复用 `resolveCurrentRepoPipeline()` 已有的详细错误提示（三级 fallback 失败时的诊断信息），用户可以根据提示添加配置。
- **[风险] 用户使用 `-w` 但忘记推送 tag** → 缓解：watch 会显示最新的 pipeline 状态，用户可以自行判断。不做强制绑定以保持灵活性。
- **[权衡] 不支持 `--id`** → tag 创建后总是 watch 最新 pipeline，无法指定特定 pipeline ID。这简化了实现，且 tag 场景下通常只需要最新 pipeline。

## Migration Plan

无需迁移。新增的 `-w/--watch` 是纯增量参数，不影响现有 tag 命令的任何行为。

## Open Questions

无。