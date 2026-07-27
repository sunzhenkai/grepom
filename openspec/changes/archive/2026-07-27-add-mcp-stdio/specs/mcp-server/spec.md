## ADDED Requirements

### Requirement: MCP Server 启动

系统 SHALL 提供 `grepom mcp serve` 子命令，启动基于 stdio 传输的 MCP Server。Server 使用 `github.com/modelcontextprotocol/go-sdk/mcp` 的 `StdioTransport` 进行 JSON-RPC 通信。

#### Scenario: 正常启动

- **WHEN** 用户执行 `grepom mcp serve`
- **THEN** 进程通过 stdin/stdout 提供 MCP JSON-RPC 服务，不输出非协议内容到 stdout

#### Scenario: 带配置启动

- **WHEN** 用户执行 `grepom mcp serve --config /path/to/.grepom.yml`
- **THEN** Server 使用指定配置文件加载仓库信息

#### Scenario: 配置缺失

- **WHEN** 无法找到 `.grepom.yml` 配置文件
- **THEN** Server 仍然启动，但工具调用返回结构化错误信息（IsError: true）

### Requirement: 工具目录查询

系统 SHALL 提供 `grepom mcp list-tools` 子命令，以 JSON 格式输出所有已注册工具的目录信息。

#### Scenario: 列出工具

- **WHEN** 用户执行 `grepom mcp list-tools`
- **THEN** 输出 JSON 数组，每项包含工具名称、描述和输入 schema

### Requirement: 无 TTY 依赖

MCP Server 运行时 SHALL NOT 依赖 TTY 交互（如密码提示、确认对话框）。stdin 专用于 JSON-RPC 协议通信。

#### Scenario: 非交互环境运行

- **WHEN** Agent 以子进程方式启动 `grepom mcp serve`（无 TTY）
- **THEN** Server 正常运行，所有需要用户输入的场景改为返回错误信息
