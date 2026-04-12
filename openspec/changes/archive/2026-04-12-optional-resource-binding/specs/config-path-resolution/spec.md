## MODIFIED Requirements

### Requirement: Group 顶层定义
系统 SHALL 支持在配置文件的 `groups` 字段下定义顶层 group 列表。每个 group 包含 `name`（唯一标识）和 `local_path`（本地映射根路径）。`resource`（引用的认证资源）为可选字段，未指定时 group 为手动管理模式。`path`（远端 group path）在未指定 resource 时为可选。Group SHALL 支持 `enabled` 布尔字段（默认 `true`）和 `exclude_repos` 字符串数组字段（默认为空）。`groups` 使用 YAML 数组格式。

#### Scenario: 定义 GitLab group
- **WHEN** 配置文件中 `groups` 包含 `{name: frontend, resource: work-gl, path: my-org/frontend, local_path: ./frontend, recursive: true}`
- **THEN** 系统加载该 group，使用 work-gl 资源认证，远端路径为 my-org/frontend，本地映射到 ./frontend

#### Scenario: 定义 GitHub group（使用 org 语义）
- **WHEN** 配置文件中 `groups` 包含 `{name: oss, resource: github, path: my-oss-org, local_path: ./open-source}`
- **THEN** 系统加载该 group，使用 github 资源认证，远端路径为 my-oss-org

#### Scenario: 定义无 resource 的手动管理 group
- **WHEN** 配置文件中 `groups` 包含 `{name: local-tools, local_path: ./tools}`
- **THEN** 系统正常加载该 group，标记为手动管理模式，repos 由用户自行维护

#### Scenario: Group name 必须唯一
- **WHEN** 配置文件中 `groups` 数组存在多个 name 相同的 group
- **THEN** 系统 SHALL 在加载配置时报错，提示 group name 冲突

#### Scenario: Group name 必填
- **WHEN** 配置文件中某 group 缺少 `name` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: Group resource 可选
- **WHEN** 配置文件中某 group 未指定 `resource` 字段
- **THEN** 系统 SHALL 正常加载该 group，该 group 为手动管理模式

#### Scenario: 有 resource 的 Group path 必填
- **WHEN** 配置文件中某 group 绑定了 resource 但缺少 `path` 字段
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: 无 resource 的 Group path 可选
- **WHEN** 配置文件中某 group 未绑定 resource 且未指定 `path` 字段
- **THEN** 系统 SHALL 正常加载该 group

#### Scenario: 定义禁用的 Group
- **WHEN** 配置文件中 group 设置 `enabled: false`
- **THEN** 系统正常加载该 group 配置，但运行时排除该 group 下所有 repo

#### Scenario: 定义带 exclude_repos 的 Group
- **WHEN** 配置文件中 group 设置 `exclude_repos: [deprecated-app, temp-repo]`
- **THEN** 系统正常加载该 group 配置，但运行时排除 `exclude_repos` 中列出的 repo
