## ADDED Requirements

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

### Requirement: 扫描支持工作区和 git 历史两种模式
系统 SHALL 默认扫描工作区文件（当前 checkout 的代码）。当用户指定 `--history` 标志时，系统 SHALL 扫描 git 历史记录中所有提交的变更内容。

#### Scenario: 默认工作区扫描
- **WHEN** 用户执行 `grepom scan`（不带 `--history`）
- **THEN** 系统扫描每个仓库工作目录下的文件内容

#### Scenario: git 历史扫描
- **WHEN** 用户执行 `grepom scan --history`
- **THEN** 系统扫描每个仓库的 git 历史（所有分支的所有提交）

#### Scenario: 组合使用历史扫描和组过滤
- **WHEN** 用户执行 `grepom scan --group backend --history`
- **THEN** 系统仅扫描 backend 组仓库的 git 历史

### Requirement: 扫描结果以终端表格形式展示
系统 SHALL 将扫描结果按仓库名称分组展示。每个仓库作为分组标题，其下的发现项以缩进列表形式展示，每行包含文件路径（过长时自动截断）、行号、规则 ID 和严重程度。表格后 SHALL 输出汇总统计（按严重程度计数）。敏感信息 SHALL 脱敏显示（前 8 字符 + `...`）。

#### Scenario: 发现敏感信息时的输出
- **WHEN** 扫描发现敏感信息
- **THEN** 系统按仓库分组输出，每行一个发现项，包含文件路径（过长截断）、行号、规则 ID、严重程度和脱敏后的 secret 片段

#### Scenario: 无发现时的输出
- **WHEN** 扫描完成且未发现任何敏感信息
- **THEN** 系统输出 "No secrets found."

#### Scenario: 汇总统计
- **WHEN** 扫描完成后存在发现项
- **THEN** 系统在所有分组后输出总计行，格式为 "Found N findings in M repos." 并附按严重程度的统计

### Requirement: 支持 JSON 格式输出
系统 SHALL 支持 `--format json` 标志，将扫描结果以 JSON 数组形式输出。每个 JSON 对象 SHALL 包含 repo（仓库名）、file（文件路径）、line（行号）、rule_id（规则 ID）、description（规则描述）、secret（脱敏后的密钥片段）、severity（严重程度）字段。

#### Scenario: JSON 格式输出
- **WHEN** 用户执行 `grepom scan --format json`
- **THEN** 系统以 JSON 数组形式输出所有发现项

### Requirement: 扫描自动感知 .gitignore
系统 SHALL 在扫描工作区文件时自动读取每个仓库根目录下的 `.gitignore` 文件，跳过被 `.gitignore` 排除的文件和目录。

#### Scenario: 跳过 .gitignore 中的文件
- **WHEN** 仓库的 `.gitignore` 包含 `node_modules/` 和 `*.log`
- **THEN** 系统不扫描 `node_modules/` 目录下的文件和 `.log` 后缀的文件

#### Scenario: 无 .gitignore 时不影响扫描
- **WHEN** 仓库根目录没有 `.gitignore` 文件
- **THEN** 系统扫描所有工作区文件（仍跳过 `.git/` 目录）

### Requirement: 支持 .gitleaksignore 白名单
系统 SHALL 识别仓库根目录下的 `.gitleaksignore` 文件，跳过其中列出的发现项。文件格式与 gitleaks 原生格式一致（每行一个 fingerprint）。

#### Scenario: 使用 .gitleaksignore 排除已知发现
- **WHEN** 仓库根目录存在 `.gitleaksignore` 且包含某发现项的 fingerprint
- **THEN** 系统不在结果中报告该发现项

### Requirement: 并行扫描多个仓库
系统 SHALL 对多个仓库并行执行扫描，提高扫描效率。系统 SHALL 显示扫描进度（已扫描/总计仓库数）。

#### Scenario: 扫描多个仓库时显示进度
- **WHEN** 用户扫描包含 10 个仓库的 group
- **THEN** 系统显示进度信息，如 "Scanning... 3/10 repos"

### Requirement: 敏感信息在输出中脱敏
系统 SHALL 对输出中的 secret 字段进行部分脱敏，仅显示前几个字符，其余用 `...` 替代。完整 secret 不得出现在终端表格输出中。

#### Scenario: 表格输出中的密钥脱敏
- **WHEN** 发现项的 secret 为 `ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890`
- **THEN** 表格输出中该字段显示为 `ghp_ABCDE...`（前 8 个字符 + 省略号）

### Requirement: 支持自定义 gitleaks 配置
系统 SHALL 支持 `--gitleaks-config` 标志，允许用户指定自定义的 gitleaks.toml 配置文件路径。未指定时使用 gitleaks 默认规则集。

#### Scenario: 使用自定义配置
- **WHEN** 用户执行 `grepom scan --gitleaks-config ./my-rules.toml`
- **THEN** 系统使用指定的配置文件中的规则进行扫描

#### Scenario: 使用默认配置
- **WHEN** 用户执行 `grepom scan`（不带 `--gitleaks-config`）
- **THEN** 系统使用 gitleaks 内置的默认规则集（100+ 规则）
