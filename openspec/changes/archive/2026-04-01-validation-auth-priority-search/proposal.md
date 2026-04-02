## Why

当前系统在 `add` 命令层面缺少前置校验，导致配置错误（如引用不存在的 resource、group 重名等）只有在后续 clone/sync 使用时才暴露。同时，认证优先级策略是 token 优先 SSH，但实际场景中用户更多使用 SSH 进行推送操作，SSH 应具有更高优先级。此外，目前缺少按名称模糊搜索仓库的能力，用户只能精确匹配仓库名。

## What Changes

- **add 命令增加前置校验**：在 `add group`、`add repo` 执行时立即校验引用的 resource 是否存在、group 名称是否重复、standalone repo 名称是否重复，而非等到 load 配置时才报错
- **调整认证优先级**：整体改为 SSH 优先，token 次之。新的优先级链为：group/repo SSH key → group/repo token → resource SSH key → resource token → 默认 SSH → 裸 HTTP
- **新增 search 命令**：支持按名称模糊搜索（子串匹配）仓库，可结合 `--group` 和 `--resource` 过滤器使用

## Capabilities

### New Capabilities
- `add-command-validation`: add 命令层面的前置校验逻辑（resource 存在性、group/repo 重名检测、group 引用有效性等）
- `search-command`: 按名称模糊搜索仓库的 CLI 命令，支持 `--group` 和 `--resource` 过滤器

### Modified Capabilities
- `clone-auth-priority`: 认证优先级从 token 优先改为 SSH 优先，整体优先级链调整为 group/repo SSH → group/repo token → resource SSH → resource token → 默认 SSH → 裸 HTTP
- `cli-commands`: 新增 `search` 子命令到 CLI 命令列表

## Impact

- **`cmd/add.go`**: 增加前置校验逻辑（加载已有配置后检查引用和重名）
- **`git/git.go`**: 修改 `Clone` 函数中认证策略的优先级顺序
- **`repo/resolver.go`**: 修改认证合并逻辑的顺序
- **`cmd/search.go`**: 新文件，实现 search 命令
- **`cmd/list.go`**: search 命令可复用 list 的表格输出样式
- **测试文件**: `config/config_test.go`、`git/git_test.go`、`repo/resolver_test.go` 需要相应更新
