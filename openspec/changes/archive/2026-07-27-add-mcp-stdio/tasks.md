## 1. 依赖与脚手架

- [x] 1.1 添加 `github.com/modelcontextprotocol/go-sdk` 依赖（go get）
- [x] 1.2 创建 `cmd/mcp.go`：注册 `mcp` 父命令及 `serve`、`list-tools`、`install` 子命令骨架

## 2. MCP Server 核心

- [x] 2.1 实现 `grepom mcp serve`：初始化 MCP Server、注册 StdioTransport、启动服务
- [x] 2.2 实现配置加载逻辑：复用 root 的 --config 标志，无配置时工具返回错误而非 panic
- [x] 2.3 实现 `grepom mcp list-tools`：遍历已注册工具，输出 JSON 目录

## 3. MCP Tools 实现

- [x] 3.1 创建 `cmd/mcp_tools.go`：定义工具注册函数和公共错误处理辅助函数
- [x] 3.2 实现 `grepom_list` 工具（含 group 过滤参数）
- [x] 3.3 实现 `grepom_status` 工具（单仓库 / 全部）
- [x] 3.4 实现 `grepom_search` 工具（模糊匹配）
- [x] 3.5 实现 `grepom_dir` 工具（路径解析）
- [x] 3.6 实现 `grepom_pull` 工具（含 60s 超时）
- [x] 3.7 实现 `grepom_clone` 工具
- [x] 3.8 实现 `grepom_groups` 工具
- [x] 3.9 实现 `grepom_scan` 工具

## 4. MCP Install 命令

- [x] 4.1 创建 `cmd/mcp_agents.go`：定义 Agent 注册表（名称、配置路径、格式）
- [x] 4.2 创建 `cmd/mcp_install.go`：实现 install 命令主逻辑（JSON/TOML upsert、备份）
- [x] 4.3 实现 `--print` dry-run 模式和 `--all` 全量安装
- [x] 4.4 实现二进制绝对路径解析（os.Executable + EvalSymlinks）

## 5. 文档与收尾

- [x] 5.1 更新 README.md 添加 MCP 使用说明（英文）
- [x] 5.2 更新 README_zh.md 添加 MCP 使用说明（中文）
- [x] 5.3 端到端验证：启动 serve、用 MCP Inspector 或 Agent 调用工具确认正常
