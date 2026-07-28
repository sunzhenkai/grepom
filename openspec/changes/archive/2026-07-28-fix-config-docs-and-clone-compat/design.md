## Context

当前配置与克隆路径存在多处「文档/直觉 vs 实现」偏差：

1. `resources` 正式类型是 `map[string]Resource`，但 README 仍展示 list；list 直接 unmarshal 失败。
2. 绑定 resource 的独立 repo 一律 `ExtractRemotePath(url)` + `resource.HTTPSURL/SSHURL` 重拼；完整 URL（尤其 `ssh://`）会拼坏；HTTPS 完整 URL 也会被拆 path 后优先走 SSH。
3. 认证链故意「不再尝试裸 HTTP」，导致无 token 的公开 HTTPS 仓无法克隆。
4. `tryClone` 丢弃 git stderr，最终只报 `all authentication methods failed`。
5. `virtual_groups` 仅引用真实 groups；`--vgroup` 过滤时独立 repo（`GroupName==""`）被排除。

约束：兼容逻辑不得破坏现有 map + 相对路径配置；文档/示例只用虚构 host，不暴露真实私密仓库信息；实现语言 Go。

## Goals / Non-Goals

**Goals:**

- 加载期兼容 resources list / map
- 完整 clone URL 原样使用；相对路径保持与 resource 拼接
- 按 URL 协议选择克隆策略，支持匿名 HTTPS
- 透出脱敏后的 git stderr；可选默认私钥兜底
- virtual_groups 可引用独立 repos，并接入现有 `--vgroup` 过滤
- 同步中英 README 与 `grepom example` 注释

**Non-Goals:**

- 不改为 go-git / 纯 Go SSH；继续调用系统 `git`
- 不自动改写用户磁盘上的旧 YAML 为 map（仅内存归一；Save 时按现有 map 写出即可）
- 不引入完整 SSH agent 协议对接；默认私钥文件探测即可
- 不支持 virtual_groups 嵌套其他 virtual_groups
- 不做交互式密码提示（保持「不触发 git 交互认证」）

## Decisions

### D1. resources list 在 Load 时归一为 map

- **做法**：自定义 `resources` 的 YAML 解码：若节点为 mapping → 直接解析为 map；若为 sequence → 每项要求有非空 `name`，转为 map key，重复 name 报错。
- **为何不用双写结构体**：正式 API 仍是 map；兼容只在边界层。
- **备选**：拒绝 list 只改文档 → 已踩坑用户配置仍挂，否决。

### D2. repo.url 分类：Relative / AbsoluteHTTP / AbsoluteSSH

解析优先级：

| 形态 | 判定 | CloneURL | SSHURL |
|------|------|----------|--------|
| 相对路径 | 无 scheme、非 `git@` | `resource.HTTPSURL(path)` | `resource.SSHURL(path)` |
| AbsoluteHTTP | `https://` / `http://` | **原样**（规范化 `.git` 可选） | 仍可从 path 推导 SSH（供有 key 时回退） |
| AbsoluteSSH | `git@` / `ssh://` | 转为可用 HTTPS（若能解析 host/path）或同 URL | **原样**（`ssh://` 可规范为 scp-like 供展示，clone 用原样或等价 scp） |

关键修正：`ssh://git@host/path` **禁止**再拼成 `git@resource.host:ssh://...`。

绑定 resource 时，resource 仍用于 token / ssh_key / provider；**不再强制用 resource.url 覆盖已声明的绝对 URL 的 host**（避免跨 host 错绑）。若相对路径，host 仍取自 resource。

### D3. 克隆策略按「首选协议」分支

在现有 5 级链基础上调整：

- **首选 HTTPS**（AbsoluteHTTP，或相对路径但用户显式只要 HTTPS——本期不做开关，相对路径保持 SSH 优先以兼容内网）：
  1. group/repo token HTTPS
  2. **匿名 HTTPS**（无 userinfo）← 新增，仅当 httpURL 非空
  3. group/repo SSH / resource SSH / default SSH / default keys
  4. resource token
- **首选 SSH**（AbsoluteSSH 或相对路径，保持现网行为）：
  1. group/repo SSH → group/repo token → resource SSH → default SSH → **default identity files** → resource token
  2. **不**在「无任何 token 且用户未写 HTTPS」时插入匿名 HTTPS（避免内网误走 HTTPS 触发交互/证书问题）；但 AbsoluteHTTP 必须能匿名成功。

「裸 HTTP」禁令收窄为：**禁止触发交互式 credential helper 的空认证探测之外的行为**；匿名 `git clone https://public/...` 明确允许。

### D4. 默认私钥兜底列表

未配置 `ssh_key` 时，在 default SSH（空 `GIT_SSH_COMMAND`）失败后，对存在的文件依次尝试：

`~/.ssh/id_ed25519`、`~/.ssh/id_rsa`、`~/.ssh/id_ecdsa`、`~/.ssh/id_ed25519_sk`

每项用与显式 `ssh_key` 相同的 `GIT_SSH_COMMAND=ssh -i ... -o IdentitiesOnly=yes`。文档仍推荐显式 `ssh_key`。

### D5. 错误诊断：读取 stderr 并汇总

- `tryClone` / `Pull`：捕获 Combined 或 Stderr 文本；`ExitError` 时包装为含 stderr 的错误。
- `Clone` 全部失败时：最终错误为 `all authentication methods failed:` + 各策略最后一行/摘要（脱敏 token URL）。
- `-v`：每步打印完整脱敏 stderr。
- 脱敏规则沿用现有 `sanitizeError` / `maskTokenURL`，并扩展覆盖 `git@` 无关泄漏。

### D6. virtual_groups.repos

```yaml
virtual_groups:
  work:
    groups: [frontend, backend]
    repos: [dotfiles, notes]
```

- `VirtualGroup` 增加 `Repos []string`
- 校验：每个名字必须存在于顶层 `repos`
- `ResolveGroupSelection` 扩展为同时返回 `standaloneRepoNames`，或新增 `ResolveScopeSelection`
- `repo.Filter` 增加 `RepoNames []string`：当 Groups/RepoNames 非空时，匹配「GroupName 在 Groups」**或**「GroupName 为空且 Name 在 RepoNames」
- `list groups` 展示 vgroup 时，repo 计数计入成员独立仓
- sync 对独立仓仍无 API 发现；`--vgroup` sync 只 sync 成员 groups（repos 无 sync 语义），但 clone/pull/list/status/search/prune 等包含独立仓

### D7. 文档与 example

- README / README_en：resources 改为 map；增加「URL 写法」「协议与认证」「ssh_key 建议」「vgroup repos」小节
- example：generic 示例同时展示相对路径与完整 HTTPS URL；注释说明行为；增加 `repos:` 于 virtual_groups
- **禁止**在文档中出现用户真实内网 host、组织名、仓库路径

## Risks / Trade-offs

- **[Risk] 匿名 HTTPS 被内网劫持 / 错误 host** → 仅对 AbsoluteHTTP 启用；相对路径默认仍 SSH 优先  
- **[Risk] list 兼容掩盖文档错误，长期双格式** → 文档以 map 为正；加载成功时可在 `-v` 提示「list 格式已兼容，建议改为 map」（可选，不强制）  
- **[Risk] 默认私钥探测增加失败尝试次数与耗时** → 仅文件存在才试；失败快；并发 clone 下可接受  
- **[Risk] AbsoluteHTTP 仍推导 SSHURL，有 key 时可能先 SSH 成功改变 remote** → 首选 HTTPS 分支把 SSH 放匿名 HTTPS 之后，公开仓优先 HTTPS remote  
- **[Risk] vgroup 含 repos 后 Filter 语义变复杂** → 单测覆盖「仅 groups / 仅 repos / 混合 / 与 --group 并集」

## Migration Plan

1. 发布含兼容逻辑的版本；旧 map 配置零改动。
2. list 格式配置无需迁移即可工作；Save/add 写出仍为 map。
3. 曾因完整 URL 拼坏而手动绕过的用户，改回标准配置即可。
4. 回滚：恢复旧二进制即可；新字段 `virtual_groups.*.repos` 在旧版本会被忽略（YAML 多余字段？需确认——当前用严格 struct，未知字段默认忽略 by yaml.v3），旧版本忽略 `repos` 键，行为退回「仅 groups」。

## Open Questions

- AbsoluteHTTP 在「已配置 ssh_key」时是否仍应先试匿名 HTTPS？（倾向：是，公开仓更快；私有 HTTPS 匿名失败后再 SSH/token）
- `ssh://` 是否统一规范为 scp-like 再 clone，还是原样交给 git？（倾向：原样交给 git，减少转换 bug）
