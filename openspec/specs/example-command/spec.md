### Requirement: example 命令导出示例配置
系统 SHALL 提供 `grepom example` 子命令，将包含全部功能和字段注释的完整示例 YAML 配置输出到 stdout。命令的 cobra 描述（Short、Long）、flag 帮助文本和输出的示例配置中的 YAML 注释 SHALL 全部使用英文。

#### Scenario: example 命令帮助信息为英文
- **WHEN** 用户运行 `grepom example --help`
- **THEN** 系统显示英文的命令描述和 flag 说明

#### Scenario: 示例配置注释为英文
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出的示例 YAML 配置中所有注释均为英文，描述每个字段的用途和可选值

#### Scenario: 输出到文件时的提示为英文
- **WHEN** 用户运行 `grepom example --output my-config.yml`
- **THEN** 系统将示例配置写入文件并输出英文提示信息

#### Scenario: 默认输出到 stdout
- **WHEN** 用户运行 `grepom example`
- **THEN** 系统将完整示例配置输出到 stdout，包含 github、gitlab、generic 三种 provider 的 resource 示例、group 示例、独立 repo 示例，以及每个字段的 YAML 注释说明

#### Scenario: 输出到文件
- **WHEN** 用户运行 `grepom example --output my-config.yml`
- **THEN** 系统将示例配置写入 `my-config.yml` 文件，并输出提示信息

#### Scenario: 输出到文件使用短标志
- **WHEN** 用户运行 `grepom example -o my-config.yml`
- **THEN** 系统行为与 `--output` 完全一致

#### Scenario: 输出文件已存在
- **WHEN** 用户运行 `grepom example -o existing.yml`，且 `existing.yml` 已存在
- **THEN** 系统报错提示文件已存在，不覆盖

### Requirement: example 命令示例配置内容完整性
示例配置 SHALL 包含所有支持的配置字段，每个字段附带 YAML 注释说明其用途和可选值。注释 SHALL 覆盖 resources map 格式、repo.url 相对/绝对写法、克隆协议行为、ssh_key 建议，以及 virtual_groups 的 repos 字段。

#### Scenario: 示例配置包含所有 provider 类型
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出的示例配置包含 `github`、`gitlab`、`generic` 三种 provider 的 resource 示例

#### Scenario: 示例配置包含所有可选字段
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出包含 `ssh_key`、`enabled`、`exclude_repos`、`recursive`、`local_path`、`token`（group/repo 级别覆盖）、`virtual_groups.repos` 等所有可选字段

### Requirement: example 说明 URL 写法与认证建议
`grepom example` 输出的示例配置 SHALL 用英文 YAML 注释说明：

1. `resources` 使用 map 格式（key 为资源名）
2. 绑定 resource 时，`repo.url` 可为相对路径（与 resource host 拼接）或完整 `https://` / `git@` / `ssh://` URL（完整 URL 不再二次拼接）
3. 完整 HTTPS URL 可匿名克隆公开仓；内网 SSH 场景建议配置 `ssh_key`
4. `virtual_groups` 可同时列出 `groups` 与顶层独立 `repos`

示例中的 host、组织与仓库路径 MUST 为虚构占位符（如 `git.example.com`、`my-org/app`），MUST NOT 包含真实私密仓库信息。

#### Scenario: example 含相对路径与完整 HTTPS 示例
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出中至少包含一条相对路径 repo.url 示例与一条完整 HTTPS repo.url 示例，并附带英文注释说明解析差异

#### Scenario: example 含 virtual_groups.repos
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出的 `virtual_groups` 示例包含 `repos` 字段注释或示例项

#### Scenario: example 推荐 ssh_key
- **WHEN** 用户运行 `grepom example`
- **THEN** 输出注释说明建议在 resource 上显式配置 `ssh_key`，尤其是 SSH agent 不可用时
