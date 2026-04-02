## 1. 重构 list 命令支持 --type 标志

- [x] 1.1 在 `cmd/list.go` 中新增 `--type` 字符串标志，支持 `repos`（默认）、`resources`、`groups` 三个值
- [x] 1.2 将现有 repo 列表逻辑提取为 `runListRepos` 函数，仅在 `--type repos` 或默认时调用
- [x] 1.3 更新 `listCmd` 的 `RunE`，根据 `--type` 值分发到不同的输出函数

## 2. 实现 resource 列表功能

- [x] 2.1 在 `cmd/list.go` 中实现 `runListResources` 函数，遍历 `cfg.Resources` 输出表格（NAME、PROVIDER、URL、SSH_KEY）
- [x] 2.2 处理 SSH_KEY 字段：已配置时显示路径，未配置时显示 `-`
- [x] 2.3 处理空 resources 场景，输出 `No resources found.`

## 3. 实现 group 列表功能

- [x] 3.1 在 `cmd/list.go` 中实现 `runListGroups` 函数，遍历 `cfg.Groups` 输出表格（NAME、RESOURCE、PATH、LOCAL_PATH、RECURSIVE、REPOS）
- [x] 3.2 RECURSIVE 列格式化：`true` 显示 `yes`，`false` 或未设置显示 `no`
- [x] 3.3 REPOS 列显示 group 下 `len(group.Repos)` 的数量
- [x] 3.4 处理空 groups 场景，输出 `No groups found.`

## 4. 更新帮助文档和示例

- [x] 4.1 更新 `listCmd` 的 `Short`、`Long` 和 `Example` 文档，体现 `--type` 标志的用法
- [x] 4.2 更新 cli-commands spec 中 list 命令的描述（如有必要）

## 5. 验证

- [x] 5.1 手动测试 `grepom list`（默认行为不变，列出 repos）
- [x] 5.2 手动测试 `grepom list --type resources`（列出所有 resources）
- [x] 5.3 手动测试 `grepom list --type groups`（列出所有 groups）
- [x] 5.4 验证向后兼容：`grepom list [name]`、`--group`、`--resource` 过滤行为不变
- [x] 5.5 验证 `grepom list --type resources some-name --group xxx` 等混合参数正确忽略过滤标志
