## Context

grepom 是一个 Go 实现的 Git 仓库管理 CLI 工具，通过 YAML 配置文件管理多个 group 下的仓库。现有 `dedup` 命令只能按 repo name 做跨组去重，将 target group 中与 reference group 同名的 repo 移除并加入 exclude_repos。

当前系统已有 `AddGroupRepo` 和 `SyncGroupRepos` 按 URL 精确匹配去重，但无法捕获 URL 格式不一致导致的重复（如 HTTPS vs SSH、有/无 .git 后缀）。组内重复只能通过手动检查发现，跨组 URL 重复完全没有检测机制。

## Goals / Non-Goals

**Goals:**
- 在 dedup 命令中新增组内 URL 去重，自动检测并清理同一 group 内 URL 指向同一仓库的重复条目
- 在 dedup 命令中新增跨组 URL 重复警告，提示用户不同 group 管理了相同的远程仓库
- 保持现有按 name 跨组排除逻辑完全向后兼容
- `--group` 从必选变为可选，使无参数运行即可完成最常见的组内去重和跨组检查

**Non-Goals:**
- 不自动删除跨组重复仓库（只警告，用户自行决策）
- 不改变现有 `exclude_repos` 的语义和行为
- 不做 URL 规范化的自动修正（只用于比较，不改写配置中的原始 URL）
- 不处理独立 repos（`Config.Repos`）的去重，只处理 group 内的 repos

## Decisions

### Decision 1: URL 规范化策略

**选择**: 规范化匹配（去掉协议、.git 后缀、统一 SSH/HTTPS 为 host/path）

**替代方案**:
- 精确匹配：最安全但会遗漏 HTTPS vs SSH 等真实重复
- 路径匹配（不含 host）：会误判不同 Git 平台上的同名项目

**规范化规则**:
1. 去掉 `https://`、`http://` 前缀
2. 去掉 `.git` 后缀
3. SSH 格式 `git@host:path` → `host/path`
4. 去掉末尾 `/`
5. host 部分转小写（DNS 大小写不敏感）
6. path 部分保留原样（Git 路径大小写敏感）
7. 保留端口号（如有）

实现位置：`config/normalize.go`，导出 `NormalizeRepoURL(url string) string`。

### Decision 2: 组内去重保留策略

**选择**: 保留第一个出现的条目，删除后续重复

**替代方案**:
- 保留 name 最规范的
- 保留 URL 最完整的

**理由**: 最简单、最可预测，YAML 中先出现的条目通常是手动维护或最先 sync 的。

### Decision 3: 命令执行流程设计

**选择**: 三步顺序执行，Step 1+2 始终运行，Step 3 按条件触发

```
Step 1: 组内去重（按 URL）— 始终执行
Step 2: 跨组 URL 警告（按 URL）— 始终执行
Step 3: 跨组 name 去重（按 name）— 仅当 --group + --reference 同时指定时触发
```

**替代方案**:
- 用 flag 切换行为（`--by-url` / `--by-name`）：增加认知负担
- 替换旧逻辑：破坏向后兼容

**理由**: 无参数时做最有用的事（组内去重+跨组警告），旧行为在指定 `--group --reference` 时自然触发。

### Decision 4: --group 从必选变为可选

**选择**: `--group` 可选，不指定时对所有 group 执行 Step 1 和 Step 2

**理由**: 用户最常见的需求是"检查我所有 group 里有没有重复"，而不是"只检查某个 group"。

### Decision 5: 组内去重不加入 exclude_repos

**选择**: 只从 repos 列表删除多余条目，不追加 exclude_repos

**理由**: 组内重复是"冗余数据"而非"应该排除的仓库"。sync 流程的 `SyncGroupRepos` 已有按 URL 去重逻辑，不会重新引入。加入 exclude_repos 语义错误且可能阻止用户后续手动添加。

## Risks / Trade-offs

- **[URL 规范化不完整]** → 规范化规则覆盖 HTTPS/HTTP/SSH 三种主要格式，对于非标准 URL 格式（如带 userinfo 的 URL）可能无法正确匹配。缓解：复用现有 `ExtractRemotePath` 的解析逻辑并扩展 host 保留。
- **[旧用户习惯变更]** → `--group` 不再是必选参数，现有脚本 `grepom dedup` 不带 `--group` 时行为从报错变为执行组内去重+跨组警告。缓解：新行为是合理的默认行为，不破坏已有带 `--group` 的用法。
- **[Step 1 组内去重后 Step 3 name 去重可能受影响]** → 如果组内去重删除了某个 repo，Step 3 的 name 匹配可能少了一个结果。缓解：Step 1 只删除真正的 URL 重复，不影响 name 匹配的正确性。
