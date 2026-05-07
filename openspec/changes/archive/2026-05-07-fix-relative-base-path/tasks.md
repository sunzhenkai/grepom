## 1. 核心实现

- [x] 1.1 在 `config/config.go` 中新增 `ResolveBasePath(cfg *Config, configDir string)` 函数：若 `cfg.Base` 非绝对路径，使用 `filepath.Join(configDir, cfg.Base)` 解析为绝对路径
- [x] 1.2 在 `cmd/root.go` 的 `loadConfig` 函数中，`config.Load` 之后调用 `config.ResolveBasePath(cfg, filepath.Dir(absPath))`

## 2. 单元测试

- [x] 2.1 在 `config/config_test.go` 中新增 `TestResolveBasePath` 测试组：覆盖相对路径、`./` 前缀、绝对路径、`~/` 路径四种场景
- [x] 2.2 在 `cmd/dir_test.go` 中新增 `TestDirCommand_RelativeBaseFromSubdir`：使用相对路径 `base` 配置，从子目录执行 `grepom dir`，验证输出为绝对路径

## 3. 验证

- [x] 3.1 运行 `go test ./config/... ./cmd/... -v` 确保所有测试通过
- [x] 3.2 手动构建并验证：使用相对路径 `base` 的配置，从项目子目录执行 `gcd <repo>` 确认跳转成功
