## Why

当前 GitLab provider 在 `sync` 时将 `group.path` 一律按 Group 处理，先调用 `/api/v4/groups/{path}`。当用户配置的是个人命名空间（如 `sunzhenkai`）而非 Group 时，GitLab 返回 `404 Group Not Found`，导致该 group 无法同步，影响个人目录场景可用性。

## What Changes

- 为 GitLab provider 增加“命名空间实体识别”能力，支持区分 Group 与 User namespace。
- 当 `group.path` 对应 Group 时，保持现有 `/groups/{id}/projects` 与递归子组行为不变。
- 当 `group.path` 对应 User namespace 时，改为使用用户仓库列表接口获取仓库，并保留现有 path 过滤与去重策略。
- 明确 404 场景的回退与错误语义，避免将“个人命名空间”误报为“group 不存在”。

## Capabilities

### New Capabilities
- `gitlab-entity-detection`: 定义 GitLab provider 对 Group/User namespace 的识别与路由规则，覆盖个人命名空间同步行为。

### Modified Capabilities
- `sync-command`: 补充 GitLab provider 在个人命名空间场景下的同步行为与错误处理预期。

## Impact

- 受影响代码：`provider/gitlab.go`、`provider/gitlab_test.go`、`cmd/sync.go`（错误展示路径可能微调）。
- 受影响文档：`openspec/specs/sync-command/spec.md`（行为补充）、新增 `openspec/specs/gitlab-entity-detection/spec.md`。
- 外部影响：无 breaking change；现有 Group 同步行为保持兼容，新增个人命名空间可同步能力。
