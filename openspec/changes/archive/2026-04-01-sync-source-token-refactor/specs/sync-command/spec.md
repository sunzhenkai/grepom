## MODIFIED Requirements

### Requirement: sync 命令
`sync` 命令 SHALL 从远程 API 发现仓库信息和子 group/org，将新发现的条目追加到配置文件。sync 命令 SHALL NOT 执行 clone 或 pull 操作。

#### Scenario: 同步所有 source 的所有 group/org
- **WHEN** 用户运行 `grepom sync`（无参数）
- **THEN** 系统对所有配置中的 source 下所有 group/org 执行远程发现，将新发现的仓库追加到配置的 repos 列表，将新发现的子 group 追加到对应 source 的 groups 列表，不执行 clone 或 pull

#### Scenario: 同步指定 source
- **WHEN** 用户运行 `grepom sync --source my-gitlab`（按名称指定 source）
- **THEN** 系统仅对该 source 下的 group/org 执行远程发现并更新配置

#### Scenario: 同步指定 group
- **WHEN** 用户运行 `grepom sync --group my-org/frontend`
- **THEN** 系统仅发现匹配该 group 路径的仓库并更新配置

#### Scenario: 同步指定 org
- **WHEN** 用户运行 `grepom sync --org my-org`
- **THEN** 系统仅发现该 org 下的仓库并更新配置

#### Scenario: 同步时无新内容
- **WHEN** 用户运行 `grepom sync` 且没有新仓库也没有新的子 group/org
- **THEN** 系统不修改配置文件，输出同步摘要
