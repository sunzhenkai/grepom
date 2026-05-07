## ADDED Requirements

### Requirement: 相对 base 自动解析为绝对路径
当 `.grepom.yml` 的 `base` 字段为相对路径时，系统 SHALL 在加载配置后将其解析为绝对路径，参照点为配置文件所在目录。`base` 为绝对路径或 `~/` 路径时行为不变。

#### Scenario: base 为相对路径时解析为绝对路径
- **WHEN** `.grepom.yml` 位于 `/home/user/projects/.grepom.yml`，`base` 配置为 `repos/my-org`
- **THEN** 系统将 `cfg.Base` 解析为 `/home/user/projects/repos/my-org`

#### Scenario: base 为 ./ 前缀的相对路径
- **WHEN** `.grepom.yml` 位于 `/home/user/projects/.grepom.yml`，`base` 配置为 `./repos`
- **THEN** 系统将 `cfg.Base` 解析为 `/home/user/projects/repos`

#### Scenario: base 为绝对路径时不改变
- **WHEN** `base` 配置为 `/opt/repos`
- **THEN** `cfg.Base` 保持 `/opt/repos` 不变

#### Scenario: base 为 ~ 路径时行为不变
- **WHEN** `base` 配置为 `~/projects`
- **THEN** `expandTilde` 将其扩展为 `/home/user/projects`，后续不再做路径解析

### Requirement: ResolveBasePath 导出方法
`config` 包 SHALL 提供 `ResolveBasePath(cfg *Config, configDir string)` 方法，接受配置对象和配置文件所在目录路径，在 `cfg.Base` 为相对路径时将其解析为绝对路径。

#### Scenario: 从 loadConfig 调用
- **WHEN** `loadConfig` 获取到配置文件绝对路径 `/home/user/projects/.grepom.yml` 和加载后的 `cfg`
- **THEN** 调用 `config.ResolveBasePath(cfg, filepath.Dir(absPath))` 确保 `cfg.Base` 为绝对路径
