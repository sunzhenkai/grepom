## Context

当前 `cmd/status.go` 遍历所有 repo，对每个已克隆的 repo 输出 `remote/path: branch, clean/dirty (N files), ahead M, behind K`，对未克隆的输出 `remote/path: not cloned`。输出使用 `r.Path`（远程路径）作为标识。

## Goals / Non-Goals

**Goals:**
- 输出顶部概要：统计 clean / dirty / not cloned / ahead / behind 的数量
- 每个 repo 精简为三列：名称（`r.Name`）、状态标记、本地路径（`fullPath`）
- 状态标记直观可读，一眼看出 repo 状况

**Non-Goals:**
- 不改变 status 命令的参数和过滤逻辑（`--group`、`--resource`、位置参数）
- 不改变 `git.Status` 数据结构

## Decisions

### Decision 1: 输出格式

选择紧凑的单行格式，概要 + repo 列表。示例：

```
12 repos: 8 clean, 2 dirty, 1 ahead, 1 behind · 3 not cloned

  design-system   clean        ~/projects/frontend/design-system
  web-app         dirty (3)    ~/projects/frontend/web-app
  api-server      ahead 2      ~/projects/frontend/api-server
  dotfiles        behind 1     ~/projects/dotfiles
  new-service     not cloned   ~/projects/backend/new-service
```

- 概要行：一行汇总所有数量
- repo 列表：固定列宽对齐，状态用语义化文字（clean / dirty (N) / ahead N / behind N / not cloned）
- 未克隆的 repo 不调用 `git status`，直接标记 `not cloned`

### Decision 2: 状态标记逻辑

当一个 repo 同时有多个状态时（如 dirty + ahead），显示优先级为：not cloned > dirty > ahead > behind > clean。只显示最高优先级的一个状态标记，保持简洁。

## Risks / Trade-offs

- **[信息精简]** 不再显示 branch 名称 → 如果用户需要 branch 信息，仍可通过 `git status` 手动查看；status 命令定位为快速概览
- **[列宽对齐]** 使用 `fmt` 对齐 → 中文状态文字宽度计算需注意，但当前状态标记均为英文，无此问题
