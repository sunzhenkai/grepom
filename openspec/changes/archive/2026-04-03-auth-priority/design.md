## Context

grepom 的 `git clone` 使用 6 级认证优先级链依次尝试克隆。当前优先级为：

```
group/repo SSH → group/repo token → resource SSH → resource token → default SSH → bare HTTP
```

用户实际使用中，SSH 认证（尤其是系统默认 SSH agent）通常比 token 认证更稳定可靠。当前 resource token 排在 resource SSH 之后但排在 default SSH 之前，导致 token 失败后才尝试 default SSH，而 default SSH 往往本就应该更早尝试。

当前实现中，`buildAuthStrategies()` 使用 `HasGroupSSHKey`/`HasGroupToken` 两个布尔标志做互斥：如果 group 级别有某种凭据，就完全跳过 resource 级别的同类型凭据。这种互斥机制在新优先级下仍然需要，但需要更精细的控制——具体来说，group 有 token 时不应跳过 resource SSH。

### 涉及的核心代码

| 文件 | 职责 |
|------|------|
| `git/git.go` | `buildAuthStrategies()` 构建认证策略列表，`Clone()` 依次尝试 |
| `repo/resolver.go` | 合并 group/repo/resource 的认证信息，传递 `HasGroupToken`/`HasGroupSSHKey` |
| `provider/provider.go` | `Repo` 结构体承载合并后的认证信息和来源标志 |

## Goals / Non-Goals

**Goals:**
- 将 clone 认证优先级调整为：group/repo 认证 → resource SSH → default SSH → resource token → bare HTTP
- 保持 group/repo 级别认证（SSH 和 token）作为最高优先级不变
- resource SSH 仅在 group/repo 未配置 SSH key 时尝试
- default SSH 总是作为 resource token 之前的回退
- 保持现有日志输出格式不变（`[N/M] 尝试 xxx...`）

**Non-Goals:**
- 不改变 resolver 的合并逻辑（token 和 SSH key 的来源追踪保持不变）
- 不改变 `CloneOptions` 结构体字段
- 不改变 token URL 构建规则（GitHub/GitLab 格式）
- 不新增配置字段

## Decisions

### Decision 1: 保持 resolver 不变，仅修改 `buildAuthStrategies()`

**选择**：不修改 `repo/resolver.go` 和 `provider.Repo` 结构体，仅调整 `git/git.go` 中 `buildAuthStrategies()` 的策略构建顺序和条件。

**理由**：
- resolver 的合并逻辑（group/repo override → resource fallback）是正确的，不需要改变
- `HasGroupToken`/`HasGroupSSHKey` 标志已经提供了足够的信息来决定策略构建
- 改动范围最小化，降低引入 bug 的风险

**备选方案**：引入更细粒度的来源标志（如分别跟踪 resource 的 SSH key 和 token 是否来自 fallback），但这会增加复杂度且没有必要。

### Decision 2: 新的策略构建逻辑

新的 `buildAuthStrategies()` 策略构建顺序和条件：

```
1. group/repo SSH key:   HasGroupSSHKey && SSHKey != "" && sshURL != ""
2. group/repo token:     HasGroupToken && Token != "" && httpURL != ""
3. resource SSH key:     !HasGroupSSHKey && SSHKey != "" && sshURL != ""
4. default SSH:          sshURL != ""（且步骤 3 未尝试过相同 SSH key）
5. resource token:       !HasGroupToken && Token != "" && httpURL != ""
6. bare HTTP:            httpURL != ""
```

**关于步骤 4 (default SSH) 的去重**：
- 如果步骤 3 已尝试 resource SSH key（使用指定 key），步骤 4 的 default SSH（使用系统默认 SSH agent）仍然应该尝试，因为系统 SSH agent 可能包含不同的 key
- 如果步骤 1 已尝试 group/repo SSH key，步骤 3 被跳过，步骤 4 的 default SSH 仍应尝试（同理）
- 因此 default SSH 只在 `sshURL != ""` 时就加入，不需要额外去重

**备选方案**：让 default SSH 与 resource SSH 互斥（如果已尝试指定 key 就不尝试默认 key），但这会减少回退机会，不符合"尽量克隆成功"的目标。

### Decision 3: 日志标签保持现有风格

策略的 `label` 字段使用与现有一致的中文格式：
- `"SSH key 认证 (group/repo)"`
- `"token 认证 (group/repo)"`
- `"SSH key 认证 (resource)"`
- `"SSH 认证 (默认)"`
- `"token 认证 (resource)"`
- `"HTTP 克隆"`

## Risks / Trade-offs

- **[行为变更]** 已有用户可能依赖当前优先级（resource token 优先于 default SSH）→ 这是期望的变更，在 proposal 中已明确
- **[重复尝试]** resource SSH key 和 default SSH 都尝试 SSH 方式，可能在 SSH 配置有问题的环境增加一次额外失败 → 可接受，因为两次失败通常很快
- **[性能]** 如果 resource token 本来就能成功，新顺序会导致多尝试 1-2 次 SSH（可能失败）→ 这是用户明确要求的优先级，SSH 失败通常很快（几秒）
