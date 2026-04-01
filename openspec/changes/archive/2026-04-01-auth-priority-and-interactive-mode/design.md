## Context

grepom 是一个 Go 编写的 Git 仓库管理 CLI 工具，基于 Cobra 框架。当前 clone 认证策略为简单的 SSH → HTTP 回退（`git/git.go`），token 仅用于 provider API 调用（sync），未用于 clone 操作。

项目结构：
- `git/git.go`：封装 git clone/pull/status 操作
- `repo/resolver.go`：从 config 构建 provider.Repo 列表，通过 `deriveSSHURL` 从 HTTP URL 推导 SSH URL
- `provider/provider.go`：`Repo` 结构包含 `CloneURL`（HTTP）和 `SSHURL`（SSH）
- `config/config.go`：Resource 定义包含 `Token` 字段（无 SSH key）；Group、Repo、GroupRepo 当前无独立认证字段
- `cmd/`：各 Cobra 子命令

当前 clone 流程：
1. `resolver.Resolve()` 将 group repos 和 standalone repos 转为 `provider.Repo`，其中 `SSHURL` 由 `deriveSSHURL` 从 `CloneURL` 推导
2. `clone.go` 调用 `gitpkg.Clone(fullPath, r.SSHURL, r.CloneURL)`
3. `git.Clone()` 先尝试 SSH URL，失败后回退到 HTTP URL

当前配置模型只有 resource 级别的 token，所有引用同一 resource 的 group/repo 共享相同的认证信息。Resource 不支持配置 SSH key。

## Goals / Non-Goals

**Goals:**
- 建立 6 级克隆认证优先级链：group/repo token → group/repo SSH key → resource token → resource SSH key → 推导 SSH → 裸 HTTP
- Resource 级别支持配置 `ssh_key`，作为 token 之后的二级认证
- 支持 Group、Repo 级别独立配置 `ssh_key` 和 `token`，覆盖 resource 默认认证
- clone 尝试每种认证方式时输出日志，让用户了解当前进度和认证回退过程
- 新增 `grepom interactive` 命令，为常见操作提供交互式引导
- 交互模式支持：init 配置、add resource/group/repo、sync、clone
- 保持现有命令行接口完全兼容（零 breaking change）

**Non-Goals:**
- 不做 GUI/TUI，仅基于终端文本交互（prompt/select）
- 不实现 token 的自动获取或 OAuth 流程
- 不修改 sync 命令的认证方式（已使用 resource token 调用 API）
- 不支持 SSH agent 转发或 SSH key 密码短语管理
- 不修改配置文件格式的核心结构（仅新增可选字段）

## Decisions

### D1: Resource 级别新增 SSH key 字段

**决定**：在 `Resource` 结构中新增可选的 `ssh_key` 字段，作为 token 之后的二级认证。

```yaml
resources:
  work-gl:
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_TOKEN}      # 一级认证：用于 API 调用 + clone token
    ssh_key: ~/.ssh/id_work     # 二级认证：clone 回退时使用的 SSH key
```

**理由**：Resource 是认证信息的聚合点，SSH key 自然属于认证配置的一部分。许多用户为不同 provider 配置不同的 SSH key（如工作密钥 vs 个人密钥），在 resource 级别配置比 group 级别更通用。

**替代方案**：
- 只在 group/repo 级别配置 SSH key：每个 group 都要重复配置，违背 resource 聚合认证的初衷

### D2: Group/Repo 级别认证字段设计

**决定**：在 `Group`、`Repo` 结构中新增两个可选 YAML 字段。GroupRepo 不单独配置，继承 Group。

```yaml
groups:
  - name: frontend
    resource: work-gl
    path: my-org/frontend
    ssh_key: ~/.ssh/id_ed25519   # 可选，覆盖 resource 的 SSH key
    token: ${FRONTEND_TOKEN}     # 可选，覆盖 resource 的 token

repos:
  - name: dotfiles
    resource: github
    url: https://github.com/me/dotfiles.git
    ssh_key: ~/.ssh/id_personal  # 可选
    token: ${DOTFILES_TOKEN}     # 可选
```

**理由**：
- Group 和独立 Repo 支持独立认证覆盖，GroupRepo 继承 Group 设置，避免重复配置
- 所有字段均为可选，不填则使用 resource 级别的认证

**替代方案**：
- 在 GroupRepo 上也加认证字段：过于细粒度，增加配置复杂度
- 使用独立的 `auth` 对象而非扁平字段：过度抽象

### D3: Token 认证 URL 构建策略

**决定**：根据 provider 类型构建不同的 token 认证 HTTPS URL。

- GitHub: `https://x-access-token:<token>@<host>/<path>.git`
- GitLab: `https://oauth2:<token>@<host>/<path>.git`

**理由**：这是 GitHub 和 GitLab 官方推荐的 token 认证克隆方式。

**替代方案**：
- Git credential helper：需要额外配置 git credential store，增加复杂度
- `.netrc` 文件：需要写入文件系统，有安全风险

### D4: SSH Key 指定方式

**决定**：通过 `GIT_SSH_COMMAND` 环境变量传递指定的 SSH key 给 git clone。

```go
cmd := exec.Command("git", "clone", sshURL, path)
cmd.Env = append(os.Environ(), "GIT_SSH_COMMAND=ssh -i "+sshKey+" -o IdentitiesOnly=yes")
```

**理由**：
- 不需要修改 `~/.ssh/config`，不影响全局 SSH 配置
- `IdentitiesOnly=yes` 确保只使用指定密钥，避免 agent 中其他密钥干扰
- Go 标准库直接支持，无需额外依赖

### D5: Clone 函数签名变更

**决定**：扩展 `git.Clone()` 函数，使用结构体选项传入认证信息。

```go
type CloneOptions struct {
    Token    string // 用于 HTTPS token 认证
    Provider string // "github" 或 "gitlab"，决定 token URL 格式
    SSHKey   string // SSH 密钥文件路径
}

func Clone(path, sshURL, httpURL string, opts CloneOptions) error
```

**理由**：使用结构体而非多个平铺参数，更清晰、可扩展。

### D6: Repo 结构扩展

**决定**：在 `provider.Repo` 中新增 `Token`、`SSHKey` 和 `Provider` 字段（Provider 已存在）。Resolver 按以下规则填充认证信息：

1. token：优先使用 group/repo 级别 → 否则使用 resource 级别
2. ssh_key：优先使用 group/repo 级别 → 否则使用 resource 级别
3. provider：始终从 resource 获取

**理由**：让每条 repo 记录携带最终确定的认证信息，clone 命令无需再做查找和合并。

### D7: 认证优先级链

**决定**：最终 clone 认证优先级（6 级）：

1. **group/repo 级别 token**（HTTPS + token URL）
2. **group/repo 级别 SSH key**（SSH + 指定 key）
3. **resource 级别 token**（HTTPS + token URL）
4. **resource 级别 SSH key**（SSH + 指定 key）
5. **推导的 SSH URL**（使用系统默认 SSH）
6. **裸 HTTP URL**

**理由**：最具体的配置优先，逐级回退到最通用的方式。Resource 的 token 用于 API 调用也是 clone 的首选认证，SSH key 作为 resource 的二级回退。

### D8: 认证尝试日志

**决定**：在 clone 过程中尝试每种认证方式时，输出一行日志到 stdout，格式如下：

```
cloning my-org/frontend/web-app...
  [1/6] 尝试 token 认证 (group)...
  [1/6] token 认证失败: <错误摘要>
  [2/6] 尝试 SSH key 认证 (group)...
  [2/6] SSH key 认证失败: <错误摘要>
  [3/6] 尝试 token 认证 (resource)...
  [3/6] 成功
```

日志策略：
- 跳过的级别（如未配置 token）不输出日志
- 失败时输出错误摘要（不含敏感信息）
- 成功时仅输出 "成功"
- verbose 模式下输出更详细的 git stderr

**理由**：用户需要看到 clone 过程的认证回退链，方便排查为什么用了某种认证方式、为什么失败。

### D9: 交互式命令实现方案

**决定**：使用 `github.com/AlecAivazis/survey/v2` 库实现交互提示。

**理由**：成熟稳定，支持 input/select/confirm 等类型，API 简洁。

### D10: 交互式命令结构

**决定**：新增 `grepom interactive` 命令作为入口，进入后显示操作菜单。

```
$ grepom interactive
? 请选择操作:
  ▸ 初始化配置 (init)
    添加资源 (add resource)
    添加组 (add group)
    添加仓库 (add repo)
    同步远程仓库 (sync)
    克隆仓库 (clone)
    查看状态 (status)
    退出
```

### D11: 交互模式与命令行模式共享逻辑

**决定**：交互模式最终调用与命令行相同的业务逻辑函数，仅输入方式不同。

## Risks / Trade-offs

- **[Token 在配置文件中可见]** → 支持 `${ENV_VAR}` 占位符语法（与 resource token 一致），避免明文存储
- **[SSH key 路径错误]** → clone 前不验证文件存在（保持简洁），失败时通过 git 错误信息自然报告
- **[Token 在进程参数中可见]** → 通过将 token 嵌入 HTTPS URL 而非命令行参数来缓解；clone 失败时日志中不打印完整 URL
- **[认证日志中泄露敏感信息]** → 日志仅输出认证方式名称和错误摘要，不输出 token 值或完整 URL
- **[survey 依赖引入]** → 该库维护活跃、API 稳定，风险可控
- **[Clone 函数签名变更影响所有调用方]** → 当前只有 `cmd/clone.go` 一个调用方，影响面小
- **[交互模式在非 TTY 环境不可用]** → 交互命令检测 tty，非交互环境给出友好提示
- **[GroupRepo 不支持独立认证]** → 继承 Group 认证是合理的默认行为
- **[认证优先级链过长导致 clone 慢]** → 跳过未配置的级别，实际尝试通常只有 1-2 种
