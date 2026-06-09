## ADDED Requirements

### Requirement: prune 支持 --vgroup
`prune` 命令 SHALL 支持 `--vgroup` 标志，通过虚拟分组选择需要扫描和清理 excluded repos 的真实 groups。`--group` 与 `--vgroup` 同时指定时，系统 SHALL 对真实 group 集合取并集；随后继续应用 `--resource`、`--apply`、`--force` 和安全检查等既有规则。

#### Scenario: prune 扫描虚拟分组
- **WHEN** 用户运行 `grepom prune --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 只扫描真实 groups `frontend` 和 `backend` 中的 excluded repos

#### Scenario: prune --group 与 --vgroup 取并集
- **WHEN** 用户运行 `grepom prune --group infra --vgroup work`
- **THEN** 系统 SHALL 扫描真实 group `infra` 以及虚拟分组 `work` 包含的真实 groups 中的 excluded repos

#### Scenario: prune --vgroup 与 --resource 组合
- **WHEN** 用户运行 `grepom prune --vgroup work --resource work-gl`
- **THEN** 系统 SHALL 只扫描虚拟分组 `work` 中引用 resource `work-gl` 的真实 groups 中的 excluded repos

#### Scenario: prune 指定不存在的虚拟分组
- **WHEN** 用户运行 `grepom prune --vgroup missing`
- **THEN** 系统 SHALL 报错提示虚拟分组 `missing` 不存在，不执行 prune
