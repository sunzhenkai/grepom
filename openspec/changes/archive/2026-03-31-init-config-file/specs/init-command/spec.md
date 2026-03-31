## ADDED Requirements

### Requirement: init 命令创建配置文件
系统 SHALL 提供 `grepom init` 命令，在当前目录创建 `.grepom.yml` 配置文件。配置文件 SHALL 包含 `base` 字段，默认值为 `~/projects`。

#### Scenario: 最小初始化
- **WHEN** 用户在当前目录运行 `grepom init`，且 `.grepom.yml` 不存在
- **THEN** 系统创建 `.grepom.yml`，内容包含 `base: ~/projects`，sources 和 repos 为空

#### Scenario: 指定 base 目录
- **WHEN** 用户运行 `grepom init --base ~/work/repos`
- **THEN** 系统创建 `.grepom.yml`，`base` 字段值为 `~/work/repos`

#### Scenario: 配置文件已存在
- **WHEN** 用户运行 `grepom init`，且 `.grepom.yml` 已存在
- **THEN** 系统报错，提示配置文件已存在，不覆盖

### Requirement: init 命令可选添加 source
系统 SHALL 支持在初始化配置文件的同时添加第一个 API source，通过以下 flag 控制：
- `--provider`: provider 类型（gitlab 或 github）
- `--url`: API base URL
- `--group`: GitLab group 路径（可多次指定）
- `--org`: GitHub org 名称（可多次指定）
- `--recursive`: 递归获取 GitLab subgroup

#### Scenario: 初始化时添加 GitLab source
- **WHEN** 用户运行 `grepom init --provider gitlab --url https://gitlab.com --group my-org/frontend`
- **THEN** 系统创建 `.grepom.yml`，包含 base 和一个 gitlab source 条目

#### Scenario: 初始化时添加 GitHub source
- **WHEN** 用户运行 `grepom init --provider github --url https://github.com --org my-org`
- **THEN** 系统创建 `.grepom.yml`，包含 base 和一个 github source 条目

#### Scenario: 指定了 provider 但未指定 group 或 org
- **WHEN** 用户运行 `grepom init --provider gitlab --url https://gitlab.com`（无 --group 或 --org）
- **THEN** 系统创建仅包含 base 的配置文件，不创建 source 条目

### Requirement: init 命令支持指定输出路径
系统 SHALL 支持 `-c` / `--config` 全局 flag 指定配置文件路径。

#### Scenario: 指定配置文件路径
- **WHEN** 用户运行 `grepom -c ~/work.yml init`
- **THEN** 系统在 `~/work.yml` 创建配置文件
