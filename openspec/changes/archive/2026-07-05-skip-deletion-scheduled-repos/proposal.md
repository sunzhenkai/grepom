## Why

Codeup（云效）将"回收站/计划删除"状态的代码库在远端重命名为 `<name>-deletion_scheduled-<id>`。grepom 在 `sync`/`clone` 时仍会拉取到这些代码库并尝试克隆，而这些库已无法访问，最终以 `all authentication methods failed` 报错收场——错误信息具有误导性（看似鉴权配置问题，实为库已废弃）。需要默认跳过这类"删除中"的代码库，避免噪音与误导，并给出清晰的跳过原因。

## What Changes

- 在 Codeup provider 的 `ListRepos` 中识别"删除中"代码库（依据 `pathWithNamespace`/`name` 含 `deletion_scheduled` 标记），默认从发现结果中剔除
- `provider.ListReposParams` 新增 `IncludeDeleted bool` 字段，用于显式保留删除中的代码库
- `provider.Repo` 的 `DisabledReason` 新增取值 `deletion_scheduled`，供 resolver 在运行时对已写入配置但实际处于删除中的代码库进行兜底拦截，输出清晰跳过原因而非抛出鉴权错误
- `grepom sync` 与 `grepom list` 新增 `--include-deleted` 标志，透传到 `ListReposParams.IncludeDeleted`
- verbose 模式下输出被跳过的删除中代码库数量
- 更新 README（中英文）说明该默认行为与 `--include-deleted` 标志

## Capabilities

### New Capabilities

- `skip-deleted-repos`: 识别并默认跳过处于"计划删除（deletion_scheduled）"状态的代码库，覆盖发现阶段过滤、运行时兜底拦截、`--include-deleted` 显式包含开关与跳过计数提示

### Modified Capabilities

- `codeup-provider`: `ListRepos` 实现新增"默认过滤 deletion_scheduled 代码库"的行为，并支持通过 `IncludeDeleted` 参数保留它们

## Impact

- 修改文件：`provider/codeup.go`（过滤逻辑）、`provider/provider.go`（`ListReposParams` 新字段）、`repo/resolver.go`（运行时兜底标记 `DisabledReason`）、`cmd/sync.go`、`cmd/list.go`（新增 `--include-deleted` 标志）
- 新增/更新测试：`provider/codeup_test.go`、`repo/resolver_test.go`
- 文档：`README.md`、`README_en.md`
- 无破坏性变更：默认行为更安全（跳过废弃库），用户可通过 `--include-deleted` 恢复旧行为
