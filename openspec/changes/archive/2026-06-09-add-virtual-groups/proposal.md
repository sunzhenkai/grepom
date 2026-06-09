## Why

当前 `groups` 只能逐个表达真实远端 group/org 或手动维护的 repo 容器。用户在 personal、work 等日常场景中需要反复查看配置并逐个操作多个 group，批量 sync、clone、pull、status、list 等操作成本较高。

引入虚拟分组后，用户可以把多个真实 group 组织成命名集合，并通过 `--vgroup` 一次性把命令应用到多个真实 group。

## What Changes

- 新增顶层配置 `virtual_groups`，用于定义虚拟分组；每个虚拟分组只包含真实 group 名称列表，不支持嵌套引用其他虚拟分组。
- 新增 `--vgroup` 命令过滤器，用于选择虚拟分组中的所有真实 group。
- 当 `--group` 与 `--vgroup` 同时指定时，目标 group 集合取并集。
- `grepom list groups` 同时展示真实 group 与虚拟分组，并通过类型列区分。
- 虚拟分组不支持 `enabled` 等运行时状态字段。
- 虚拟分组允许与真实 group 同名，因为两者处于不同命名空间，并通过不同命令参数选择。
- 更新 README 与示例配置，说明 `virtual_groups` 配置和 `--vgroup` 用法。

## Capabilities

### New Capabilities

- `virtual-groups`: 定义虚拟分组配置、校验规则、`--vgroup` 选择语义，以及命令对虚拟分组的批量操作行为。

### Modified Capabilities

- `group-list`: `grepom list groups` 需要展示真实 group 与虚拟分组，并用类型区分。
- `cli-commands`: clone/list/status/pull/search 等现有 group 过滤命令需要支持 `--vgroup`。
- `sync-command`: sync 需要支持通过 `--vgroup` 同步虚拟分组包含的真实 groups。
- `list-remote-repos`: `list --remote` 需要支持通过 `--vgroup` 查询虚拟分组包含的真实 groups。
- `dedup-command`: dedup 需要支持通过 `--vgroup` 对虚拟分组包含的真实 groups 执行检查和去重流程。
- `prune-command`: prune 需要支持通过 `--vgroup` 限定虚拟分组包含的真实 groups。
- `search-command`: search 需要支持通过 `--vgroup` 限定虚拟分组包含的真实 groups。

## Impact

- 配置模型：`config.Config` 新增虚拟分组字段及加载校验。
- 仓库解析：`repo.Filter` 和过滤逻辑需要支持多个 group 名称。
- 命令层：新增公共 `--vgroup` 解析逻辑，避免各命令重复展开虚拟分组。
- 用户界面：help、example、README、`list groups` 输出格式需要更新。
- 测试：覆盖配置校验、虚拟分组展开、`--group` 与 `--vgroup` 并集、同名真实/虚拟 group、相关命令过滤行为。
