## Why

每次查看 CI/CD pipeline 状态都需要输入完整命令 `grepom pipeline watch <repo-name>`，而在日常开发中，用户通常已经 `cd` 到目标仓库目录，手动输入 repo-name 是冗余操作。`mr` 命令已经实现了基于当前 git 目录的零参数自动检测，`pipeline watch` 应该享有同样的便利。同时，pipeline 结束后用户常需要打开网页查看 job 详情，但当前 watch 输出不包含 URL，需要手动去网页查找。

## What Changes

- 新增顶级快捷命令 `grepom watch [repo-name]`，等价于 `grepom pipeline watch <repo-name>`，当省略 repo-name 时自动从当前 git 仓库推断
- 自动推断采用三级 fallback 策略：配置精确匹配 → Host 匹配 + Path 推导 → 已知公共域名环境变量
- 自动推断失败时输出详细的诊断信息（当前仓库、远程地址、主机名）和可操作的建议（添加配置、设置环境变量）
- watch 输出中新增 pipeline Web URL 打印（开始时和结束时各一次），方便用户直接打开浏览器查看 job 详情
- 现有 `pipeline watch` 子命令同样受益于 URL 打印改进（复用同一 `runWatchLoop` 函数）

## Capabilities

### New Capabilities
- `watch-shortcut`: 顶级 `watch` 快捷命令，支持 repo-name 自动推断、显式指定、`--id` 标志
- `pipeline-url-display`: watch 输出中打印 pipeline Web URL

### Modified Capabilities
- `pipeline-watch`: 重构为 `WatchTarget` + `runWatchLoop` 架构，使 watch 循环逻辑可被顶级 `watch` 命令复用；增加 URL 打印输出

## Impact

- **代码**: `cmd/pipeline.go`（重构 + 新增 URL 打印）、新增 `cmd/watch.go`（顶级命令 + 自动推断逻辑）
- **文档**: README.md、README_en.md 需要同步更新，添加 `watch` 命令说明
- **向后兼容**: 完全兼容，`pipeline watch` 子命令行为不变（仅增加 URL 打印）
