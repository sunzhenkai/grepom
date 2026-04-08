## Context

grepom 最近新增了 `generic` provider，但 interactive 模式和 `add resource` CLI 命令中的 provider 列表仍硬编码为 `["gitlab", "github"]`。此外用户反馈子命令名称较长，希望支持前缀匹配。

涉及文件：
- `cmd/interactive.go`：3 处硬编码 provider 列表（第 150、214 行的 Select Options，第 175-178 行的自动命名逻辑）
- `cmd/add.go`：第 38-41 行 provider 验证和帮助文本
- `cmd/root.go`：前缀匹配配置

## Goals / Non-Goals

**Goals:**
- interactive 模式和 `add resource` 命令中 provider 选项包含 `generic`
- `interactiveInit` 中 generic provider 的默认资源名为 `generic`，默认 URL 为空（需用户输入）
- 所有子命令支持唯一前缀匹配（如 `grepom i` → `interactive`，`grepom cl` → `clone`）

**Non-Goals:**
- 不修改 provider 包的注册机制
- 不实现模糊匹配或自动补全（仅前缀匹配）
- 不修改 `config.go` 中的 provider 验证（已在上一个变更中完成）

## Decisions

### 1. provider 列表来源：硬编码 vs 动态获取

**选择：硬编码 `[]string{"gitlab", "github", "generic"}`**

理由：`provider.AvailableProviders()` 返回的顺序不确定（map 遍历），而 interactive 的 Select 菜单需要稳定顺序。硬编码简单可靠，新增 provider 时只需在一处更新。

替代方案：调用 `provider.AvailableProviders()` 后排序。增加了不必要的复杂度。

### 2. 前缀匹配：Cobra 原生 vs 自定义

**选择：Cobra 原生前缀匹配**

Cobra 默认支持唯一前缀匹配（如输入 `cl` 匹配 `clone`），无需额外代码。只需确认未设置 `DisableSuggestions` 或 `SuggestionsMinimumDistance`。如果有歧义（如 `s` 同时匹配 `sync`、`status`、`search`），Cobra 会报错并提示可能的命令。

### 3. generic provider 默认 URL

**选择：不设默认 URL，提示用户输入**

gitlab 默认 `https://gitlab.com`，github 默认 `https://github.com`，但 generic 没有通用默认值。interactive 中 URL 输入框不设 Default，强制用户填写。

## Risks / Trade-offs

- [风险] Cobra 前缀匹配在命令前缀有歧义时报错 → 可接受，Cobra 会给出建议列表
- [风险] 硬编码 provider 列表需要手动维护 → 低风险，provider 类型极少变动
