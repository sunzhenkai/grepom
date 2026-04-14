## ADDED Requirements

### Requirement: 配置文件支持自定义 YAML 缩进空格数量
系统 SHALL 支持在配置文件中通过 `yaml_indent` 字段指定 YAML 写入时的缩进空格数量。该字段为可选整数，未设置时默认使用 2 空格缩进。

#### Scenario: 配置中指定 4 空格缩进
- **WHEN** 配置文件中设置 `yaml_indent: 4` 且系统写入配置文件（如 add、sync 操作）
- **THEN** 生成的 YAML 文件使用 4 空格缩进

#### Scenario: 未设置 yaml_indent 时使用默认值
- **WHEN** 配置文件中未设置 `yaml_indent` 字段
- **THEN** 系统写入配置文件时使用默认的 2 空格缩进

#### Scenario: yaml_indent 设置为 2
- **WHEN** 配置文件中设置 `yaml_indent: 2`
- **THEN** 生成的 YAML 文件使用 2 空格缩进

### Requirement: yaml_indent 字段不写入配置文件
系统 SHALL 在写入配置文件时不将 `yaml_indent` 字段本身写入文件。该字段仅作为运行时行为控制，不应出现在持久化的配置文件中。

#### Scenario: yaml_indent 不出现在输出文件中
- **WHEN** 用户配置中设置 `yaml_indent: 4` 并执行 `grepom add` 操作触发配置写入
- **THEN** 生成的 YAML 文件中不包含 `yaml_indent` 字段，但缩进为 4 空格

#### Scenario: yaml_indent 值在加载后被保留
- **WHEN** 用户加载包含 `yaml_indent: 4` 的配置文件后修改配置并保存
- **THEN** 保存的文件使用 4 空格缩进，但不包含 `yaml_indent` 字段本身
