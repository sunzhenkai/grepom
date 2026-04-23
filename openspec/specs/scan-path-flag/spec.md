## ADDED Requirements

### Requirement: scan 命令支持 -p/--path 标志指定扫描路径
系统 SHALL 提供 `-p/--path` 标志，允许用户指定任意目录路径进行扫描。当指定 `-p` 时，系统 SHALL 完全忽略配置文件的存在，直接扫描指定路径。`-p` 指定时 `[name]` 位置参数 SHALL 被忽略。

#### Scenario: 指定路径扫描
- **WHEN** 用户执行 `grepom scan -p /tmp/my-project`
- **THEN** 系统扫描 `/tmp/my-project` 目录下的文件，不查找或加载任何配置文件

#### Scenario: -p 指定为当前目录
- **WHEN** 用户执行 `grepom scan -p .`
- **THEN** 系统扫描当前工作目录，不查找或加载任何配置文件

#### Scenario: -p 与 name 位置参数同时使用
- **WHEN** 用户执行 `grepom scan -p /tmp myrepo`
- **THEN** 系统扫描 `/tmp` 目录，忽略 `myrepo` 位置参数

#### Scenario: -p 指定的路径不存在
- **WHEN** 用户执行 `grepom scan -p /nonexistent/path`
- **THEN** 系统报错提示路径不存在

#### Scenario: -p 指定时仍支持其他 flag
- **WHEN** 用户执行 `grepom scan -p /tmp --format json --gitleaks-config rules.toml`
- **THEN** 系统使用自定义规则以 JSON 格式扫描 `/tmp` 目录

### Requirement: scan 命令在扫描开始时打印目标摘要
系统 SHALL 在开始扫描前在 stderr 打印扫描目标摘要信息。路径模式（`-p` 指定或无配置回退）时打印路径；配置模式时打印仓库名称列表。仓库数量超过 5 个时，显示前 5 个名称后加"...及 N 个仓库"。

#### Scenario: -p 模式打印路径摘要
- **WHEN** 用户执行 `grepom scan -p /tmp/my-project`
- **THEN** 系统在 stderr 输出类似 "Scanning /tmp/my-project..." 的信息

#### Scenario: 无配置模式打印当前目录摘要
- **WHEN** 用户执行 `grepom scan` 且当前目录无 `.grepom.yml`
- **THEN** 系统在 stderr 输出类似 "Scanning current directory..." 的信息

#### Scenario: 配置模式打印仓库列表（≤5 个）
- **WHEN** 配置中有 3 个已克隆的仓库
- **THEN** 系统在 stderr 输出类似 "Scanning repo1, repo2, repo3..." 的信息

#### Scenario: 配置模式打印仓库列表（>5 个）
- **WHEN** 配置中有 10 个已克隆的仓库
- **THEN** 系统在 stderr 输出类似 "Scanning repo1, repo2, repo3, repo4, repo5, ...and 5 more" 的信息
