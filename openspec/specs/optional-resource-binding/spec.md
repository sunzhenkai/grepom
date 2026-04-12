### Requirement: Group 的 resource 字段可选
Group 的 `resource` 字段 SHALL 为可选。当未指定 resource 时，该 group 视为手动管理模式，repos 列表由用户自行维护。

#### Scenario: 定义无 resource 的 group
- **WHEN** 配置文件中 group 未指定 `resource` 字段
- **THEN** 系统正常加载该 group，标记为手动管理模式

#### Scenario: 无 resource 的 group 手动维护 repos
- **WHEN** group 未绑定 resource，用户手动在 repos 列表中添加了 repo 条目（包含 name、url、path）
- **THEN** 系统正常加载这些 repo，clone/pull 时直接使用 repo 的 url 字段

### Requirement: Standalone repo 的 resource 字段可选
Standalone repo 的 `resource` 字段 SHALL 为可选。当未指定 resource 时，`url` 字段 SHALL 为必填。

#### Scenario: 定义无 resource 的 standalone repo
- **WHEN** 配置文件中 standalone repo 未指定 `resource` 字段但指定了 `url: git@github.com:user/repo.git`
- **THEN** 系统正常加载该 repo，clone/pull 时使用 url 字段的值

#### Scenario: 无 resource 且无 url 的 standalone repo 报错
- **WHEN** 配置文件中 standalone repo 未指定 `resource` 且未指定 `url`
- **THEN** 系统 SHALL 在加载配置时报错，提示必须提供 resource 或 url

### Requirement: 无 resource 时 clone/pull 使用系统默认认证
当 group 或 repo 未绑定 resource 时，clone 和 pull 操作 SHALL 使用系统默认的 SSH/HTTPS 认证，不注入额外 token 或 SSH key。

#### Scenario: 无 resource 的 repo 使用 SSH clone
- **WHEN** 无 resource 的 repo 的 url 为 `git@github.com:user/repo.git`，系统 SSH agent 已配置对应密钥
- **THEN** clone/pull 使用系统 SSH agent 认证，正常完成操作

#### Scenario: 无 resource 的 repo 使用 HTTPS clone
- **WHEN** 无 resource 的 repo 的 url 为 `https://github.com/user/repo.git`
- **THEN** clone/pull 使用系统 git credential 认证（或公开仓库无需认证）

### Requirement: 依赖 resource 的操作在缺失时跳过并提示
sync、list --remote 等依赖 resource 的操作，在 group 或 repo 未绑定 resource 时 SHALL 跳过执行并输出提示信息。

#### Scenario: sync 跳过无 resource 的 group
- **WHEN** 用户运行 `grepom sync`，group `frontend` 未绑定 resource
- **THEN** 系统跳过 `frontend` 的远程发现，输出提示"group frontend: 未绑定 resource，跳过 sync"

#### Scenario: list --remote 跳过无 resource 的 group
- **WHEN** 用户运行 `grepom list --remote`，group `frontend` 未绑定 resource
- **THEN** 系统跳过 `frontend` 的远程列表查询，输出提示信息

#### Scenario: 多个 group 中部分无 resource
- **WHEN** 用户运行 `grepom sync`，group A 绑定了 resource，group B 未绑定
- **THEN** 系统正常执行 group A 的 sync，跳过 group B 并输出提示
