### Requirement: sync 命令
`sync` 命令 SHALL 从远程 API 发现仓库信息，将新发现的条目追加到对应 group 的 repos 字段。sync 命令 SHALL NOT 执行 clone 或 pull 操作。

#### Scenario: 同步所有 group
- **WHEN** 用户运行 `grepom sync`（无参数）
- **THEN** 系统对所有配置中的 group 执行远程发现，将新发现的仓库追加到各自 group 的 repos 列表，不执行 clone 或 pull

#### Scenario: 同步指定 group
- **WHEN** 用户运行 `grepom sync --group frontend`
- **THEN** 系统仅对 name 为 `frontend` 的 group 执行远程发现并更新其 repos 列表

#### Scenario: 同步指定 resource 下的所有 group
- **WHEN** 用户运行 `grepom sync --resource work-gl`
- **THEN** 系统对所有引用 `work-gl` 资源的 group 执行远程发现并更新各自 repos 列表

#### Scenario: 同步时无新内容
- **WHEN** 用户运行 `grepom sync` 且没有新仓库
- **THEN** 系统不修改配置文件，输出同步摘要

#### Scenario: GitLab recursive group 发现所有层级项目
- **WHEN** GitLab group `frontend`（recursive: true）下有直接项目和子 group `ui` 下的项目
- **THEN** 系统将所有层级的项目作为 repo 条目追加到 `frontend` group 的 repos 列表，repo 的 path 保持远端完整路径（如 `my-org/frontend/ui/design-system`）

#### Scenario: 发现新 repo 不在全局 repos 中
- **WHEN** sync 从 group `frontend` 发现新 repo
- **THEN** 系统 SHALL 将该 repo 追加到 group `frontend` 的 `repos` 字段中，而非顶层全局 `repos`

### Requirement: sync 配置更新策略（只增不删）
sync 命令在更新配置文件时 SHALL 仅追加新发现的 repo 条目到对应 group 的 repos 列表，不删除或修改已有条目。

#### Scenario: 远程新增仓库
- **WHEN** group `frontend` 下远程新增了 repo `new-app`
- **THEN** 系统将 `new-app` 追加到 group `frontend` 的 repos 列表

#### Scenario: Repo 已存在于 group repos
- **WHEN** 远程 repo `shared-utils` 已存在于 group `frontend` 的 repos 列表中（按 URL 匹配）
- **THEN** 系统不重复追加

#### Scenario: 远程仓库被删除不影响配置
- **WHEN** 远程某仓库已被删除，但配置文件中对应 group 的 repos 仍有该条目
- **THEN** 系统不从配置中删除该 repo 条目

#### Scenario: 非 recursive group 仅发现直接项目
- **WHEN** GitLab group 配置为 `recursive: false`（或未设置）
- **THEN** 系统仅发现该 group 直接包含的项目，不递归子 group

### Requirement: sync 并发写入保护
当多个 sync 实例同时运行时，系统 SHALL 使用文件锁防止配置文件的并发写入冲突。

#### Scenario: 多个 sync 实例同时运行
- **WHEN** 两个 `grepom sync` 进程同时运行并都需要写入配置文件
- **THEN** 第二个进程等待第一个进程完成写入后再执行自己的写入，不会导致配置文件损坏或数据丢失

#### Scenario: 获取锁超时
- **WHEN** sync 进程无法在合理时间内获取配置文件锁
- **THEN** 系统报告错误并退出，提示用户另一个 sync 正在运行
