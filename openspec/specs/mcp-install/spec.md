## ADDED Requirements

### Requirement: 安装到指定 Agent

系统 SHALL 提供 `grepom mcp install <agent>` 子命令，将 MCP Server 配置写入指定 Agent 的配置文件。

#### Scenario: 安装到 claude-code

- **WHEN** 用户执行 `grepom mcp install claude-code`
- **THEN** 在 `~/.claude/settings.json` 中 upsert `mcpServers.grepom` 配置项，command 为 grepom 二进制绝对路径，args 为 `["mcp", "serve"]`

#### Scenario: 安装到 cursor

- **WHEN** 用户执行 `grepom mcp install cursor`
- **THEN** 在 `~/.cursor/mcp.json` 中 upsert `mcpServers.grepom` 配置项

#### Scenario: 安装到 codex（TOML 格式）

- **WHEN** 用户执行 `grepom mcp install codex`
- **THEN** 在 `~/.codex/config.toml` 中 upsert `[mcp_servers.grepom]` 表

#### Scenario: 不支持的 Agent

- **WHEN** 用户执行 `grepom mcp install unknown-agent`
- **THEN** 输出错误信息，列出所有支持的 Agent 名称

### Requirement: 配置文件备份

系统 SHALL 在修改 Agent 配置文件前创建 `.bak` 备份文件。

#### Scenario: 已有配置文件

- **WHEN** Agent 配置文件已存在且执行 install
- **THEN** 先创建 `<原文件名>.bak` 备份，再修改原文件

#### Scenario: 配置文件不存在

- **WHEN** Agent 配置文件不存在
- **THEN** 直接创建新文件，不创建备份

### Requirement: Dry-run 模式

系统 SHALL 支持 `--print` 标志，仅输出将要写入的配置片段而不实际修改文件。

#### Scenario: 打印配置

- **WHEN** 用户执行 `grepom mcp install claude-code --print`
- **THEN** 输出 JSON/TOML 配置片段到 stdout，不修改任何文件

### Requirement: 全量安装

系统 SHALL 支持 `--all` 标志，将配置写入所有支持的 Agent。

#### Scenario: 安装到所有 Agent

- **WHEN** 用户执行 `grepom mcp install --all`
- **THEN** 依次为每个支持的 Agent 执行安装，跳过配置文件目录不存在的 Agent

### Requirement: 二进制路径解析

系统 SHALL 使用 `os.Executable()` + `filepath.EvalSymlinks` 解析当前 grepom 二进制的绝对路径，确保 Agent 启动 MCP Server 时不依赖 PATH。

#### Scenario: 通过符号链接调用

- **WHEN** grepom 通过 `~/.local/bin/grepom`（符号链接）调用 install
- **THEN** 配置中写入解析后的真实绝对路径
