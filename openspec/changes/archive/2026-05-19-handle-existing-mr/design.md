## Context

`grepom mr` 命令用于从命令行创建 Merge Request (GitLab) 或 Pull Request (GitHub)。当前实现在调用平台 API 创建 MR/PR 时，如果同源分支已有打开的 MR/PR，会直接返回错误：

- **GitLab**: 返回 `409 Conflict`，错误消息为 "Another open merge request already exists for this source branch: !N"
- **GitHub**: 返回 `422 Unprocessable Entity`，错误消息包含 "A pull request already exists for..."

当前代码在 `gitlab.go` 第 93-94 行和 `github.go` 第 116-122 行统一将非成功状态码视为错误，没有对"已存在"场景做特殊处理。

用户的使用诉求是幂等的：执行 `grepom mr` 后能得到一个可用的 MR 地址，无论它是刚创建的还是之前已存在的。

## Goals / Non-Goals

**Goals:**

- 当 MR/PR 已存在时，自动查找并返回已有记录的地址，而非报错退出
- 在输出中区分"新建"和"已存在"两种情况，让用户明确知道结果
- 同时覆盖 GitLab (409) 和 GitHub (422) 两个 Provider

**Non-Goals:**

- 不修改 `CreateMergeRequest` 的函数签名（保持返回 `(*MergeRequest, error)` 不变）
- 不实现"更新已有 MR"的功能（如修改标题、描述等）
- 不处理 MR/PR 审批、合并等后续操作

## Decisions

### Decision 1: 在 CreateMergeRequest 内部处理 409/422，而非新增接口方法

**选择**: 在 `CreateMergeRequest` 方法内部透明处理

**替代方案**:
- A) 新增 `EnsureMergeRequest` 方法 → 增加 API 表面积，调用方需要知道何时用哪个
- B) 新增 `FindMergeRequest` 方法让调用方自行处理 → 调用方代码复杂化

**理由**: 从用户视角看，`grepom mr` 的意图是"让我有一个 MR"，而不是"严格创建新的"。内部处理最简洁，对调用方透明。

### Decision 2: 通过 `MergeRequest.AlreadyExists` 字段区分新建 vs 已存在

**选择**: 在现有 `MergeRequest` 结构体上加 `AlreadyExists bool` 字段

**替代方案**:
- A) 返回自定义 error 类型 → 调用方需要类型断言，不够 Go-idiomatic
- B) 不区分，统一输出 → 丢失了有用信息（用户不知道是新建的还是已有的）

**理由**: 加字段侵入性最小，不改接口签名，调用方可根据字段自由选择输出方式。

### Decision 3: 409/422 时通过搜索 API 查找已有 MR/PR

**选择**: 收到 409/422 后，调用对应平台的搜索 API 查找已有记录

- GitLab: `GET /api/v4/projects/:id/merge_requests?source_branch=xxx&state=opened`
- GitHub: `GET /repos/:owner/:repo/pulls?head=:owner::branch&state=open`

**替代方案**:
- A) 从 409 错误体解析 MR iid → 不可靠，GitLab 只在 message 字符串里包含 `!1`，非结构化数据
- B) 不搜索，只提示"MR 已存在，请去网页查看"→ 用户体验差，没有给出具体地址

**理由**: 搜索 API 返回完整的 MR/PR 对象，可以精确获取标题、URL 等信息。

### Decision 4: 搜索失败时回退到原始错误

**选择**: 如果搜索 API 也失败（如权限不足、网络问题），返回原始的 409/422 错误信息

**理由**: 搜索失败不应该掩盖原始错误。用户至少能看到"已有 MR"的提示信息，即使无法获取详情。

## Risks / Trade-offs

- **[搜索 API 权限]** GitLab 搜索 API 需要至少 `read_api` 权限的 token → 如果创建 MR 的 token 没有 `read_api` 权限，搜索会失败。但创建 MR 本身需要更高权限，通常包含 `read_api`，风险较低。
- **[多次 API 调用]** 409/422 场景会多一次搜索请求 → 这是边缘场景（正常创建只走一次 POST），可接受。
- **[race condition]** 理论上在搜索时原 MR 可能已被关闭/合并 → 极低概率，且此时用户可以再次执行命令。不在本次处理范围内。
- **[GitHub head 参数格式]** GitHub 的 `head` 参数格式为 `owner:branch`，需从 `RepoPath`（`owner/repo`）中提取 owner → 实现时需注意处理。