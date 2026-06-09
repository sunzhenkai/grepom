## Why

`grepom svc list` 默认输出 6 列（含完整 PATH、COMMAND、LOG 路径），在 macOS 终端中极易换行、难以扫读；同时服务运行态数据存放在 `~/Library/Application Support/grepom/`，路径含空格且不符合开发者习惯的 XDG 目录布局。此外 CLI 尚无 shell 补全，输入服务名时无法 Tab 补全，影响日常使用效率。

## What Changes

- **紧凑列表**：`grepom svc list` 默认输出 4 列（NAME、STATUS、PID、PATH），与 TUI 列表视图对齐；新增 `-v`/`--verbose` 标志输出完整 6 列。
- **状态目录迁移至 XDG**：服务 registry 和日志改存 `$XDG_STATE_HOME/grepom/services/<scope>/`（未设置时 fallback 为 `~/.local/state/grepom/services/<scope>/`）；**不迁移**旧 `Application Support` 路径下的数据，旧记录需用户自行 `clean` 或手动处理。
- **Shell 补全**：新增 `grepom completion` 子命令（bash/zsh/fish），并为 `svc` 相关子命令的服务名参数提供动态补全（合并 `.grepom.yml` 配置名与 registry 中的服务名）。
- 更新中英文 README 中的路径说明、list 用法和补全安装指引。

## Capabilities

### New Capabilities

- `shell-completion`: grepom 全命令 shell 补全及 svc 服务名动态补全。

### Modified Capabilities

- `service-process-management`: list 默认表格列变更、verbose 标志、状态目录路径规范。
- `cli-commands`: 新增 `completion` 子命令。

## Impact

- 影响 `service/scope.go`：状态目录解析逻辑从 `UserConfigDir` 改为 XDG state home。
- 影响 `cmd/svc.go`：`printSvcTable` 紧凑/verbose 模式、list 标志、服务名 `ValidArgsFunction`。
- 影响 `cmd/root.go` 或新文件：注册 `completion` 子命令。
- 影响 `openspec/specs/service-process-management/spec.md` 和 `openspec/specs/cli-commands/spec.md` 需求描述。
- 影响 `README.md`、`README_en.md`。
- **BREAKING（弱）**：新启动的服务写入新路径；旧路径下的 registry 不再被读取，用户需知悉。
