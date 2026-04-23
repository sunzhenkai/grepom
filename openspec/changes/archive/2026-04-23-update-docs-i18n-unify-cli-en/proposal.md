## Why

当前项目文档和 CLI 输出存在两个问题：

1. **文档过时** — README.md（中文）和 README_en.md（英文）缺少近期新增的功能（scan、push、search、prune、dedup、pipeline、init、example 等命令），且缺少新引入的 flags 和配置字段说明。
2. **CLI 输出语言不统一** — 部分命令输出为英文，部分为中文，部分中英混用。具体表现为：
   - `cmd/interactive.go` 整个交互式 TUI 为中文（107+ 个中文字符串）
   - `cmd/scan.go`、`cmd/push.go`、`cmd/example.go` 的 cobra 描述为中文
   - `git/git.go` 中认证策略标签为中文
   - `scanner/scanner.go` 中错误信息为中文
   - `cmd/sync.go`、`cmd/list.go` 中混合中英文

用户要求将所有 CLI 工具的命令输出统一为英文，同时更新文档使其包含中英文双语版本，默认展示中文。

## What Changes

- 更新 README.md（中文默认），补充缺失的功能文档：scan、push、search、prune、dedup、pipeline、init、example 等命令
- 更新 README_en.md（英文），与中文版保持同步
- 将 `cmd/interactive.go` 中所有用户可见的中文字符串翻译为英文
- 将 `cmd/scan.go` 中 cobra 描述（Short/Long/Example）和 flag 帮助文本改为英文
- 将 `cmd/push.go` 中 cobra 描述和 flag 帮助文本改为英文
- 将 `cmd/example.go` 中 cobra 描述、flag 帮助文本及示例配置注释改为英文
- 将 `cmd/sync.go`、`cmd/list.go` 中混合中英文的输出统一为英文
- 将 `git/git.go` 中认证策略标签和进度输出改为英文
- 将 `scanner/scanner.go` 中错误信息改为英文
- 更新相关测试文件中引用中文字符串的断言

## Capabilities

### New Capabilities

无新增能力。

### Modified Capabilities

- `cli-commands`: 更新 cobra 命令描述，统一为英文
- `interactive-mode`: 交互式 TUI 所有用户可见字符串改为英文
- `secret-scanning`: scan 命令描述和错误信息改为英文
- `push-guard`: push 命令描述和错误信息改为英文
- `example-command`: example 命令描述和示例配置注释改为英文

## Impact

- **代码变更**: `cmd/`、`git/`、`scanner/` 目录下的 Go 源文件
- **文档变更**: README.md、README_en.md
- **测试变更**: `git/git_test.go`、`scanner/scanner_test.go`、`scanner/finding_test.go` 中引用中文字符串的断言需同步更新
- **无 API 变更**: 仅字符串内容变更，不影响接口和结构
- **无依赖变更**: 不引入新的第三方库
