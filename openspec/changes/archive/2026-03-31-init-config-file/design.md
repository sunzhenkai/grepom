## Context

grepom 是一个 Go CLI 工具，用于管理多个 git 仓库。当前 `init` 命令执行 clone 操作，与 CLI 惯例（`git init`、`cargo init`、`npm init` 等均表示初始化项目/配置）不符。用户首次使用 grepom 时需要手动创建 `.grepom.yml`，无法通过工具自身完成配置初始化。

现有代码中，`cmd/init.go` 包含 clone 逻辑，`cmd/add.go` 支持向配置文件追加 source/repo（配置文件不存在时会自动创建），但缺少一个显式的"初始化"入口。

## Goals / Non-Goals

**Goals:**
- 将 `init` 命令语义改为"初始化配置文件"
- 提供交互式引导，帮助用户快速创建第一个配置
- 将原 clone 逻辑迁移到语义更准确的 `clone` 命令
- 保持与现有 `add` 命令的互补关系：`init` 创建骨架，`add` 追加内容

**Non-Goals:**
- 不做 TUI 或富交互界面，使用简单的 flag + prompt 模式
- 不自动检测或填充 API token（用户手动提供）
- 不修改现有 `add source` / `add repo` 的行为
- 不引入外部交互式输入库（使用标准库）

## Decisions

### 1. `init` 命令行为：非交互式 flag 模式

**选择**: 使用 `--base`、`--provider`、`--url`、`--group`、`--org` 等 flag 控制初始化内容，不使用交互式 prompt
**理由**:
- 与现有 `add source` 命令风格一致（flag-driven）
- CLI 工具更适合脚本化和自动化，交互式 prompt 在管道场景中不友好
- 简单场景 `grepom init` 即可创建最小配置，复杂场景通过 flag 补充

**替代方案**: 交互式 wizard（用户体验更好但不适合脚本、实现更复杂）

### 2. 最小初始化 vs 完整初始化

**选择**: `grepom init` 仅创建包含 `base` 的最小配置文件，source 通过 flag 可选添加
**理由**:
- 保持命令简单，`init` 只负责"初始化"，添加 source 可以通过 `add` 命令完成
- 允许用户 `grepom init && grepom add source ...` 的工作流
- 同时支持 `grepom init --provider gitlab --url https://gitlab.com --group my-org` 一步到位

### 3. `clone` 命令：直接迁移

**选择**: 将 `cmd/init.go` 的逻辑原样迁移到 `cmd/clone.go`，仅修改命令名和 help 文本
**理由**:
- 零行为变更，降低风险
- 用户只需将 `grepom init` 替换为 `grepom clone`

### 4. 配置文件已存在时的行为

**选择**: 报错并提示，不覆盖
**理由**: 防止用户意外覆盖已有配置

## Risks / Trade-offs

- **[BREAKING 变更]** → `init` 语义变更会影响已有用户。缓解：在 clone 不存在时给出提示 "did you mean `grepom clone`?"
- **[无交互引导]** → 新用户可能不知道如何填写参数。缓解：`grepom init --help` 提供清晰的示例
