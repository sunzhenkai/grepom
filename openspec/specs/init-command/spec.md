### Requirement: init 命令创建配置文件
系统 SHALL 提供 `grepom init` 命令，在当前目录创建 `.grepom.yml` 配置文件。配置文件 SHALL 使用新格式包含 `base`、`resources`、`groups`、`repos` 字段。

#### Scenario: 最小初始化
- **WHEN** 用户在当前目录运行 `grepom init`，且 `.grepom.yml` 不存在
- **THEN** 系统创建 `.grepom.yml`，内容包含 `base: ~/projects`，resources 为空 map，groups 为空数组，repos 为空数组

#### Scenario: 指定 base 目录
- **WHEN** 用户运行 `grepom init --base ~/work/repos`
- **THEN** 系统创建 `.grepom.yml`，`base` 字段值为 `~/work/repos`

#### Scenario: 配置文件已存在
- **WHEN** 用户运行 `grepom init`，且 `.grepom.yml` 已存在
- **THEN** 系统报错，提示配置文件已存在，不覆盖

### Requirement: init 命令可选添加 resource
系统 SHALL 支持在初始化配置文件的同时添加第一个 resource，通过以下 flag 控制：
- `--provider`: provider 类型（gitlab 或 github）
- `--url`: API base URL
- `--token`: 认证令牌（可使用 ${ENV_VAR} 语法）

#### Scenario: 初始化时添加 GitLab resource
- **WHEN** 用户运行 `grepom init --provider gitlab --url https://gitlab.com --token ${GITLAB_TOKEN}`
- **THEN** 系统创建 `.grepom.yml`，包含 base 和一个名为自动生成名称的 gitlab resource 条目

#### Scenario: 指定了 provider 但未指定 token
- **WHEN** 用户运行 `grepom init --provider gitlab --url https://gitlab.com`（无 --token）
- **THEN** 系统创建仅包含 base 的配置文件，不创建 resource 条目

### Requirement: init 命令支持指定输出路径
系统 SHALL 支持 `-c` / `--config` 全局 flag 指定配置文件路径。

#### Scenario: 指定配置文件路径
- **WHEN** 用户运行 `grepom -c ~/work.yml init`
- **THEN** 系统在 `~/work.yml` 创建配置文件
