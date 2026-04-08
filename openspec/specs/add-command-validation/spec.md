## ADDED Requirements

### Requirement: add group 校验引用的 resource 存在
`add group` 命令在执行时 SHALL 先加载已有配置，检查 `--resource` 指定的 resource 名称是否存在于配置文件的 `resources` 中。若不存在，命令 SHALL 立即报错退出，不写入配置。

#### Scenario: 添加 group 引用不存在的 resource
- **WHEN** 用户运行 `grepom add group --name frontend --resource nonexistent --path my-org/frontend`
- **THEN** 系统报错 "resource \"nonexistent\" not found"，不修改配置文件

#### Scenario: 添加 group 引用已存在的 resource
- **WHEN** 用户运行 `grepom add group --name frontend --resource work-gl --path my-org/frontend`，且配置中存在 resource `work-gl`
- **THEN** 系统正常追加 group 条目到配置文件

### Requirement: add group 校验名称唯一性
`add group` 命令 SHALL 检查待添加的 group 名称是否与已有 group 名称重复。若重复，命令 SHALL 立即报错退出。

#### Scenario: 添加重名 group
- **WHEN** 配置中已存在 group `frontend`，用户运行 `grepom add group --name frontend --resource work-gl --path other/frontend`
- **THEN** 系统报错 "group \"frontend\" already exists"，不修改配置文件

#### Scenario: 添加不重名 group
- **WHEN** 配置中已存在 group `frontend`，用户运行 `grepom add group --name backend --resource work-gl --path my-org/backend`
- **THEN** 系统正常追加 group 条目

### Requirement: add repo（standalone）校验引用的 resource 存在
`add repo` 命令（非 group repo）在执行时 SHALL 检查 `--resource` 指定的 resource 名称是否存在于配置中。若不存在，命令 SHALL 立即报错退出。

#### Scenario: 添加 standalone repo 引用不存在的 resource
- **WHEN** 用户运行 `grepom add repo --name dotfiles --resource nonexistent --url https://github.com/me/dotfiles.git`
- **THEN** 系统报错 "resource \"nonexistent\" not found"，不修改配置文件

#### Scenario: 添加 standalone repo 引用已存在的 resource
- **WHEN** 用户运行 `grepom add repo --name dotfiles --resource github --url https://github.com/me/dotfiles.git`，且配置中存在 resource `github`
- **THEN** 系统正常追加 repo 条目

### Requirement: add repo（standalone）校验名称唯一性
`add repo` 命令（非 group repo）SHALL 检查待添加的 repo 名称是否与已有 standalone repo 名称重复。若重复，命令 SHALL 立即报错退出。

#### Scenario: 添加重名 standalone repo
- **WHEN** 配置中已存在 standalone repo `dotfiles`，用户运行 `grepom add repo --name dotfiles --resource github --url https://github.com/other/dotfiles.git`
- **THEN** 系统报错 "repo \"dotfiles\" already exists"，不修改配置文件

#### Scenario: 添加不重名 standalone repo
- **WHEN** 配置中已存在 standalone repo `dotfiles`，用户运行 `grepom add repo --name config --resource github --url https://github.com/me/config.git`
- **THEN** 系统正常追加 repo 条目

### Requirement: add repo to group 校验 group 存在
`add repo --group <name>` 命令 SHALL 检查指定的 group 是否存在于配置中。若不存在，命令 SHALL 立即报错退出。

#### Scenario: 添加 repo 到不存在的 group
- **WHEN** 用户运行 `grepom add repo --name special --group nonexistent --url https://...`
- **THEN** 系统报错 "group \"nonexistent\" not found"，不修改配置文件

#### Scenario: 添加 repo 到已存在的 group
- **WHEN** 用户运行 `grepom add repo --name special --group frontend --url https://...`，且配置中存在 group `frontend`
- **THEN** 系统正常追加 repo 到 group 的 repos 列表

### Requirement: add repo to group 校验 resource 存在（通过 group 间接引用）
当通过 `--group` 添加 repo 时，命令 SHALL 通过 group 的 resource 字段间接验证配置有效性。由于 AddGroupRepo 已通过 Load 加载配置，resource 引用有效性由 Load.validate 保证。

#### Scenario: group 引用有效 resource 时正常添加 repo
- **WHEN** group `frontend` 引用了有效的 resource `work-gl`，用户运行 `grepom add repo --name special --group frontend --url https://...`
- **THEN** 系统正常追加 repo 到 group `frontend` 的 repos 列表

### Requirement: add resource 命令的 provider 验证
`add resource` 命令 SHALL 验证 `--provider` 参数值为合法的 provider 类型（gitlab、github 或 generic）。不合法时 SHALL 报错并列出所有支持的 provider。

#### Scenario: 使用 generic provider 添加资源
- **WHEN** 用户运行 `grepom add resource --name my-git --provider generic --url git.internal.com --token ${GIT_TOKEN}`
- **THEN** 系统正常添加资源，provider 为 `generic`

#### Scenario: 使用不支持的 provider 添加资源
- **WHEN** 用户运行 `grepom add resource --name test --provider unknown --url example.com`
- **THEN** 系统报错提示不支持的 provider，并列出 `gitlab`、`github`、`generic`

#### Scenario: provider 参数缺失
- **WHEN** 用户运行 `grepom add resource --name test --url example.com`（未指定 `--provider`）
- **THEN** 系统报错提示 `--provider` 是必填参数
