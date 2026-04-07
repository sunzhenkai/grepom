## 1. generic provider 实现

- [x] 1.1 新建 `provider/generic.go`，实现 `Provider` 接口，`ListRepos` 和 `ListGroups` 返回空列表
- [x] 1.2 在 `generic.go` 的 `init()` 中调用 `Register("generic", ...)` 注册 provider

## 2. 配置验证更新

- [x] 2.1 修改 `config/config.go` 中的 provider 验证逻辑，改为从 `provider.AvailableProviders()` 动态获取合法列表
- [x] 2.2 更新验证错误信息，列出所有支持的 provider 名称

## 3. git 认证映射更新

- [x] 3.1 在 `git/git.go` 的 token 用户名 switch 中为 `generic` provider 添加 `token` 用户名映射

## 4. example 命令实现

- [x] 4.1 新建 `cmd/example.go`，定义包含全部字段和 YAML 注释的示例配置常量字符串（含 github、gitlab、generic 三种 provider）
- [x] 4.2 实现 `grepom example` 命令，默认输出到 stdout
- [x] 4.3 添加 `--output` / `-o` 标志，支持写入文件（文件已存在时报错）
- [x] 4.4 在 `cmd/root.go` 中注册 `example` 子命令

## 5. 验证

- [x] 5.1 运行 `go build ./...` 确认编译通过
- [x] 5.2 手动验证 `grepom example` 输出包含三种 provider 示例
- [x] 5.3 手动验证配置文件中使用 `provider: generic` 可通过验证，`provider: unknown` 报错
