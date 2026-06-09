## ADDED Requirements

### Requirement: list --remote 支持 --vgroup
`list --remote` 在默认 repos 模式下 SHALL 支持 `--vgroup` 标志，通过虚拟分组选择需要查询远程仓库的真实 groups。`--group` 与 `--vgroup` 同时指定时，系统 SHALL 对真实 group 集合取并集；随后继续应用 `--resource`、`--all`、禁用状态和 `exclude_repos` 等既有规则。

#### Scenario: 远程查询虚拟分组
- **WHEN** 用户运行 `grepom list --remote --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 仅查询真实 groups `frontend` 和 `backend` 关联的远程仓库并展示

#### Scenario: list --remote --group 与 --vgroup 取并集
- **WHEN** 用户运行 `grepom list --remote --group infra --vgroup work`
- **THEN** 系统 SHALL 查询真实 group `infra` 以及虚拟分组 `work` 包含的真实 groups 关联的远程仓库

#### Scenario: list --remote --vgroup 与 --resource 组合
- **WHEN** 用户运行 `grepom list --remote --vgroup work --resource work-gl`
- **THEN** 系统 SHALL 仅查询虚拟分组 `work` 中引用 resource `work-gl` 的真实 groups

#### Scenario: list --remote 指定不存在的虚拟分组
- **WHEN** 用户运行 `grepom list --remote --vgroup missing`
- **THEN** 系统 SHALL 报错提示虚拟分组 `missing` 不存在，不执行远程查询
