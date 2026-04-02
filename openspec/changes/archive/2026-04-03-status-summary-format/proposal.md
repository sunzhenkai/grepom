## Why

当前 `grepom status` 命令逐行输出每个仓库的完整路径和详细 git 状态，缺少整体概览。当仓库数量较多时，用户需要自己扫描全部输出才能了解整体状况。需要在顶部先给出汇总统计，然后每个 repo 只显示精简的三列信息（名称、状态、本地路径），提高可读性。

## What Changes

- status 命令顶部新增概要行，统计各状态的 repo 数量（clean、dirty、not cloned、ahead、behind）
- 每个 repo 的输出精简为三列：名称、状态标记、本地路径（不再显示完整 remote path 和 branch 等细节）

## Capabilities

### New Capabilities

（无）

### Modified Capabilities

- `cli-commands`: status 命令输出格式变更——新增概要统计，repo 详情精简为名称/状态/路径三列

## Impact

- 仅影响 `cmd/status.go`，不涉及 config、git、repo 等包的接口变更
