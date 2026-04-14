## Context

grepom 是一个 Go 语言编写的 Git 仓库管理 CLI 工具。当前 `scan` 命令使用 `text/tabwriter` 输出扫描结果表格，在仓库路径和文件路径较长时列对齐混乱（REPO 列中 `repos/github/sunzhenkai/...` 这样的路径占位过大，导致后续列错位）。配置文件使用 `gopkg.in/yaml.v3` 的 `yaml.Marshal` 写入，固定 2 空格缩进。

当前代码结构：
- `cmd/scan.go`：扫描命令入口，包含 `outputTable` 和 `outputJSON` 两个输出函数
- `scanner/finding.go`：Finding 数据结构和脱敏函数
- `config/config.go`：配置加载/保存，`writeConfig` 使用 `yaml.Marshal` 写入

## Goals / Non-Goals

**Goals:**
- 改善 scan 表格输出的可读性，使结果在终端中整齐对齐
- 支持 `--output` 参数将扫描结果写入指定文件
- 配置文件写入时支持自定义 YAML 缩进空格数量

**Non-Goals:**
- 不改变扫描引擎本身的行为
- 不引入新的输出格式（如 CSV、SARIF），仅优化现有的 table 和 json
- 不改变配置文件 schema，仅增加一个可选的顶层字段

## Decisions

### 1. 表格输出优化：按 repo 分组 + 自动截断

**决策**：将结果按 repo 分组展示，每组内使用缩进的文件列表格式。过长的文件路径自动截断中间部分（保留首尾），移除单独的 SECRET 列（已在之前的实现中脱敏，信息密度低）。

**理由**：用户反馈的乱主要来自单行内容过长导致 tabwriter 换行或错位。按 repo 分组后，repo 名称只出现一次（作为分组标题），大幅减少每行宽度。截断文件路径可进一步控制列宽。

**替代方案**：
- 使用 `text/tabwriter` 的 `AlignRight` 和固定宽度 → 对中文和变宽字体仍不够友好
- 使用第三方表格库（如 `tablewriter`）→ 引入新依赖，增加复杂度

### 2. 输出到文件：新增 `--output` flag

**决策**：在 `scan` 命令中新增 `--output` / `-o` 标志，接受文件路径参数。当指定时，将格式化后的结果写入该文件；未指定时保持原行为输出到 stdout。

**理由**：最小侵入式改动，仅需在 `outputTable` 和 `outputJSON` 中将 `os.Stdout` 替换为可配置的 `io.Writer`。

**实现**：
- 在 `cmd/scan.go` 的 `init()` 中注册 `scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "", ...)`
- `runScan` 中根据 `scanOutput` 是否为空决定输出目标
- 创建/打开文件时使用 `os.Create`，覆盖已有文件

### 3. YAML 缩进配置：配置文件顶层字段

**决策**：在 `Config` 结构体中新增可选字段 `YAMLIndent int` `yaml:"yaml_indent,omitempty"`。在 `writeConfig` 中使用 `yaml.Encoder` 并设置 `Encoder.SetIndent(cfg.YAMLIndent)` 来控制缩进。

**理由**：用户在配置文件中直接设定偏好，一次配置长期有效。默认不设置时保持 2 空格（yaml.v3 默认行为），完全向后兼容。

**替代方案**：
- 命令行 flag（如 `--yaml-indent 4`）→ 每次操作都需要指定，不方便
- 环境变量 → 不如配置文件直观
- 硬编码 4 空格 → 不够灵活

## Risks / Trade-offs

- **[截断路径可能导致信息丢失]** → 缓解：保留完整路径的前后部分，截断中间用 `...` 标记，同时 JSON 输出中始终保留完整路径
- **[输出文件已存在时被覆盖]** → 缓解：使用 `os.Create` 语义明确，文档说明会覆盖
- **[yaml_indent 字段出现在配置文件中可能困惑]** → 缓解：字段为 `omitempty`，不设置时不写入文件，不影响已有配置
