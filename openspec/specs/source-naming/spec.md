### Requirement: resource 支持 name 作为 map key
`resources` 使用 YAML map 格式定义，map key 即为资源名称。系统 SHALL 通过 map key 标识和引用资源。

#### Scenario: 通过 resource name 引用
- **WHEN** 配置文件中 `resources` 下存在 key `work-gl`
- **THEN** 用户可通过 `--resource work-gl` 或配置中的 `resource: work-gl` 引用该资源

#### Scenario: resource name 不可为空
- **WHEN** 配置文件中 `resources` map 的 key 为空字符串
- **THEN** 系统 SHALL 在加载配置时报错

#### Scenario: sync 按 resource 过滤
- **WHEN** 用户运行 `grepom sync --resource work-gl`
- **THEN** 系统仅同步引用 `work-gl` 资源的 group
