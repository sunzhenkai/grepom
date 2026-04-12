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

### Requirement: Group local_path 字段
Group 的 `local_path` 字段指定该 group 在本地的映射根路径，相对于配置文件的 `base` 字段。如果省略，默认使用 group name 作为 local_path。

#### Scenario: 显式指定 local_path
- **WHEN** 配置文件中 group 的 `local_path: ./frontend`
- **THEN** 该 group 下的 repo 映射到 `<base>/frontend/` 下

#### Scenario: 省略 local_path 使用默认值
- **WHEN** 配置文件中 group 缺少 `local_path` 字段，group name 为 `frontend`
- **THEN** 该 group 的 local_path 默认为 `./frontend`

### Requirement: Group recursive 字段
Group 的 `recursive` 字段（布尔值，默认 false）控制是否递归发现子 group 下的项目。仅对 GitLab provider 有效。

#### Scenario: GitLab recursive=true 发现所有层级
- **WHEN** GitLab group `my-org/frontend`（recursive: true）下有子 group `ui` 和 `ui/components`
- **THEN** sync 发现 my-org/frontend 下所有层级的项目，包括 `ui/` 和 `ui/components/` 下的项目

#### Scenario: GitLab recursive=false 仅发现直接项目
- **WHEN** GitLab group `my-org/frontend`（recursive: false）
- **THEN** sync 仅发现 my-org/frontend 直接包含的项目，不包括子 group 下的项目

#### Scenario: GitHub group 忽略 recursive 字段
- **WHEN** GitHub group 设置了 `recursive: true`
- **THEN** 该字段被忽略，GitHub 按 org 语义列出所有仓库

### Requirement: Group 内 repos 列表
每个 group 可以包含 `repos` 列表，存储该 group 下 sync 发现的或手动添加的 repo。每个 repo 包含 `name`、`url` 和 `path`（远端完整路径）。Group 内 repo 的本地路径通过自动推导得出，不存储 local_path 字段。

#### Scenario: Group 内有多个 repo
- **WHEN** group `frontend` 的 repos 包含三个条目，path 分别为 `my-org/frontend/shared-utils`、`my-org/frontend/ui/design-system`、`my-org/frontend/api/gateway`
- **THEN** 系统正常加载这些 repo，本地路径分别为 `./frontend/shared-utils`、`./frontend/ui/design-system`、`./frontend/api/gateway`

#### Scenario: Group 内 repo 的 path 必须以 group path 为前缀
- **WHEN** group path 为 `my-org/frontend`，但某 repo 的 path 为 `other-org/backend`
- **THEN** 系统 SHALL 在加载配置时报错，提示 repo path 不匹配 group path

#### Scenario: Group repos 列表可以为空
- **WHEN** 配置文件中某 group 的 `repos` 字段为空或省略
- **THEN** 系统正常加载该 group，视为尚无发现的 repo

#### Scenario: Group 内 repo 按 URL 去重
- **WHEN** sync 发现的 repo URL 已存在于该 group 的 repos 列表中
- **THEN** 系统 SHALL 不重复追加该 repo
