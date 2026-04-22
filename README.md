# grepom

[English](./README_en.md) | 简体中文

Git 仓库编排器与管理器 — 通过单个 YAML 配置文件管理 GitLab 组和 GitHub 组织中的多个 Git 仓库。

## 功能特性

- **声明式配置** — 在 YAML 中定义 GitLab 组和 GitHub 组织，grepom 自动发现仓库
- **批量操作** — 一次性克隆、拉取和检查所有仓库的状态
- **分层布局** — 在本地保留组/子组的目录结构
- **多提供者** — 同时支持 GitLab 和 GitHub API
- **灵活过滤** — 按名称、组或提供者过滤

## 安装

```bash
go install github.com/wii/grepom@latest
```

或者从源代码构建：

```bash
make install
```

## 使用方法

创建配置文件（默认为 `.grepom.yml`）：

```yaml
base: ~/projects

sources:
  - name: my-gitlab
    provider: gitlab
    url: https://gitlab.com
    token: ${GITLAB_TOKEN}
    groups:
      - path: my-org/frontend
        recursive: true

  - name: my-github
    provider: github
    url: https://github.com
    token: ${GITHUB_TOKEN}
    orgs:
      - name: my-org
```

### 命令

```bash
grepom sync                            # 发现仓库并更新配置元数据
grepom sync --source my-gitlab         # 按名称同步特定的源
grepom sync --source 0                 # 按索引同步特定的源

grepom clone                           # 克隆所有已发现的仓库
grepom clone web-app                   # 克隆特定的仓库
grepom clone --group my-org/frontend   # 克隆组中的所有仓库

grepom list                            # 列出所有已发现的仓库
grepom list --source gitlab            # 按提供者过滤
grepom list --group my-org/frontend    # 按组过滤

grepom status                          # 检查所有已克隆仓库的状态
grepom status web-app                  # 特定仓库的状态

grepom pull                            # 拉取所有已克隆仓库的更新
grepom pull web-app                    # 拉取特定仓库的更新
```

### 交互式添加源/仓库

```bash
grepom add source --name my-gitlab --provider gitlab --url https://gitlab.com --group my-org/backend --recursive --token '${GITLAB_TOKEN}'
grepom add source --provider github --url https://github.com --org my-org
grepom add repo --name special --url https://gitlab.com/other/special.git
```

### Token 环境变量

Token 字段支持 `${ENV_VAR}` 占位符语法。实际值在运行时从环境变量中解析，并且在写入配置文件时保留占位符。

```yaml
sources:
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
| `--config` | `-c` | `.grepom.yml` | 配置文件路径 |
| `--verbose` | `-v` | `false` | 启用详细输出 |

## 构建

```bash
make build    # 构建二进制文件
make test     # 运行测试
make lint     # 运行 vet 和格式检查
make install  # 构建并安装到 ~/.local/bin
make clean    # 删除二进制文件
```
