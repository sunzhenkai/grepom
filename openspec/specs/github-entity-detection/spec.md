### Requirement: GitHub provider 检测实体类型
GitHub provider 在列举仓库前 SHALL 先通过 `GET /users/{name}` 获取实体的 `type` 字段，判断目标是 `Organization` 还是 `User`。

#### Scenario: Organization 类型的实体
- **WHEN** GitHub `/users/{name}` 返回 `type` 为 `"Organization"`
- **THEN** SHALL 使用 `GET /orgs/{name}/repos?type=all` 列举所有仓库（含私有）

#### Scenario: User 类型的实体
- **WHEN** GitHub `/users/{name}` 返回 `type` 为 `"User"`
- **THEN** SHALL 使用 `GET /users/{name}/repos?type=all` 列举所有仓库

#### Scenario: 实体类型检测返回 404
- **WHEN** `GET /users/{name}` 返回 HTTP 404
- **THEN** SHALL 回退到 `GET /orgs/{name}/repos?type=all` 尝试列举仓库

### Requirement: GitHub provider 发现私有仓库
当 group 关联的 GitHub 实体为 Organization 时，sync 操作 SHALL 能发现该 Organization 下的私有仓库，前提是配置的 token 具有相应权限。

#### Scenario: Organization 含私有仓库
- **WHEN** 一个 Organization 拥有 22 个公开仓库和 2 个私有仓库
- **AND** 配置的 token 具有该 Organization 的读取权限
- **THEN** `grepom sync` SHALL 发现全部 24 个仓库，包括私有仓库

#### Scenario: Token 无权限访问私有仓库
- **WHEN** 配置的 token 不具有某私有仓库的读取权限
- **THEN** GitHub API 不返回该私有仓库，sync 仅发现 token 可见的仓库（非报错行为）

### Requirement: 向后兼容
修改后的 GitHub provider SHALL 对已有公开仓库的同步行为保持不变，不引入 breaking change。

#### Scenario: 纯公开仓库的 Organization
- **WHEN** 一个 Organization 仅拥有公开仓库
- **THEN** sync 结果与修改前完全一致（同样数量的仓库和相同的内容）

#### Scenario: 个人用户仓库
- **WHEN** group 的 path 指向一个个人 GitHub 用户
- **THEN** sync 仍通过 `/users/{name}/repos` 列举仓库，行为不变
