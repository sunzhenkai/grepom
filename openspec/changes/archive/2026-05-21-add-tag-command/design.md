## Context

grepom 是一个用 Go 编写的 git 仓库编排管理 CLI 工具，基于 cobra 框架。当前已有 `list`、`clone`、`pull`、`push`、`status`、`scan`、`search`、`sync`、`watch` 等命令。

其中 `push` 命令的设计模式与 `tag` 命令高度一致：
- 不依赖 `.grepom.yml` 配置文件
- 直接在当前目录的 git 仓库操作
- 使用 `gitpkg.IsCloned(".")` 前置检查

`git/git.go` 中已有 `Pull()`、`Push()`、`PushBranch()`、`GetCurrentBranch()`、`GetRemoteURL()` 等函数，但缺少 tag 相关操作。

项目使用 `github.com/AlecAivazis/survey/v2` 进行交互式提示（见 `cmd/interactive.go`），`cmd/interactive.go` 中有 `isTerminal()` 函数可用于 TTY 检测。

## Goals / Non-Goals

**Goals:**

- 新增 `grepom tag` 子命令，在当前仓库自动创建版本 tag
- 支持 `v` 版本（默认）和 `t` 版本（`-t` 标志）
- 自动从所有 remote 拉取最新 tags，按时间排序查找最新 v* tag
- 版本号解析容错（截取/补齐到 3 位），PATCH 溢出（>99）自动进位
- tag 冲突时自动递增寻找可用版本号
- 无 v tag 时默认创建 `v0.0.1`
- 默认创建本地 tag，交互询问是否推送；`-p` 直接推送；非 TTY 跳过询问
- `--dry-run` 预览模式
- `-m` 创建附注 tag

**Non-Goals:**

- 不实现多仓库批量打 tag（tag 命令仅操作当前目录）
- 不实现删除 tag
- 不实现 tag 列表/查询（用户可直接用 `git tag -l`）
- 不实现从 `.grepom.yml` 配置读取信息
- 不实现 changelog 生成

## Design

### 命令行接口

```
Usage:
  grepom tag          创建下一个 v 版本 tag
  grepom tag -t       创建 t 版本 tag（测试版）

Flags:
  -t, --test          创建 t 前缀版本
  -p, --push          创建后直接推送到所有 remote
      --dry-run       预览，不实际创建
  -m, --message       附注 tag 的消息（默认创建轻量 tag）

Examples:
  grepom tag                      # v0.1.5 → v0.1.6
  grepom tag -t                   # t0.1.6.0（首个 t 版本）
  grepom tag -t                   # t0.1.6.1（递增第 4 位）
  grepom tag -p                   # 创建 v 并推送到所有 remote
  grepom tag -t -p                # 创建 t 并推送
  grepom tag --dry-run            # 只看结果
  grepom tag -m "release v0.2.0"  # 附注 tag
```

### 核心流程

```
1. 前置检查：当前目录是 git 仓库？
2. git fetch --tags --all（从所有 remote 拉取最新 tags）
3. git tag --sort=-creatordate --list "v*"（按时间排序）
4. 取时间最新的 v* tag → 解析版本号
   - 无 v tag → baseVersion = (0, 0, 0)
   - 有 v tag → 解析数字部分，截取/补齐到 3 位
5. 计算下一个版本：PATCH + 1，PATCH > 99 则 MINOR + 1, PATCH = 0
6. 无 -t → 生成 v{MAJOR}.{MINOR}.{PATCH}
   有 -t → 前 3 位同 v 计算，查找 t{M}.{m}.{p}.* 按时间最新取第 4 位 + 1（无则从 0 起）
7. 冲突检测：tag 已存在 → 自动递增继续寻找
8. --dry-run → 仅输出预览信息
9. 创建 tag（轻量或附注）
10. 推送处理：
    - 有 -p → 遍历所有 remote 逐一推送
    - 无 -p + 有 TTY → 询问用户是否推送
    - 无 -p + 无 TTY → 提示使用 -p
```

### 版本号解析规则

| 输入 tag | 截取/补齐 | 解析结果 |
|----------|-----------|----------|
| `v0.0.1` | `v0.0.1` | `(0, 0, 1)` |
| `v0.1` | `v0.1.0` | `(0, 1, 0)` |
| `v1` | `v1.0.0` | `(1, 0, 0)` |
| `v0.1.2.3` | `v0.1.2` | `(0, 1, 2)` |
| `v2.3.4.5.6` | `v2.3.4` | `(2, 3, 4)` |

算法：
1. 去掉前缀（`v` 或 `t`），按 `.` 分割
2. 每段转为 int，忽略非数字段
3. 多于目标位数（v=3, t=4）→ 截取前 N 位
4. 少于目标位数 → 末尾补 0

### 版本递进规则

```
PATCH 范围: 0 ~ 99（99 上限仅对 PATCH/第 3 位生效）
MINOR 范围: 无上限

递进示例:
  (0, 0, 1)  → v0.0.2
  (0, 0, 98) → v0.0.99
  (0, 0, 99) → v0.1.0     ← PATCH 溢出, MINOR + 1
  (0, 1, 99) → v0.2.0
  (3, 99, 99)→ v4.0.0
  (无 tag)   → v0.0.1     ← 首个 tag
```

### t 版本第 4 位规则

```
t 版本的前 3 位始终跟随 v 版本的计算结果
第 4 位独立递增，无上限

查找规则:
  1. 计算出新的前 3 位（如 0.1.2）
  2. 查找 t0.1.2.* 中按时间最新的 tag
  3. 取其第 4 位 + 1
  4. 找不到匹配的 t tag → 从 0 开始

示例（最新 v tag = v0.1.1）:
  grepom tag     → v0.1.2
  grepom tag -t  → t0.1.2.0
  grepom tag -t  → t0.1.2.1
  grepom tag -t  → t0.1.2.2

  v 更新后（最新 v tag = v0.1.3）:
  grepom tag -t  → t0.1.3.0     ← 前 3 位变了，第 4 位重置
```

### 冲突自动解决

```
v tag 冲突:
  计算得 v0.1.6 → 已存在 → 尝试 v0.1.7 → 已存在 → v0.1.8 ✓
  不断 PATCH + 1（带 99 进位），直到找到空位

t tag 冲突:
  计算得 t0.2.3.6 → 已存在 → 尝试 t0.2.3.7 ✓
  第 4 位不断 + 1，无上限

安全上限: 最大循环 10000 次，超出报错
```

### 推送交互设计

```
场景 1: grepom tag（无 -p，有 TTY）
  Latest v tag: v0.1.5
  New tag:      v0.1.6
  ✓ created locally

  Push to all remotes? [y/N]: y
    ✓ origin   pushed
    ✓ mirror   pushed

场景 2: grepom tag -p（有 -p，跳过询问）
  Latest v tag: v0.1.5
  New tag:      v0.1.6
  ✓ created locally
  Pushing to all remotes...
    ✓ origin   pushed
    ✓ mirror   pushed

场景 3: grepom tag（无 -p，无 TTY）
  Latest v tag: v0.1.5
  New tag:      v0.1.6
  ✓ created locally
  (no TTY, skipping push prompt. Use -p to push.)

场景 4: grepom tag --dry-run
  [dry-run] Latest v tag: v0.1.5
  [dry-run] New tag:      v0.1.6
  [dry-run] Would create tag v0.1.6 locally
```

## Decisions

### Decision 1: tag 命令不依赖 .grepom.yml

**选择**: 类似 `push` 命令，`tag` 直接操作当前目录的 git 仓库，不需要 grepom 配置文件。

**理由**:
- tag 操作是单仓库行为，与 grepom 的多仓库编排职责正交
- 用户可能在任意 git 仓库中使用此功能
- 与 `push` 命令的模式一致

### Decision 2: 按时间排序而非版本号排序

**选择**: 使用 `git tag --sort=-creatordate` 按创建时间排序，取最新的 v* tag。

**理由**:
- 版本号排序在非标准格式 tag 场景下可能出错
- 时间排序更符合"最后一次操作"的直觉
- 用户可能在不连续的版本号上工作（如 hotfix 分支）

### Decision 3: 冲突时自动递增而非报错

**选择**: 计算出的 tag 已存在时，自动递增继续寻找，不中断用户流程。

**理由**:
- tag 冲突在多人协作或 hotfix 场景下常见
- 自动解决比手动重试更友好
- 设置最大循环次数（10000）防止无限循环

### Decision 4: fetch 从所有 remote 拉取

**选择**: 使用 `git fetch --tags --all` 从所有 remote 拉取最新 tags。

**理由**:
- 仓库可能有多个 remote（origin、mirror 等）
- 确保获取到最完整的 tag 集合用于版本计算
- 与 `-p` 推送到所有 remote 的行为对称

### Decision 5: 使用 survey 库进行推送确认

**选择**: 复用 `github.com/AlecAivazis/survey/v2` 的 `Confirm` 组件进行交互式确认。

**理由**:
- 项目已有此依赖（`cmd/interactive.go`）
- 一致的交互风格
- 自动处理 TTY/非 TTY 场景

## Risks / Trade-offs

**[fetch --all 耗时]** → 仓库有大量 remote 或网络慢时 fetch 耗时较长。缓解：显示 fetching 状态提示。

**[时间排序可能不等于版本递增]** → 如果用户在不同分支上打了多个 v tag，时间最新的不一定是版本号最大的。但这是用户明确选择的行为。

**[大量冲突循环]** → 极端情况下 tag 密集（如 CI 自动打 tag），冲突循环次数多。缓解：设置 10000 次上限。

**[轻量 tag 与附注 tag 的 creatordate 差异]** → 轻量 tag 的 creatordate 是指向的 commit 日期，附注 tag 是 tagger 日期。可能导致排序不直观。这是 git 本身的行为，不额外处理。
