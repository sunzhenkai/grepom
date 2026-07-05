### Requirement: 删除中代码库的识别
系统 SHALL 通过辅助函数 `isDeletionScheduled(name, pathWithNamespace string) bool` 识别处于"计划删除"状态的代码库：当 `name` 或 `pathWithNamespace` 包含子串 `deletion_scheduled` 时返回 `true`。该检测 SHALL 被 provider 发现阶段与 resolver 运行时兜底共用，保证行为一致。

#### Scenario: name 含 deletion_scheduled 标记
- **WHEN** 调用 `isDeletionScheduled("creative-matching-deletion_scheduled-499", "dsp/creative-matching-deletion_scheduled-499")`
- **THEN** 返回 `true`

#### Scenario: 仅 pathWithNamespace 含标记（组被删除）
- **WHEN** 调用 `isDeletionScheduled("some-repo", "dsp-services-deletion_scheduled-452/some-repo")`
- **THEN** 返回 `true`

#### Scenario: 正常代码库
- **WHEN** 调用 `isDeletionScheduled("grepom", "wii/solo/grepom")`
- **THEN** 返回 `false`

### Requirement: 发现阶段默认剔除删除中代码库
所有 provider 的 `ListRepos` SHALL 在 `ListReposParams.IncludeDeleted` 为 `false`（默认）时，从返回结果中剔除被识别为删除中的代码库。`ListReposParams` SHALL 新增 `IncludeDeleted bool` 字段。

#### Scenario: 默认剔除删除中代码库
- **WHEN** Codeup `ListRepos` 发现 3 个代码库，其中 1 个 `pathWithNamespace` 含 `deletion_scheduled`，且未设置 `IncludeDeleted`
- **THEN** 返回结果仅包含 2 个正常代码库

#### Scenario: IncludeDeleted 保留删除中代码库
- **WHEN** `ListReposParams.IncludeDeleted` 为 `true`
- **THEN** 返回结果包含全部 3 个代码库（含删除中者）

#### Scenario: 非 Codeup provider 不受影响
- **WHEN** GitLab/GitHub provider 调用 `ListRepos`
- **THEN** 仅依据各自既有逻辑返回结果，不因本字段改变行为（这些 provider 的代码库命名不含 `deletion_scheduled` 标记，检测函数对它们恒为 `false`）

### Requirement: 运行时兜底拦截删除中代码库
`repo/resolver.go` 的 `resolveInternal` SHALL 对配置中已存在但被 `isDeletionScheduled` 命中的代码库，设置 `DisabledReason = "deletion_scheduled"`。该 reason 与 `"disabled"`/`"excluded"` 同等对待：`Resolve()` 默认剔除，`ResolveAndFilter` 在 `Filter.IncludeDisabled=true` 时保留。

#### Scenario: 配置中存在删除中代码库时默认跳过
- **WHEN** 配置中某 group repo 的 `Path` 含 `deletion_scheduled`，调用 `Resolve()`
- **THEN** 该 repo 被剔除，不出现在返回列表中

#### Scenario: --all 展示删除中代码库并标注
- **WHEN** 配置中存在删除中代码库，`grepom list --all`（即 `Filter.IncludeDisabled=true`）
- **THEN** 该 repo 出现在列表中，`DisabledReason` 为 `"deletion_scheduled"`

#### Scenario: clone 跳过删除中代码库不报鉴权错误
- **WHEN** 配置中存在删除中代码库，运行 `grepom clone`
- **THEN** 该 repo 被跳过，不触发 `git clone`，不产生 `all authentication methods failed` 错误

### Requirement: sync 与 list 的 --include-deleted 标志
`grepom sync` 与 `grepom list` SHALL 支持 `--include-deleted` 布尔标志，设为 `true` 时将 `ListReposParams.IncludeDeleted` 置 `true`，使发现结果保留删除中代码库。

#### Scenario: sync 默认不写入删除中代码库
- **WHEN** 运行 `grepom sync`（不带 `--include-deleted`）
- **THEN** 被识别为删除中的代码库不会被写入 `.grepom.yml`

#### Scenario: sync --include-deleted 写入删除中代码库
- **WHEN** 运行 `grepom sync --include-deleted`
- **THEN** 删除中的代码库被写入 `.grepom.yml`

### Requirement: 跳过计数提示
`grepom sync` 在 verbose 模式下 SHALL 输出本次发现阶段被跳过的删除中代码库数量。

#### Scenario: verbose 输出跳过数量
- **WHEN** 运行 `grepom sync -v`，发现阶段跳过了 3 个删除中代码库
- **THEN** 输出包含跳过数量（如 `skipped 3 deletion_scheduled repos`）
