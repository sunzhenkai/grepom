# grepom

[English](./README_en.md) | 简体中文

Git 仓库编排器与管理器 — 通过单个 YAML 配置文件管理 GitLab 组和 GitHub 组织中的多个 Git 仓库。

## 功能特性

- **声明式配置** — 在 YAML 中定义 GitLab 组和 GitHub 组织，grepom 自动发现仓库
- **批量操作** — 一次性克隆、拉取和检查所有仓库的状态
- **分层布局** — 在本地保留组/子组的目录结构
- **多提供者** — 同时支持 GitLab、GitHub、Codeup 和 Generic
- **灵活过滤** — 按名称、组或提供者过滤
- **敏感信息扫描** — 内置 gitleaks 引擎，支持工作区和 git 历史扫描
- **推送保护** — 推送前自动检测敏感信息，防止泄露
- **交互式模式** — 菜单驱动的交互式操作界面
- **MR/PR 创建** — 在 CLI 中直接创建 GitHub Pull Request 或 GitLab Merge Request，已有 MR/PR 时自动返回地址
- **服务进程管理** — 后台启动本地开发服务，查看状态、日志、停止进程，并提供 TUI 管理界面

## 安装

```bash
go install github.com/wii/grepom@latest
```

或者从源代码构建：

```bash
make install
```

## 快速开始

```bash
grepom init                     # 初始化配置文件
grepom example -o .grepom.yml   # 导出示例配置（含完整字段说明）
grepom add resource ...         # 添加认证资源
grepom add group ...            # 添加远程组
grepom sync                     # 发现仓库并更新配置
grepom clone                    # 克隆所有仓库
```

## 使用方法

创建配置文件（默认为 `.grepom.yml`）。grepom 会从当前目录沿父目录链自动向上查找配置文件（类似 git 查找 `.git` 的行为），因此你可以在任意子目录中执行命令。

```yaml
base: ~/projects

resources:
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_TOKEN}
    ssh_key: ~/.ssh/id_work        # 可选

  - name: my-github
    provider: github
    url: https://github.com
    token: ${GITHUB_TOKEN}

groups:
  - name: frontend
    resource: my-gitlab
    path: my-org/frontend
    recursive: true
    exclude_repos:                 # 可选：排除指定仓库
      - archived-repo

  - name: my-org
    resource: my-github
    path: my-github-org

repos:                             # 独立仓库（不属于任何组）
  - name: dotfiles
    resource: my-github
    url: https://github.com/me/dotfiles.git

services:                          # 可选：本地开发服务定义
  api:
    cwd: ./backend
    command: make dev
  web:
    cwd: ./frontend
    command:
      - pnpm
      - dev
```

### 命令

```bash
# 初始化与配置
grepom init                         # 初始化配置文件
grepom example                      # 导出完整示例配置
grepom interactive                  # 进入交互式操作模式

# 同步与发现
grepom sync                         # 发现仓库并更新配置元数据
grepom sync --source my-gitlab      # 按资源名同步
grepom sync --group frontend        # 按组同步

# 克隆与拉取
grepom clone                        # 克隆所有已发现的仓库
grepom clone web-app                # 克隆特定仓库
grepom clone --group frontend       # 克隆组中的所有仓库
grepom clone --concurrency 8        # 8 个 worker 并行克隆

grepom pull                         # 拉取所有已克隆仓库的更新
grepom pull web-app                 # 拉取特定仓库
grepom pull --force                 # 跳过安全检查强制拉取
grepom pull --concurrency 8         # 8 个 worker 并行拉取

# 查询与过滤
grepom list                         # 列出需要关注的仓库（未推送/未提交）
grepom list --all                   # 列出所有仓库及状态
grepom list --no-push               # 仅显示有未推送提交的仓库
grepom list --no-commit             # 仅显示有未提交更改的仓库
grepom list --group frontend        # 按组过滤
grepom list --resource my-gitlab    # 按资源过滤
grepom list groups                  # 列出已配置的组
grepom list resources               # 列出已配置的资源
grepom list --remote                # 从远程 API 查询仓库列表
grepom list --remote --type groups  # 从远程 API 查询组列表

grepom status                       # 检查所有已克隆仓库的状态
grepom status web-app               # 特定仓库的状态

grepom search web                   # 按名称模糊搜索仓库
grepom search web --group frontend  # 在指定组内搜索

grepom dir                          # 输出 base 目录路径
grepom dir web-app                  # 输出仓库的本地路径
grepom dir web --group fe           # 在指定组内搜索并输出路径
cd "$(grepom dir web-app)"          # 快速跳转到仓库目录

# 敏感信息扫描
grepom scan                         # 扫描所有已克隆仓库的工作区
grepom scan -p /path/to/project     # 直接扫描指定目录（无需配置文件）
grepom scan --group frontend        # 仅扫描 frontend 组
grepom scan --history               # 扫描工作区 + git 历史
grepom scan --format json           # JSON 格式输出
grepom scan --output results.txt    # 输出到文件
grepom scan --gitleaks-config rules.toml  # 使用自定义规则

# 推送保护
grepom push                         # 扫描后推送（无敏感信息时）
grepom push -f                      # 发现敏感信息仍强制推送
grepom push -- origin main          # 透传参数给 git push

# MR/PR 创建
grepom mr                           # 自动检测并创建 MR/PR（已有则返回地址）
grepom mr --from feat-x --to main   # 指定源分支和目标分支
grepom mr --title "Add dark mode"   # 自定义标题
grepom mr --draft                   # 创建草稿 MR/PR
grepom mr --web                     # 在浏览器中打开创建页面
grepom pr                           # 'mr' 的别名

# CI/CD 管道
grepom watch                        # 自动推断当前仓库，监控最新管道
grepom watch web-app                # 监控指定仓库的最新管道
grepom watch --id 1234              # 监控指定管道 ID
grepom pipeline list <repo-name>    # 列出仓库的管道
grepom pipeline watch <repo-name>   # 实时监控管道状态
grepom tag -w                       # 创建版本标签后自动监控管道状态

# 服务进程管理
grepom svc run -- make dev         # 在当前目录后台启动服务（默认服务名为目录名）
grepom svc run api                  # 从 .grepom.yml 读取 api 服务定义并启动
grepom svc list                     # 表格展示服务名、状态、PID、路径、命令和日志路径
grepom svc status api               # 查看单个服务状态
grepom svc logs -f api              # 持续查看服务日志
grepom svc logs --open api          # 用编辑器打开日志文件
grepom svc kill api                 # 停止服务
grepom svc kill -9 api              # 强制停止服务
grepom svc clean                    # 清理已退出服务的记录
grepom svc dir api                  # 输出服务工作目录
grepom svc tui                      # 打开 TUI 管理界面
eval "$(grepom svc --shell)"        # 启用 gsvc 快捷跳转服务目录

# 维护
grepom prune                        # 删除配置中不存在的已克隆仓库
grepom dedup                        # 检查所有组的组内重复和跨组警告
grepom dedup --group core-team      # 仅检查 core-team 组
grepom dedup --group core-team --reference infra-team  # 额外执行按名称跨组去重
grepom dedup --apply                # 执行实际写入

# 添加资源/组/仓库
grepom add resource --name my-gl --provider gitlab --url https://gitlab.com --token '${GITLAB_TOKEN}'
grepom add group --name frontend --resource my-gl --path my-org/frontend --recursive
grepom add repo --name special --url https://gitlab.com/other/special.git
```

### Token 环境变量

Token 字段支持 `${ENV_VAR}` 占位符语法。实际值在运行时从环境变量中解析，并且在写入配置文件时保留占位符。

```yaml
resources:
  - provider: gitlab
    token: ${GITLAB_TOKEN}   # 在运行时从 $GITLAB_TOKEN 解析
```

```bash
export GITLAB_TOKEN=glpat-xxxxxxxxxxxx
grepom sync   # 使用解析后的 token 值
```

### 标志

| 标志 | 简写 | 默认值 | 描述 |
|------|------|--------|------|
| `--config` | `-c` | 自动查找 | 配置文件路径（默认从当前目录向上查找 `.grepom.yml`） |
| `--verbose` | `-v` | `false` | 启用详细输出 |

## 构建

```bash
make build    # 构建二进制文件
make test     # 运行测试
make lint     # 运行 vet 和格式检查
make install  # 构建并安装到 ~/.local/bin
make clean    # 删除二进制文件
```
