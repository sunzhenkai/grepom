## Purpose

定义 `grepom search` 命令按名称模糊搜索仓库的行为。
## Requirements
### Requirement: search 命令
系统 SHALL 提供 `grepom search <keyword>` 命令，按名称模糊搜索仓库。搜索使用大小写不敏感的子串匹配。

#### Scenario: 搜索匹配的仓库
- **WHEN** 用户运行 `grepom search web`
- **THEN** 系统显示所有名称包含 "web"（大小写不敏感）的仓库，包括 group 内 repo 和 standalone repo

#### Scenario: 搜索无匹配结果
- **WHEN** 用户运行 `grepom search xyz`
- **THEN** 系统输出 "no matching repos found"

#### Scenario: 搜索关键字为空
- **WHEN** 用户运行 `grepom search`（无参数）
- **THEN** 系统报错提示需要提供搜索关键字

### Requirement: search 结合 group 过滤器
`search` 命令 SHALL 支持 `--group` 过滤器，仅在指定 group 的范围内搜索仓库。

#### Scenario: 在指定 group 内搜索
- **WHEN** 用户运行 `grepom search web --group frontend`
- **THEN** 系统仅在 group `frontend` 的 repos 中搜索名称包含 "web" 的仓库

#### Scenario: 指定的 group 不存在
- **WHEN** 用户运行 `grepom search web --group nonexistent`
- **THEN** 系统输出 "no matching repos found"

### Requirement: search 结合 resource 过滤器
`search` 命令 SHALL 支持 `--resource` 过滤器，仅在引用指定 resource 的仓库中搜索。

#### Scenario: 在指定 resource 范围内搜索
- **WHEN** 用户运行 `grepom search web --resource work-gl`
- **THEN** 系统仅在引用 resource `work-gl` 的仓库中搜索名称包含 "web" 的仓库

### Requirement: search 输出格式
`search` 命令的输出格式 SHALL 与 `list` 命令保持一致，以表格形式显示仓库名称、路径、group、resource 和克隆状态。

#### Scenario: search 输出包含完整信息
- **WHEN** 用户运行 `grepom search web` 且找到匹配仓库
- **THEN** 输出包含仓库名称、本地路径、所属 group、关联 resource 和克隆状态，格式与 `grepom list` 一致

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

