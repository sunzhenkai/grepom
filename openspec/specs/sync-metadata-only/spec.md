### Requirement: sync 仅同步元数据
`sync` 命令 SHALL 仅从远程 API 读取仓库列表和子 group/org 信息，将发现的条目保存到配置文件中。sync 命令 SHALL NOT 执行 clone 或 pull 操作。

#### Scenario: 发现新仓库并保存到配置
- **WHEN** 用户运行 `grepom sync`，远程 API 返回配置中不存在的仓库
- **THEN** 系统将新仓库以 `repos` 条目追加到配置文件，不执行 clone

#### Scenario: 发现新子 group 并追加
- **WHEN** 用户运行 `grepom sync`，远程 API 返回配置中不存在的子 group
- **THEN** 系统将新子 group 追加到对应 source 的 groups 列表，不执行 clone

#### Scenario: 无新内容时配置不变
- **WHEN** 用户运行 `grepom sync` 且没有新仓库和新的子 group/org
- **THEN** 系统不修改配置文件，仅输出同步摘要

#### Scenario: sync 输出提示后续操作
- **WHEN** sync 发现并保存了新仓库到配置
- **THEN** 系统在摘要中提示用户运行 `grepom clone` 来克隆新仓库
