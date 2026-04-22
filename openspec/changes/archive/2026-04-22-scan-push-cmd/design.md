## Context

grepom 是一个 CLI 工具，管理多个 git 仓库。当前架构：

- **配置驱动**：所有命令（clone、pull、scan 等）都依赖 YAML 配置文件（`.grepom.yml`），通过 `loadConfig()` 加载配置解析仓库列表
- **scanner 包**：封装 gitleaks 引擎，提供 `ScanDir(ctx, path)` 和 `ScanGitHistory(ctx, path)` 两个核心方法，返回 `[]Finding`
- **scan 命令**：当前通过配置解析仓库列表，过滤出已克隆仓库，调用 `scanner.ScanDir()` 并行扫描，输出表格或 JSON

当前痛点：
1. 在一个普通 git 项目中（非 grepom 管理），想快速扫描敏感信息需要先创建配置文件
2. 推送代码前没有安全检查，密钥泄露风险高

## Goals / Non-Goals

**Goals:**
- `grepom scan` 在无配置文件时自动扫描当前目录，零配置可用
- 新增 `grepom push` 命令，推送前自动扫描当前目录敏感信息
- push 发现敏感信息时默认拒绝推送，`-f` 强制推送但仍打印警告
- 复用现有 scanner 包，不引入新的外部依赖

**Non-Goals:**
- 不修改 push 命令的 git 推送行为（直接调用 `git push`）
- 不实现 pre-push hook 的自动安装（后续可扩展）
- 不修改 scan 命令在已有配置文件时的行为
- 不增加新的扫描规则或修改 scanner 包核心逻辑

## Decisions

### Decision 1: scan 无配置时的回退策略

**选择**：修改 `cmd/scan.go` 的 `runScan` 函数，当 `loadConfig()` 因配置文件不存在而失败时，回退到扫描当前工作目录（`.`）。

**理由**：
- 改动最小，只需在 `runScan` 中捕获配置加载失败的错误并判断是否为"文件不存在"类型
- config.FindConfig 在找不到配置文件时返回明确错误信息，可以据此判断
- 复用现有的 `scanner.ScanDir()` 方法，无需修改 scanner 包

**备选方案**：新建一个 `scanCurrentDir` 函数独立处理——增加了代码重复，不如在现有流程中增加回退分支。

### Decision 2: push 命令的实现方式

**选择**：在 `cmd/push.go` 中实现独立的 `push` 子命令，不依赖配置文件，直接操作当前目录。

**理由**：
- push 的语义是"在当前项目中执行 git push 前做安全检查"，与 grepom 的多仓库管理场景不同
- 独立命令结构清晰，不与现有的配置驱动命令耦合
- 当前目录扫描可以直接调用 `scanner.ScanDir(ctx, ".")`

**流程设计**：
1. 检测当前目录是否为 git 仓库（存在 `.git` 目录）
2. 执行 `scanner.ScanDir(ctx, ".")` 扫描工作区
3. 判断是否有发现项：
   - 有发现 → 打印详情，检查 `-f` 标志
     - 无 `-f`：退出码非零，拒绝推送
     - 有 `-f`：打印警告，继续推送
   - 无发现 → 直接执行 `git push`
4. 调用 `exec.Command("git", "push")` 执行推送

### Decision 3: push 命令的输出格式

**选择**：复用 scan 命令的 `outputTable` 函数输出扫描结果。

**理由**：
- 一致的输出体验
- 减少代码重复
- 用户已在 scan 命令中熟悉该格式

### Decision 4: 配置文件不存在的判断方式

**选择**：在 `runScan` 中检查 `loadConfig()` 返回的错误字符串，或引入 `config.IsConfigNotFoundError(err)` 辅助函数。

**理由**：当前 `FindConfig` 返回 `fmt.Errorf("no config file found...")` 格式的错误，可以：
- 方案 A：字符串匹配（简单但有脆弱性）
- 方案 B：自定义错误类型（更健壮）

选择方案 B：在 config 包中定义 `ErrConfigNotFound` 变量，让 `FindConfig` 返回该错误，scan 和 push 中用 `errors.Is` 判断。

## Risks / Trade-offs

- **[扫描性能]** → 当前目录可能很大（如包含 node_modules），扫描时间可能较长。缓解：scanner 已支持 .gitignore 排除，且 push 只扫描工作区不扫描历史。
- **[误报导致推送阻塞]** → gitleaks 可能产生误报，阻止正常推送。缓解：提供 `-f` 强制推送选项，用户可以快速绕过。
- **[扫描范围]** → push 只扫描工作区文件，不扫描 git 历史。已在 git 历史中的泄露不会被检测。缓解：这是合理的权衡，历史扫描应使用 `grepom scan --history`。
- **[git push 代理]** → push 命令直接调用系统 git，不处理认证。缓解：这是设计意图，push 只关注安全检查，认证由 git 自身处理。
