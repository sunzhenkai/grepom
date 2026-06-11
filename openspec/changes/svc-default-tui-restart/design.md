## Context

当前 `grepom svc` 不带子命令时仅打印帮助文本（`cmd.Help()`），用户必须显式执行 `grepom svc tui` 才能进入 TUI。TUI 列表视图支持 stop（`s`）、kill-9（`S`）、clean（`c`）等操作，但缺少 restart 快捷键，而 CLI 层的 `Manager.Restart()` 方法已完整实现并测试通过。

本次变更涉及两个独立但相关的改进：
1. 将 `svc` 无参数的默认行为从帮助文本改为启动 TUI
2. 在 TUI 中新增 restart 操作

## Goals / Non-Goals

**Goals:**
- `grepom svc` 无参数时直接启动 TUI，降低用户操作步骤
- `grepom svc --help` 仍然正常显示帮助信息
- TUI 列表视图中支持 `R`（大写）快捷键重启选中服务
- TUI 底栏帮助文本显示新的 restart 快捷键提示
- 复用已有的 `Manager.Restart()` 方法，无新增依赖

**Non-Goals:**
- 不修改 `service` alias 的行为（它共享 `svcCmd.RunE`，会自动继承）
- 不添加 TUI 中的批量 restart 功能
- 不修改 `Manager.Restart()` 的内部实现逻辑
- 不添加 restart 前的确认提示（与 kill 操作保持一致的无确认模式）

## Decisions

### 1. 默认 TUI 的实现方式：复用 `runSvcTui` 函数

**选择**：修改 `svcCmd.RunE`，在非 `--shell` 模式下直接调用 `runSvcTui` 的核心逻辑。

**替代方案**：
- 方案 B：移除 `svcCmd.RunE`，让 Cobra 报 "unknown subcommand" 错误 — 用户体验差
- 方案 C：新建子命令 `_default` 作为隐式默认 — 过度设计

**理由**：`runSvcTui` 已包含 TTY 检测（`EnsureTTY`）、Manager 初始化和 TUI 启动的完整流程，直接复用最简洁。`--help` 由 Cobra 自动处理，不受 `RunE` 影响。

### 2. Restart 快捷键选择 `R`（大写）

**选择**：使用大写 `R`。

**理由**：
- 小写 `r` 已被 refresh 占用
- 与现有的大小写模式一致（`s` stop / `S` kill-9），`R` restart 符合直觉
- `Manager.Restart()` 是重型操作（kill + run），大写键暗示更强的操作

### 3. Restart 后的 UI 行为

**选择**：restart 成功后设置 `m.message` 并刷新列表（与 kill 操作保持一致的 pattern）。

**理由**：复用 `kill()` 方法的 message + refresh 模式，保持代码一致性。失败时通过 `m.message` 显示错误信息。

## Risks / Trade-offs

- **[行为变更]** 无参数 `svc` 从打印帮助变为启动 TUI，可能影响脚本中 `grepom svc` 的使用 → 缓解：脚本通常使用子命令（`svc list` 等），不会受影响；如有需要可通过 `svc --help` 获取帮助
- **[TTY 检测]** 在无 TTY 环境（如 CI/CD）中 `EnsureTTY()` 会报错 → 缓解：已有 TTY 检测机制，错误信息明确（"svc tui requires a TTY"）
- **[快捷键冲突]** 未来可能引入更多操作导致快捷键不够用 → 缓解：当前键位空间充足，如需扩展可引入多键组合或 `:` 命令模式
