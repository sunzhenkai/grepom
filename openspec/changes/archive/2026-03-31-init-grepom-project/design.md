## Context

grepom 是一个全新的 Go CLI 项目，用于管理多个 git 仓库。目标用户是同时维护多个 GitLab group/subgroup 和 GitHub org 仓库的开发者。当前项目为空仓库，无任何现有代码。

## Goals / Non-Goals

**Goals:**
- 提供声明式 YAML 配置，用户通过定义 API 源自动发现仓库
- 支持 GitLab（含递归 subgroup）和 GitHub（org/team）的仓库获取
- 提供 5 个核心命令覆盖完整生命周期：添加配置、clone、列表、状态查看、更新
- group/subgroup 嵌套层级映射为本地目录嵌套结构
- 多配置文件支持，可按场景切换

**Non-Goals:**
- 不做本地缓存/离线模式 — 每次运行从 API 动态获取
- 不做 SSH key 管理 — 用户自行配置 git credentials
- 不做仓库 fork/mirror/CI 相关操作
- 不做 TUI（终端 UI）— 纯 CLI
- 不支持 GitLab/GitHub 以外的平台（预留扩展点即可）

## Decisions

### 1. CLI 框架: cobra

**选择**: cobra
**理由**: Go 生态中最成熟的 CLI 框架，子命令、flag、自动补全、帮助文档均为一等公民。urfave/cli 更简洁但子命令嵌套和帮助生成不如 cobra。
**替代方案**: urfave/cli（更轻量但生态较小）、标准库 flag（过于底层）

### 2. 配置文件格式: YAML

**选择**: YAML（`gopkg.in/yaml.v3`）
**理由**: 人类可读性好，支持注释，适合手写维护的配置场景。
**替代方案**: TOML（结构清晰但注释支持弱）、JSON（不可读）。

### 3. Git 操作: 直接调用 git 命令

**选择**: 通过 `os/exec` 调用本地 `git` 命令
**理由**: `go-git` 是纯 Go 实现，但部分操作（如 SSH、submodule、credential helper）支持不完善。直接调用 git 命令兼容性最好，用户环境已安装 git。
**替代方案**: go-git（纯 Go，但功能不完整，调试困难）

### 4. API 客户端: 直接 HTTP 调用

**选择**: 使用 `net/http` 直接调用 GitLab API v4 和 GitHub REST API
**理由**: 只需要 list projects/repositories 这几个端点，引入完整 SDK 过重。GitLab Go SDK（`gitlab-go`）和 GitHub Go SDK（`go-github`）都很大，我们只需要 group traversal + repo listing。
**替代方案**: `github.com/xanzy/go-gitlab` + `github.com/google/go-github`（功能完整但依赖重）

### 5. 环境变量引用: `${VAR}` 语法

**选择**: YAML 中使用 `${VAR}` 语法引用环境变量
**理由**: 直观，与 shell/Docker Compose 习惯一致。实现上在 YAML 解析后做一轮字符串替换。
**替代方案**: os.ExpandEnv（标准库，行为一致）

### 6. Provider 接口设计

```go
type Provider interface {
    ListRepos(ctx context.Context, source Source) ([]Repo, error)
}
```

每个 provider 实现该接口。`Repo` 包含 `Name`、`CloneURL`（HTTP+SSH）、`Path`（相对路径）。

### 7. 项目目录结构

```
grepom/
├── main.go                 # 入口
├── cmd/
│   ├── root.go             # 根命令 + -c flag
│   ├── add.go              # add source/repo
│   ├── init.go             # clone 仓库
│   ├── list.go             # 列出仓库
│   ├── status.go           # 查看 git 状态
│   └── pull.go             # pull 更新
├── config/
│   └── config.go           # 配置结构定义 + YAML 解析
├── provider/
│   ├── provider.go         # Provider 接口 + Repo 模型
│   ├── gitlab.go           # GitLab API 实现
│   └── github.go           # GitHub API 实现
├── repo/
│   └── resolver.go         # 仓库路径解析 + 列表聚合
└── git/
    └── git.go              # git clone/pull/status 封装
```

## Risks / Trade-offs

- **[每次运行都调 API]** → 大量 group 时可能较慢。缓解：合理使用并发请求，后续可加缓存层
- **[直接调用 git 命令]** → 依赖用户本地 git 安装，输出解析可能因 git 版本不同而异。缓解：要求 git >= 2.x，使用 `--porcelain` 格式输出
- **[直接 HTTP 调用 API]** → 需要自行处理分页、错误码。缓解：封装通用的分页遍历函数
- **[无缓存]** → API 限流场景下频繁调用可能被限。缓解：尊重 Rate-Limit 头，提供合理重试
