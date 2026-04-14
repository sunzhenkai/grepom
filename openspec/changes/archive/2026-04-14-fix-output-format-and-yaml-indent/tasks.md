## 1. 表格输出格式优化

- [x] 1.1 在 `scanner/finding.go` 中新增 `TruncatePath(path string, maxLen int) string` 函数，用于截断过长的文件路径（保留前部分和文件名，中间用 `...` 替代）
- [x] 1.2 在 `cmd/scan.go` 中重写 `outputTable` 函数，按仓库名称分组展示结果：每个仓库作为分组标题，发现项以缩进列表形式展示（含截断后的文件路径、行号、规则 ID、严重程度、脱敏 secret）
- [x] 1.3 确保 `outputTable` 末尾仍输出汇总统计信息（总发现数、仓库数、按严重程度计数）

## 2. 扫描结果输出到文件

- [x] 2.1 在 `cmd/scan.go` 的 `init()` 中新增 `--output` / `-o` flag（`scanOutput string`）
- [x] 2.2 修改 `outputTable` 和 `outputJSON` 函数签名，接受 `io.Writer` 参数替代硬编码 `os.Stdout`
- [x] 2.3 在 `runScan` 中根据 `scanOutput` 是否为空决定输出目标：为空时使用 `os.Stdout`，否则使用 `os.Create` 打开目标文件
- [x] 2.4 添加输出文件创建失败的错误处理，确保目录不存在等情况给出明确错误提示

## 3. YAML 缩进配置

- [x] 3.1 在 `config/config.go` 的 `Config` 结构体中新增 `YAMLIndent int` 字段（`yaml:"yaml_indent,omitempty"`）
- [x] 3.2 修改 `writeConfig` 函数：使用 `yaml.NewEncoder` 替代 `yaml.Marshal`，根据 `cfg.YAMLIndent` 设置 `encoder.SetIndent`（默认 2）
- [x] 3.3 在 `writeConfig` 中确保 `yaml_indent` 字段本身不被写入输出文件（写入前将字段清零，写入后恢复）

## 4. 测试验证

- [x] 4.1 为 `TruncatePath` 函数编写单元测试，覆盖短路径不截断、长路径截断、边界情况
- [x] 4.2 验证 `outputTable` 分组输出格式正确（手动运行 `grepom scan` 确认终端展示）
- [x] 4.3 验证 `--output` flag 功能：结果正确写入文件、JSON 格式输出到文件、无发现时文件内容正确
- [x] 4.4 验证 YAML 缩进配置：设置 `yaml_indent: 4` 后执行 add 操作，确认生成的 YAML 文件使用 4 空格缩进且不包含 `yaml_indent` 字段
