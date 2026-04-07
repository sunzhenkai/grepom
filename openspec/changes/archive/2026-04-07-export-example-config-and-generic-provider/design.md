## Context

grepom 是一个用 Go 编写的多仓库管理 CLI 工具，通过 YAML 配置文件（`.grepom.yml`）声明式管理跨 GitLab/GitHub 的仓库。当前有两个不足：

1. **配置发现性差**：`grepom init` 仅生成最小空配置，用户无法快速了解所有可用字段（如 `ssh_key`、`enabled`、`exclude_repos`、`recursive` 等）。需要一种方式输出完整的示例配置。
2. **Provider 局限**：仅支持 GitHub 和 GitLab，不能管理自建 Git 服务器或无 API 的 Git 仓库。

现有架构中，Provider 使用注册表模式（`provider.Register`）自注册，配置验证硬编码了合法 provider 列表。

## Goals / Non-Goals

**Goals:**
- 提供 `grepom example` 命令，输出包含全部字段和注释的完整示例配置
- 实现 `generic` provider，支持不依赖平台 API 的纯 Git URL 仓库管理
- 保持完全向后兼容

**Non-Goals:**
- 不实现 Gitea、Bitbucket 等具体第三方平台 provider
- `generic` provider 不支持自动发现仓库（无远程 API）
- `example` 命令不生成可直接使用的配置（仅为参考示例）

## Decisions

### 1. `example` 命令输出方式

**决策**：默认输出到 stdout，支持 `--output` / `-o` 标志写入文件。

**理由**：输出到 stdout 方便用户管道处理（`grepom example > my-config.yml`），也可通过 `--output` 直接写入文件。比仅支持写文件更灵活。

**备选方案**：
- 仅写文件：不够灵活，无法与管道配合
- 仅 stdout：已包含在决策中，`--output` 为便利补充

### 2. 示例配置的内容策略

**决策**：使用硬编码的 Go 字符串常量生成示例配置，包含 YAML 注释说明每个字段的用途和可选值。

**理由**：
- 硬编码保证示例配置的格式美观、注释丰富（YAML 的 `gopkg.in/yaml.v3` marshal 不支持注释）
- 示例内容固定，不需要动态生成
- 维护成本低，新增字段时只需更新常量

**备选方案**：
- 从 struct 反射生成：无法添加注释，格式不美观
- 外部模板文件：增加分发复杂度

### 3. `generic` provider 的设计

**决策**：`generic` provider 实现 `Provider` 接口，但 `ListRepos` 和 `ListGroups` 均返回空列表和 nil 错误。

**理由**：
- 遵循接口隔离原则，无需修改 Provider 接口
- `generic` 的仓库通过配置文件的 `repos` 或 `groups.repos` 显式声明，不需要 API 发现
- `sync` 命令对 `generic` resource 会静默跳过（返回空列表）

### 4. `generic` provider 的认证映射

**决策**：在 `git/git.go` 的 token 用户名映射中，`generic` 使用 `token` 作为默认用户名（与当前 `default` 分支一致）。

**理由**：通用 Git 服务器没有统一的 token 认证用户名约定，使用 `token` 是最通用的选择。用户使用 SSH key 时此设置无影响。

### 5. 配置验证的 provider 列表管理

**决策**：将 `config.go` 中硬编码的 provider 列表改为从 `provider.AvailableProviders()` 动态获取。

**理由**：避免在多处维护 provider 列表，新增 provider 时只需注册一次，减少遗漏风险。

**备选方案**：
- 继续硬编码：每次新增 provider 都要改 config.go，容易遗漏
- 配置文件定义：过度工程化

## Risks / Trade-offs

- **[示例配置维护]** → 新增字段时需同步更新示例配置常量。通过代码审查和测试保证一致性。
- **[generic provider 能力有限]** → 用户可能期望 `generic` 支持仓库发现。通过文档和注释明确说明 `generic` 仅支持显式声明的仓库。
- **[动态 provider 列表]** → 如果 `provider` 包初始化顺序异常，可能导致验证时列表不完整。Go 的 `init()` 机制保证同一 binary 内的初始化顺序可靠。
