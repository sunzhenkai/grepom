## ADDED Requirements

### Requirement: scan 命令在无配置文件时扫描当前目录
系统 SHALL 在执行 `grepom scan` 时，如果配置文件不存在，自动回退到扫描当前工作目录（`.`）模式，使用与有配置时相同的 gitleaks 引擎和输出格式。系统 SHALL 在此模式下提示用户正在扫描当前目录。

#### Scenario: 无配置文件时扫描当前目录
- **WHEN** 用户在某个项目目录中执行 `grepom scan`，且当前目录及默认路径下没有 `.grepom.yml` 配置文件
- **THEN** 系统扫描当前工作目录下的所有文件（遵循 .gitignore 排除规则），输出扫描结果

#### Scenario: 无配置文件时显示提示信息
- **WHEN** 系统回退到当前目录扫描模式
- **THEN** 系统在 stderr 输出提示信息（如 "Scanning current directory (no config file found)..."）

#### Scenario: 有配置文件时行为不变
- **WHEN** 用户执行 `grepom scan`，且配置文件存在
- **THEN** 系统按现有逻辑解析配置中的仓库列表并扫描，行为与之前完全一致

#### Scenario: 当前目录不是 git 仓库时仍可扫描
- **WHEN** 用户在非 git 仓库的目录中执行 `grepom scan`（无配置文件）
- **THEN** 系统仍扫描该目录下的文件内容，返回发现的敏感信息

### Requirement: scan 命令无配置时支持所有输出选项
系统 SHALL 在无配置文件模式下支持与有配置时相同的输出选项，包括 `--format`（table/json）、`--output`（输出到文件）和 `--gitleaks-config`（自定义规则）。

#### Scenario: 无配置时使用 JSON 格式输出
- **WHEN** 用户执行 `grepom scan --format json`（无配置文件）
- **THEN** 系统以 JSON 数组格式输出当前目录的扫描结果

#### Scenario: 无配置时使用自定义 gitleaks 规则
- **WHEN** 用户执行 `grepom scan --gitleaks-config ./rules.toml`（无配置文件）
- **THEN** 系统使用指定的自定义规则扫描当前目录
