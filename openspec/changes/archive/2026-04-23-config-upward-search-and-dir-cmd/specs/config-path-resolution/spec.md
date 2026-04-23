## MODIFIED Requirements

### Requirement: 配置文件查找策略
配置文件查找 SHALL 支持从当前工作目录沿父目录链向上遍历。查找优先级为：显式指定路径（`-c`） > 当前目录 `.grepom.yml` > 父目录链向上遍历。

#### Scenario: 仅当前目录查找（原有行为）
- **WHEN** 用户在 `/home/user/projects/` 目录执行 `grepom status`，该目录存在 `.grepom.yml`
- **THEN** 直接使用该配置文件

#### Scenario: 当前目录无配置，向上找到（新增行为）
- **WHEN** 用户在 `/home/user/projects/my-org/web-app` 目录执行 `grepom status`，该目录无 `.grepom.yml`，但 `/home/user/projects/` 存在
- **THEN** 使用 `/home/user/projects/.grepom.yml`

#### Scenario: 显式指定路径覆盖一切（原有行为不变）
- **WHEN** 用户执行 `grepom -c /custom/path/config.yml status`
- **THEN** 使用 `/custom/path/config.yml`，不进行任何查找

#### Scenario: 未找到配置文件时的错误提示（更新）
- **WHEN** 从当前目录遍历到根目录均未找到 `.grepom.yml`
- **THEN** 报错提示包含建议：使用 `-c` 指定配置文件，或在任意父目录创建 `.grepom.yml`
