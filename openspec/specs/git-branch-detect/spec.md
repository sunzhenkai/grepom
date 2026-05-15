## ADDED Requirements

### Requirement: 获取当前分支名
系统 SHALL 提供 `GetCurrentBranch(path string) (string, error)` 函数，通过 `git rev-parse --abbrev-ref HEAD` 获取指定目录的当前分支名。

#### Scenario: 正常获取分支名
- **WHEN** 在 `/path/to/repo` 目录中调用 `GetCurrentBranch`，当前处于 `feature-x` 分支
- **THEN** SHALL 返回 `"feature-x"` 和 nil error

#### Scenario: 分离 HEAD 状态
- **WHEN** 仓库处于 detached HEAD 状态
- **THEN** SHALL 返回 commit hash 和 nil error

#### Scenario: 非 git 仓库
- **WHEN** 目录不是 git 仓库
- **THEN** SHALL 返回空字符串和描述性 error

### Requirement: 获取 remote URL
系统 SHALL 提供 `GetRemoteURL(path string, remote string) (string, error)` 函数，通过 `git remote get-url` 获取指定 remote 的 URL。

#### Scenario: 获取 origin URL（HTTPS）
- **WHEN** 调用 `GetRemoteURL("/path/to/repo", "origin")`，origin 为 `https://github.com/myorg/repo.git`
- **THEN** SHALL 返回 `"https://github.com/myorg/repo.git"`

#### Scenario: 获取 origin URL（SSH）
- **WHEN** origin 为 `git@github.com:myorg/repo.git`
- **THEN** SHALL 返回 `"git@github.com:myorg/repo.git"`

#### Scenario: remote 不存在
- **WHEN** 调用 `GetRemoteURL` 但指定的 remote 不存在
- **THEN** SHALL 返回空字符串和描述性 error

### Requirement: 检测未推送 commit
系统 SHALL 提供 `HasUnpushedCommits(path string, branch string) (bool, int, error)` 函数，检查指定分支是否有未推送到远端的 commit，并返回未推送 commit 数量。

#### Scenario: 有未推送 commit
- **WHEN** `feature-x` 分支有 3 个 commit 未推送到 `origin/feature-x`
- **THEN** SHALL 返回 `true, 3, nil`

#### Scenario: 无未推送 commit
- **WHEN** `feature-x` 分支所有 commit 都已推送
- **THEN** SHALL 返回 `false, 0, nil`

#### Scenario: 远端分支不存在（新分支从未 push）
- **WHEN** `feature-x` 分支是本地新建的，远端尚不存在
- **THEN** SHALL 返回 `true, N, nil`（N 为本地所有 commit 数量或使用 fallback 方式计数）

### Requirement: 获取 HEAD commit message
系统 SHALL 提供 `GetHeadCommitMessage(path string) (string, error)` 函数，通过 `git log -1 --format=%B` 获取最新 commit 的完整 message。

#### Scenario: 正常获取 commit message
- **WHEN** HEAD commit message 为 "feat: add dark mode\n\nImplement dark mode toggle"
- **THEN** SHALL 返回完整的多行 message 字符串

#### Scenario: 单行 commit message
- **WHEN** HEAD commit message 为 "fix: typo"
- **THEN** SHALL 返回 `"fix: typo"`

#### Scenario: 空仓库
- **WHEN** 仓库没有任何 commit
- **THEN** SHALL 返回空字符串和描述性 error

### Requirement: 从 remote URL 解析 owner/repo 路径
系统 SHALL 提供 `ParseRemotePath(remoteURL string) string` 函数，从 git remote URL 中提取 owner/repo 格式的路径。

#### Scenario: HTTPS URL
- **WHEN** remoteURL 为 `https://github.com/myorg/myrepo.git`
- **THEN** SHALL 返回 `"myorg/myrepo"`

#### Scenario: SSH URL
- **WHEN** remoteURL 为 `git@github.com:myorg/myrepo.git`
- **THEN** SHALL 返回 `"myorg/myrepo"`

#### Scenario: 多级路径（GitLab 风格）
- **WHEN** remoteURL 为 `https://gitlab.com/myorg/team/myrepo.git`
- **THEN** SHALL 返回 `"myorg/team/myrepo"`

#### Scenario: 已有工具函数兼容
- **WHEN** 需要从 remote URL 提取路径
- **THEN** SHALL 优先复用 `repo.ExtractRemotePath` 函数（已存在于 `repo/resolver.go`）
