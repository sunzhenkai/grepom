## MODIFIED Requirements

### Requirement: 用户可以执行 scan 命令扫描仓库敏感信息
系统 SHALL 提供 `grepom scan` 命令，对配置中已克隆的仓库执行敏感信息扫描。命令 SHALL 支持 `--group` 和 `--resource` 标志过滤扫描范围，支持通过位置参数指定单个 repo 名称。未克隆的仓库 SHALL 被跳过并提示。当配置文件不存在时，系统 SHALL 自动回退到扫描当前工作目录模式，使用相同的 gitleaks 引擎和输出格式，并在 stderr 输出提示信息。

#### Scenario: 扫描所有仓库
- **WHEN** 用户执行 `grepom scan`
- **THEN** 系统扫描配置中所有已克隆的仓库的工作区文件，输出发现结果

#### Scenario: 按组扫描
- **WHEN** 用户执行 `grepom scan --group frontend`
- **THEN** 系统仅扫描属于 "frontend" 组的已克隆仓库

#### Scenario: 按资源扫描
- **WHEN** 用户执行 `grepom scan --resource work-gl`
- **THEN** 系统仅扫描使用 "work-gl" 资源的已克隆仓库

#### Scenario: 扫描单个仓库
- **WHEN** 用户执行 `grepom scan web-app`
- **THEN** 系统仅扫描名称为 "web-app" 的仓库

#### Scenario: 仓库未克隆时跳过
- **WHEN** 用户扫描的某个仓库尚未克隆到本地
- **THEN** 系统跳过该仓库并在输出中标注 "not cloned"

#### Scenario: 无配置文件时扫描当前目录
- **WHEN** 用户执行 `grepom scan` 且当前目录及默认路径下没有配置文件
- **THEN** 系统扫描当前工作目录下的所有文件，在 stderr 输出 "Scanning current directory (no config file found)..." 提示信息，并输出扫描结果
