## ADDED Requirements

### Requirement: watch 顶级快捷命令

系统 SHALL 提供顶级命令 `grepom watch [repo-name]`，作为 `grepom pipeline watch <repo-name>` 的快捷方式。repo-name 为可选参数，`--id` 标志可选。

#### Scenario: 显式指定 repo-name
- **WHEN** 用户运行 `grepom watch rt-merger-service`
- **THEN** 系统 SHALL 执行与 `grepom pipeline watch rt-merger-service` 完全相同的逻辑

#### Scenario: 显式指定 repo-name 和 pipeline ID
- **WHEN** 用户运行 `grepom watch rt-merger-service --id 1234`
- **THEN** 系统 SHALL 监控 ID 为 1234 的 pipeline

#### Scenario: 仅指定 pipeline ID（自动推断 repo）
- **WHEN** 用户运行 `grepom watch --id 1234`
- **THEN** 系统 SHALL 自动推断当前仓库的 repo 信息，并监控 ID 为 1234 的 pipeline

#### Scenario: 无参数自动推断
- **WHEN** 用户在 git 仓库目录中运行 `grepom watch`
- **THEN** 系统 SHALL 自动从当前 git 仓库推断 repo 信息，监控最新的 pipeline

#### Scenario: 参数限制
- **WHEN** 用户运行 `grepom watch a b`（超过 1 个位置参数）
- **THEN** 系统 SHALL 报错并退出

### Requirement: watch 自动推断三级 fallback

当用户省略 repo-name 时，系统 SHALL 通过三级 fallback 策略自动推断当前仓库的 pipeline 查询信息。

#### Scenario: Level 1 - 配置精确匹配成功
- **WHEN** 当前目录是 git 仓库，remote URL 为 `git@gitlab.mycompany.com:msa/rt-merger-service.git`，.grepom.yml 中存在某个 repo 条目的 CloneURL 或 SSHURL 提取出 `msa/rt-merger-service`
- **THEN** 系统 SHALL 使用该 repo 条目的 resource 信息获取 provider、serverURL、token，并通过 remotePath 构建 pipeline 请求

#### Scenario: Level 2 - Host 匹配成功
- **WHEN** 当前目录的 remote URL 的 host 为 `gitlab.mycompany.com`，.grepom.yml 中没有匹配的 repo 条目，但存在 resource 的 URL 为 `gitlab.mycompany.com`
- **THEN** 系统 SHALL 使用该 resource 的 provider 和 token，从 remote URL 提取 remotePath，构建 pipeline 请求

#### Scenario: Level 2 - Host 匹配但无 token
- **WHEN** 当前目录的 remote URL 的 host 匹配到配置中的 resource，但该 resource 的 token 解析失败
- **THEN** 系统 SHALL 输出包含 resource 名称和错误原因的详细错误信息

#### Scenario: Level 3 - 公共域名环境变量成功
- **WHEN** 当前目录的 remote URL 的 host 为 `github.com`，配置中无匹配，但环境变量 `GREPOM_GITHUB_TOKEN` 已设置
- **THEN** 系统 SHALL 使用 GitHub provider + 环境变量 token + remote URL 的 remotePath 构建 pipeline 请求

#### Scenario: Level 3 - 公共域名但无环境变量
- **WHEN** 当前目录的 remote URL 的 host 为 `gitlab.com`，配置中无匹配，且环境变量 `GREPOM_GITLAB_TOKEN` 未设置
- **THEN** 系统 SHALL 输出包含 host 名称和 token 设置建议的详细错误信息

#### Scenario: 所有 fallback 均失败
- **WHEN** 当前目录的 remote URL 的 host 无法在任何级别匹配到 provider
- **THEN** 系统 SHALL 输出详细错误信息，包含当前仓库信息、远程地址、主机名，以及添加配置或设置环境变量的建议

### Requirement: watch 自动推断前置检查

自动推断 SHALL 在执行 fallback 前进行前置检查。

#### Scenario: 当前目录不是 git 仓库
- **WHEN** 用户在非 git 仓库目录运行 `grepom watch`
- **THEN** 系统 SHALL 输出错误信息说明当前目录不是 git 仓库，建议 cd 到项目目录或使用 `grepom pipeline watch <repo-name>`

#### Scenario: 当前仓库没有 remote origin
- **WHEN** 用户在没有配置 remote origin 的 git 仓库中运行 `grepom watch`
- **THEN** 系统 SHALL 输出错误信息说明没有 remote origin，建议先 `git remote add origin <url>` 或使用完整命令

### Requirement: watch 自动推断失败详细提示

当自动推断失败时，系统 SHALL 输出结构化的详细错误信息，包含诊断和建议。

#### Scenario: 配置中无匹配的详细提示
- **WHEN** Level 1 和 Level 2 均未匹配，且 host 不是已知公共域名
- **THEN** 系统 SHALL 输出包含以下信息的错误：
  - 当前仓库名称（从 remote URL 推导）
  - 远程地址
  - 远程路径
  - 主机名
  - 诊断结论（未在配置中找到匹配的 repo 或 resource）
  - 建议（在 .grepom.yml 中添加 resource 条目，或设置环境变量）

#### Scenario: 已知域名但无 token 的详细提示
- **WHEN** Level 3 匹匹配到已知域名但缺少 token
- **THEN** 系统 SHALL 输出包含以下信息的错误：
  - 当前仓库名称
  - 远程地址
  - Provider 名称
  - 建议设置的环境变量名称
  - 或在 .grepom.yml 中添加 resource 的示例

### Requirement: watch 命令复用 pipeline watch 的 watch 循环逻辑

`grepom watch`、`grepom pipeline watch` 和 `grepom tag -w` SHALL 共享同一 watch 轮询循环实现，不重复实现轮询、状态渲染、Ctrl+C 处理等逻辑。

#### Scenario: watch 循环行为一致
- **WHEN** 用户通过 `grepom watch web-app`、`grepom pipeline watch web-app` 或 `grepom tag -w` 监控同一个 pipeline
- **THEN** 三者的轮询间隔、状态行格式、终态退出行为 SHALL 完全一致

#### Scenario: Ctrl+C 行为一致
- **WHEN** 用户在 `grepom watch` 或 `grepom tag -w` 运行过程中按 Ctrl+C
- **THEN** 系统 SHALL 与 `pipeline watch` 相同的方式优雅退出
