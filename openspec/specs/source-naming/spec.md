### Requirement: source 支持 name 字段
`Source` 结构体 SHALL 支持可选的 `name` 字段，用于通过名称标识和引用 source。

#### Scenario: 配置中指定 source name
- **WHEN** 配置文件中某 source 包含 `name: my-gitlab`
- **THEN** 用户可通过 `--source my-gitlab` 引用该 source

#### Scenario: 通过 name 引用 source
- **WHEN** 用户运行 `grepom sync --source my-gitlab`
- **THEN** 系统仅同步 name 为 `my-gitlab` 的 source

#### Scenario: name 未指定时回退到索引
- **WHEN** 配置文件中的 source 未指定 `name` 字段
- **THEN** 用户可通过 `--source 0` 按数组索引引用该 source

#### Scenario: source name 重复时报错
- **WHEN** 配置文件中存在多个 name 相同的 source
- **THEN** 系统 SHALL 在加载配置时报错，提示 name 冲突

#### Scenario: name 匹配优先于索引
- **WHEN** 用户运行 `grepom sync --source 0`，且某 source 的 name 为 `"0"`
- **THEN** 系统优先按 name 匹配

### Requirement: add source 支持 --name 参数
`grepom add source` 命令 SHALL 支持 `--name` 参数，用于设置 source 的名称标识。

#### Scenario: 添加带 name 的 source
- **WHEN** 用户运行 `grepom add source --name my-gitlab --provider gitlab --url https://gitlab.com --group my-org`
- **THEN** 系统在配置文件中创建包含 `name: my-gitlab` 的 source 条目

#### Scenario: 不指定 name 时省略字段
- **WHEN** 用户运行 `grepom add source --provider gitlab --url https://gitlab.com --group my-org`（无 --name）
- **THEN** 系统创建的 source 条目不包含 `name` 字段
