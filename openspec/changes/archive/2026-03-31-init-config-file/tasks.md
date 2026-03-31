## 1. 命令迁移

- [x] 1.1 将 `cmd/init.go` 重命名为 `cmd/clone.go`，将命令名从 `init` 改为 `clone`，更新 Short/Long/Example 文本
- [x] 1.2 重写 `cmd/init.go`，实现配置文件初始化逻辑（创建最小 `.grepom.yml`，支持 `--base` flag）

## 2. init 命令：source 支持

- [x] 2.1 为 `init` 命令添加 `--provider`、`--url`、`--group`、`--org`、`--recursive` flag
- [x] 2.2 实现 init 时可选创建第一个 source 条目的逻辑

## 3. 兼容性处理

- [x] 3.1 为 `init` 命令添加位置参数检测，当用户传入 `[name]` 时提示 "did you mean `grepom clone`?"
- [x] 3.2 更新 `root.go` 的 Example 文本，将 `init` 示例改为 `clone`

## 4. 配置层支持

- [x] 4.1 在 `config/config.go` 中添加 `InitConfig(path, base string)` 函数，创建最小配置文件并处理已存在检测
