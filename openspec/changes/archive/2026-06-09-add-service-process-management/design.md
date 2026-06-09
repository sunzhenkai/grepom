## Context

`grepom` 当前是 Go/Cobra CLI，核心能力围绕 `.grepom.yml` 中声明的仓库、组和资源展开。已有 `dir` 命令和 `gcd()` shell helper 能把仓库名解析为本地路径，但没有管理本地开发服务运行态的能力。

服务进程管理属于本机开发时的运行态管理：它需要记录某个目录下用某条命令启动的进程、日志位置和当前状态。运行态数据不应写回 `.grepom.yml`，否则会把机器相关的 PID、日志路径和临时状态混入声明式仓库配置。

本变更还引入 TUI 管理界面。TUI 不应拥有独立业务逻辑，而应复用 CLI 子命令背后的服务管理接口，避免 CLI 和 TUI 状态判断不一致。

## Goals / Non-Goals

**Goals:**

- 提供 `grepom svc`/`grepom service` 子命令，覆盖启动、列表、状态、日志、停止、清理、目录定位和 TUI 管理。
- 支持命令行直接传入启动命令，也支持从 `.grepom.yml` 的 `services` 配置读取服务定义。
- 默认后台运行服务，并将 stdout/stderr 写入服务日志文件。
- 使用表格展示服务列表，包含名称、状态、PID、路径、命令和日志路径。
- 在 `list`/`status` 时检查真实进程状态，识别 running、exited、stale 等状态。
- 使用进程组管理服务，普通 kill 和强制 kill 都尽量作用于整个服务树。
- 提供类似 `tail -n`、`tail -f` 和编辑器打开日志的能力。
- 提供 TUI 管理界面，支持浏览服务、查看日志、停止服务、清理退出服务和获取服务路径。
- 更新中英文 README，说明配置和命令用法。

**Non-Goals:**

- 不实现守护进程、自动重启、健康检查、端口探测或服务依赖编排。
- 不支持同名服务的多实例并发管理；同名服务再次启动默认应阻止或要求显式覆盖。
- 不把运行态 PID、日志路径和退出状态写回 `.grepom.yml`。
- 不在第一版承诺 Windows 上完整的进程组语义；可提供降级行为或明确错误。
- 不把 TUI 发展为完整交互式仓库管理替代品；它只覆盖服务管理。

## Decisions

### 运行态独立存储

服务运行态写入用户状态目录，并按配置文件路径隔离。建议路径形态：

```text
<user-state-dir>/grepom/services/<config-hash>/
  registry.json
  logs/
    <service-name>.log
```

`user-state-dir` 优先使用 `os.UserConfigDir`/`os.UserCacheDir`/平台约定时需保持可测试；实现中可封装为可注入函数。状态目录按 `.grepom.yml` 绝对路径 hash 隔离，避免不同工作区服务同名冲突。没有配置文件的直接目录运行，可以使用当前工作目录绝对路径 hash。

备选方案是把状态写入项目目录的 `.grepom/`。这方便跟项目放在一起，但会污染仓库、容易被误提交，也不适合记录本机 PID，所以不采用。

### 配置只声明服务定义

`.grepom.yml` 新增可选顶层字段：

```yaml
services:
  api:
    cwd: ./backend
    command: make dev
  web:
    cwd: ./frontend
    command:
      - pnpm
      - dev
```

`cwd` 相对配置文件所在目录解析。`command` 支持字符串和字符串数组：字符串通过 shell 执行，适合开发命令中的环境变量、管道和复合命令；数组直接执行，适合精确参数传递。

备选方案是只支持数组，解析最稳但配置体验差；只支持字符串，体验好但测试和转义更复杂。两者都支持能覆盖更多开发场景。

### 服务记录模型

registry 中每条记录包含：

- `name`
- `pid`
- `pgid`
- `cwd`
- `command`
- `command_args`
- `log_path`
- `started_at`
- `last_status`
- `exit_status`
- `config_path`

`list`/`status` 读取 registry 后实时探测进程状态，而不是只展示上次写入的 `last_status`。MVP 可用 `kill(pid, 0)` 判断进程是否存在，并结合保存的命令、cwd、启动时间能力逐步增强 PID 复用识别。

### 进程组优先

启动服务时创建独立进程组。`kill` 默认发送 `SIGTERM` 到进程组，`kill -9` 发送 `SIGKILL` 到进程组。若平台或启动方式无法获取进程组，则降级为对 PID 发信号，并在输出中说明。

这样可以覆盖 `make dev`、`pnpm dev`、热加载工具等会启动子进程的常见场景。

### 命令行形态

主入口为：

```bash
grepom svc run [name] -- <command> [args...]
grepom svc run [name]
grepom svc list
grepom svc status [name]
grepom svc logs [name]
grepom svc logs -n 200 [name]
grepom svc logs -f [name]
grepom svc logs --open [name]
grepom svc kill [name]
grepom svc kill -9 [name]
grepom svc clean
grepom svc dir [name]
grepom svc tui
```

`service` 是 `svc` 的别名。`run` 无显式名称时使用当前目录名；有名称但无命令时从配置读取该服务；无配置命令时返回可理解的错误。可以保留 `-r/--run` 作为根命令快捷方式，但文档主推 `svc run -- ...`，避免空格命令解析歧义。

### 日志读取与打开

服务启动时将 stdout 和 stderr 追加写入同一个日志文件，并在每次启动前写入一段启动分隔信息。`logs` 默认输出末尾若干行，`-n` 控制行数，`-f` 持续跟随文件新增内容。`--open` 使用 `$VISUAL`、`$EDITOR`、平台默认打开命令的优先级打开日志；都不可用时打印日志路径。

### TUI 复用服务管理接口

新增内部服务管理接口，例如：

```go
type Manager interface {
    Run(...)
    List(...)
    Status(...)
    Logs(...)
    Kill(...)
    Clean(...)
    Dir(...)
}
```

CLI 和 TUI 都调用同一实现。TUI 展示服务表格，支持键盘选择服务、刷新状态、查看日志尾部、停止服务、强制停止服务、清理退出服务和复制/打印服务路径。TUI 依赖应集中在 UI 包中，避免污染核心管理逻辑。可选依赖优先考虑 Bubble Tea 生态；若引入新依赖，需要在任务中包含依赖评估和 README 更新。

## Risks / Trade-offs

- PID 复用导致状态误判 → 保存更多进程身份信息，并将 MVP 状态标记为 best-effort；后续可按平台读取启动时间增强判断。
- 进程组行为在不同平台不一致 → Unix 平台完整支持，其他平台降级并在文档注明。
- shell 字符串命令存在转义和注入风险 → 只执行用户本地显式配置或命令；数组命令作为精确执行选项；文档说明差异。
- 日志文件持续增长 → MVP 提供日志路径和清理选项；日志轮转不纳入第一版。
- TUI 增加依赖和测试复杂度 → 核心逻辑保持无 TUI 依赖，TUI 只做薄交互层，优先测试 manager 和命令行为。
- 同名服务冲突 → registry 按 config/cwd 隔离，同一 registry 内同名服务默认唯一；重复启动时给出明确错误。
