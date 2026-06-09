## ADDED Requirements

### Requirement: search 支持 --vgroup
`search` 命令 SHALL 支持 `--vgroup` 标志，通过虚拟分组限定搜索范围。`--group` 与 `--vgroup` 同时指定时，系统 SHALL 对真实 group 集合取并集；随后继续应用 `--resource` 和关键字匹配等既有规则。

#### Scenario: 在虚拟分组内搜索
- **WHEN** 用户运行 `grepom search api --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 仅在真实 groups `frontend` 和 `backend` 的 repos 中搜索名称包含 `api` 的仓库

#### Scenario: search --group 与 --vgroup 取并集
- **WHEN** 用户运行 `grepom search api --group infra --vgroup work`
- **THEN** 系统 SHALL 在真实 group `infra` 以及虚拟分组 `work` 包含的真实 groups 中搜索名称包含 `api` 的仓库

#### Scenario: search --vgroup 与 --resource 组合
- **WHEN** 用户运行 `grepom search api --vgroup work --resource work-gl`
- **THEN** 系统 SHALL 仅在虚拟分组 `work` 中引用 resource `work-gl` 的真实 groups 中搜索

#### Scenario: search 指定不存在的虚拟分组
- **WHEN** 用户运行 `grepom search api --vgroup missing`
- **THEN** 系统 SHALL 报错提示虚拟分组 `missing` 不存在，不执行搜索
