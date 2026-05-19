## Context

grepom 的 `pipeline watch` 子命令需要用户显式指定 repo-name 参数，而用户日常开发中通常已经 `cd` 到目标仓库目录。`mr` 命令已经实现了基于当前 git 目录的零参数自动检测，验证了"从当前目录推断仓库信息"的模式是可行且受用户欢迎的。

当前 `pipeline watch` 的代码结构将 repo 解析、pipeline 查询和 watch 循环全部耦合在 `runPipelineWatch` 函数中，不利于复用。同时，watch 输出不包含 pipeline Web URL，用户需要手动去网页查找 job 详情。

### 当前代码结构

```
cmd/pipeline.go
├── pipelineCmd (cobra parent)
├── pipelineListCmd
├── pipelineWatchCmd (Args: ExactArgs(1))
├── resolvePipelineInput()  → (provider, serverURL, remotePath, token)
├── runPipelineList()
├── runPipelineWatch()      ← 所有逻辑耦合在这里
└── formatWatchDuration()
```

### 现有自动检测参考：mr 命令

`cmd/mr.go` 中的 `detectProvider()` 展示了从 git remote 推导 provider + token 的模式：
1. 从 remote URL 提取 host
2. 先尝试与 config resources 中的 URL 匹配
3. 再 fallback 到已知公共域名（github.com / gitlab.com）
4. 最终从环境变量获取 token

## Goals / Non-Goals

**Goals:**
- `grepom watch` 作为顶级快捷命令，支持省略 repo-name 自动推断
- 自动推断失败时输出详细的诊断信息和可操作建议
- watch 输出中打印 pipeline Web URL
- 现有 `pipeline watch` 子命令同样受益于 URL 打印
- 代码复用：拆出共享的 `WatchTarget` + `runWatchLoop`，避免逻辑重复

**Non-Goals:**
- 不改变 `pipeline list` 的输出格式（后续可独立考虑添加 URL 列）
- 不实现"自动打开浏览器"功能（用户自行复制 URL）
- 不为 `pipeline list` 添加自动推断（当前只针对 watch 场景）
- 不实现 Level 3（已知公共域名）之外的更多 fallback 策略

## Decisions

### Decision 1: `watch` 为顶级命令而非 pipeline 子命令改造

**选择**: 新增顶级 `grepom watch [repo-name]` 命令
**替代方案**: 让 `pipeline watch` 的 repo-name 变为可选参数
**理由**: 
- 顶级命令更短，符合"快捷"定位
- 不破坏 `pipeline watch` 的现有 `ExactArgs(1)` 约束和向后兼容性
- `mr` 和 `pr` 已经有顶级命令的先例

### Decision 2: 三级 fallback 自动推断策略

**选择**: 
1. **Level 1** - 配置精确匹配：遍历配置中所有 repo，用 `ExtractRemotePath(CloneURL/SSHURL)` 与当前 remote 的 remotePath 比对
2. **Level 2** - Host 匹配 + Path 推导：遍历 config resources，比对 host；用 resource 的 provider + token + remote URL 的 remotePath
3. **Level 3** - 已知公共域名：github.com / gitlab.com + 环境变量 token

**替代方案**: 只做 Level 1 + 简单目录名兜底
**理由**: Level 2 是关键创新——用户可能在自托管 GitLab 上工作，配置中有 resource 但没有把每个 repo 都列出来。Level 2 能覆盖"resource 的 URL 是 `gitlab.company.com`，remote URL 也指向同一 host"的场景。Level 3 与 `mr` 命令的 `detectProvider()` fallback 一致。

### Decision 3: WatchTarget 结构体 + runWatchLoop 拆分

**选择**: 引入 `WatchTarget` 结构体，将 watch 循环逻辑抽入 `runWatchLoop()`
**结构**:
```go
type WatchTarget struct {
    Provider   cicd.PipelineProvider
    ServerURL  string
    RepoPath   string
    Token      string
    RepoName   string  // 用于显示
}
```
**理由**: `runPipelineWatch` 和 `runWatch` 需要共享 watch 循环逻辑，但入口解析逻辑不同。`WatchTarget` 是两者的统一接口。

### Decision 4: URL 打印位置

**选择**: 在 watch 开始和 pipeline 结束时各打印一次 URL
**替代方案**: 只在结束时打印 / 混在状态行里打印
**理由**: 
- 开始时打印：用户等待过程中可能想打开网页看详细日志
- 结束时打印：pipeline 失败时是最高频的网页访问需求
- 不混在 `\r` 覆盖的状态行里：终端中 URL 可被点击/选中复制，动态更新行会覆盖

### Decision 5: 错误提示模板设计

**选择**: 每种失败场景有独立模板，包含：
- 当前仓库信息（repo name、remote URL、host）
- 诊断结论（为什么推断失败）
- 可操作建议（添加配置、设置环境变量、使用完整命令）

**替代方案**: 简单的错误消息 + 通用建议
**理由**: 详细的诊断信息能大幅降低用户的困惑。用户最常见的失败场景是"在公司自托管 GitLab 上，配置中有 resource 但没有这个 repo 条目"，详细提示能直接指导他们修复。

### Decision 6: 只查 origin remote

**选择**: `git remote get-url origin`
**替代方案**: 遍历所有 remote
**理由**: 与 `mr` 命令一致，origin 是最标准的 remote 名称，简单可靠。

## Risks / Trade-offs

- **[风险] Level 2 Host 匹配可能误匹配** → 多个 resource 指向同一 host 时，取第一个匹配。这是极低概率场景，实际中同一 host 通常只有一个 resource。如果出现，用户可以通过显式传 repo-name 规避。
- **[风险] `WatchTarget` 结构体与 `resolvePipelineInput` 返回值重叠** → `resolvePipelineInput` 返回 4 个值，`WatchTarget` 是 5 个字段的 struct。长期应统一，但本次变更不改变 `resolvePipelineInput` 的签名以控制改动范围。`runWatch` 中通过 `resolveCurrentRepoPipeline` 构造 `WatchTarget`，`runPipelineWatch` 中通过 `resolvePipelineInput` 的返回值构造。
- **[权衡] 不为 pipeline list 添加自动推断** → list 是表格输出，用户通常一次性查看多个 repo，自动推断的收益不如 watch 明显。留作后续独立变更。
- **[权衡] URL 打印仅在 watch 中，不在 list 中** → list 是表格输出，加 URL 列会破坏现有列宽。后续可考虑添加 `--url` flag。
