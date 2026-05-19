## Why

当同一源分支已经存在打开的 MR/PR 时，`grepom mr` 命令会报错退出：
- GitLab 返回 409 Conflict: "Another open merge request already exists for this source branch"
- GitHub 返回 422 Unprocessable Entity: "A pull request already exists for..."

用户期望的行为是：如果 MR/PR 已存在，应该直接返回现有 MR/PR 的地址，而不是报错。这是一种幂等性需求——"确保我有一个 MR"比"严格创建一个新的"更符合实际使用场景。

## What Changes

- GitLab Provider: 收到 409 时，通过搜索 API 查找同源分支的已打开 MR，找到后返回该 MR 而非报错
- GitHub Provider: 收到 422 时，通过搜索 API 查找同源分支的已打开 PR，找到后返回该 PR 而非报错
- `MergeRequest` 结构体新增 `AlreadyExists` 布尔字段，区分"新建"和"已存在"
- `cmd/mr.go` 输出逻辑区分两种情况：新建显示 ✅，已存在显示 ℹ️ 并提示地址
- 同步更新多语言 README 文档

## Capabilities

### New Capabilities
- `idempotent-mr`: MR/PR 创建的幂等性处理——当同源分支已有打开的 MR/PR 时，自动查找并返回已有记录，而非报错退出

### Modified Capabilities
<!-- 无现有 capability 需要修改 -->

## Impact

- 代码影响范围: `mergerequest/` 包（mergerequest.go、gitlab.go、github.go）、`cmd/mr.go` 及对应测试文件
- API 行为变化: `CreateMergeRequest` 返回值语义从"严格创建"变为"创建或获取已有"，但返回类型不变，不影响外部接口签名
- 用户体验: 已有 MR 场景从报错变为友好提示 + 展示已有 MR 地址
- 依赖: 无新外部依赖，仅使用已有的 GitLab/GitHub REST 搜索 API