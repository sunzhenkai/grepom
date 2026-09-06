## ADDED Requirements

### Requirement: pipeline list 命令列出指定仓库最近的 CI/CD pipelines

系统 SHALL 提供 `grepom pipeline list <repo-name>` 命令，列出指定仓库最近的 CI/CD pipeline 运行记录。

#### Scenario: 列出最近 5 条 pipeline（默认）
- **WHEN** 用户运行 `grepom pipeline list web-app`
- **THEN** 系统 SHALL 通过 resolver 查找名为 "web-app" 的 repo，从其 CloneURL 反推远程路径，获取 Resource 的 ServerURL 和 Provider 类型，调用对应 PipelineProvider 的 ListPipelines 方法，返回最近 5 条 pipeline 并格式化输出

#### Scenario: 自定义显示条数
- **WHEN** 用户运行 `grepom pipeline list web-app -n 10`
- **THEN** 系统 SHALL 返回最近 10 条 pipeline

#### Scenario: repo 不存在
- **WHEN** 用户运行 `grepom pipeline list nonexistent`
- **THEN** 系统 SHALL 输出错误信息 "repo not found: nonexistent" 并以非零退出码退出

#### Scenario: repo 没有 Resource 绑定
- **WHEN** 用户指定的 repo 未绑定任何 resource（无 CloneURL 或 Provider 信息）
- **THEN** 系统 SHALL 输出错误信息 "repo has no resource binding, cannot query pipelines" 并以非零退出码退出

### Requirement: pipeline list 输出格式

系统 SHALL 以表格形式输出 pipeline 列表，每行包含：ID、分支、commit SHA（短格式）、状态图标+文本、持续时间。

#### Scenario: 正常输出
- **WHEN** pipeline list 成功获取到 pipelines
- **THEN** 系统 SHALL 输出如下格式的表格：
  ```
  ID     BRANCH   SHA       STATUS          DURATION
  #1234  main     abc1234   ✅ success      2m34s
  #1233  main     def5678   ❌ failed       5m12s
  #1232  feat     901abcd   🔄 running      1m03s
  ```
  状态映射：running→🔄, pending→⏳, success→✅, failed→❌, canceled→🚫

#### Scenario: 无 pipeline 记录
- **WHEN** 指定仓库没有任何 pipeline 记录
- **THEN** 系统 SHALL 输出 "No pipelines found for <repo-name>."

### Requirement: pipeline list Token 获取

系统 SHALL 复用现有 token 优先级链路获取 API 访问令牌。`resolvePipelineInput` 函数 SHALL 通过 `Resource.ResolvedToken()` 获取已解析的 token，不再直接读取 `Resource.Token` 字段。

#### Scenario: Token 解析
- **WHEN** 执行 pipeline list
- **THEN** 系统 SHALL 通过 `repo.Resolver.ResolveAndFilter` 查找 repo，从其 Resource 获取 ServerURL 和 Provider 类型，并通过 `res.ResolvedToken()` 获取已解析的 token

#### Scenario: pipeline token 环境变量未设置时报错
- **WHEN** 用户运行 `grepom pipeline list web-app`，resource token 为 `${GITLAB_TOKEN}`，环境变量 `GITLAB_TOKEN` 未设置
- **THEN** 系统通过 `res.ResolvedToken()` 获得错误，输出包含 resource 名称和环境变量名的错误信息

### Requirement: pipeline list 支持 GitLab Provider

系统 SHALL 通过 GitLab Pipelines API 获取 pipeline 数据。

#### Scenario: GitLab API 调用
- **WHEN** repo 的 Provider 为 "gitlab"
- **THEN** 系统 SHALL 调用 `GET /projects/:id/pipelines` API，`:id` 为 URL-encoded 远程路径，使用 `PRIVATE-TOKEN` header 认证

### Requirement: pipeline list 支持 GitHub Provider

系统 SHALL 通过 GitHub Actions API 获取 pipeline 数据。

#### Scenario: GitHub API 调用
- **WHEN** repo 的 Provider 为 "github"
- **THEN** 系统 SHALL 调用 `GET /repos/:owner/:repo/actions/runs` API，将 `github.com` 转换为 `api.github.com`，使用 `Bearer` token 认证

### Requirement: pipeline list 支持 Codeup Provider（云效 Flow）
系统 SHALL 通过云效 Flow OAPI v1 获取 Codeup 仓库关联的流水线运行记录。API 基础地址 SHALL 映射为 `https://openapi-rdc.aliyuncs.com/oapi/v1/flow/organizations/{orgId}`，认证 SHALL 使用 `x-yunxiao-token` 请求头。由于 Flow 流水线是组织级资源，系统 SHALL 先列出组织流水线，按流水线代码源地址与目标仓库匹配，再查询该流水线的运行记录。

#### Scenario: Codeup API 调用
- **WHEN** repo 的 Provider 为 "codeup"，且 resource 配置了 organization_id
- **THEN** 系统 SHALL 列出该组织的 Flow 流水线，匹配代码源地址指向目标仓库的流水线，查询其最近运行记录并返回

#### Scenario: 无关联流水线
- **WHEN** 组织内没有任何 Flow 流水线的代码源指向目标仓库
- **THEN** 系统 SHALL 输出 "No pipelines found for <repo-name>."

#### Scenario: 缺少 organization_id
- **WHEN** repo 的 Provider 为 "codeup"，但 resource 未配置 organization_id
- **THEN** 系统 SHALL 报错并提示在 config 的 codeup resource 中配置 `organization_id`

#### Scenario: 多条流水线匹配同一仓库
- **WHEN** 组织内有多条 Flow 流水线的代码源指向目标仓库
- **THEN** 系统 SHALL 合并各流水线的运行记录，按开始时间倒序返回前 N 条
