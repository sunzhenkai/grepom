## Why

grepom 当前缺乏快速创建版本 tag 的能力。用户在发布流程中需要手动执行多条 git 命令：先 `git fetch --tags` 拉取最新标签，再 `git tag -l "v*"` 查看现有版本，手动计算下一个版本号，创建 tag，最后逐个 remote 推送。这个过程繁琐且容易出错——特别是在 tag 已存在需要递增、仓库有多个 remote 需要逐一推送的场景下。

此外，团队在开发过程中需要一个独立的测试版本（t 前缀）来区分正式版本和临时测试版本。t 版本的前三位跟随 v 版本计算，第四位独立递增，用于同一版本下的多次迭代测试。

## What Changes

- **新增 `grepom tag` 子命令**：在当前 git 仓库中自动计算并创建下一个版本 tag
- **双版本格式支持**：
  - `v` 版本（默认）：格式 `vMAJOR.MINOR.PATCH`，如 `v0.1.2`
  - `t` 版本（`-t` 标志）：格式 `tMAJOR.MINOR.PATCH.ITER`，如 `t0.1.2.3`
- **自动版本计算**：基于按时间排序的最新 `v*` tag，自动递增版本号
  - 版本号解析容错：多于 3 位截取前 3 位，少于 3 位补齐到 3 位
  - PATCH 范围 0~99，溢出时自动进位到 MINOR（MINOR 无上限）
  - tag 冲突时自动递增寻找下一个可用版本号
  - 无任何 v tag 存在时默认创建 `v0.0.1`
- **灵活的推送控制**：
  - 默认仅创建本地 tag，交互式询问是否推送到远程
  - `-p` / `--push` 标志直接推送到所有远程仓库
  - 非 TTY 环境跳过询问，提示用户使用 `-p`
- **预览模式**：`--dry-run` 仅显示将创建的 tag，不实际执行
- **附注 tag 支持**：`-m` / `--message` 创建带消息的附注 tag

## Capabilities

### New Capabilities

- `tag-command`: 版本标签管理能力——自动计算下一个 v/t 版本号，创建本地 tag，可选推送到所有远程仓库

### Modified Capabilities

（无已有能力需要修改）

## Impact

- **新增文件**：`cmd/tag.go`（CLI 入口）、`git/tag.go`（tag 操作封装）
- **已有文件**：`git/git.go` 可能新增少量 tag 相关辅助函数
- **新增依赖**：无（使用现有 Go 标准库 + cobra）
- **无 breaking change**：纯新增功能，不影响已有命令行为
- **不依赖 .grepom.yml**：类似 `push` 命令，直接操作当前目录的 git 仓库
