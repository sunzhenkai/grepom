## ADDED Requirements

### Requirement: `dir` 命令输出 base 目录路径
`grepom dir` 无参数时 SHALL 输出配置文件中定义的 `base` 目录的绝对路径到 stdout。

#### Scenario: 无参数输出 base 目录
- **WHEN** 用户执行 `grepom dir`（无参数），配置中 `base` 为 `~/projects`（解析为 `/home/user/projects`）
- **THEN** 输出 `/home/user/projects` 到 stdout

#### Scenario: 配置文件通过向上探测找到
- **WHEN** 当前目录为 `/home/user/projects/my-org/web-app`，配置文件在 `/home/user/projects/.grepom.yml`
- **THEN** `grepom dir` 成功输出 base 目录路径（不报错"配置未找到"）

### Requirement: `dir` 命令按名称模糊搜索仓库路径
`grepom dir <name>` SHALL 使用大小写不敏感子串匹配在所有已启用的仓库中搜索，输出匹配仓库的本地绝对路径。

#### Scenario: 精确匹配单个仓库
- **WHEN** 用户执行 `grepom dir web-app`，且仅有一个仓库名称包含 `web-app`
- **THEN** 输出该仓库的本地绝对路径到 stdout

#### Scenario: 模糊匹配单个结果
- **WHEN** 用户执行 `grepom dir web`，仅有一个仓库名称（大小写不敏感）包含子串 `web`
- **THEN** 输出该仓库的本地绝对路径到 stdout

#### Scenario: 模糊匹配多个结果
- **WHEN** 用户执行 `grepom dir web`，存在 `web-app`、`web-api`、`frontend-web` 三个匹配
- **THEN** 以表格形式列出所有匹配仓库（NAME、PATH、GROUP、RESOURCE），并返回非零退出码

#### Scenario: 无匹配结果
- **WHEN** 用户执行 `grepom dir nonexistent`
- **THEN** 输出错误信息到 stderr 并返回非零退出码

#### Scenario: 搜索范围包含 group 内仓库和独立仓库
- **WHEN** 存在 group `frontend` 下的 `web-app` 和独立仓库 `web-tool`
- **THEN** `grepom dir web` 同时搜索两者，返回所有匹配结果

### Requirement: 路径输出到 stdout、错误输出到 stderr
`dir` 命令 SHALL 将目标路径输出到 stdout，所有错误和诊断信息输出到 stderr，确保 stdout 内容可被 `cd "$(grepom dir ...)"` 安全使用。

#### Scenario: 路径输出到 stdout
- **WHEN** `grepom dir web-app` 成功找到仓库
- **THEN** stdout 仅包含路径字符串，不包含任何额外文本或格式

#### Scenario: 错误信息输出到 stderr
- **WHEN** `grepom dir nonexistent` 未找到匹配
- **THEN** 错误提示输出到 stderr，stdout 为空
