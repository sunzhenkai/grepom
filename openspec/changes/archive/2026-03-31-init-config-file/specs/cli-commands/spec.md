## MODIFIED Requirements

### Requirement: init command
系统 SHALL 将原 `grepom init`（clone 仓库）功能迁移至 `grepom clone` 命令。`clone` 命令 SHALL 将仓库 clone 到本地文件系统 `<base>/<path>`，按需创建目录。

#### Scenario: Clone all repos
- **WHEN** 用户运行 `grepom clone`（无参数）
- **THEN** 系统从所有 sources clone 所有仓库到各自 base 下的路径

#### Scenario: Clone single repo
- **WHEN** 用户运行 `grepom clone web-app`
- **THEN** 系统仅 clone 名为 `web-app` 的仓库

#### Scenario: Clone by group
- **WHEN** 用户运行 `grepom clone --group my-org/frontend`
- **THEN** 系统仅 clone `my-org/frontend` 下的所有仓库

#### Scenario: Repo already exists
- **WHEN** 目标目录已包含 git 仓库
- **THEN** 系统跳过 clone 并打印提示（非错误）

## ADDED Requirements

### Requirement: clone 命令兼容提示
当用户运行 `grepom init [name]`（带位置参数）时，系统 SHALL 提示用户 "did you mean `grepom clone`?" 并退出，帮助已有用户迁移。

#### Scenario: 用户误用 init clone 语法
- **WHEN** 用户运行 `grepom init web-app`（带位置参数）
- **THEN** 系统提示 "did you mean `grepom clone`?" 并以非零状态码退出
