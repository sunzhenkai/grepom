## 1. 配置包增强

- [x] 1.1 在 `config/config.go` 中定义 `ErrConfigNotFound` 错误变量，修改 `FindConfig` 函数在找不到配置文件时返回 `ErrConfigNotFound`（而非普通 fmt.Errorf）
- [x] 1.2 在 `config/config.go` 中添加 `IsConfigNotFound(err error)` 辅助函数，使用 `errors.Is` 判断

## 2. scan 命令无配置回退

- [x] 2.1 修改 `cmd/scan.go` 的 `runScan` 函数：当 `loadConfig()` 返回配置文件不存在错误时，回退到扫描当前目录模式
- [x] 2.2 在回退模式下，在 stderr 输出提示信息 "Scanning current directory (no config file found)..."
- [x] 2.3 在回退模式下，调用 `scanner.ScanDir(ctx, ".")` 扫描当前目录，复用现有的 `outputTable` / `outputJSON` 输出结果
- [x] 2.4 确保回退模式下 `--format`、`--output`、`--gitleaks-config` 标志仍然生效

## 3. push 命令实现

- [x] 3.1 在 `git/git.go` 中添加 `Push(path string, args ...string)` 函数，在指定目录执行 `git push` 及额外参数
- [x] 3.2 创建 `cmd/push.go`，定义 `pushCmd` cobra 命令，注册 `-f`/`--force` 和 `--gitleaks-config` 标志
- [x] 3.3 实现 `runPush` 函数：检测当前目录是否为 git 仓库，调用 `scanner.ScanDir` 扫描当前目录
- [x] 3.4 当发现敏感信息时，调用 `outputTable` 输出详情，无 `-f` 时退出码非零拒绝推送
- [x] 3.5 当发现敏感信息且有 `-f` 时，打印警告信息后继续执行 `git push`
- [x] 3.6 当未发现敏感信息时，直接执行 `git push`
- [x] 3.7 在 `cmd/push.go` 的 `init()` 中将 `pushCmd` 注册到 `rootCmd`

## 4. 测试

- [x] 4.1 为 `config.IsConfigNotFound` 编写单元测试
- [x] 4.2 为 `git.Push` 函数编写单元测试
- [x] 4.3 验证 `grepom scan` 在无配置文件时正确扫描当前目录
- [x] 4.4 验证 `grepom push` 在有/无敏感信息时行为正确
- [x] 4.5 验证 `grepom push -f` 强制推送时打印警告信息
