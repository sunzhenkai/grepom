## Purpose

定义 `grepom sync` 命令从远程 API 发现仓库并更新配置的行为。
## Requirements
### Requirement: sync 命令
`sync` 命令 SHALL 从远程 API 发现仓库信息，将新发现的条目追加到对应 group 的 repos 字段。sync 命令 SHALL NOT 执行 clone 或 pull 操作。sync 命令 SHALL 通过 `Resource.ResolvedToken()` 获取已解析的 token 用于 provider API 认证，不再直接读取 `Resource.Token` 字段。

sync 命令 SHALL 跳过 `enabled: false` 的 group，不对禁用的 group 执行远程发现。sync 命令 SHALL 跳过未绑定 resource 的 group，不对手动管理的 group 执行远程发现，并输出提示信息。sync 命令 SHALL 在写入配置时保留 group 的 `exclude_repos` 列表，不被覆盖或清空。sync 命令 SHALL 在发现仓库后跳过匹配 group `exclude_repos` 列表的仓库，不将其写入配置。

对于 Codeup provider，sync 命令 SHALL 使用 Groups 查询模式（与 GitLab 一致），通过 `ListRepos` 全量拉取仓库后按代码组路径过滤。

#### Scenario: 同步所有 group
- **WHEN** 用户运行 `grepom sync`（无参数）
- **THEN** 系统对所有配置中启用（`enabled: true`）且绑定了 resource 的 group 执行远程发现，将新发现的仓库追加到各自 group 的 repos 列表，不执行 clone 或 pull

#### Scenario: sync 使用已解析的 token
- **WHEN** 用户运行 `grepom sync`，resource token 为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 已设置为 `glpat-xxx`
- **THEN** sync 命令通过 `res.ResolvedToken()` 获取 `glpat-xxx`，使用该值作为 provider API 认证 token

#### Scenario: sync token 环境变量未设置时报错
- **WHEN** 用户运行 `grepom sync`，resource token 为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 未设置
- **THEN** sync 命令在调用 provider API 前通过 `res.ResolvedToken()` 获得错误，向 stderr 输出包含 resource 名称和环境变量名的错误信息，继续处理其他 group

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

#### Scenario: GitLab 个人命名空间同步成功
- **WHEN** 用户运行 `grepom sync`，group `sunzhenkai` 的 path 为 `sunzhenkai`，provider 为 `gitlab`
- **THEN** 系统 SHALL 不因 `404 Group Not Found` 直接失败
- **AND** 系统 SHALL 发现并写入 `path` 以 `sunzhenkai/` 开头的仓库

#### Scenario: GitLab 个人命名空间无可见仓库
- **WHEN** 用户运行 `grepom sync`，group path 指向个人命名空间，但 token 无可见仓库
- **THEN** 系统 SHALL 将该 group 视为“同步成功但无新增仓库”，并继续处理其他 group

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
sync 命令在更新配置文件时 SHALL 仅追加新发现的 repo 条目到对应 group 的 repos 列表，不删除或修改已有条目。sync 命令 SHALL 在写入配置时保留 group 的 `exclude_repos` 列表。sync 命令 SHALL 在写入前过滤掉匹配 group `exclude_repos` 的新发现仓库。sync 命令 SHALL 对单次批次内的新增仓库进行自去重，确保同一 URL 的仓库在批次内不重复写入。

#### Scenario: 远程新增仓库
- **WHEN** group `frontend` 下远程新增了 repo `new-app`
- **THEN** 系统将 `new-app` 追加到 group `frontend` 的 repos 列表

#### Scenario: Repo 已存在于 group repos
- **WHEN** 远程 repo `shared-utils` 已存在于 group `frontend` 的 repos 列表中（按 URL 匹配）
- **THEN** 系统不重复追加

#### Scenario: 批次内包含重复 URL 的仓库
- **WHEN** 单次 sync 从 provider 获取的仓库列表中，同一 URL 出现多次（如 `https://gitlab.com/org/app.git` 出现两次）
- **THEN** 系统 SHALL 仅保留第一个出现的条目，不将重复条目写入配置文件

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

#### Scenario: sync 跳过 path 不匹配 group path 的远程仓库
- **WHEN** 用户同步 group `topon-bidder`（path 为 `topon-bidder`），远程 API 返回 path 为 `zhangfeixiang/ai-coding` 的共享仓库
- **THEN** 系统 SHALL 跳过该仓库，不写入配置文件，避免后续加载配置时报 path 校验错误

