## Context

当前 `config.Load()` 延迟解析 token 环境变量的改动只覆盖了 Resolver 这条路径。`Resource.Token` 字段在 Load 后保留 `${VAR}` 原始字符串，Resolver 在组装 `provider.Repo` 时调用 `config.ResolveToken()` 解析。但以下命令直接从 `cfg.Resources[name].Token` 取值，绕过了 Resolver：

| 命令 | 文件:行号 | 取值方式 |
|------|-----------|----------|
| sync | cmd/sync.go:101 | `res.Token` |
| list --remote | cmd/list.go:316 | `res.Token` |
| list --remote -t groups | cmd/list.go:428 | `rq.res.Token` |
| pipeline list/watch | cmd/pipeline.go:96 | `res.Token` |
| interactive sync | cmd/interactive.go:675 | `res.Token` |

这些路径拿到的 token 仍是 `${GITLAB_TOKEN}` 字符串，传给 provider 后导致 401（而非清晰的"环境变量未设置"报错）。

此外，`ResolveToken()` 的正则 `^\$\{([A-Za-z_][A-Za-z0-9_]*)}$` 不会匹配被引号包裹的值，如 `'${GITLAB_TOKEN}'` 或 `"${GITLAB_TOKEN}"`。

## Goals / Non-Goals

**Goals:**
- 所有使用 token 的代码路径通过统一方法获取已解析的 token
- `ResolveToken()` 自动清理成对的首尾引号（单引号、双引号）
- 不改变 `config.Load()` 的行为——仍然保留原始值不解析
- 保持 write-back 行为——配置文件回写时仍保留 `${VAR}` 占位符
- 错误信息包含 resource/group/repo 上下文，方便定位问题

**Non-Goals:**
- 不重构 Resolver 使其成为所有命令的统一入口（改动量过大）
- 不改变 token 在 provider 中的使用方式
- 不处理 token 缓存或过期机制
- 不处理非 token 字段的环境变量展开

## Decisions

### Decision 1: 在 `Resource` 上新增 `ResolvedToken()` 方法

**选择**: `Resource.ResolvedToken() (string, error)` 封装"获取原始 token → 去引号 → 解析环境变量"。

```go
func (r Resource) ResolvedToken() (string, error) {
    return ResolveToken(r.Token)
}
```

**替代方案**: 在 `Config` 上提供 `ResolveResourceToken(name string)` 方法。
**否决原因**: 需要传 name 字符串间接查找，不如直接挂在 Resource 上自然。Resource 是值类型，`ResolvedToken()` 是纯函数，不需要修改状态。

**理由**: `ResolvedToken()` 语义清晰——调用方知道自己在获取"已解析的 token"。返回 `(string, error)` 强制调用方处理解析失败。Go 编译器不会让你忽略 error。

### Decision 2: 引号清理放在 `ResolveToken()` 内部

**选择**: `ResolveToken()` 在正则匹配前先调 `stripQuotes()` 去除成对首尾引号。

```go
func stripQuotes(s string) string {
    if len(s) >= 2 {
        if (s[0] == '\'' && s[len(s)-1] == '\'') ||
           (s[0] == '"' && s[len(s)-1] == '"') {
            return s[1 : len(s)-1]
        }
    }
    return s
}

func ResolveToken(token string) (string, error) {
    token = stripQuotes(token)
    // ... 原有正则匹配逻辑
}
```

**替代方案**: 在 `config.Load()` 或 `ResolvedToken()` 里清理引号。
**否决原因**: `ResolveToken()` 是所有解析路径的必经入口（`ResolvedToken()` 内部也调用它、Resolver 里对 group/repo override token 也直接调用它）。放在这里一步到位，不需要每个调用方都记得先去引号。

**理由**: Group/Repo 的 `Token` 字段也可能带引号，且它们没有 `ResolvedToken()` 方法。将引号清理放在 `ResolveToken()` 内部可以覆盖所有场景。

### Decision 3: Resolver 中的 override token 也统一经过 `ResolveToken()`

**选择**: Resolver 中 group/repo 的 token override（`g.Token`、`repo.Token`）仍直接调用 `config.ResolveToken()`，但 `ResolveToken()` 已内置引号清理。Resource 默认 token 改为调用 `res.ResolvedToken()`。

**替代方案**: 给 Group/Repo 也加 `ResolvedToken()` 方法。
**否决原因**: Go 的值接收者方法需要每个类型都定义一遍，且 Group/Repo 的 token override 是可选字段，语义不同。直接调 `config.ResolveToken()` 更简单直接。

### Decision 4: pipeline.go 的 `resolvePipelineInput` 改用已解析的 token

**选择**: `pipeline.go:96` 当前通过 Resolver 获取 repo 信息后再从 `cfg.Resources[r.Resource]` 取 token。改为调用 `res.ResolvedToken()`。

**注意**: pipeline 已经经过 Resolver（`resolver.ResolveAndFilter`），Resolver 内部会解析 token 并存入 `provider.Repo.Token`。但 `resolvePipelineInput` 没有使用 `repos[0].Token`，而是重新从 Resource 取。有两个修复路径：

1. 改为使用 `repos[0].Token`（Resolver 已解析）
2. 继续从 Resource 取，但调 `res.ResolvedToken()`

选择路径 2——保持 Resource 级别获取的语义（pipeline 需要 Resource 级别的 APIURL 等信息），与 Resource 取 token 保持一致。

## Risks / Trade-offs

- **[引号清理误伤]** → 如果 token 值本身以引号开头结尾且是有效 token（极不可能），会被误去引号。缓解：只去除成对的首尾引号，不对称的不处理。明文 token 通常不以引号开头。
- **[Group/Repo override token 不走 ResolvedToken]** → Group/Repo 没有 `ResolvedToken()` 方法，但 `ResolveToken()` 已内置引号清理，效果等价。
- **[AddResource 立即解析行为不变]** → `AddResource()` 调用 `ResolveToken()`，同样受益于引号清理。
