## Context

当前 `grepom` 在两个独立场景中存在问题：

1. **`mr` 命令的 provider 检测**：当仓库的 git remote URL 包含嵌入凭据（`https://oauth2:TOKEN@gitlab.company.com/...`）时，`extractHost()` 函数会错误地将 `oauth2:TOKEN@gitlab.company.com` 整体作为 host 返回，导致 provider 匹配失败。

2. **`sync` 命令的组内去重**：`SyncGroupRepos()` 仅将新增仓库与配置文件中已存在的仓库列表进行 URL 对比去重，不对 `newRepos` 批次本身进行自去重。如果 provider 返回的数据包含重复条目（例如边界情况），同一仓库可能被重复写入。

## Goals / Non-Goals

**Goals:**
- `extractHost()` 正确处理 HTTPS/HTTP URL 中的 `user:password@` 和 SSH URL 中的 `user@` 格式，提取纯净的 host
- `grepom sync` 在同一组内不产生重复的仓库记录（URL 维度去重）
- 修改最小化、局部化，不引入新的依赖
- 向后兼容：所有现有 URL 格式（不带 userinfo 的）保持原有行为

**Non-Goals:**
- 不修改 provider 发现逻辑本身（`detectProvider` 的策略 1/2 保持不变）
- 不修改跨组去重行为（`dedup` 命令职责不变）
- 不改变配置文件格式
- 不对 URL 进行规范化（如 `.git` 后缀、协议统一等）—— 仅做 userinfo 剥离和批次内 URL 精确去重

## Decisions

### D1: 在 `extractHost()` 中剥离 userinfo

**选项**:
- A. 在 `extractHost()` 内部，去除 scheme 后、查找 `/` 之前，先搜索 `@` 并剥离
- B. 在 `detectProvider()` 中调用 `extractHost()` 之后再处理 userinfo
- C. 使用 Go 标准库 `net/url` 解析

**决策**: 选 **A**（在 `extractHost()` 内部处理）。

**理由**:
- `extractHost()` 是整个代码库中唯一的主机名提取点，在此修复一处即可保障全局
- 选项 C（`net/url.Parse`）可以正确处理，但会引入额外的错误处理和结构体分配，且 SSH 格式（`git@host:path`）不是标准 URL，需要特殊处理。选项 C 过于侵入性
- 选项 B 会导致重复逻辑（每个调用者都需处理）
- 选项 A 是最小改动，在现有字符串操作逻辑中增加一个 `@` 搜索步骤

**实现细节**:

```
原逻辑（HTTPS 路径）:
  rest = TrimPrefix(url, "https://")
  idx = Index(rest, "/"); return rest[:idx]

新逻辑:
  rest = TrimPrefix(url, "https://")
  // 1. 剥离 userinfo@
  if atIdx := Index(rest, "@"); atIdx >= 0 {
      slashIdx := Index(rest, "/")
      if slashIdx < 0 || atIdx < slashIdx {  // @ 在 host 部分
          rest = rest[atIdx+1:]
      }
  }
  // 2. 提取 host
  idx = Index(rest, "/"); return rest[:idx]
```

`ssh://` 路径同理。对于 `git@host:path` 的 SCP 风格 URL，当前逻辑已正确处理（在 `@` 之后查找 `:`），无需修改。

### D2: 双重去重保障

**选项**:
- A. 仅在 `cmd/sync.go` 的 `newGroupRepos` 构建循环中去重
- B. 仅在 `config.SyncGroupRepos()` 中去重
- C. 两处都去重（双重保障）

**决策**: 选 **C**（双重保障）。

**理由**:
- `cmd/sync.go` 中的去重是"第一道防线"—— 防止重复数据流入 `SyncGroupRepos()`
- `config.SyncGroupRepos()` 中的去重是"最后一道防线"—— 无论调用者如何，确保写入文件的数据无重复
- `SyncGroupRepos()` 是一个公共 API，未来可能有其他调用者。自去重应作为其契约的一部分
- 性能开销极小（O(n²) 对少量 repo 几乎没有影响）
- 选项 C 符合防御性编程原则

**实现细节**:

在 `cmd/sync.go` 中：
```go
// 在 newGroupRepos 构建循环内
exists := false
for _, existing := range newGroupRepos {
    if existing.URL == r.CloneURL {
        exists = true
        break
    }
}
if !exists {
    newGroupRepos = append(newGroupRepos, ...)
}
```

在 `config.SyncGroupRepos()` 中：
```go
// 在遍历 newRepos 的外层，增加一个 seen map
seen := make(map[string]bool)
for _, nr := range newRepos {
    if seen[nr.URL] {
        continue  // 跳过 newRepos 内部重复
    }
    seen[nr.URL] = true
    // ... 现有逻辑: 与 group.Repos 对比 ...
}
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|---------|
| `extractHost()` 修改可能影响其他调用者（status, clone 等命令也使用该函数） | 该函数当前仅用于 `mr` 命令，`go grep` 确认无其他调用者。若未来扩展，行为变更仍安全：去重后的 host 才是正确值 |
| `@` 符号可能出现在非 userinfo 位置（如 GitLab subgroup URL 不含 `@`） | userinfo 的 `@` 必然在第一个 `/` 之前。我们通过 `atIdx < slashIdx` 条件保护。正常情况下 URL 的 path 部分不含 `@` |
| `seen` map 增加内存开销 | 对每个 group 的 repo 批次，数量级在百到千，内存开销可忽略 |

## Open Questions

无。两个修复的范围和实现路径均已明确。
