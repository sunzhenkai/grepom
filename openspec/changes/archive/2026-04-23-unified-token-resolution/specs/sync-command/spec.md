## MODIFIED Requirements

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
