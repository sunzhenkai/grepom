## Why

grepom 目前只能通过 CLI 交互式使用。随着 AI Agent（Claude Code、Cursor、Codex 等）成为日常开发工具，让 Agent 能直接调用 grepom 管理多仓库（查询状态、拉取代码、搜索仓库等）可以显著提升自动化效率。通过 MCP（Model Context Protocol）stdio 传输协议暴露工具接口，Agent 无需解析 CLI 输出即可结构化地与 grepom 交互。

## What Changes

- 新增 `grepom mcp serve` 子命令，启动 stdio MCP Server
- 新增 `grepom mcp list-tools` 子命令，离线打印工具目录
- 新增 `grepom mcp install <agent>` 子命令，将 MCP Server 配置写入指定 Agent 的配置文件
- 暴露一组 MCP Tools，覆盖核心只读/安全操作：仓库列表、状态查询、搜索、目录解析、拉取、克隆等
- 引入 `github.com/modelcontextprotocol/go-sdk` 依赖
- 同步更新 README 文档（多语言）

## Capabilities

### New Capabilities

- `mcp-server`: MCP stdio Server 生命周期管理（serve / list-tools），工具注册与调度
- `mcp-tools`: 暴露给 Agent 的工具集定义（输入 schema、输出格式、错误处理）
- `mcp-install`: 将 MCP Server 配置自动写入各 Agent 配置文件（JSON/TOML），支持备份与 dry-run

### Modified Capabilities

（无现有 spec 需要修改，MCP 层复用已有 cmd/config/git/repo 包逻辑）

## Impact

- **新增文件**: `cmd/mcp.go`、`cmd/mcp_tools.go`、`cmd/mcp_agents.go`、`cmd/mcp_install.go`
- **依赖**: 新增 `github.com/modelcontextprotocol/go-sdk`
- **构建**: Makefile 无需修改（标准 go build）
- **文档**: README.md / README_zh.md 新增 MCP 使用说明
- **兼容性**: 纯新增功能，不影响现有 CLI 行为
