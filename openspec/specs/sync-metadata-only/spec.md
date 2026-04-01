### Requirement: sync 仅同步元数据
`sync` 命令 SHALL 仅从远程 API 读取仓库列表信息，将发现的条目保存到对应 group 的 repos 字段中。sync 命令 SHALL NOT 执行 clone 或 pull 操作，也 SHALL NOT 在配置中创建 subgroup 条目。

#### Scenario: 发现新仓库并保存到 group repos
- **WHEN** 用户运行 `grepom sync`，远程 API 返回配置中不存在的仓库
- **THEN** 系统将新仓库追加到对应 group 的 `repos` 字段中，不执行 clone

#### Scenario: recursive group 不创建 subgroup 配置条目
- **WHEN** 用户运行 `grepom sync`，GitLab recursive group 发现了子 group
- **THEN** 系统 SHALL NOT 在配置文件中创建 subgroup 的 group 条目，而是将子 group 下的项目直接写入父 group 的 repos，保持 repo 的远端完整 path

#### Scenario: 无新内容时配置不变
- **WHEN** 用户运行 `grepom sync` 且没有新仓库
- **THEN** 系统不修改配置文件，仅输出同步摘要

#### Scenario: sync 输出提示后续操作
- **WHEN** sync 发现并保存了新仓库到某 group 的 repos
- **THEN** 系统在摘要中提示用户运行 `grepom clone` 来克隆新仓库
