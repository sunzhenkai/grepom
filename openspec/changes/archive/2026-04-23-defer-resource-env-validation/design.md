## Context

当前 `config.Load()` 在加载配置时（`config/config.go` 第 161-201 行）立即遍历所有 resource、group、repo 的 token 字段，调用 `resolveToken()` 将 `${ENV_VAR}` 占位符解析为实际值。如果环境变量未设置，加载直接报错失败。

这造成的问题是：
1. 用户在只使用 GitHub 的环境下，如果配置中还有 GitLab resource 且 `GITLAB_TOKEN` 未设置，整个 `Load()` 失败
2. 即使某个 resource 被标记为 `enabled: false`，其 token 对应的环境变量仍需存在
3. `cmd/add.go` 不得不引入 `loadExistingConfig()` 变通方案来绕过这个问题（第 261 行）

**关键数据流**：
```
config.Load() → resolveToken()（eager，所有 token）→ validate() → Resolver（使用已解析的 token）
```

**目标数据流**：
```
config.Load() → 保存原始 ${VAR} → validate() → Resolver → resolveToken()（lazy，仅启用的 resource）
```

## Goals / Non-Goals

**Goals:**
- `config.Load()` 不再因环境变量未设置而失败，始终能成功加载配置
- 仅在实际使用 token 时（Resolver 解析阶段、AddResource 等）才解析环境变量
- 保持 write-back 行为不变：配置文件回写时仍保留 `${VAR}` 占位符
- `cmd/add.go` 的 `loadExistingConfig()` 变通方案可以简化或移除

**Non-Goals:**
- 不改变 `${ENV_VAR}` 占位符语法
- 不改变 token 在 provider 中的使用方式
- 不处理非 token 字段的环境变量展开
- 不实现 token 缓存或过期机制

## Decisions

### Decision 1: token 在 `config.Load()` 中保留原始值，不解析

**选择**: `Load()` 中移除 `resolveToken()` 调用，将原始 `${VAR}` 字符串直接存入 `Resource.Token`、`Group.Token`、`Repo.Token` 字段。rawTokens map 仍保留用于 write-back 场景（区分明文 token 和占位符 token）。

**替代方案**: 引入新类型 `LazyToken` 封装延迟解析逻辑。
**否决原因**: 改动面过大，需要修改所有使用 Token 字段的代码。

**理由**: 最小改动方案。rawTokens map 继续承担"记录原始值用于 write-back"的职责。但需要确保在使用 token 前进行解析。

### Decision 2: 在 Resolver 层面解析 token

**选择**: `repo.Resolver.resolveInternal()` 中获取 token 后调用 `resolveToken()` 进行延迟解析。仅对实际参与解析的（即未被 disabled 过滤掉的）resource 的 token 进行解析。

**替代方案**: 在 provider 层面延迟解析。
**否决原因**: provider 接口简洁，不应感知配置细节。Resolver 是"配置 → 运行时"的桥梁，更合适。

**理由**: Resolver 已经是 auth priority chain 的组装点，在这里解析 token 是最自然的位置。

### Decision 3: AddResource 仍立即解析

**选择**: `config.AddResource()` 函数在添加新 resource 时仍立即调用 `resolveToken()`，验证用户输入的 token 是否有效。

**理由**: `add resource` 是用户主动操作，立即反馈 token 是否有效是合理的。如果延迟到使用时才发现 token 无效，用户体验更差。

### Decision 4: 错误信息增加上下文

**选择**: 延迟解析失败时，错误信息需包含 resource name / group name 等上下文，帮助用户定位是哪个配置项的环境变量缺失。

**理由**: eager 解析时错误能直接关联到 Load 阶段的 resource。延迟解析后，需要确保错误信息同样清晰。

## Risks / Trade-offs

- **[运行时才报错而非启动时]** → 用户可能在执行 clone/sync 中途才遇到环境变量未设置的报错。缓解：在 Resolver 解析阶段统一报错，此时还未开始实际 IO 操作。
- **[rawTokens map 的双重职责变化]** → rawTokens 原来既保存原始值又标记"已解析"。延迟解析后 rawTokens 仅用于 write-back，Token 字段本身可能就是 `${VAR}` 字符串。缓解：resolveToken 的幂等性保证——对非占位符字符串直接返回原值。
- **[测试覆盖]** → 需要确保 disabled resource 的 token 不会触发解析，这是最容易出遗漏的地方。
