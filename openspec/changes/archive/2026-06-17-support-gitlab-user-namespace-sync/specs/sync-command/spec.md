## ADDED Requirements

### Requirement: sync 支持 GitLab 个人命名空间路径
当 group 绑定的 resource provider 为 GitLab 且 group path 指向个人命名空间时，`grepom sync` SHALL 能发现并保存该命名空间下 token 可见的仓库。

#### Scenario: GitLab 个人命名空间同步成功
- **WHEN** 用户运行 `grepom sync`，group `sunzhenkai` 的 path 为 `sunzhenkai`，provider 为 `gitlab`
- **THEN** 系统 SHALL 不因 `404 Group Not Found` 直接失败
- **AND** 系统 SHALL 发现并写入 `path` 以 `sunzhenkai/` 开头的仓库

#### Scenario: GitLab 个人命名空间无可见仓库
- **WHEN** 用户运行 `grepom sync`，group path 指向个人命名空间，但 token 无可见仓库
- **THEN** 系统 SHALL 将该 group 视为“同步成功但无新增仓库”，并继续处理其他 group
