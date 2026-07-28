## ADDED Requirements

### Requirement: 虚拟分组可引用独立 repos
系统 SHALL 允许 `virtual_groups.<name>` 包含可选的 `repos` 字符串数组，用于引用顶层独立仓库（`repos:` 列表中的 `name`）。`repos` 成员 MUST 存在于顶层独立仓库；不得引用 group 内仓库名或其他虚拟分组。`--vgroup` 在 list/status/clone/pull/search/prune 等过滤场景 SHALL 同时包含成员真实 groups 下的仓库与成员独立 repos。

#### Scenario: 定义含 repos 的虚拟分组
- **WHEN** 配置包含独立仓 `dotfiles`，且 `virtual_groups.tools.repos: [dotfiles]`
- **THEN** 系统加载成功；`grepom list --vgroup tools` 包含 `dotfiles`

#### Scenario: groups 与 repos 混合
- **WHEN** 虚拟分组 `work` 的 `groups` 含 `frontend`，`repos` 含 `notes`
- **THEN** `--vgroup work` 同时处理 `frontend` 下仓库与独立仓 `notes`

#### Scenario: 引用不存在的独立仓
- **WHEN** `virtual_groups.tools.repos` 包含 `missing-repo`，但顶层 `repos` 中不存在该 name
- **THEN** 系统 SHALL 在加载配置时报错，提示虚拟分组引用了不存在的独立仓库

#### Scenario: 仅 repos 无 groups
- **WHEN** 虚拟分组只配置 `repos: [dotfiles]`，`groups` 为空或省略
- **THEN** 系统加载成功；`--vgroup` 仅过滤到这些独立仓

#### Scenario: sync --vgroup 仅作用于成员 groups
- **WHEN** 用户运行 `grepom sync --vgroup tools`，该 vgroup 同时含 groups 与 repos
- **THEN** 系统仅对成员真实 groups 执行远程发现；独立 repos 无 API sync 语义，保持既有声明不变

## MODIFIED Requirements

### Requirement: 虚拟分组配置
系统 SHALL 支持顶层 `virtual_groups` 配置字段，用于定义虚拟分组到真实 group 与可选独立 repo 的映射。`virtual_groups` SHALL 使用 map 结构，key 为虚拟分组名称，value 包含可选的 `groups` 字符串数组与可选的 `repos` 字符串数组。虚拟分组的 `groups` 成员 SHALL 只引用真实 group，不支持引用其他虚拟分组；`repos` 成员 SHALL 只引用顶层独立仓库。

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

### Requirement: --vgroup 选择语义
支持 group 过滤的命令 SHALL 提供 `--vgroup` 标志，用于选择虚拟分组包含的真实 groups 与独立 repos。`--group` 与 `--vgroup` 同时指定时，系统 SHALL 对两者得到的真实 group 集合取并集，并将 vgroup 中的独立 repos 一并纳入过滤结果；如果两者都未指定，系统 SHALL 保持现有的全部 group/repo 行为。

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
- **THEN** 系统 SHALL 先展开虚拟分组 `work` 的真实 groups（及独立 repos），再仅保留其中引用 resource `work-gl` 的仓库

#### Scenario: --vgroup 包含独立仓
- **WHEN** 用户运行 `grepom clone --vgroup tools`，虚拟分组 `tools` 的 `repos` 包含 `dotfiles`
- **THEN** 系统 SHALL 将独立仓 `dotfiles` 纳入克隆集合
