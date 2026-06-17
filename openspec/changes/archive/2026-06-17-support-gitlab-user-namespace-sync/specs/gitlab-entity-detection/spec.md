## ADDED Requirements

### Requirement: GitLab provider 检测命名空间实体类型
GitLab provider 在按 `group.path` 列举仓库前 SHALL 识别该路径对应的命名空间实体类型（Group 或 User），并基于识别结果选择正确的仓库列举接口。

#### Scenario: 路径对应 Group
- **WHEN** `group.path` 能被 GitLab Group 查询接口解析为有效 Group
- **THEN** provider SHALL 使用 Group 仓库接口列举仓库
- **AND** 当 `recursive: true` 时 SHALL 继续递归子组并汇总仓库

#### Scenario: 路径对应 User namespace
- **WHEN** `group.path` 在 Group 查询中返回“Group Not Found”
- **AND** 该路径可被识别为有效 GitLab 用户命名空间
- **THEN** provider SHALL 使用用户仓库接口列举该用户可见仓库
- **AND** 返回的 repo `path` SHALL 保持远端命名空间路径格式（如 `sunzhenkai/grepom`）

#### Scenario: 路径既不是 Group 也不是 User
- **WHEN** `group.path` 无法匹配 Group 且无法匹配 User namespace
- **THEN** provider SHALL 返回明确的“path 不存在或不可访问”错误

### Requirement: GitLab Group 现有行为向后兼容
对已存在的 Group 路径配置，GitLab provider SHALL 保持与当前实现一致的仓库发现行为，不引入 breaking change。

#### Scenario: 组织 Group 同步保持不变
- **WHEN** `group.path` 为组织 Group 且包含子组仓库
- **THEN** 同步结果中的仓库集合与改动前保持等价（仅允许顺序差异）
