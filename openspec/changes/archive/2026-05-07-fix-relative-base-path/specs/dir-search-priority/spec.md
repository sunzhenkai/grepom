## MODIFIED Requirements

### Requirement: 无参数输出配置文件所在目录
当 `grepom dir` 不带任何参数调用时，系统 SHALL 输出 `.grepom.yml` 配置文件所在的目录的绝对路径。

#### Scenario: 配置文件在当前目录
- **WHEN** 用户在 `/home/user/projects` 目录下执行 `grepom dir`，且该目录存在 `.grepom.yml`
- **THEN** 系统输出 `/home/user/projects`

#### Scenario: 配置文件在父目录
- **WHEN** 用户在 `/home/user/projects/subdir` 目录下执行 `grepom dir`，且 `.grepom.yml` 位于 `/home/user/projects/.grepom.yml`
- **THEN** 系统输出 `/home/user/projects`

#### Scenario: 通过 -c 指定配置文件
- **WHEN** 用户执行 `grepom -c /custom/path/.grepom.yml dir`
- **THEN** 系统输出 `/custom/path`

### Requirement: 有参数时输出绝对路径
当 `grepom dir <keyword>` 带参数调用时，无论 `base` 配置为绝对路径或相对路径，系统 SHALL 输出匹配仓库的绝对本地路径。

#### Scenario: base 为相对路径时仍输出绝对路径
- **WHEN** `.grepom.yml` 的 `base` 为 `repos/my-org`（相对路径），用户在项目子目录执行 `grepom dir web-app`
- **THEN** 系统输出绝对路径（如 `/home/user/projects/repos/my-org/frontend/web-app`）

#### Scenario: base 为相对路径时从子目录调用
- **WHEN** `.grepom.yml` 的 `base` 为 `repos/my-org`，用户在 `repos/my-org/frontend/web-app` 子目录中执行 `grepom dir web-api`
- **THEN** 系统输出绝对路径（如 `/home/user/projects/repos/my-org/backend/web-api`），而非相对路径
