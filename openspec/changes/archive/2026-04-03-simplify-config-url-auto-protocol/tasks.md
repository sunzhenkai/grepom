## 1. Config URL 解析改造

- [x] 1.1 在 `config/config.go` 中新增 `stripScheme(url string) string` 函数，剥离 `http://` 和 `https://` 前缀，返回纯 host:port
- [x] 1.2 修改 `config.go` 的 `validate()` 方法，将 resource `url` 统一转换为 host:port 格式存储（替换原来的 `https://` 前缀补全逻辑）
- [x] 1.3 新增 `Resource.APIURL() string` 方法，返回 `https://` + host 用于 provider API 调用
- [x] 1.4 新增 `Resource.SSHURL(path string) string` 方法，返回 `git@<host>:<path>.git`
- [x] 1.5 新增 `Resource.HTTPSURL(path string) string` 方法，返回 `https://<host>/<path>.git`
- [x] 1.6 新增 `Resource.HTTPURL(path string) string` 方法，返回 `http://<host>/<path>.git`
- [x] 1.7 更新 `config_test.go` 中的 URL 相关测试用例，覆盖新旧格式

## 2. Resolver 适配

- [x] 2.1 修改 `repo/resolver.go` 的 `deriveSSHURL()` 函数，改为接收 host:port 而非 HTTPS URL
- [x] 2.2 更新 resolver 中对 `CloneURL` 的构建逻辑，使用 `Resource.HTTPSURL(path)` 替代直接存储的 URL
- [x] 2.3 更新 resolver 中对 `SSHURL` 的构建逻辑，使用 `Resource.SSHURL(path)`
- [x] 2.4 更新 `resolver_test.go` 中的相关测试用例

## 3. Provider API URL 适配

- [x] 3.1 修改 `provider/gitlab.go` 中 `ListRepos` 使用 `params.ServerURL`（已由调用方确保为 HTTPS API URL）
- [x] 3.2 修改 `provider/github.go` 中 `ListRepos` 使用 `params.ServerURL`
- [x] 3.3 检查 `cmd/sync.go` 中构建 `ListReposParams.ServerURL` 的逻辑，确保使用 `Resource.APIURL()`

## 4. Clone 认证链改造

- [x] 4.1 修改 `git/git.go` 的 `buildAuthStrategies()` 函数，移除策略 #6（裸 HTTP URL）
- [x] 4.2 修改 `Clone()` 函数的日志输出，在策略标签后追加打印实际 URL（token URL 中的 token 用 `***` 替代）
- [x] 4.3 新增 `maskTokenURL(url string) string` 辅助函数，将 token URL 中的 token 部分替换为 `***`
- [x] 4.4 更新 `git_test.go` 中的认证策略相关测试

## 5. 集成测试与验证

- [x] 5.1 运行全部单元测试 `go test ./...`，确保通过
- [x] 5.2 手动验证：使用仅 host 的配置文件执行 `grepom clone`，确认日志打印实际 URL
- [x] 5.3 手动验证：使用含 `https://` 前缀的旧格式配置文件，确认仍能正常工作
- [x] 5.4 确认无 token 且无 SSH 时 clone 直接报错，不会触发交互式认证提示
