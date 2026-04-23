## MODIFIED Requirements

### Requirement: 用户可以执行 scan 命令扫描仓库敏感信息
系统 SHALL 提供 `grepom scan` 命令，对配置中已克隆的仓库执行敏感信息扫描。命令的 cobra 描述（Short、Long、Example）、flag 帮助文本、错误信息和用户可见输出 SHALL 全部使用英文。命令 SHALL 支持 `--group` 和 `--resource` 标志过滤扫描范围，支持通过位置参数指定单个 repo 名称。命令 SHALL 支持 `-p/--path` 标志直接指定扫描目录路径（此时忽略配置文件和位置参数）。系统 SHALL 仅在当前目录查找 `.grepom.yml` 配置文件（不沿父目录链向上遍历）。未克隆的仓库 SHALL 被跳过并提示。当配置文件不存在时，系统 SHALL 自动回退到扫描当前工作目录模式。扫描开始时 SHALL 打印扫描目标摘要信息。

#### Scenario: scan 命令帮助信息为英文
- **WHEN** 用户运行 `grepom scan --help`
- **THEN** 系统显示英文的命令描述、flag 说明（包括 `-p/--path`）和示例

#### Scenario: 扫描所有仓库
- **WHEN** 用户执行 `grepom scan`（当前目录有 `.grepom.yml`）
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
- **WHEN** 用户执行 `grepom scan` 且当前目录没有 `.grepom.yml`
- **THEN** 系统扫描当前工作目录下的所有文件，在 stderr 输出英文提示信息

#### Scenario: 使用 -p 指定路径扫描
- **WHEN** 用户执行 `grepom scan -p /some/path`
- **THEN** 系统忽略配置文件，直接扫描 `/some/path` 目录
