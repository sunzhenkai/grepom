## 1. 前缀匹配

- [x] 1.1 在 `cmd/root.go` 的 `init()` 中启用 `cobra.EnablePrefixMatching = true`

## 2. interactive 模式 provider 支持 generic

- [x] 2.1 修改 `interactiveInit` 中 provider 选择列表，从 `["gitlab", "github"]` 改为 `["gitlab", "github", "generic"]`
- [x] 2.2 修改 `interactiveInit` 中默认 URL 推导逻辑，`generic` 不设默认 URL
- [x] 2.3 修改 `interactiveInit` 中资源名自动推导逻辑，`generic` 默认资源名为 `generic`
- [x] 2.4 修改 `interactiveAddResource` 中 provider 选择列表，从 `["gitlab", "github"]` 改为 `["gitlab", "github", "generic"]`
- [x] 2.5 修改 `interactiveAddResource` 中默认 URL 推导逻辑，`generic` 不设默认 URL

## 3. add resource 命令 provider 验证

- [x] 3.1 修改 `cmd/add.go` 中 `addResourceCmd` 的 provider 验证，支持 `generic`
- [x] 3.2 更新 provider 验证错误信息和帮助文本，列出三种 provider

## 4. 验证

- [x] 4.1 运行 `go build ./...` 确认编译通过
- [x] 4.2 验证 `grepom cl` 前缀匹配执行 `clone`
- [x] 4.3 验证 `grepom s` 歧义前缀报错
