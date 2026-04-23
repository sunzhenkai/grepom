## MODIFIED Requirements

### Requirement: scan 命令在无配置文件时扫描当前目录
系统 SHALL 在执行 `grepom scan` 时，仅在当前目录查找 `.grepom.yml` 配置文件（不沿父目录链向上遍历）。如果当前目录没有 `.grepom.yml`，系统 SHALL 自动回退到扫描当前工作目录（`.`）模式，使用与有配置时相同的 gitleaks 引擎和输出格式。系统 SHALL 在此模式下提示用户正在扫描当前目录。

#### Scenario: 当前目录无配置文件时扫描当前目录
- **WHEN** 用户在某个项目目录中执行 `grepom scan`，且当前目录没有 `.grepom.yml` 配置文件
- **THEN** 系统扫描当前工作目录下的所有文件（遵循 .gitignore 排除规则），输出扫描结果

#### Scenario: 当前目录无配置文件时显示提示信息
- **WHEN** 系统回退到当前目录扫描模式
- **THEN** 系统在 stderr 输出提示信息（如 "Scanning current directory (no config file found)..."）

#### Scenario: 当前目录有配置文件时正常扫描
- **WHEN** 用户执行 `grepom scan`，且当前目录存在 `.grepom.yml`
- **THEN** 系统按现有逻辑解析配置中的仓库列表并扫描，行为与之前完全一致

#### Scenario: 祖先目录有配置文件但当前目录没有
- **WHEN** 用户在子目录中执行 `grepom scan`，当前目录没有 `.grepom.yml`，但某个父目录有
- **THEN** 系统不使用父目录的配置文件，而是回退到扫描当前工作目录

#### Scenario: 当前目录不是 git 仓库时仍可扫描
- **WHEN** 用户在非 git 仓库的目录中执行 `grepom scan`（无配置文件）
- **THEN** 系统仍扫描该目录下的文件内容，返回发现的敏感信息
