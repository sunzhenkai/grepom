## ADDED Requirements

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

### Requirement: 精确匹配优先于子串匹配
当 `grepom dir <keyword>` 带参数调用时，系统 SHALL 先执行 case-insensitive 精确匹配。只有在精确匹配无结果时，才退回 case-insensitive 子串匹配。

#### Scenario: 精确匹配恰好一个仓库
- **WHEN** 仓库列表包含 `ranker` 和 `ranker-service`，用户执行 `grepom dir ranker`
- **THEN** 系统精确匹配到 `ranker`，输出其路径到 stdout（单行），退出码 0

#### Scenario: 精确匹配多个仓库
- **WHEN** 不同 group 下存在两个同名的 `web-app` 仓库，用户执行 `grepom dir web-app`
- **THEN** 系统精确匹配到两个 `web-app`，输出两个路径到 stdout（每行一个），退出码 0

#### Scenario: 无精确匹配退回子串匹配
- **WHEN** 仓库列表包含 `ranker`、`ranker-service`、`ad-ranker`，用户执行 `grepom dir rank`
- **THEN** 系统无精确匹配，退回子串匹配，输出三个路径到 stdout（每行一个），退出码 0

#### Scenario: 精确和子串均无匹配
- **WHEN** 用户执行 `grepom dir nonexistent`
- **THEN** 系统输出错误信息到 stderr，退出码非 0

### Requirement: 多匹配时输出所有路径到 stdout
当搜索匹配到多个仓库时，系统 SHALL 将每个仓库的完整路径输出到 stdout（每行一个），退出码为 0。不再输出表格到 stderr。

#### Scenario: 两个匹配
- **WHEN** 搜索匹配到 `web-app`（路径 `/base/frontend/web-app`）和 `web-api`（路径 `/base/backend/web-api`）
- **THEN** stdout 输出两行路径，退出码 0

#### Scenario: 配合 --group 过滤后多匹配
- **WHEN** 用户执行 `grepom dir web --group frontend`，在 frontend group 中匹配到多个仓库
- **THEN** stdout 输出所有匹配路径，退出码 0

### Requirement: Shell function 运行时检测 fzf
`grepom dir --shell` SHALL 输出固定的 gcd() shell function，在 shell 端运行时检测 fzf 是否可用，不依赖 Go 端检测。

#### Scenario: 输出统一的 shell function
- **WHEN** 用户执行 `grepom dir --shell`
- **THEN** 输出的函数包含 `command -v fzf` 运行时检测，无条件分支固定

#### Scenario: 无参数时 gcd 调用 grepom dir
- **WHEN** 用户执行 `gcd`（无参数）
- **THEN** gcd 函数调用 `grepom dir`（无参数），获取配置文件所在目录并 cd

#### Scenario: 有参数且有 fzf 时 gcd 使用 fzf 选择
- **WHEN** 系统安装了 fzf，用户执行 `gcd web`
- **THEN** gcd 函数通过 `grepom dir web | fzf --select-1` 选择目标目录

#### Scenario: 有参数且无 fzf 时 gcd 使用 head -n 1
- **WHEN** 系统未安装 fzf，用户执行 `gcd web`
- **THEN** gcd 函数通过 `grepom dir web | head -n 1` 取第一个匹配目录

### Requirement: 有参数时输出绝对路径
当 `grepom dir <keyword>` 带参数调用时，无论 `base` 配置为绝对路径或相对路径，系统 SHALL 输出匹配仓库的绝对本地路径。

#### Scenario: base 为相对路径时仍输出绝对路径
- **WHEN** `.grepom.yml` 的 `base` 为 `repos/my-org`（相对路径），用户在项目子目录执行 `grepom dir web-app`
- **THEN** 系统输出绝对路径（如 `/home/user/projects/repos/my-org/frontend/web-app`）

#### Scenario: base 为相对路径时从子目录调用
- **WHEN** `.grepom.yml` 的 `base` 为 `repos/my-org`，用户在 `repos/my-org/frontend/web-app` 子目录中执行 `grepom dir web-api`
- **THEN** 系统输出绝对路径（如 `/home/user/projects/repos/my-org/backend/web-api`），而非相对路径
