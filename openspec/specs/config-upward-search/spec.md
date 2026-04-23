### Requirement: 配置文件向上遍历查找
当 `FindConfig` 未指定显式路径且当前目录不存在 `.grepom.yml` 时，系统 SHALL 沿父目录链向上遍历查找 `.grepom.yml`，直到找到配置文件或到达文件系统根目录。

#### Scenario: 当前目录存在配置文件
- **WHEN** 当前目录存在 `.grepom.yml`
- **THEN** 直接返回当前目录的 `.grepom.yml` 路径，不进行向上遍历

#### Scenario: 父目录存在配置文件
- **WHEN** 当前目录为 `/home/user/projects/my-org/web-app`，当前目录无 `.grepom.yml`，但 `/home/user/projects/` 存在 `.grepom.yml`
- **THEN** 返回 `/home/user/projects/.grepom.yml`

#### Scenario: 多级向上查找
- **WHEN** 当前目录为 `/a/b/c/d`，仅 `/a/` 存在 `.grepom.yml`
- **THEN** 依次检查 `/a/b/c/d`、`/a/b/c`、`/a/b`，最终在 `/a/` 找到并返回

#### Scenario: 遍历到文件系统根仍未找到
- **WHEN** 从当前目录一直遍历到文件系统根目录，均未找到 `.grepom.yml`
- **THEN** 返回 `ErrConfigNotFound` 错误，提示用户使用 `-c` 指定或创建配置文件

#### Scenario: 显式路径优先
- **WHEN** 用户通过 `-c` 指定了配置文件路径
- **THEN** 直接使用指定路径，不进行向上遍历

#### Scenario: 嵌套配置就近原则
- **WHEN** `/home/user/projects/.grepom.yml` 和 `/home/user/projects/personal/.grepom.yml` 均存在，当前目录为 `/home/user/projects/personal/blog`
- **THEN** 返回 `/home/user/projects/personal/.grepom.yml`（最近的配置文件）
