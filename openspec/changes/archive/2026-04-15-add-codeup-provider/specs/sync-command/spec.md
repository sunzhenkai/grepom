## MODIFIED Requirements

### Requirement: sync 命令
`sync` 命令 SHALL 从远程 API 发现仓库信息，将新发现的条目追加到对应 group 的 repos 字段。sync 命令 SHALL NOT 执行 clone 或 pull 操作。sync 命令 SHALL 跳过 `enabled: false` 的 group，不对禁用的 group 执行远程发现。sync 命令 SHALL 跳过未绑定 resource 的 group，不对手动管理的 group 执行远程发现，并输出提示信息。sync 命令 SHALL 在写入配置时保留 group 的 `exclude_repos` 列表，不被覆盖或清空。sync 命令 SHALL 在发现仓库后跳过匹配 group `exclude_repos` 列表的仓库，不将其写入配置。

对于 Codeup provider，sync 命令 SHALL 使用 Groups 查询模式（与 GitLab 一致），通过 `ListRepos` 全量拉取仓库后按代码组路径过滤。

#### Scenario: 同步所有 group
- **WHEN** 用户运行 `grepom sync`（无参数）
- **THEN** 系统对所有配置中启用（`enabled: true`）且绑定了 resource 的 group 执行远程发现，将新发现的仓库追加到各自 group 的 repos 列表，不执行 clone 或 pull

#### Scenario: 同步指定 group
- **WHEN** 用户运行 `grepom sync --group frontend`
- **THEN** 系统仅对 name 为 `frontend` 的 group 执行远程发现并更新其 repos 列表

#### Scenario: 同步指定 resource 下的所有 group
- **WHEN** 用户运行 `grepom sync --resource work-gl`
- **THEN** 系统对所有引用 `work-gl` 资源且启用（`enabled: true`）的 group 执行远程发现并更新各自 repos 列表

#### Scenario: 同步时无新内容
- **WHEN** 用户运行 `grepom sync` 且没有新仓库
- **THEN** 系统不修改配置文件，输出同步摘要

#### Scenario: GitLab recursive group 发现所有层级项目
- **WHEN** GitLab group `frontend`（recursive: true）下有直接项目和子 group `ui` 下的项目
- **THEN** 系统将所有层级的项目作为 repo 条目追加到 `frontend` group 的 repos 列表，repo 的 path 保持远端完整路径（如 `my-org/frontend/ui/design-system`）

#### Scenario: Codeup group 发现所有层级代码库
- **WHEN** Codeup group `solo`（path: `wii/solo`）下有 `wii/solo/grepom` 和 `wii/solo/deep/project`
- **THEN** 系统将所有 pathWithNamespace 以 `wii/solo/` 开头的代码库作为 repo 条目追加到 group 的 repos 列表

#### Scenario: 发现新 repo 不在全局 repos 中
- **WHEN** sync 从 group `frontend` 发现新 repo
- **THEN** 系统 SHALL 将该 repo 追加到 group `frontend` 的 `repos` 字段中，而非顶层全局 `repos`

#### Scenario: 同步时跳过禁用的 group
- **WHEN** 用户运行 `grepom sync`，group `frontend` 设置 `enabled: false`
- **THEN** 系统跳过 `frontend`，不对其执行远程发现

#### Scenario: 同步时跳过禁用 resource 下的 group
- **WHEN** 用户运行 `grepom sync`，resource `work-gl` 设置 `enabled: false`，group `frontend` 和 `backend` 引用该 resource
- **THEN** 系统跳过 `frontend` 和 `backend`，不执行远程发现

#### Scenario: 同步时跳过无 resource 的 group
- **WHEN** 用户运行 `grepom sync`，group `manual-group` 未绑定 resource
- **THEN** 系统跳过 `manual-group`，不对其执行远程发现，并输出提示"group manual-group: 未绑定 resource，跳过 sync"

#### Scenario: 同步时跳过 exclude_repos 中的仓库
- **WHEN** 用户运行 `grepom sync`，group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`，远程发现了 `deprecated-app` 和 `new-app`
- **THEN** 系统仅将 `new-app` 写入配置，跳过 `deprecated-app`

#### Scenario: 同步跳过被排除仓库时 verbose 输出
- **WHEN** 用户运行 `grepom sync -v`，group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`，远程发现了 5 个仓库其中 1 个被排除
- **THEN** 系统输出该 group 跳过了 1 个被排除的仓库

#### Scenario: 同步摘要包含被排除仓库数量
- **WHEN** 用户运行 `grepom sync`，group `frontend` 的 `exclude_repos` 包含 2 个仓库，远程发现了 10 个仓库
- **THEN** 系统的同步摘要中显示发现了 10 个仓库，但仅保存了 8 个（非排除的）新仓库

#### Scenario: Codeup provider 使用 Groups 模式
- **WHEN** 用户运行 `grepom sync`，group 的 resource provider 为 `codeup`
- **THEN** 系统使用 Groups 查询模式构建 `ListReposParams`，传入 `Groups` 字段（非 `Orgs`）

### Requirement: sync 配置更新策略（只增不删）
sync 命令在更新配置文件时 SHALL 仅追加新发现的 repo 条目到对应 group 的 repos 列表，不删除或修改已有条目。sync 命令 SHALL 在写入配置时保留 group 的 `exclude_repos` 列表。sync 命令 SHALL 在写入前过滤掉匹配 group `exclude_repos` 的新发现仓库。

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

#### Scenario: sync 保留 exclude_repos 配置
- **WHEN** group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`，sync 发现新 repo 并写入配置
- **THEN** 写入后的配置中 `exclude_repos` 仍为 `["deprecated-app"]`，不被清空或覆盖

#### Scenario: sync 不写入被排除的仓库
- **WHEN** group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`，远程新发现了 `deprecated-app`
- **THEN** `deprecated-app` 不被写入配置文件

#### Scenario: 已在配置中的被排除仓库不受 sync 影响
- **WHEN** group `frontend` 的 repos 中已有 `deprecated-app`，`exclude_repos` 为 `["deprecated-app"]`
- **THEN** sync 不删除已有的 `deprecated-app` 条目（sync 只增不删策略不变）

### Requirement: sync 并发写入保护
当多个 sync 实例同时运行时，系统 SHALL 使用文件锁防止配置文件的并发写入冲突。

#### Scenario: 多个 sync 实例同时运行
- **WHEN** 两个 `grepom sync` 进程同时运行并都需要写入配置文件
- **THEN** 第二个进程等待第一个进程完成写入后再执行自己的写入，不会导致配置文件损坏或数据丢失

#### Scenario: 获取锁超时
- **WHEN** sync 进程无法在合理时间内获取配置文件锁
- **THEN** 系统报告错误并退出，提示用户另一个 sync 正在运行
