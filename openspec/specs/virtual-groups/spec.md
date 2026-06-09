# virtual-groups Specification

## Purpose

将多个真实 group 组织为命名虚拟分组，并通过 `--vgroup` 对成员 group 批量执行命令。
## Requirements
### Requirement: 虚拟分组配置
系统 SHALL 支持顶层 `virtual_groups` 配置字段，用于定义虚拟分组到真实 group 的映射。`virtual_groups` SHALL 使用 map 结构，key 为虚拟分组名称，value 包含 `groups` 字符串数组字段。虚拟分组的 `groups` 成员 SHALL 只引用真实 group，不支持引用其他虚拟分组。

#### Scenario: 定义虚拟分组
- **WHEN** 配置文件包含 `virtual_groups.work.groups: [frontend, backend]`
- **THEN** 系统加载虚拟分组 `work`，其成员为真实 group `frontend` 和 `backend`

#### Scenario: 未配置虚拟分组
- **WHEN** 配置文件未包含 `virtual_groups`
- **THEN** 系统正常加载配置，虚拟分组列表视为空

#### Scenario: 虚拟分组引用不存在的真实 group
- **WHEN** 虚拟分组 `work` 的 `groups` 包含 `missing-group`，但 `groups` 中不存在名为 `missing-group` 的真实 group
- **THEN** 系统 SHALL 在加载配置时报错，提示虚拟分组引用了不存在的真实 group

#### Scenario: 虚拟分组不支持嵌套
- **WHEN** 配置文件中存在虚拟分组 `all` 和 `work`，且 `all.groups` 包含 `work`，但真实 group 中不存在名为 `work` 的 group
- **THEN** 系统 SHALL 按真实 group 引用校验失败，不把 `work` 解析为另一个虚拟分组

### Requirement: 虚拟分组名称命名空间
虚拟分组名称 SHALL 与真实 group 名称处于独立命名空间。系统 SHALL 允许虚拟分组与真实 group 同名，并通过 `--vgroup` 选择虚拟分组，通过 `--group` 选择真实 group。

#### Scenario: 虚拟分组与真实 group 同名
- **WHEN** 配置中存在真实 group `work`，同时存在虚拟分组 `work`
- **THEN** 系统 SHALL 正常加载配置，不报告名称冲突

#### Scenario: 同名时 --group 选择真实 group
- **WHEN** 用户运行 `grepom list --group work`
- **THEN** 系统 SHALL 仅按真实 group `work` 过滤仓库

#### Scenario: 同名时 --vgroup 选择虚拟分组
- **WHEN** 用户运行 `grepom list --vgroup work`
- **THEN** 系统 SHALL 展开虚拟分组 `work` 中列出的真实 groups，并按这些真实 groups 过滤仓库

### Requirement: --vgroup 选择语义
支持 group 过滤的命令 SHALL 提供 `--vgroup` 标志，用于选择虚拟分组包含的真实 groups。`--group` 与 `--vgroup` 同时指定时，系统 SHALL 对两者得到的真实 group 集合取并集；如果两者都未指定，系统 SHALL 保持现有的全部 group 行为。

#### Scenario: 仅指定 --vgroup
- **WHEN** 用户运行 `grepom status --vgroup work`，虚拟分组 `work` 包含 `frontend` 和 `backend`
- **THEN** 系统 SHALL 只处理真实 group `frontend` 和 `backend` 下的仓库

#### Scenario: --group 与 --vgroup 取并集
- **WHEN** 用户运行 `grepom status --group infra --vgroup work`，虚拟分组 `work` 包含 `frontend` 和 `backend`
- **THEN** 系统 SHALL 处理真实 group `infra`、`frontend` 和 `backend` 下的仓库

#### Scenario: 并集结果去重
- **WHEN** 用户运行 `grepom status --group frontend --vgroup work`，虚拟分组 `work` 包含 `frontend` 和 `backend`
- **THEN** 系统 SHALL 只处理一次真实 group `frontend`，并同时处理 `backend`

#### Scenario: 指定不存在的虚拟分组
- **WHEN** 用户运行 `grepom status --vgroup missing`
- **THEN** 系统 SHALL 报错提示虚拟分组 `missing` 不存在

#### Scenario: --vgroup 与 --resource 组合
- **WHEN** 用户运行 `grepom list --vgroup work --resource work-gl`
- **THEN** 系统 SHALL 先展开虚拟分组 `work` 的真实 groups，再仅保留其中引用 resource `work-gl` 的仓库

