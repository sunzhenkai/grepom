## Context

当前 `provider/gitlab.go` 在 `ListRepos` 中对每个 `group.path` 固定执行 Group 流程：先通过 `/api/v4/groups/{path}` 解析 group，再通过 `/groups/{id}/projects` 拉取仓库。该实现对 GitLab 组织/子组场景有效，但对个人命名空间（user namespace）不成立，导致 `404 Group Not Found`。  
在 `cmd/sync.go` 中，单个 group 失败会记录错误并继续其他 group，因此用户会看到“部分报错 + 最终完成摘要”。

目标是在不破坏现有 Group 递归同步能力的前提下，补齐个人命名空间同步能力，并明确错误语义。

## Goals / Non-Goals

**Goals:**
- 支持 GitLab `group.path` 指向用户命名空间时的仓库发现。
- 保持 Group 路径（含 `recursive`）行为与结果不变。
- 对无法识别的 path 提供更可理解的错误语义，减少误导性提示。
- 保持 `sync` 侧现有过滤逻辑（path 过滤、exclude、去重）兼容。

**Non-Goals:**
- 不改动配置模型（仍使用 `groups[].path`）。
- 不引入新的 CLI 参数或 provider 配置字段。
- 不处理 GitLab 之外 provider 的行为变化。
- 不在本次引入并发或缓存优化。

## Decisions

### 1) 引入 GitLab 命名空间实体识别层
- 决策：在 GitLab provider 内新增“path 识别 -> 路由”的中间层，先判断 path 对应 Group 还是 User，再选择仓库列举接口。
- 理由：与现有 GitHub provider 的实体识别模式一致，能够将“路径语义”与“仓库拉取实现”解耦，便于后续扩展。
- 备选方案：
  - 方案 A：先调 Group 接口，404 时直接失败（现状）。问题是无法支持用户命名空间。
  - 方案 B：在 `sync` 层做 provider 特判。问题是破坏 provider 封装，增加 cmd 层复杂度。

### 2) Group 优先、User 回退的路由策略
- 决策：默认先尝试 Group 解析；若确认是“Group 不存在”类响应，再尝试 User 路径对应的仓库列表接口。
- 理由：兼容现有 Group 主路径，且对已有行为变更最小；只在明确不命中 Group 时进入 User 流程。
- 备选方案：
  - 方案 A：先 User 再 Group。会增加企业组织场景的额外请求。
  - 方案 B：总是并发请求两类接口。实现复杂且浪费配额。

### 3) 保持下游过滤与去重逻辑不变
- 决策：provider 返回统一 `Repo{Path, CloneURL...}`，继续沿用 `cmd/sync.go` 现有 path 前缀校验、exclude_repos 过滤和批次去重。
- 理由：减少行为漂移风险；无需修改写配置策略。
- 备选方案：
  - 在 provider 内额外做一次 path 过滤。会与 cmd 层逻辑重复，增加维护成本。

## Risks / Trade-offs

- [风险] GitLab 各部署版本对用户仓库接口细节差异较大  
  → 通过单元测试覆盖主要分支（Group 命中、User 命中、双 404、鉴权失败），并优先采用通用 REST 路径。

- [风险] 404 文案判断可能误判（仅靠字符串）  
  → 在错误分类时优先使用状态码 + 上下文路径，字符串仅作辅助。

- [风险] 新增分支导致请求次数上升  
  → 仅在 Group 未命中时触发 User 回退，常见 Group 场景不增加请求。
