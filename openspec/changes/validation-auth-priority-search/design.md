## Context

grepom 是一个 Go CLI 工具，通过单个 YAML 配置文件管理多个 git 仓库。当前架构为三层配置模型：Resource（认证连接）→ Group（远端分组）→ Repo（仓库条目）。

当前存在三个问题：
1. `add` 命令（`add group`、`add repo`）未在执行时校验引用的 resource 是否存在或名称是否重复，错误延迟到 clone/sync 时才发现
2. clone 认证优先级是 token → SSH，但用户实际更多通过 SSH 推送变更，SSH 应优先
3. 缺少按名称模糊搜索仓库的功能，`list` 命令仅支持精确匹配

## Goals / Non-Goals

**Goals:**
- 在 `add group`、`add repo` 命令执行时立即校验配置一致性（resource 存在性、名称唯一性）
- 将 clone 认证优先级调整为 SSH 优先策略
- 新增 `search` 子命令，支持子串匹配搜索仓库名称

**Non-Goals:**
- 不修改 `add resource` 的逻辑（resource 是 map 结构，天然保证唯一性）
- 不修改配置文件的 YAML schema
- 不引入正则或 glob 搜索（仅支持子串匹配）
- 不修改 interactive 模式（可后续迭代）

## Decisions

### 1. add 命令校验策略：加载-校验-追加-写回

**决定**: 在 `add group` 和 `add repo` 命令中，先 `Load()` 已有配置，校验通过后再执行追加操作。

**理由**: 项目已有 `Load()` 函数内含完整校验（validate），但 `AddGroup()` 和 `AddRepo()` 使用的是 `ensureConfigFile()` 跳过了 validate。改为先 Load 再追加可以利用现有校验逻辑，同时也需要增加 add 层面的特有校验（如引用 resource 存在性、名称重复检测）。

**替代方案**:
- 在 `ensureConfigFile` 中加入 validate → 侵入性较大，且 Load 会解析 token 环境变量，对仅添加操作可能有副作用
- 在 config 包新增 `ValidateAddGroup()` / `ValidateAddRepo()` 函数 → 更好的隔离，保持 Add 函数的简洁性

**最终方案**: 在 `cmd/add.go` 层面加载已有配置进行校验（检查 resource 存在性、名称重复），config 包的 Add 函数保持不变。校验逻辑放在 cmd 层，因为它需要读取现有配置来执行跨条目检查。

### 2. 认证优先级：SSH 优先

**决定**: 新的 6 级优先级链：
1. group/repo SSH key（指定密钥）
2. group/repo token（HTTPS + token URL）
3. resource SSH key
4. resource token
5. 默认 SSH（推导 URL）
6. 裸 HTTP

**理由**: SSH 是推送变更的主要方式，优先使用 SSH 可以减少认证失败重试次数。对于需要 push 的场景，SSH key 认证是最可靠的方式。

**实现位置**: 
- `git/git.go` 的 `Clone` 函数：调整 strategies 构建顺序
- `repo/resolver.go`：Resolver 不需要修改，因为它只负责合并认证信息，不决定优先级
- `clone-auth-priority` spec：更新优先级链定义

### 3. search 命令：复用 Resolver + 子串匹配过滤

**决定**: 新增 `cmd/search.go`，复用 `repo.Resolver` 获取全量仓库列表，然后用 `strings.Contains` 进行大小写不敏感的子串匹配。

**理由**: 
- 复用已有的 Resolver 架构，避免重复代码
- 子串匹配足够满足用户需求，无需引入正则表达式
- `--group` 和 `--resource` 过滤器复用已有的 `Filter` 结构体（先子串搜索再按 group/resource 精确过滤）

**实现**:
- 新增 `repo.ApplySearchFilter()` 函数，支持子串匹配
- search 命令的输出格式复用 list 命令的表格样式

## Risks / Trade-offs

- **[Breaking] 认证优先级变更** → 已有用户如果依赖 token 优先的行为可能受影响。缓解：在 release notes 中明确说明优先级变更
- **add 命令校验需要加载配置** → 如果 token 环境变量未设置，Load 会报错。缓解：add 操作只检查结构引用关系，使用 `ensureConfigFile` 读取后单独校验，不调用完整 Load
- **search 子串匹配性能** → 配置文件中仓库数量通常在百级别以内，全量遍历性能无问题
