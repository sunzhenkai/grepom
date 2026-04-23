## Context

当前 `grepom scan` 命令（`cmd/scan.go`）的配置查找依赖于 `config.FindConfig()`，该函数会从当前目录沿父目录链向上查找 `.grepom.yml`。这在实际使用中导致两个问题：

1. 用户在子目录（如仓库根目录内）运行 scan 时，可能匹配到祖先目录中属于另一个工作区的配置文件，导致路径推算错误（所有 repo 报 "not cloned"）。
2. 无配置文件时只能扫描当前目录（`runScanCurrentDir()`），无法指定其他路径。

当前的代码结构：
- `runScan()` 先调用 `loadConfig()`（使用 `FindConfig` 向上查找）
- 找到配置 → 走 resolver 路径
- 未找到 → `runScanCurrentDir()` 扫描 `.`
- `[name]` 位置参数用于按名称过滤 resolver 结果

## Goals / Non-Goals

**Goals:**
- 新增 `-p/--path` 标志，允许指定任意目录路径进行扫描
- 简化配置查找：scan 仅在当前目录查找 `.grepom.yml`，不再向上遍历
- 扫描开始时打印目标摘要（路径或仓库名列表）

**Non-Goals:**
- 不改变其他命令（clone、pull、status 等）的配置查找行为
- 不实现递归发现子目录中的 git 仓库
- 不改变扫描引擎（gitleaks）的行为
- `-p` 指定时不支持 `[name]` 过滤（二者互斥）

## Decisions

### Decision 1: `-p` 指定时完全忽略配置文件

**选择**：当用户指定 `-p <path>` 时，不调用 `loadConfig()`，直接将 `<path>` 作为扫描目标。

**理由**：`-p` 的语义是"我要扫描这个目录"，与配置文件中的仓库管理无关。混合使用会导致混乱（配置中的 base 路径、过滤逻辑都不适用）。

**替代方案**：`-p` 作为配置仓库的路径前缀覆盖 → 过于复杂，违背 `-p` 的直觉语义。

### Decision 2: 配置查找仅限当前目录

**选择**：`grepom scan`（无 `-p`）只检查 `os.Stat(".grepom.yml")`，不调用 `FindConfig()` 向上遍历。

**理由**：scan 的核心是"扫描我的代码"，用户应该在项目根目录（有 .grepom.yml）运行它。向上查找会导致在子目录误触发上级工作区的配置。其他命令（clone、pull）保留向上查找，因为它们是"管理我的仓库集合"，适合从一个工作区的任意子目录操作。

**实现方式**：在 `cmd/scan.go` 内直接用 `os.Stat(".grepom.yml")` 检查，而非调用全局的 `loadConfig()`。不影响 `root.go` 中的 `loadConfig()` 函数本身。

### Decision 3: `-p` 与 `[name]` 互斥

**选择**：`-p` 指定时 `[name]` 位置参数被忽略（不报错，静默忽略）。

**理由**：避免引入复杂的互斥校验逻辑。实际使用中 `-p` 指定一个目录时不存在"按名称过滤"的概念。

### Decision 4: 扫描目标摘要格式

**选择**：
- 路径模式（`-p` 或无配置）：`Scanning /path/to/dir...`
- 配置模式（仓库列表 ≤5）：逐行打印 `Scanning repo1, repo2, repo3...`
- 配置模式（仓库列表 >5）：`Scanning repo1, repo2, repo3, ...and N more`

**理由**：用户需要确认扫描范围是否正确，尤其是在配置模式下。超过 5 个时截断避免刷屏。

## Risks / Trade-offs

- **[行为变更] 向上查找被移除** → 某些用户可能依赖在子目录运行 `grepom scan` 匹配上级配置。但这恰恰是当前 bug 的根源，且正确的使用方式是在配置文件所在目录运行。风险可控。
- **[兼容性] `-p` 是新 flag** → 不影响现有用法，纯增量功能。无破坏性。
- **[静默忽略] `-p` + `[name]` 不报错** → 用户可能困惑为什么 `grepom scan -p /tmp myrepo` 忽略了 myrepo。可在 stderr 打印提示信息缓解。
