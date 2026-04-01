## 1. 扩展配置结构体

- [x] 1.1 在 `config/config.go` 的 `Resource` 结构体中新增 `SSHKey string \`yaml:"ssh_key,omitempty"\`` 字段
- [x] 1.2 在 `config/config.go` 的 `Group` 结构体中新增 `SSHKey string \`yaml:"ssh_key,omitempty"\`` 和 `Token string \`yaml:"token,omitempty"\`` 字段
- [x] 1.3 在 `config/config.go` 的 `Repo` 结构体中新增 `SSHKey string \`yaml:"ssh_key,omitempty"\`` 和 `Token string \`yaml:"token,omitempty"\`` 字段
- [x] 1.4 更新 `config.Load()` 中的 token 解析逻辑，对 Group.Token 和 Repo.Token 也执行 `${ENV_VAR}` 解析
- [x] 1.5 更新 `config.writeConfig()` 确保 Group.Token、Repo.Token 和 Resource.SSHKey 的 raw 值正确保留（类似 resource token 的处理方式）

## 2. 扩展 provider.Repo 结构携带认证信息

- [x] 2.1 在 `provider/provider.go` 的 `Repo` 结构体中新增 `Token string` 和 `SSHKey string` 字段
- [x] 2.2 修改 `repo/resolver.go` 的 `Resolve()` 方法，按优先级合并认证信息：group/repo 级别 token/ssh_key → resource 级别 token/ssh_key → resource provider

## 3. 改造 git.Clone 支持认证优先级链

- [x] 3.1 在 `git/git.go` 中定义 `CloneOptions` 结构体：`Token string`、`Provider string`、`SSHKey string`
- [x] 3.2 修改 `Clone` 函数签名为 `Clone(path, sshURL, httpURL string, opts CloneOptions) error`
- [x] 3.3 新增 `buildTokenURL(httpURL, token, provider string) string` 函数，根据 provider 类型构建 token 认证 HTTPS URL（GitHub: `x-access-token:<token>@`，GitLab: `oauth2:<token>@`）
- [x] 3.4 新增 `cloneWithSSHKey(path, sshURL, sshKey string) error` 辅助函数，通过 `GIT_SSH_COMMAND` 环境变量指定 SSH key 执行 git clone
- [x] 3.5 实现 6 级认证优先级链：group/repo token → group/repo SSH key → resource token → resource SSH key → 推导 SSH → 裸 HTTP
- [x] 3.6 每种认证方式尝试时输出日志（`[N/M] 尝试 <方式> (<级别>)...`），失败输出错误摘要，成功输出 "成功"，跳过未配置的级别
- [x] 3.7 确保日志中不泄露 token 或包含 token 的完整 URL
- [x] 3.8 为 `buildTokenURL`、`cloneWithSSHKey` 和优先级链逻辑编写单元测试

## 4. 更新 clone 命令

- [x] 4.1 修改 `cmd/clone.go`，在调用 `gitpkg.Clone` 时传入 `r.Token`、`r.Provider`、`r.SSHKey` 构建的 `CloneOptions`

## 5. 更新 add 命令

- [x] 5.1 在 `cmd/add.go` 的 `addResourceCmd` 新增 `--ssh-key` flag，写入 `Resource.SSHKey`
- [x] 5.2 在 `cmd/add.go` 的 `addGroupCmd` 新增 `--ssh-key` 和 `--token` flag，写入 `Group.SSHKey` 和 `Group.Token`
- [x] 5.3 在 `cmd/add.go` 的 `addRepoCmd` 新增 `--ssh-key` 和 `--token` flag，写入 `Repo.SSHKey` 和 `Repo.Token`
- [x] 5.4 更新 `config.AddGroup` 和 `config.AddRepo` 确保新字段正确写入配置文件

## 6. 引入 survey 依赖

- [x] 6.1 执行 `go get github.com/AlecAivazis/survey/v2` 添加交互式 prompt 库依赖

## 7. 实现交互式命令框架

- [x] 7.1 新建 `cmd/interactive.go`，注册 `interactive` 子命令到 rootCmd
- [x] 7.2 实现 TTY 检测：非交互式终端运行时提示错误并退出
- [x] 7.3 实现主菜单循环：显示操作列表（初始化配置、添加资源、添加组、添加仓库、同步、克隆、查看状态、退出），使用 survey Select prompt

## 8. 实现交互式子流程

- [x] 8.1 实现交互式 init：逐步提示配置文件路径、base 目录、是否添加资源、provider 类型选择、URL、token、是否配置 SSH key，最终调用 `config.InitConfig` 和 `config.AddResource`
- [x] 8.2 实现交互式添加资源：提示资源名称、provider 选择、URL、token、是否配置 SSH key（可选），调用 `config.AddResource`
- [x] 8.3 实现交互式添加组：提示组名称、从已配置资源中选择、远程路径、本地路径、是否递归、是否配置 SSH key（可选）、是否配置 token（可选），调用 `config.AddGroup`
- [x] 8.4 实现交互式添加仓库：选择独立仓库或组内仓库，分别提示不同字段（独立仓库含 SSH key 和 token 选项），调用对应 `config.AddRepo` 或 `config.AddGroupRepo`
- [x] 8.5 实现交互式同步：提示同步范围（全部/按组/按资源），复用 sync 命令逻辑
- [x] 8.6 实现交互式克隆：提示克隆范围，复用 clone 命令逻辑（含认证优先级链和日志）
- [x] 8.7 实现交互式查看状态：提示查看范围，复用 status 命令逻辑

## 9. 测试与验证

- [x] 9.1 编写 `git/git_test.go` 中 token URL 构建的单元测试（GitHub 和 GitLab 两种 provider）
- [x] 9.2 编写 `git/git_test.go` 中 SSH key clone 的单元测试
- [x] 9.3 编写 `repo/resolver_test.go` 中 Resolver 合并认证信息的单元测试（覆盖 group/repo 覆盖 resource、resource SSH key fallback 等场景）
- [ ] 9.4 手动验证：构建项目 `go build`，测试 `grepom interactive` 各菜单项流程
- [ ] 9.5 手动验证：测试完整 6 级认证优先级链和日志输出
