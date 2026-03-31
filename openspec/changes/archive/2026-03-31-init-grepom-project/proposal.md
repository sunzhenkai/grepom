## Why

管理多个 git 仓库（尤其是跨 GitLab group/subgroup 和 GitHub org）是日常开发中的重复劳动。缺少一个统一工具来：按组织结构批量 clone/pull 仓库、查看所有仓库状态、通过配置文件复用仓库组织方式。grepom 解决这个问题，提供声明式配置 + API 自动发现 + 统一 CLI 操作。

## What Changes

- 初始化 Go CLI 项目骨架（cobra 框架、模块结构）
- 实现 YAML 配置解析，支持多配置文件（`-c` flag）、环境变量引用、API 源定义
- 实现 GitLab/GitHub provider，通过 API 递归获取 group/subgroup/org 下的仓库列表
- 实现仓库路径解析，将 group/subgroup 嵌套层级映射为本地目录嵌套结构
- 实现 5 个核心子命令：`add`、`init`、`list`、`status`、`pull`，均支持 one/all 模式
- 实现 git 操作封装（clone、pull、status）

## Capabilities

### New Capabilities

- `config-management`: YAML 配置文件解析、多文件支持、环境变量引用、配置结构定义
- `provider-api`: GitLab/GitHub API 客户端，递归获取 group/subgroup/org 下的仓库列表
- `repo-resolution`: 仓库模型定义、路径解析（group/subgroup 嵌套映射为目录结构）、仓库列表聚合
- `cli-commands`: cobra 子命令实现（add/init/list/status/pull），支持 one/all/target 过滤模式
- `git-operations`: git 命令封装（clone/pull/status），提供统一的 git 操作接口

### Modified Capabilities

（无，新项目）

## Impact

- **代码**: 全新 Go 项目，无现有代码受影响
- **依赖**: cobra (CLI)、yaml.v3 (配置解析)、GitLab/GitHub Go SDK 或 HTTP 客户端
- **外部系统**: 调用 GitLab API v4、GitHub REST API
- **用户**: 需要提供 API token（通过环境变量），配置 YAML 文件
