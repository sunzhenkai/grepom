## Context

grepom 是一个 Go 语言编写的 CLI 工具，使用 cobra 框架。当前项目有以下现状：

1. **文档缺失**：README.md（中文）和 README_en.md（英文）只覆盖了 sync、clone、list、status、pull、add 等早期命令，缺少 scan、push、search、prune、dedup、pipeline、init、example 等新命令的文档。
2. **CLI 输出语言混乱**：
   - 大部分命令的 cobra 描述为英文（root、init、clone、list、pull、sync、status、search、prune、dedup、pipeline）
   - `scan.go`、`push.go`、`example.go` 的 cobra 描述为中文
   - `interactive.go` 整个 TUI 界面为中文（107+ 字符串）
   - `git/git.go` 认证策略标签为中文
   - `scanner/scanner.go` 错误信息为中文
   - `sync.go`、`list.go` 输出混合中英文
3. **无 i18n 基础设施**：项目中没有国际化框架，所有字符串硬编码。

## Goals / Non-Goals

**Goals:**
- 将所有 CLI 工具的用户可见输出统一为英文
- 更新 README.md（中文）和 README_en.md（英文），补全所有命令文档
- 确保文档中英文版内容同步

**Non-Goals:**
- 不引入 i18n 框架或本地化系统
- 不实现运行时语言切换
- 不修改代码注释的语言（注释保持中文即可）
- 不修改内部日志/调试信息（仅修改用户可见输出）

## Decisions

### 1. 直接替换字符串而非引入 i18n 框架

**选择**：直接将中文字符串替换为英文字符串，硬编码在源码中。

**理由**：
- 项目规模较小，CLI 字符串数量有限（约 130+ 处）
- 引入 i18n 框架（如 go-i18n）会增加复杂度和依赖
- 用户明确要求统一为英文，不需要多语言支持
- Go CLI 工具的惯例是英文输出

**替代方案**：引入 go-i18n + locale 文件 — 过度工程化，不符合项目规模

### 2. 文档结构保持双文件模式

**选择**：继续维护 README.md（中文，默认）和 README_en.md（英文）两个文件。

**理由**：
- 现有结构已建立，用户已习惯
- 双文件模式对小型项目来说简单直观
- 无需构建系统支持

### 3. 测试中的中文字符串断言同步更新

**选择**：同步更新所有引用中文字符串的测试断言。

**理由**：
- 测试断言引用的是用户可见的输出字符串
- 如果只改源码不改测试，测试会失败
- 保持测试与代码一致

### 4. 示例配置中的注释改为英文

**选择**：`cmd/example.go` 中的示例 YAML 配置注释改为英文。

**理由**：
- 示例配置是 CLI 输出的一部分，应与 CLI 语言统一
- YAML 配置是通用格式，英文注释更符合国际惯例

## Risks / Trade-offs

- **[风险] 文档更新后可能再次过时** → 后续新增功能时应同步更新 README
- **[风险] 中文字符串翻译可能不够准确** → 使用简洁、标准的 CLI 英文表述，参考主流 Go CLI 工具（如 git、kubectl、docker）的措辞风格
- **[权衡] 不保留中文 CLI 输出** → 用户明确要求统一英文，这是正确的权衡；文档仍保持中文版
