## ADDED Requirements

### Requirement: scan 命令支持输出到指定文件
系统 SHALL 支持 `--output` / `-o` 标志，接受文件路径参数。指定后，系统 SHALL 将格式化后的扫描结果写入该文件。未指定时，结果输出到 stdout。

#### Scenario: 输出到指定文件（table 格式）
- **WHEN** 用户执行 `grepom scan --output results.txt`
- **THEN** 系统将表格格式的扫描结果写入 `results.txt` 文件

#### Scenario: 输出到指定文件（json 格式）
- **WHEN** 用户执行 `grepom scan --format json --output results.json`
- **THEN** 系统将 JSON 格式的扫描结果写入 `results.json` 文件

#### Scenario: 未指定 output 时保持原行为
- **WHEN** 用户执行 `grepom scan`（不带 `--output`）
- **THEN** 系统将扫描结果输出到 stdout

#### Scenario: 无发现时输出到文件
- **WHEN** 用户执行 `grepom scan --output empty.txt` 且扫描无发现
- **THEN** 系统将 "No secrets found." 写入 `empty.txt`

### Requirement: 输出文件已存在时覆盖
系统 SHALL 在指定的输出文件已存在时直接覆盖写入，不提示确认。

#### Scenario: 输出文件已存在
- **WHEN** 用户执行 `grepom scan --output results.txt` 且 `results.txt` 已存在
- **THEN** 系统用新的扫描结果覆盖文件内容

### Requirement: 输出文件写入失败时报告错误
系统 SHALL 在无法创建或写入输出文件时报告错误信息并退出。

#### Scenario: 输出路径目录不存在
- **WHEN** 用户执行 `grepom scan --output /nonexistent/dir/results.txt`
- **THEN** 系统报告错误 "无法创建输出文件" 并以非零退出码退出
