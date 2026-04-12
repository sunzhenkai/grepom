## Context

当前 grepom 的 group 和 standalone repo 都强制绑定一个 resource（认证资源）。Resource 提供 provider 类型、host、token 和 SSH key，用于远程 API 发现（sync）和 clone/pull 认证。

实际场景中，用户可能只想用 grepom 管理一组手动维护的仓库（如内部工具链、个人项目集合），不需要通过 API 自动发现。此时强制配置一个 resource 是多余的。

当前架构中 resource 被使用的场景：
- **sync 命令**：需要 resource 的 provider + url + token 调用远程 API
- **list --remote**：需要 resource 查询远程仓库列表
- **clone/pull**：需要 resource 构建 clone URL 和提供认证
- **配置验证**：validate() 强制要求 resource 字段非空且引用存在

## Goals / Non-Goals

**Goals:**
- Group 可以不绑定 resource，此时 repos 由用户手动维护
- Standalone repo 可以不绑定 resource，此时必须提供 `url` 字段
- 依赖 resource 的操作（sync、list --remote）在缺少 resource 时优雅跳过并提示用户
- clone/pull 在无认证信息时使用系统默认 SSH/HTTPS

**Non-Goals:**
- 不支持"部分绑定 resource"的 group（即 group 不绑定 resource，但 group 内的某个 repo 可以有自己的 resource）——这个可以作为后续扩展
- 不改变 resource 本身的定义方式
- 不引入新的 resource 类型或 provider

## Decisions

### Decision 1: resource 字段变为可选，而非引入新类型

**选择**：直接将 Group 和 Repo 的 `resource` 字段从必填改为可选。

**理由**：这是最小变更方案。不需要引入新的数据结构或"手动管理"类型，只需要在验证和运行时逻辑中处理空值。

**替代方案**：引入 `manual` provider 类型 —— 增加了概念复杂度，用户需要理解 manual provider 与不填 resource 的区别。

### Decision 2: 无 resource 的 group，sync 命令直接跳过

**选择**：sync 命令检测到 group 无 resource 时，输出提示信息并跳过该 group。

**理由**：sync 的核心功能是远程发现，没有 resource 就无法调用 API。跳过是自然的处理方式。

### Decision 3: 无 resource 的 standalone repo 必须提供 url

**选择**：standalone repo 不绑定 resource 时，`url` 字段变为必填。url 提供完整的 clone 地址（如 `git@github.com:user/repo.git`）。

**理由**：没有 resource 时系统无法自动构建 clone URL，必须由用户显式提供。

### Decision 4: 无 resource 的 group 内 repo 仍使用 url 字段

**选择**：无 resource 的 group，其 repo 条目的 `url` 字段由用户手动填写完整 clone URL。resolver 解析时直接使用该 URL，不再从 resource 推导。

**理由**：保持数据模型一致。GroupRepo 已有 url 字段，只是当前由 sync 自动填充。

### Decision 5: 认证链路保持兼容

**选择**：clone/pull 时，如果 repo 没有关联 resource，使用系统默认的 SSH/HTTPS 认证（即 git 命令本身的行为），不注入额外的 token 或 SSH key。

**理由**：用户可能已经在 SSH config 或 git credential 中配置了认证，不需要 grepom 重复处理。

## Risks / Trade-offs

- **[向后兼容]** 已有配置文件都绑定了 resource，不受影响。仅在移除 resource 或新增无 resource 的 group/repo 时触发新逻辑。 → 无迁移风险。
- **[用户体验]** 用户可能忘记配置 resource 导致 sync 不执行。 → 通过明确的输出提示来缓解，告知用户该 group 未绑定 resource，sync 已跳过。
- **[认证失败]** 无 resource 的 repo 在 clone 时如果系统没有配置对应的 SSH/credential，git 命令会失败。 → 这是预期行为，错误信息由 git 本身提供，足够清晰。
