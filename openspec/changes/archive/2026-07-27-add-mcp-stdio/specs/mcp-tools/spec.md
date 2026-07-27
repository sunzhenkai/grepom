## ADDED Requirements

### Requirement: 仓库列表工具

系统 SHALL 提供 `grepom_list` MCP Tool，返回配置中所有仓库的结构化列表。

#### Scenario: 列出全部仓库

- **WHEN** Agent 调用 `grepom_list` 无过滤参数
- **THEN** 返回 JSON 数组，每项包含 name、group、url、local_path 字段

#### Scenario: 按组过滤

- **WHEN** Agent 调用 `grepom_list` 并传入 `group` 参数
- **THEN** 仅返回属于该组的仓库

### Requirement: 仓库状态工具

系统 SHALL 提供 `grepom_status` MCP Tool，查询指定仓库或全部仓库的 Git 状态。

#### Scenario: 查询单个仓库状态

- **WHEN** Agent 调用 `grepom_status` 并传入 `repo` 参数
- **THEN** 返回该仓库的分支、是否 clean、ahead/behind 信息

#### Scenario: 查询全部状态

- **WHEN** Agent 调用 `grepom_status` 不传 `repo` 参数
- **THEN** 返回所有已克隆仓库的状态摘要

### Requirement: 仓库搜索工具

系统 SHALL 提供 `grepom_search` MCP Tool，按名称模糊搜索仓库。

#### Scenario: 模糊匹配

- **WHEN** Agent 调用 `grepom_search` 并传入 `query` 参数
- **THEN** 返回名称包含 query 子串的仓库列表（不区分大小写）

### Requirement: 仓库路径工具

系统 SHALL 提供 `grepom_dir` MCP Tool，返回指定仓库的本地绝对路径。

#### Scenario: 已克隆仓库

- **WHEN** Agent 调用 `grepom_dir` 并传入已克隆仓库名
- **THEN** 返回该仓库的本地绝对路径

#### Scenario: 未克隆仓库

- **WHEN** Agent 调用 `grepom_dir` 并传入未克隆仓库名
- **THEN** 返回错误信息（IsError: true），说明仓库尚未克隆

### Requirement: 拉取工具

系统 SHALL 提供 `grepom_pull` MCP Tool，对指定仓库执行 git pull。

#### Scenario: 拉取成功

- **WHEN** Agent 调用 `grepom_pull` 并传入已克隆仓库名
- **THEN** 执行 git pull 并返回操作结果摘要

#### Scenario: 拉取超时

- **WHEN** git pull 操作超过 60 秒未完成
- **THEN** 终止操作并返回超时错误

### Requirement: 克隆工具

系统 SHALL 提供 `grepom_clone` MCP Tool，克隆指定仓库到本地。

#### Scenario: 克隆成功

- **WHEN** Agent 调用 `grepom_clone` 并传入配置中存在的仓库名
- **THEN** 执行克隆并返回本地路径

#### Scenario: 仓库不存在

- **WHEN** Agent 调用 `grepom_clone` 并传入配置中不存在的仓库名
- **THEN** 返回错误信息（IsError: true）

### Requirement: 组信息工具

系统 SHALL 提供 `grepom_groups` MCP Tool，列出配置中的所有组及其仓库数量。

#### Scenario: 列出组

- **WHEN** Agent 调用 `grepom_groups`
- **THEN** 返回 JSON 数组，每项包含 group 名称、provider、仓库数量

### Requirement: 密钥扫描工具

系统 SHALL 提供 `grepom_scan` MCP Tool，对指定仓库执行密钥扫描。

#### Scenario: 扫描无发现

- **WHEN** Agent 调用 `grepom_scan` 并传入仓库名，仓库无密钥泄露
- **THEN** 返回空发现列表

#### Scenario: 扫描有发现

- **WHEN** Agent 调用 `grepom_scan` 并传入仓库名，仓库存在密钥泄露
- **THEN** 返回发现列表，每项包含文件路径、行号、规则 ID

### Requirement: 工具错误处理

所有 MCP Tool SHALL 在出错时返回 `IsError: true` 的文本内容，包含可操作的错误描述，使 Agent 能自行修正调用。

#### Scenario: 参数缺失

- **WHEN** Agent 调用工具但缺少必填参数
- **THEN** 返回 IsError 结果，说明缺少哪个参数及期望格式
