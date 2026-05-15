## ADDED Requirements

### Requirement: mr 和 pr 命令别名
系统 SHALL 注册 `mr` 和 `pr` 两个 cobra 命令，二者 SHALL 指向同一个命令处理函数，行为完全一致。

#### Scenario: mr 命令执行
- **WHEN** 用户执行 `grepom mr`
- **THEN** 系统 SHALL 创建 Merge Request / Pull Request

#### Scenario: pr 命令执行
- **WHEN** 用户执行 `grepom pr`
- **THEN** 系统 SHALL 执行与 `grepom mr` 完全相同的逻辑

### Requirement: 无参数时智能检测分支
当用户不提供 `--from` 和 `--to` 参数时，系统 SHALL 自动检测：`from` 取当前分支名，`to` 取默认分支名（通过 `refs/remotes/origin/HEAD` 解析）。

#### Scenario: 自动检测成功
- **WHEN** 用户在 `feature-x` 分支上执行 `grepom mr`，且默认分支为 `main`
- **THEN** 系统 SHALL 使用 `from=feature-x`、`to=main` 创建 MR

#### Scenario: 当前分支等于默认分支
- **WHEN** 用户在 `main` 分支（即默认分支）上执行 `grepom mr`，未指定 `--from`
- **THEN** 系统 SHALL 报错并提示用户通过 `--from` 指定源分支

#### Scenario: 默认分支检测失败
- **WHEN** `origin/HEAD` 未设置，系统无法自动检测默认分支
- **THEN** 系统 SHALL 报错并提示用户通过 `--to` 指定目标分支

### Requirement: 手动指定分支
系统 SHALL 支持 `--from` 和 `--to` 参数手动指定源分支和目标分支。

#### Scenario: 手动指定两个分支
- **WHEN** 用户执行 `grepom mr --from feature-x --to develop`
- **THEN** 系统 SHALL 使用 `from=feature-x`、`to=develop` 创建 MR

#### Scenario: 仅指定 from
- **WHEN** 用户执行 `grepom mr --from feature-x`
- **THEN** 系统 SHALL 使用 `from=feature-x`，`to` 取默认分支

#### Scenario: 仅指定 to
- **WHEN** 用户执行 `grepom mr --to develop`
- **THEN** 系统 SHALL 使用 `from` 取当前分支，`to=develop`

### Requirement: MR/PR 标题和正文
系统 SHALL 自动从 HEAD commit message 中提取 title（第一行）和 body（后续行）。用户可通过 `--title` 和 `--body` 覆盖。系统 SHALL 支持 `--body-file` 从文件读取正文。

#### Scenario: 自动提取 commit message
- **WHEN** 用户执行 `grepom mr`，HEAD commit message 为 "feat: add dark mode\n\nImplement dark mode toggle"
- **THEN** 系统 SHALL 使用 title="feat: add dark mode"，body="Implement dark mode toggle"

#### Scenario: 手动指定标题
- **WHEN** 用户执行 `grepom mr --title "Custom PR title"`
- **THEN** 系统 SHALL 使用 title="Custom PR title"

#### Scenario: 从文件读取正文
- **WHEN** 用户执行 `grepom mr --body-file pr-description.md`
- **THEN** 系统 SHALL 读取 `pr-description.md` 文件内容作为 MR 正文

### Requirement: 草稿 MR/PR
系统 SHALL 支持 `--draft` 标志，创建的 MR/PR 标记为草稿。

#### Scenario: 创建草稿 MR
- **WHEN** 用户执行 `grepom mr --draft`
- **THEN** 系统 SHALL 创建一个标记为 draft 的 MR/PR

### Requirement: 浏览器打开创建页面
系统 SHALL 支持 `--web` 标志，不调用 API 而是在浏览器中打开对应平台的 MR/PR 创建页面。URL 中 SHALL 包含 from/to 分支信息和 draft 标志。

#### Scenario: --web 模式打开浏览器
- **WHEN** 用户执行 `grepom mr --web`，当前仓库为 GitHub 上的 `myorg/myrepo`，from=`feature-x`，to=`main`
- **THEN** 系统 SHALL 在浏览器中打开 `https://github.com/myorg/myrepo/compare/main...feature-x?expand=1`

#### Scenario: --web --draft 模式
- **WHEN** 用户执行 `grepom mr --web --draft`
- **THEN** 系统 SHALL 在浏览器中打开包含 draft 参数的 URL

### Requirement: 未推送 commit 交互提示
当检测到 `from` 分支有未推送到远端的 commit 时，系统 SHALL 提示用户是否先 push。在非 TTY 环境下 SHALL 直接报错。

#### Scenario: 有未推送 commit，用户确认 push
- **WHEN** `feature-x` 分支有 3 个未推送 commit，用户在 TTY 环境下执行 `grepom mr`
- **THEN** 系统 SHALL 提示 "分支 feature-x 有 3 个未推送的 commit，是否先 push?"
- **WHEN** 用户选择 Yes
- **THEN** 系统 SHALL 先执行 `git push origin feature-x`，然后创建 MR

#### Scenario: 有未推送 commit，用户拒绝 push
- **WHEN** 用户选择 No
- **THEN** 系统 SHALL 报错退出，提示用户先手动 push

#### Scenario: 无 TTY 环境下有未推送 commit
- **WHEN** stdin 不是 TTY，且分支有未推送 commit
- **THEN** 系统 SHALL 报错 "分支有未推送 commit，请先 push 或在 TTY 环境下运行"

#### Scenario: 无未推送 commit
- **WHEN** `feature-x` 分支所有 commit 都已推送到远端
- **THEN** 系统 SHALL 直接创建 MR，不提示 push

### Requirement: Provider 识别
系统 SHALL 从 `git remote get-url origin` 的 URL 中解析 host，按优先级识别 provider：
1. 与 config 中 resource URL 匹配 → 使用 config 中的 provider
2. 知名域名匹配（github.com → github，gitlab.com → gitlab，codeup.aliyun.com → codeup）
3. 无法识别 → 报错

#### Scenario: 通过 config 匹配 provider
- **WHEN** remote URL 为 `https://gitlab.mycompany.com/team/app.git`，config 中有 resource URL 为 `gitlab.mycompany.com`
- **THEN** 系统 SHALL 识别为 gitlab provider，使用该 resource 的 token

#### Scenario: 通过知名域名匹配
- **WHEN** remote URL 为 `https://github.com/myorg/myrepo.git`，config 中无匹配 resource
- **THEN** 系统 SHALL 识别为 github provider

#### Scenario: 无法识别 provider
- **WHEN** remote URL 为 `https://git.example.com/team/app.git`，config 中无匹配 resource
- **THEN** 系统 SHALL 报错并提示用户在 config 中添加对应 resource

### Requirement: Token 获取策略
系统 SHALL 按以下优先级获取 API token：
1. config 中匹配的 resource token（支持 `${ENV_VAR}` 占位符）
2. 环境变量 `GREPOM_GITHUB_TOKEN`（GitHub）或 `GREPOM_GITLAB_TOKEN`（GitLab）
3. 无法获取 → 报错

#### Scenario: 从 config 获取 token
- **WHEN** remote URL 匹配 config 中的 resource，且该 resource 有 token 配置
- **THEN** 系统 SHALL 使用该 resource 的 token

#### Scenario: 从环境变量获取 token
- **WHEN** remote URL 未匹配 config 中的 resource，但设置了 `GREPOM_GITHUB_TOKEN` 环境变量
- **THEN** 系统 SHALL 使用环境变量中的 token

#### Scenario: 无法获取 token
- **WHEN** remote URL 未匹配 config，也没有对应的环境变量
- **THEN** 系统 SHALL 报错并提示设置环境变量或在 config 中添加 resource

### Requirement: Codeup 不支持提示
当检测到 provider 为 codeup 时，系统 SHALL 输出友好提示，告知用户暂不支持 API 创建 MR，并提供浏览器端创建页面的 URL。

#### Scenario: Codeup 仓库执行 mr
- **WHEN** 用户在 Codeup 仓库中执行 `grepom mr`
- **THEN** 系统 SHALL 输出 "Codeup 暂不支持通过 API 创建 Merge Request"
- **THEN** 系统 SHALL 输出浏览器端创建页面 URL 并退出

### Requirement: 成功创建后的输出
MR/PR 创建成功后，系统 SHALL 输出创建结果的编号和 Web URL。

#### Scenario: GitHub PR 创建成功
- **WHEN** PR 创建成功
- **THEN** 系统 SHALL 输出类似 "PR #42: <title>\nhttps://github.com/..." 的信息

#### Scenario: GitLab MR 创建成功
- **WHEN** MR 创建成功
- **THEN** 系统 SHALL 输出类似 "MR !123: <title>\nhttps://gitlab.com/..." 的信息
