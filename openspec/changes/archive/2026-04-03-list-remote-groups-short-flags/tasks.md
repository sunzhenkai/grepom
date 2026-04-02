## 1. Provider 接口扩展

- [x] 1.1 在 `provider/provider.go` 中新增 `RemoteGroup` 结构体（Name、Path、Provider 字段）
- [x] 1.2 在 `provider/provider.go` 中新增 `ListGroupsParams` 结构体（ServerURL、Token 字段）
- [x] 1.3 在 `Provider` 接口中新增 `ListGroups(ctx context.Context, params ListGroupsParams) ([]RemoteGroup, error)` 方法

## 2. GitLab Provider 实现

- [x] 2.1 在 `provider/gitlab.go` 中实现 `ListGroups` 方法：调用 `GET /api/v4/groups?per_page=100` 分页查询所有可见 groups
- [x] 2.2 解析 GitLab group 响应（id、path、full_path），映射为 `RemoteGroup`
- [x] 2.3 复用现有的 `getWithPagination` 分页逻辑和 `checkGitLabRateLimit` 速率限制检测

## 3. GitHub Provider 实现

- [x] 3.1 在 `provider/github.go` 中实现 `ListGroups` 方法：调用 `GET /user/orgs?per_page=100` 分页查询用户所属 orgs
- [x] 3.2 解析 GitHub org 响应（login 字段），映射为 `RemoteGroup`
- [x] 3.3 复用现有的 `getWithPagination` 分页逻辑和 `checkGitHubRateLimit` 速率限制检测

## 4. list 命令远程 groups 支持

- [x] 4.1 在 `cmd/list.go` 中新增 `runListRemoteGroups(cfg *config.Config)` 函数：遍历所有 resources 调用 `ListGroups`
- [x] 4.2 支持 `--resource` 过滤：仅查询指定 resource 的 groups
- [x] 4.3 支持 `--group` 过滤：在结果中按名称过滤
- [x] 4.4 输出表格：NAME、RESOURCE、PATH 三列
- [x] 4.5 修改 `--remote` 限制逻辑：从 `listType != "repos"` 改为 `listType == "resources"`，在 switch 中为 `groups` + `remote` 添加分支调用 `runListRemoteGroups`

## 5. Flag 短别名

- [x] 5.1 `cmd/list.go`：`--group` 添加 `-g`、`--resource` 添加 `-R`、`--type` 添加 `-t`、`--remote` 添加 `-r`
- [x] 5.2 `cmd/clone.go`：`--group` 添加 `-g`、`--resource` 添加 `-R`、`--concurrency` 添加 `-n`
- [x] 5.3 `cmd/status.go`：`--group` 添加 `-g`、`--resource` 添加 `-R`
- [x] 5.4 `cmd/pull.go`：`--group` 添加 `-g`、`--resource` 添加 `-R`、`--force` 添加 `-f`、`--concurrency` 添加 `-n`
- [x] 5.5 `cmd/search.go`：`--group` 添加 `-g`、`--resource` 添加 `-R`
- [x] 5.6 `cmd/sync.go`：`--group` 添加 `-g`、`--resource` 添加 `-R`
- [x] 5.7 `cmd/add.go`：`addResourceCmd` 的 `--name` 添加 `-n`、`--provider` 添加 `-p`、`--url` 添加 `-u`、`--token` 添加 `-k`、`--ssh-key` 添加 `-s`
- [x] 5.8 `cmd/add.go`：`addGroupCmd` 的 `--name` 添加 `-n`、`--resource` 添加 `-R`、`--path` 添加 `-p`、`--local-path` 添加 `-l`、`--recursive` 添加 `-r`、`--ssh-key` 添加 `-s`、`--token` 添加 `-k`
- [x] 5.9 `cmd/add.go`：`addRepoCmd` 的 `--name` 添加 `-n`、`--resource` 添加 `-R`、`--url` 添加 `-u`、`--local-path` 添加 `-l`、`--group` 添加 `-g`、`--path` 添加 `-p`、`--ssh-key` 添加 `-s`、`--token` 添加 `-k`
- [x] 5.10 `cmd/init.go`：`--base` 添加 `-b`、`--provider` 添加 `-p`、`--url` 添加 `-u`、`--token` 添加 `-k`

## 6. 验证

- [x] 6.1 确认 `go build` 编译通过
- [x] 6.2 手动测试 `grepom list --remote --type groups` 输出正常
- [x] 6.3 手动测试短别名组合（如 `grepom list -r -t groups -R work-gl`）行为与长 flag 一致
- [x] 6.4 手动测试 `grepom list --remote --type resources` 返回预期错误信息
