## MODIFIED Requirements

### Requirement: sync 命令
`sync` 命令 SHALL 从远程 API 发现仓库信息，将新发现的条目追加到对应 group 的 repos 字段。sync 命令 SHALL NOT 执行 clone 或 pull 操作。sync 命令 SHALL 跳过 `enabled: false` 的 group，不对禁用的 group 执行远程发现。sync 命令 SHALL 跳过未绑定 resource 的 group，不对手动管理的 group 执行远程发现，并输出提示信息。sync 命令 SHALL 在写入配置时保留 group 的 `exclude_repos` 列表，不被覆盖或清空。sync 命令 SHALL 在发现仓库后跳过匹配 group `exclude_repos` 列表的仓库，不将其写入配置。

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
