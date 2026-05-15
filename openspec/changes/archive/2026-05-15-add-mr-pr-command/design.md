## Context

grepom 是一个 Go 编写的 Git 多仓库管理 CLI 工具，基于 Cobra 框架构建命令体系。当前已支持 clone、pull、push（含密钥扫描）、sync、status、pipeline 等命令。项目采用了清晰的分层架构：

- `cmd/` — Cobra 命令层
- `provider/` — 仓库发现接口（GitHub、GitLab、Codeup）
- `cicd/` — CI/CD Pipeline 查询接口（GitHub、GitLab），独立注册表模式
- `git/` — Git 操作封装
- `config/` — YAML 配置管理

`cicd/` 包建立了一个很好的先例：独立包、独立注册表、独立接口，Codeup 不实现自然就"不支持"。PR/MR 功能应沿用此模式。

## Goals / Non-Goals

**Goals:**

- 提供 `mr`（及别名 `pr`）命令，支持通过 CLI 直接创建 MR/PR
- 支持 GitHub（Pull Request）和 GitLab（Merge Request），Codeup 给出不支持提示
- 无参数时智能检测当前分支、默认分支、remote URL、provider 类型
- 支持 `--from`/`--to` 手动指定源分支和目标分支
- 支持 `--draft` 创建草稿
- 支持 `--web` 打开浏览器创建页面
- 分支有未推送 commit 时交互式提示是否先 push
- 自动从 HEAD commit 提取 title/body，支持 `--title`/`--body` 覆盖
- 兼容「有 config」和「无 config」两种使用场景

**Non-Goals:**

- 不实现 MR/PR 的列表、查看、评论、合并等后续操作（后续迭代）
- 不支持 Codeup 的 MR/PR API（当前仅提示不支持）
- `mr` 命令的自动 push 不做密钥扫描（用户应已通过 `grepom push` 操作）
- 不支持批量多仓库 MR/PR（本次仅在当前仓库或指定仓库内操作）
- 不实现 MR/PR 模板系统（可用 `--body-file` 传入）

## Decisions

### D1: 独立 `mergerequest/` 包，沿用 `cicd/` 注册表模式

**选择**: 新建 `mergerequest/` 包，内含 `MergeRequestProvider` 接口和注册表。

**理由**: 与 `cicd/` 包的模式完全一致——独立包、独立接口、独立注册表。Codeup 不注册实现即"不支持"。这比在 `provider.Provider` 接口上扩展更干净（不会让所有 provider 都被迫实现 MR 能力）。

**替代方案**:
- 在 `provider.Provider` 接口上加方法 → Codeup 需返回 error，不优雅
- 不用接口，直接在 cmd 里硬编码 → 无法测试，无法扩展

### D2: 命令模式——当前目录优先，兼容 config 指定

**选择**: 默认在当前 git 仓库目录操作（类似 `push` 命令）。如果当前目录不在 grepom 管理范围内，仍可工作（通过环境变量获取 token）。

**理由**: 开发者创建 MR 的最常见场景是在项目目录里直接操作。不需要先切到 grepom 管理的目录结构。

**Token 获取优先级**:
1. 从 config 匹配（remote URL host == config resource URL）→ 使用 config 中的 token
2. 环境变量 `GREPOM_GITHUB_TOKEN` / `GREPOM_GITLAB_TOKEN` → 兜底方案
3. 都没有 → 报错提示

### D3: `mr` 和 `pr` 双命令别名

**选择**: 注册两个 cobra Command，底层共享同一个 `runMR` 函数。

**理由**: GitHub 社区习惯 `pr`，GitLab 社区习惯 `mr`。两个名字指向同一功能避免认知负担。

### D4: 分支检测策略

**选择**:
- `from` 默认 = 当前分支（`git rev-parse --abbrev-ref HEAD`）
- `to` 默认 = 默认分支（`git symbolic-ref refs/remotes/origin/HEAD` → 取 `refs/remotes/origin/` 后的部分）
- 如果 from == to → 报错提示

**理由**: 覆盖最常见的开发流程——在 feature 分支上开发完毕，创建 MR 合入主分支。

### D5: `--web` 模式仅构建 URL 并打开浏览器

**选择**: `--web` 不调用 API，仅拼接 URL 后用 `open`/`xdg-open` 打开浏览器。

**理由**: 与 `gh pr create --web` 行为一致。浏览器端可以提供更丰富的编辑体验（选择 reviewer、label 等）。URL 格式：
- GitHub: `https://github.com/{owner}/{repo}/compare/{to}...{from}?expand=1`（draft 加 `&draft=1`）
- GitLab: `{server}/{path}/-/merge_requests/new?merge_request[source_branch]={from}&merge_request[target_branch]={to}`

### D6: Provider 识别策略

**选择**: 从 `git remote get-url origin` 解析 host，按以下优先级判断 provider：

1. 与 config 中 resource URL 匹配 → 使用 config 中的 provider 类型
2. 知名域名匹配（`github.com` → GitHub，`gitlab.com` → GitLab，`codeup.aliyun.com` → Codeup）
3. 都不匹配 → 报错

**理由**: 大多数用户使用知名平台。自建 GitLab 实例通过 config 匹配解决。

### D7: `mergerequest/` 接口设计

**选择**: 接口包含两个方法：
- `CreateMergeRequest(ctx, params) (*MergeRequest, error)` — API 创建
- `BuildWebURL(params) string` — 浏览器 URL 构建（纯计算，无 IO）

**理由**: `BuildWebURL` 虽然不需要 API 调用，但每个平台的 URL 格式不同，放在 provider 接口里最自然。Codeup 的 `BuildWebURL` 可以返回手动创建 MR 的页面 URL。

## Risks / Trade-offs

**[自建 GitLab 识别失败]** → 如果用户没有在 config 中配置自建 GitLab 的 resource，且域名不在知名列表中，将无法识别 provider。缓解：报错时明确提示用户在 config 中添加对应的 resource。

**[GitHub API 路径格式与 GHE]** → GitHub Enterprise Server 的 API 路径是 `https://ghe.example.com/api/v3/repos/...`，与 github.com 的 `https://api.github.com/repos/...` 不同。沿用 `cicd/github.go` 中的 `githubAPIURL` 函数处理此差异。

**[GitLab URL-encoded path]** → GitLab API 中项目路径需要 URL 编码（`my%2Forg%2Fmyrepo`）。已在 `cicd/gitlab.go` 中有先例，沿用 `url.PathEscape` 处理。

**[无 TTY 时 push 提示]** → 如果没有 TTY（管道/CI 环境），交互式提示无法工作。此时如果有未推送 commit，直接报错并提示用户先手动 push。

**[分支名含特殊字符]** → URL 构建（`--web`）中分支名需要 URL 编码。使用 `url.QueryEscape` 处理。

**[默认分支检测失败]** → `origin/HEAD` 在某些 clone 方式下未设置。fallback 策略：依次尝试 `main`、`master`，如果都没有则要求用户通过 `--to` 指定。
