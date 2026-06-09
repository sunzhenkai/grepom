## ADDED Requirements

### Requirement: sync 支持 --vgroup
`sync` 命令 SHALL 支持 `--vgroup` 标志，通过虚拟分组选择需要同步的真实 groups。`--group` 与 `--vgroup` 同时指定时，系统 SHALL 对真实 group 集合取并集；随后继续应用 `--resource` 过滤、禁用 group/resource 跳过、无 resource group 跳过和 exclude 过滤等既有规则。

#### Scenario: 同步虚拟分组
- **WHEN** 用户运行 `grepom sync --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 仅对真实 groups `frontend` 和 `backend` 中启用且绑定 resource 的 groups 执行远程发现

#### Scenario: sync --group 与 --vgroup 取并集
- **WHEN** 用户运行 `grepom sync --group infra --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 对真实 groups `infra`、`frontend` 和 `backend` 中启用且绑定 resource 的 groups 执行远程发现

#### Scenario: sync --vgroup 与 --resource 组合
- **WHEN** 用户运行 `grepom sync --vgroup work --resource work-gl`
- **THEN** 系统 SHALL 仅同步虚拟分组 `work` 中引用 resource `work-gl` 的真实 groups

#### Scenario: sync 指定不存在的虚拟分组
- **WHEN** 用户运行 `grepom sync --vgroup missing`
- **THEN** 系统 SHALL 报错提示虚拟分组 `missing` 不存在，不执行同步
