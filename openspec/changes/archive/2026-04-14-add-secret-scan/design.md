## Context

grepom 是一个用 Go 编写的 git 仓库编排管理 CLI 工具，基于 cobra 框架。它通过 YAML 配置文件管理分布在不同 GitLab groups 和 GitHub orgs 下的多个仓库。当前已有 `list`、`clone`、`pull`、`status`、`search`、`sync` 等命令。

用户管理的仓库集合包含公司项目、个人隐私项目和个人开源项目，存在意外泄露敏感信息的风险。需要新增 `scan` 命令，利用成熟的 gitleaks 规则引擎对仓库进行私密信息扫描。

gitleaks v8 提供了可编程的 Go API：
- `detect.NewDetectorDefaultConfig()` 创建默认规则检测器
- `sources.Files` 扫描文件目录
- `sources.Git` 扫描 git 历史
- `detector.DetectSource(ctx, source)` 执行扫描并返回 `[]report.Finding`

当前 grepom 的 `repo.Resolver` 可解析出所有 repo 的本地路径列表，`git.IsCloned()` 可判断是否已克隆。这些基础设施将直接复用。

## Goals / Non-Goals

**Goals:**

- 新增 `grepom scan` 命令，支持按 group/repo 粒度扫描已克隆仓库中的敏感信息
- 集成 gitleaks v8 detect/sources 包作为扫描引擎，利用其 100+ 内置检测规则
- 支持两种扫描模式：工作区文件扫描（默认）和 git 历史扫描（`--history`）
- 以终端表格形式输出扫描结果，按严重程度统计
- 支持 `.gitleaksignore` 白名单机制（gitleaks 原生支持）
- 扫描时自动感知 `.gitignore`，跳过被忽略的文件
- 并行扫描多个 repo，利用已有的并行框架
- 支持 `--format json` 输出 JSON 格式，便于管道和自动化

**Non-Goals:**

- 不实现自定义规则编辑（用户可通过 gitleaks.toml 配置文件实现）
- 不实现自动修复/清除泄露（仅检测和报告）
- 不实现 CI/CD 集成（grepom 定位为本地开发工具）
- 不实现远程仓库扫描（不 clone 到临时目录再扫，只扫描本地已 clone 的）
- 不实现 pre-commit hook 集成
- 不实现扫描结果持久化/趋势分析

## Decisions

### Decision 1: 采用 gitleaks Go library 而非 CLI wrapper 或自写引擎

**选择**: 直接 import `github.com/zricethezav/gitleaks/v8` 的 `detect`、`sources`、`config`、`report` 包

**理由**:
- gitleaks 有 100+ 内置检测规则（SSH 私钥、AK/SK、GitHub/GitLab token、数据库连接串、kubeconfig 等），覆盖面远超自写
- Go library 集成避免外部二进制依赖，用户体验更好（一个 grepom 二进制搞定）
- gitleaks 的 `detect.Detector` 提供了干净的 programmatic API：`NewDetectorDefaultConfig()` + `DetectSource()`
- gitleaks 原生支持 `.gitleaksignore`，无需自建白名单系统

**代价**:
- go.mod 依赖树膨胀（aho-corasick、semgroup、zerolog、viper、lipgloss、go-gitdiff 等）
- 编译时间增加
- 耦合 gitleaks 的 API 稳定性

**备选方案**:
- **CLI wrapper**: 需要用户预装 gitleaks 或自动下载，增加运维复杂度
- **自写 regex scanner**: 零依赖但规则维护成本高，覆盖面不足

### Decision 2: scanner 包作为 gitleaks 的薄封装层

**选择**: 新建 `scanner/` 包，封装 gitleaks API 调用，提供 grepom 友好的接口

**理由**:
- 解耦 CLI 命令和 gitleaks API，降低替换扫描引擎的成本
- 集中处理 gitleaks 配置加载、`.gitignore` 感知、结果格式化等逻辑
- 如果未来 gitleaks API 发生 breaking change，只需修改 `scanner/` 包

**接口设计**:
```
scanner.Scanner
├── ScanDir(ctx, repoPath) → []scanner.Finding      // 工作区扫描
├── ScanGitHistory(ctx, repoPath) → []scanner.Finding // 历史扫描
└── 内部使用 gitleaks detect.Detector + sources.Files/Git
```

### Decision 3: .gitignore 感知通过 gitleaks 的 allowlist 机制实现

**选择**: 读取 `.gitignore` 文件内容，构建路径正则表达式，注入到 gitleaks 的全局 allowlist 中

**理由**:
- gitleaks 的 `config.Allowlist` 原生支持路径排除（`Paths` 字段为 `[]*regexp.Regexp`）
- 无需修改 gitleaks 源码
- 性能友好，allowlist 在 prefilter 阶段就跳过匹配路径

### Decision 4: 并行扫描复用 grepom 已有模式

**选择**: 在 `cmd/scan.go` 中使用 goroutine + sync.WaitGroup 并行扫描多个 repo，与 `cmd/pull.go`、`cmd/clone.go` 一致的模式

**理由**:
- grepom 已有 `git/parallel.go` 可参考
- 无需引入新的并发框架
- 与已有命令的代码风格一致

### Decision 5: 输出格式

**选择**: 默认终端表格 + 可选 JSON

**理由**:
- 表格输出与 grepom 已有命令（`list`、`search`、`status`）风格一致，使用 `text/tabwriter`
- JSON 输出满足自动化场景（管道到 jq、写入文件等）
- 不实现 SARIF 等复杂格式，保持简洁

## Risks / Trade-offs

**[依赖膨胀]** → gitleaks v8 引入大量传递依赖。缓解：隔离在 `scanner/` 包中，如果未来不可接受可切换到 CLI wrapper 方案。

**[gitleaks API 不稳定]** → gitleaks 未承诺 library API 稳定性。缓解：`scanner/` 包作为薄封装层，限制 API 变更影响范围。

**[大量 repo 时的扫描耗时]** → 50+ repo 全量扫描可能耗时数分钟。缓解：并行扫描 + 默认只扫工作区（不扫历史）+ 进度指示器。

**[git 历史扫描的资源消耗]** → `git log -p` 对大仓库消耗大量内存和 CPU。缓解：默认不开启，需要 `--history` 显式启用。

**[误报率]** → gitleaks 的通用规则可能产生误报（测试文件中的 mock token 等）。缓解：支持 `.gitleaksignore` 和自定义 gitleaks.toml 配置。
