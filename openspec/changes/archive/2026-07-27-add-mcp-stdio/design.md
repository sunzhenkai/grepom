## Context

grepom 是一个 Go CLI 工具，通过 cobra 框架管理多 Git 仓库。当前所有功能通过终端命令交互，无法被 AI Agent 程序化调用。

MCP（Model Context Protocol）是 Agent 生态的标准工具协议，采用 JSON-RPC over stdio 传输。参考项目 senv 已使用 `github.com/modelcontextprotocol/go-sdk` 成功实现 MCP Server，验证了该方案在 Go + cobra 体系下的可行性。

## Goals / Non-Goals

**Goals:**

- 通过 `grepom mcp serve` 启动 stdio MCP Server，供本地 Agent 调用
- 暴露核心仓库管理能力为 MCP Tools（只读查询 + 安全写操作）
- 提供 `grepom mcp install` 一键配置主流 Agent
- 工具返回结构化 JSON，错误使用 `IsError: true` 让 Agent 自行修正

**Non-Goals:**

- 不实现 SSE / HTTP 传输（仅 stdio）
- 不暴露破坏性操作（如 prune、dedup 删除）为 MCP Tool
- 不实现 MCP Resources / Prompts（仅 Tools）
- 不做远程 MCP Server 部署

## Decisions

### 1. 使用官方 Go SDK

选择 `github.com/modelcontextprotocol/go-sdk/mcp`。

- 替代方案：自行实现 JSON-RPC 协议 → 维护成本高，协议演进风险大
- 替代方案：hashicorp/go-plugin → 非 MCP 标准，Agent 不兼容

### 2. 子命令结构

```
grepom mcp serve          # 启动 stdio server
grepom mcp list-tools     # 打印工具目录（JSON）
grepom mcp install <agent>  # 写入 agent 配置
```

与 senv 保持一致的命令结构，降低用户认知成本。serve 命令复用 root 的 `--config` 持久标志加载配置。

### 3. 工具集设计

| 工具名 | 对应能力 | 类型 |
|--------|---------|------|
| `grepom_list` | 列出仓库 | 只读 |
| `grepom_status` | 查询仓库状态 | 只读 |
| `grepom_search` | 模糊搜索仓库 | 只读 |
| `grepom_dir` | 获取仓库本地路径 | 只读 |
| `grepom_pull` | 拉取指定仓库 | 写（安全） |
| `grepom_clone` | 克隆指定仓库 | 写（安全） |
| `grepom_groups` | 列出组信息 | 只读 |
| `grepom_scan` | 密钥扫描 | 只读 |

工具命名使用 `grepom_` 前缀避免冲突。输入 schema 使用 Go struct + jsonschema tag 自动生成。

### 4. 配置加载策略

MCP Server 启动时通过 `--config` 标志或默认向上搜索 `.grepom.yml` 加载配置。stdin 是 JSON-RPC 通道，不能用于交互提示，因此：
- 无配置时返回错误 JSON（不 panic）
- 需要 token 的操作如果 token 缺失，返回明确错误信息

### 5. Install 命令实现

支持 Agent：claude-code、claude-desktop、cursor、codex、zcode、kimi、pi。

- JSON 格式配置：upsert `mcpServers.grepom` 字段
- TOML 格式（codex）：upsert `[mcp_servers.grepom]` 表
- 写入前创建 `.bak` 备份
- 使用 `os.Executable()` + `EvalSymlinks` 解析绝对路径

## Risks / Trade-offs

- **[go-sdk 版本演进]** → 锁定 minor 版本，关注 breaking change
- **[stdio 阻塞]** → MCP SDK 内部处理并发，工具 handler 中避免无限阻塞（pull/clone 设超时）
- **[配置路径歧义]** → serve 启动时 cwd 可能非项目目录；通过 `--config` 显式指定或 `--dir` 标志解决
- **[Agent 兼容性]** → 各 Agent 配置文件路径/格式不同；通过 agents 注册表集中管理，新增 Agent 只需加一条记录
