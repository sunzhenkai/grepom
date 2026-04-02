## MODIFIED Requirements

### Requirement: status command
系统 SHALL 提供 `grepom status` 命令，显示已克隆仓库的 git 状态概要和每个仓库的精简状态。

输出分为两部分：
1. **概要行**：统计各状态 repo 数量（clean、dirty、ahead、behind、not cloned），格式如 `12 repos: 8 clean, 2 dirty, 1 ahead, 1 behind · 3 not cloned`
2. **仓库列表**：每个 repo 一行，包含名称、状态标记、本地路径，三列对齐显示

状态标记优先级（仅显示最高优先级）：not cloned > dirty (N) > ahead N > behind N > clean

#### Scenario: Status of all repos with概要
- **WHEN** 用户运行 `grepom status`
- **THEN** 系统先输出概要行统计各状态数量，然后列出每个 repo 的名称、状态标记和本地路径

#### Scenario: Status by group
- **WHEN** 用户运行 `grepom status --group frontend`
- **THEN** 系统仅显示 group `frontend` 下的 repo，概要行也仅统计该 group 的 repo

#### Scenario: Status of not-yet-cloned repo
- **WHEN** 某 repo 未克隆
- **THEN** 系统在列表中显示该 repo，状态标记为 `not cloned`，不调用 git status

#### Scenario: 所有 repo 均为 clean
- **WHEN** 所有 repo 均 clean，无 ahead/behind
- **THEN** 概要行显示 `N repos: N clean`，无 behind/ahead 部分

#### Scenario: 无仓库
- **WHEN** 过滤后无仓库匹配
- **THEN** 系统输出 `No repositories found.`
