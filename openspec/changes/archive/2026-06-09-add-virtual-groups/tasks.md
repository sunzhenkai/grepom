## 1. 配置模型与校验

- [x] 1.1 在 `config.Config` 中新增 `VirtualGroups` 顶层字段，并定义虚拟分组结构体。
- [x] 1.2 更新配置加载默认值，未配置 `virtual_groups` 时初始化为空集合。
- [x] 1.3 在配置校验中允许虚拟分组与真实 group 同名。
- [x] 1.4 在配置校验中要求虚拟分组成员必须引用已存在的真实 group。
- [x] 1.5 为虚拟分组配置加载、缺省字段、同名真实 group、缺失成员引用添加单元测试。

## 2. Group 选择与仓库过滤

- [x] 2.1 实现公共 group 选择展开函数，支持 `--group` 和 `--vgroup` 并集、去重、错误提示。
- [x] 2.2 扩展 `repo.Filter` 支持多个真实 group 名称，并保持现有单值 `Group` 行为兼容。
- [x] 2.3 更新 `repo.ApplyFilter`、搜索过滤等逻辑，使多 group 过滤与 `--resource` 过滤按预期组合。
- [x] 2.4 为 group 选择展开、并集去重、缺失虚拟分组、多 group 仓库过滤添加单元测试。

## 3. 命令接入

- [x] 3.1 为 `clone`、`list`、`status`、`pull`、`scan` 添加 `--vgroup` 标志并接入公共选择逻辑。
- [x] 3.2 为 `search` 添加 `--vgroup` 标志并接入公共选择逻辑。
- [x] 3.3 为 `prune` 添加 `--vgroup` 标志并接入 excluded repos 过滤逻辑。
- [x] 3.4 为 `sync` 添加 `--vgroup` 标志，并在直接遍历 `cfg.Groups` 的流程中使用展开后的真实 group 集合。
- [x] 3.5 为 `list --remote` 添加 `--vgroup` 支持，并确保与 `--resource`、`--all`、禁用状态、`exclude_repos` 规则组合正确。
- [x] 3.6 为 `dedup` 添加 `--vgroup` 标志，限定组内 URL 去重和跨组 URL 警告范围；保持 `--reference` 只解析真实 group。

## 4. 分组列表展示与示例

- [x] 4.1 更新 `grepom list groups` 输出，新增 TYPE 与 GROUPS 列，展示真实 group 和虚拟分组。
- [x] 4.2 计算虚拟分组成员真实 groups 的 repo 总数，并在 REPOS 列展示。
- [x] 4.3 更新 `grepom example` 配置示例，加入 `virtual_groups` 示例。
- [x] 4.4 更新 `README.md` 和 `README_en.md`，说明 `virtual_groups` 配置、`--vgroup` 用法、与 `--group` 的并集语义。

## 5. 命令级测试与验证

- [x] 5.1 为 `list groups` 添加真实 group 与虚拟分组混合展示测试。
- [x] 5.2 为常用仓库命令添加 `--vgroup` 过滤测试，至少覆盖 `list`、`status`、`clone` 或 `pull` 中的代表路径。
- [x] 5.3 为 `sync --vgroup`、`list --remote --vgroup` 添加命令级测试或 provider mock 测试。
- [x] 5.4 为 `dedup --vgroup`、`prune --vgroup`、`search --vgroup` 添加行为测试。
- [x] 5.5 运行 Go 测试套件，确认现有 group、resource、enabled、exclude 行为未回归。
