## 1. 依赖引入与项目结构

- [x] 1.1 将 gitleaks v8 添加到 go.mod：`go get github.com/zricethezav/gitleaks/v8`，确认编译通过
- [x] 1.2 创建 `scanner/` 包目录结构：`scanner/scanner.go`（核心引擎）、`scanner/finding.go`（结果类型）

## 2. scanner 包核心实现

- [x] 2.1 在 `scanner/finding.go` 中定义 `Finding` 结构体（Repo、File、Line、RuleID、Description、Secret、Severity 字段）和 `MaskSecret(secret string) string` 脱敏函数
- [x] 2.2 在 `scanner/scanner.go` 中实现 `Scanner` 结构体，包含 `NewScanner(opts Options)` 构造函数，Options 包含 GitleaksConfigPath、MaxTargetMegaBytes 等配置
- [x] 2.3 实现 `Scanner.initDetector()` 方法：加载 gitleaks 配置（自定义 toml 或默认规则），创建 `detect.Detector` 实例，配置 Redact 脱敏级别
- [x] 2.4 实现 `Scanner.ScanDir(ctx, repoPath) ([]Finding, error)` 方法：使用 `sources.Files` 扫描工作区，调用 `detector.DetectSource()`，将 gitleaks `report.Finding` 转换为 grepom `scanner.Finding`
- [x] 2.5 实现 `Scanner.ScanGitHistory(ctx, repoPath) ([]Finding, error)` 方法：使用 `sources.Git`（通过 `NewGitLogCmd`）扫描 git 历史，同样转换结果
- [x] 2.6 实现 `.gitignore` 感知：读取 repo 根目录 `.gitignore`，将匹配路径构建为 gitleaks `config.Allowlist.Paths` 正则，注入到全局 allowlist
- [x] 2.7 实现 `.gitleaksignore` 支持：调用 `detector.AddGitleaksIgnore(path)` 加载仓库根目录的 ignore 文件

## 3. cmd/scan.go CLI 命令

- [x] 3.1 创建 `cmd/scan.go`，定义 `scanCmd` cobra.Command（Use: "scan [name]"），注册 `--group`、`--resource`、`--history`、`--format`、`--gitleaks-config` 标志
- [x] 3.2 实现 RunE 主逻辑：加载 config → resolver 解析 repos → 过滤已克隆的 repo → 创建 Scanner → 按模式调用 ScanDir/ScanGitHistory → 收集结果
- [x] 3.3 实现并行扫描：使用 goroutine + sync.WaitGroup 并发扫描多个 repo，带进度输出（"Scanning... N/M repos"）
- [x] 3.4 实现终端表格输出：使用 `text/tabwriter` 输出 REPO/FILE/LINE/RULE/SEVERITY 列，末尾输出汇总统计
- [x] 3.5 实现 JSON 输出：`--format json` 时将 []Finding 序列化为 JSON 数组输出到 stdout
- [x] 3.6 在 `cmd/root.go` 中注册 `scanCmd`（`rootCmd.AddCommand(scanCmd)`）

## 4. 测试与验证

- [x] 4.1 为 `scanner/finding.go` 编写单元测试：验证 Finding 结构体字段、MaskSecret 脱敏逻辑
- [x] 4.2 为 `scanner/scanner.go` 编写集成测试：创建包含模拟敏感信息的临时 git repo，验证 ScanDir 和 ScanGitHistory 能检出
- [x] 4.3 验证 `grepom scan` 端到端流程：配置一个测试 group，clone 仓库，执行 scan，验证表格和 JSON 输出
- [x] 4.4 验证 `.gitignore` 感知：在测试 repo 中添加 `.gitignore`，确认排除的文件不被扫描
- [x] 4.5 验证 `.gitleaksignore`：创建 ignore 文件，确认指定 fingerprint 的发现项被过滤
- [x] 4.6 编译验证：`go build ./...` 确认无编译错误，`go vet ./...` 无警告
