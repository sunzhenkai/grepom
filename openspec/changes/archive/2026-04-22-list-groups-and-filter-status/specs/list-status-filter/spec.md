## ADDED Requirements

### Requirement: list 按未推送状态筛选仓库
`grepom list` 命令 SHALL 支持 `--no-push` 标志。当使用 `--no-push` 时，系统 SHALL 仅展示有未推送提交（ahead > 0）的已克隆仓库。未克隆的仓库 SHALL 不在结果中展示。

#### Scenario: 使用 --no-push 筛选未推送仓库
- **WHEN** 用户运行 `grepom list --no-push`
- **THEN** 系统获取所有已克隆仓库的 git 状态，仅展示 ahead > 0 的仓库，输出格式与默认 list 一致（NAME、PATH、GROUP、RESOURCE、CLONED 列）

#### Scenario: --no-push 与 --group 组合使用
- **WHEN** 用户运行 `grepom list --no-push --group frontend`
- **THEN** 系统仅在 group `frontend` 范围内筛选，展示 ahead > 0 的仓库

#### Scenario: --no-push 与 --resource 组合使用
- **WHEN** 用户运行 `grepom list --no-push --resource work-gl`
- **THEN** 系统仅在引用 resource `work-gl` 的仓库中筛选，展示 ahead > 0 的仓库

#### Scenario: --no-push 无匹配仓库
- **WHEN** 用户运行 `grepom list --no-push` 且所有已克隆仓库均无未推送提交
- **THEN** 系统输出 `No repositories found.`

#### Scenario: --no-push 所有仓库均未克隆
- **WHEN** 用户运行 `grepom list --no-push` 且所有仓库均未克隆
- **THEN** 系统输出 `No repositories found.`

### Requirement: list 按未提交状态筛选仓库
`grepom list` 命令 SHALL 支持 `--no-commit` 标志。当使用 `--no-commit` 时，系统 SHALL 仅展示有未提交更改（dirty > 0）的已克隆仓库。未克隆的仓库 SHALL 不在结果中展示。

#### Scenario: 使用 --no-commit 筛选未提交仓库
- **WHEN** 用户运行 `grepom list --no-commit`
- **THEN** 系统获取所有已克隆仓库的 git 状态，仅展示 dirty > 0 的仓库，输出格式与默认 list 一致

#### Scenario: --no-commit 与 --group 组合使用
- **WHEN** 用户运行 `grepom list --no-commit --group frontend`
- **THEN** 系统仅在 group `frontend` 范围内筛选，展示 dirty > 0 的仓库

#### Scenario: --no-commit 无匹配仓库
- **WHEN** 用户运行 `grepom list --no-commit` 且所有已克隆仓库均为 clean
- **THEN** 系统输出 `No repositories found.`

### Requirement: --no-push 与 --no-commit 组合使用
`--no-push` 和 `--no-commit` SHALL 支持组合使用。组合时 SHALL 展示满足任一条件（ahead > 0 或 dirty > 0）的仓库（并集逻辑）。

#### Scenario: --no-push 与 --no-commit 组合筛选
- **WHEN** 用户运行 `grepom list --no-push --no-commit`
- **THEN** 系统展示 ahead > 0 或 dirty > 0 的仓库（并集）

#### Scenario: --no-push --no-commit --group 组合
- **WHEN** 用户运行 `grepom list --no-push --no-commit --group backend`
- **THEN** 系统在 group `backend` 范围内展示 ahead > 0 或 dirty > 0 的仓库

### Requirement: --remote 模式下忽略状态筛选标志
当使用 `--remote` 标志时，`--no-push` 和 `--no-commit` SHALL 不生效（静默忽略），系统 SHALL 正常执行远程仓库列表查询。

#### Scenario: --remote --no-push 组合
- **WHEN** 用户运行 `grepom list --remote --no-push`
- **THEN** 系统正常执行远程仓库列表查询，忽略 `--no-push`，输出与 `grepom list --remote` 一致

#### Scenario: --remote --no-commit 组合
- **WHEN** 用户运行 `grepom list --remote --no-commit`
- **THEN** 系统正常执行远程仓库列表查询，忽略 `--no-commit`，输出与 `grepom list --remote` 一致
