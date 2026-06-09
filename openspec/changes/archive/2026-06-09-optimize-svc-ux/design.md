## Context

`grepom svc` 已实现服务启动、列表、日志、TUI 等能力。当前实现中：

- `printSvcTable` 固定输出 6 列，PATH/COMMAND/LOG 均为完整绝对路径，终端可读性差。
- TUI `listView` 仅展示 NAME、STATUS、PID、PATH 四列，与 CLI list 不一致。
- 状态目录通过 `os.UserConfigDir()` 解析，macOS 上为 `~/Library/Application Support/grepom/services/<scope>/`，路径含空格且不符合 XDG 惯例。
- 项目尚无 Cobra shell completion 支持。

用户已明确决策：紧凑 list + `-v` 全量、XDG state 目录且不迁移旧数据、全命令补全 + svc 服务名动态补全，三个优化一并交付。

## Goals / Non-Goals

**Goals:**

- `grepom svc list` 默认 4 列（NAME、STATUS、PID、PATH），`-v`/`--verbose` 输出完整 6 列。
- `grepom svc status <name>` 继续输出完整元数据（含 command、log path）。
- 状态目录改为 `$XDG_STATE_HOME/grepom/services/<scope>/`（fallback `~/.local/state/grepom/services/<scope>/`）。
- 新增 `grepom completion bash|zsh|fish`，并为 svc 服务名参数提供动态补全。
- 更新中英文 README。

**Non-Goals:**

- 不迁移 `Application Support` 或旧路径下的 registry/logs。
- 不为旧路径提供双读或自动合并。
- 不为非 svc 命令实现复杂的动态补全（如 repo 名、group 名）；除 svc 服务名外，依赖 Cobra 默认静态补全。
- 不改动 TUI 列表布局（已与目标默认 list 对齐）。

## Decisions

### 紧凑列表与 verbose 模式

`printSvcTable` 增加 `verbose bool` 参数：

```text
默认 (verbose=false):
  NAME | STATUS | PID | PATH

verbose (svc list -v):
  NAME | STATUS | PID | PATH | COMMAND | LOG
```

PATH 列对 home 目录做 `~` 前缀替换以缩短显示；verbose 模式下 LOG 仍输出完整路径（便于脚本和复制）。

`svc status <name>` 始终调用 verbose 表格或等价详情输出，不因 list 默认紧凑而缩水。

备选：单独 `svc list --wide` 而非 `-v`——用户已选定 `-v`/`--verbose`，与常见 CLI 惯例一致。

### XDG State 目录解析

在 `service/scope.go` 新增可注入的 `StateHomeFunc`，默认实现：

```go
func defaultStateHome() (string, error) {
    if v := os.Getenv("XDG_STATE_HOME"); v != "" {
        return v, nil
    }
    home, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(home, ".local", "state"), nil
}
```

`StateDir(scopeID)` 改为 `filepath.Join(stateHome, "grepom", "services", scopeID)`。

移除对 `UserConfigDirFunc` 的依赖（或保留变量但不再用于 StateDir）。测试通过注入 `StateHomeFunc` 指向 temp 目录。

Windows：若 `XDG_STATE_HOME` 未设置，fallback 为 `%LOCALAPPDATA%\grepom\state` 或 `filepath.Join(home, ".local", "state")`——与 Neovim 在 Windows 上的 data/state 合并策略类似，保持单一实现路径。

不迁移：旧 registry 留在原路径，新 scope 解析只读新路径；用户需对旧服务执行 `kill`/`clean` 或手动删除旧目录。README 中简短说明。

### Shell 补全架构

使用 Cobra 内置 completion：

1. 在 `cmd/completion.go`（或 `cmd/root.go`）注册：

```go
rootCmd.AddCommand(completionCmd) // cobra.CompletionCmd 或自定义包装
```

2. svc 服务名补全函数 `completeSvcNames`：

```text
resolveServiceManager() → mgr
  ├─ registry 中的服务名 (mgr.List)
  └─ .grepom.yml services 键 (mgr.Services)
合并去重、字典序排序后返回
```

3. 注册到需要服务名的命令：

| 命令 | 参数位置 |
|------|----------|
| `svc logs` | Arg 0 |
| `svc kill` | Arg 0 |
| `svc dir` | Arg 0 |
| `svc status` | Arg 0（可选参数） |
| `svc run` | Arg 0（可选，配置名） |

通过 `cmd.RegisterFlagCompletionFunc` 仅用于有动态 flag 值的场景；服务名用 `ValidArgsFunction`。

4. `svc` 和 `service` 别名命令共享同一套 `ValidArgsFunction`（在 `registerSvcSubcommands` 中统一挂载）。

补全失败时（无配置、无 registry）返回 `cobra.ShellCompDirectiveNoFileComp` 且不报错，避免 Tab 时打印错误到终端。

### 文档

README 补充：

- 状态目录位置说明（XDG_STATE_HOME / ~/.local/state/grepom）
- 旧路径不迁移的简短提示
- `grepom svc list -v` 用法
- Shell 补全安装示例（zsh/bash）

## Risks / Trade-offs

- **[弱 BREAKING] 旧 registry 不可见** → README 说明；用户 clean 旧数据后在新路径重新 run。
- **XDG 目录不存在** → `Run`/`StateDir` 时 `MkdirAll` 确保创建（现有逻辑应已有 mkdir，需确认 registry 写入路径）。
- **补全性能** → 每次 Tab 触发一次 manager 解析；scope 内服务数量有限，可接受；避免在补全中做进程探测。
- **PATH 的 `~` 缩短** → 仅影响显示，不改变 registry 存储的绝对路径；无 home 时原样输出。

## Migration Plan

无自动迁移。发布说明 / README 提示：

1. 升级后 `grepom svc list` 可能显示为空（新路径无 registry）。
2. 旧服务需手动停止（若仍在运行）并删除 `~/Library/Application Support/grepom/` 下对应目录（可选）。
3. 重新 `grepom svc run ...` 在新路径创建记录。

回滚：还原代码后新路径下的 registry 不再被读取，影响与升级对称。

## Open Questions

无——用户已确认全部关键决策。
