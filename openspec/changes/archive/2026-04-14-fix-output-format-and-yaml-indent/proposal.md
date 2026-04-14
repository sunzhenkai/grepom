## Why

`grepom scan` 的表格输出在仓库路径和文件路径较长时列对齐混乱，可读性差；同时缺少将扫描结果持久化到文件的能力，不便于后续审计和集成；此外，配置文件写入时 YAML 缩进固定为 2 空格，无法满足用户自定义格式偏好。

## What Changes

- 优化 `scan` 命令的表格输出格式，使用对齐更友好的展示方式（自动截断过长路径、按 repo 分组显示等），提升可读性
- 新增 `--output` / `-o` 标志，支持将扫描结果输出到指定文件（支持 table 和 json 两种格式）
- 配置文件写入时支持自定义 YAML 缩进空格数量（通过配置字段或命令行参数）

## Capabilities

### New Capabilities
- `scan-output-format`: 优化扫描结果表格输出格式，支持结果按 repo 分组展示，过长内容自动截断，提升终端可读性
- `scan-output-file`: scan 命令新增 `--output` 参数，支持将扫描结果写入指定文件
- `yaml-indent-config`: 配置文件写入时支持自定义 YAML 缩进空格数量

### Modified Capabilities
- `secret-scanning`: 扫描结果输出格式优化，新增输出到文件的能力

## Impact

- `cmd/scan.go`: 修改输出逻辑，新增 `--output` flag
- `scanner/finding.go`: 可能需要调整 JSON 输出函数以支持写入文件
- `config/config.go`: `writeConfig` 函数需要支持可配置缩进；新增缩进配置字段
- `config/verbose.go`: 可能需要增加缩进相关配置
- 无 **BREAKING** 变更，所有改动向后兼容
