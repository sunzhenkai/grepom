## ADDED Requirements

### Requirement: dedup 支持 --vgroup
`dedup` 命令 SHALL 支持 `--vgroup` 标志，通过虚拟分组选择参与检查的真实 groups。仅指定 `--vgroup` 时，组内 URL 去重和跨组 URL 警告 SHALL 限定在虚拟分组展开得到的真实 groups 范围内。`--group` 与 `--vgroup` 同时指定时，目标真实 group 集合 SHALL 取并集。

#### Scenario: dedup 检查虚拟分组
- **WHEN** 用户运行 `grepom dedup --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 对真实 groups `frontend` 和 `backend` 执行组内 URL 去重检查，并只在该范围内输出跨组 URL 警告

#### Scenario: dedup --group 与 --vgroup 取并集
- **WHEN** 用户运行 `grepom dedup --group infra --vgroup work`
- **THEN** 系统 SHALL 对真实 group `infra` 以及虚拟分组 `work` 包含的真实 groups 执行适用的 dedup 检查

#### Scenario: dedup --apply 应用于虚拟分组
- **WHEN** 用户运行 `grepom dedup --vgroup work --apply`
- **THEN** 系统 SHALL 仅对虚拟分组 `work` 展开的真实 groups 写入组内去重变更

#### Scenario: dedup 指定不存在的虚拟分组
- **WHEN** 用户运行 `grepom dedup --vgroup missing`
- **THEN** 系统 SHALL 报错提示虚拟分组 `missing` 不存在，不执行 dedup

### Requirement: dedup reference 保持真实 group 语义
`dedup --reference` SHALL 继续表示逗号分隔的真实 reference groups，不解析虚拟分组。虚拟分组仅通过 `--vgroup` 选择目标检查范围。

#### Scenario: reference 不解析虚拟分组
- **WHEN** 用户运行 `grepom dedup --group core --reference work`，且 `work` 只存在于虚拟分组中、不存在同名真实 group
- **THEN** 系统 SHALL 按现有 reference group 规则报错提示 reference group 不存在
