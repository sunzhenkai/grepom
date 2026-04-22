## Context

`grepom` 是一个 Git 仓库编排管理工具，通过 YAML 配置文件管理跨 GitLab、GitHub、Codeup 等平台的多个仓库。当前 `grepom list` 命令支持通过 `--type` 标志切换列出 repos/resources/groups，支持 `--group`、`--resource`、`--all`、`--remote` 等过滤标志。

`grepom status` 命令已能获取每个仓库的 git 状态（dirty/ahead/behind/clean），但 `list` 命令无法按状态筛选仓库。用户需要组合使用 `status` + 人工查找来定位"有未推送提交"或"有未提交更改"的仓库，效率低。

当前 `list` 命令使用 Cobra 框架，位置参数仅支持仓库名称过滤，不支持 `groups` 关键字作为 `--type groups` 的快捷方式。

## Goals / Non-Goals

**Goals:**
- 提供 `grepom list groups` 快捷命令，等价于 `grepom list --type groups`
- 在 `grepom list` 中新增 `--no-push` 和 `--no-commit` 状态筛选标志
- 两个状态标志可组合使用（并集逻辑）
- 保持现有命令行为完全向后兼容

**Non-Goals:**
- 不改变 `grepom status` 命令的任何行为
- 不改变 `--remote` 模式的行为（状态筛选仅适用于本地仓库列表）
- 不增加其他 git 状态筛选（如 behind、clean）
- 不改变输出格式

## Decisions

### 1. 位置参数 `groups` 的解析策略

**决策**: 在 `RunE` 函数中，检查位置参数是否为 `groups`（或 `resources`），若是则自动设置 `listType` 为对应值。

**理由**: 这是最小侵入的改动。不需要拆分为 Cobra 子命令（那样会引入大量结构变化），只需在 `RunE` 入口处增加一层判断。同时也可以顺带支持 `resources` 作为位置参数，提升一致性。

**替代方案**: 将 `list` 拆分为 `list repos`、`list groups`、`list resources` 子命令 —— 改动过大，且与现有 `--type` 标志冲突。

### 2. 状态筛选在 cmd 层处理

**决策**: 状态筛选逻辑放在 `cmd/list.go` 的 `runListRepos` 函数中，在获取仓库列表后、输出表格前进行过滤，不修改 `repo.Filter` 和 `repo.Resolver`。

**理由**: `repo.Resolver` 是纯配置层的解析器，不应依赖 `git` 包。状态筛选需要调用 `git.GetStatus`，属于运行时操作。在 cmd 层处理保持了关注点分离：Resolver 负责配置解析，cmd 层负责运行时状态过滤。

**替代方案**: 在 `repo.Filter` 中增加状态字段，让 Resolver 负责过滤 —— 这会引入 `git` 包到 `repo` 包的循环依赖风险。

### 3. 组合使用为并集（OR）逻辑

**决策**: `--no-push` + `--no-commit` 组合时展示满足任一条件的仓库。

**理由**: 用户最常见的场景是"帮我找出所有需要关注的仓库"——不管是需要 push 还是需要 commit。并集逻辑更符合这个使用意图。交集逻辑（同时未 push 且未 commit）的场景极少。

### 4. `--remote` 模式下忽略状态标志

**决策**: 当使用 `--remote` 标志时，`--no-push` 和 `--no-commit` 不生效（因为远程仓库没有本地 git 状态）。不报错，静默忽略。

**理由**: `--remote` 查询的是远程 API 返回的仓库列表，无法获取本地 git 状态。报错会打断用户工作流，且用户很可能只是在 `--remote` 命令中附带了这个标志。

## Risks / Trade-offs

- **[性能] 状态筛选需要遍历所有仓库调用 git status** → 缓解：仅在指定了 `--no-push` 或 `--no-commit` 时才调用 `git.GetStatus`，不影响默认 `grepom list` 的性能。且 `GetStatus` 内部仅执行一次 `git status --porcelain=v2 --branch`，开销较小。

- **[兼容性] 位置参数 `groups` 可能与仓库名冲突** → 缓解：如果用户恰好有一个名为 `groups` 的仓库，此改动会导致行为变化。但实际中极少有仓库命名为 `groups`。可以在文档中说明此特殊情况，并建议使用 `--type groups` 替代。
