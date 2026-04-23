## 1. 核心方法：stripQuotes + ResolvedToken

- [x] 1.1 在 `config/config.go` 中新增 `stripQuotes(s string) string` 函数：当字符串首尾为成对的单引号或双引号时去除，否则原样返回
- [x] 1.2 修改 `ResolveToken()` 函数：在正则匹配前先调用 `stripQuotes(token)` 清理引号
- [x] 1.3 在 `config/config.go` 中为 `Resource` 类型新增 `ResolvedToken() (string, error)` 方法，内部调用 `ResolveToken(r.Token)`

## 2. 修复 cmd 包中的直接 token 使用

- [x] 2.1 修改 `cmd/sync.go:101`：将 `Token: res.Token` 改为先调用 `res.ResolvedToken()` 获取已解析 token，处理错误后传入 `params`
- [x] 2.2 修改 `cmd/list.go:316`（`runListRemoteRepos`）：将 `Token: res.Token` 改为先调用 `res.ResolvedToken()`，处理错误
- [x] 2.3 修改 `cmd/list.go:428`（`runListRemoteGroups`）：将 `Token: rq.res.Token` 改为先调用 `rq.res.ResolvedToken()`，处理错误
- [x] 2.4 修改 `cmd/pipeline.go:96`（`resolvePipelineInput`）：将 `res.Token` 改为调用 `res.ResolvedToken()`，处理错误
- [x] 2.5 修改 `cmd/interactive.go:675`：将 `Token: res.Token` 改为先调用 `res.ResolvedToken()`，处理错误

## 3. 统一 Resolver 中的 token 解析方式

- [x] 3.1 修改 `repo/resolver.go` group 循环中的 resource 默认 token 获取：将 `config.ResolveToken(token)` 改为对 resource 默认 token 使用 `res.ResolvedToken()`，group/repo override token 仍使用 `config.ResolveToken()`（已内置引号清理）
- [x] 3.2 修改 `repo/resolver.go` standalone repo 循环中的 token 获取，同上逻辑

## 4. 测试

- [x] 4.1 新增 `stripQuotes` 函数测试：覆盖成对单引号、双引号、不对称引号、空字符串、无引号、单字符等情况
- [x] 4.2 修改 `TestResolveToken_*` 系列测试：新增引号包裹场景（`'${MY_TOKEN}'`、`"${MY_TOKEN}"`、`"glpat-xxx"`）
- [x] 4.3 新增 `TestResolvedToken_*` 系列测试：验证 `Resource.ResolvedToken()` 的环境变量解析、明文透传、未设置报错、空值返回
- [x] 4.4 修改 `TestResolve_LazyTokenResolution_*` 测试：验证 Resolver 仍通过 `ResolvedToken` 正确解析
- [x] 4.5 运行全部测试确保无回归
