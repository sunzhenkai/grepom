## Why

grepom 管理着公司项目、个人隐私项目和个人开源项目的 git 仓库集合。这些仓库中极易意外泄露敏感信息——SSH 私钥、API Token（GitHub/GitLab）、AWS AK/SK、数据库连接串、kubeconfig 等。当前缺乏一种统一、自动化的手段来批量检查所有仓库中的私密信息泄露。用户需要一个内建命令，能够按 group/repo 维度扫描，覆盖工作区和 git 历史，在泄露发生时及时发现。

## What Changes

- **新增 `grepom scan` 命令**：对指定 group、repo 或全部仓库执行敏感信息扫描
- **集成 gitleaks Go library**（`github.com/zricethezav/gitleaks/v8/detect`）：利用其成熟的规则引擎进行内容检测，覆盖 SSH 私钥、AK/SK、Token、密码、数据库连接串、kubeconfig 等数十种密钥类型
- **双模式扫描**：支持工作区文件扫描（当前 checkout 的代码）和 git 历史扫描（包括已删除提交中的泄露）
- **终端表格输出**：以结构化表格展示扫描结果（repo、文件、规则、严重程度）
- **白名单机制**：支持 `.gitleaksignore` 和自定义忽略规则，允许标记已知/合法的密钥
- **.gitignore 感知**：扫描时自动跳过 `.gitignore` 中排除的文件
- **并行扫描**：利用 grepom 已有的并行框架对多个 repo 并发扫描

## Capabilities

### New Capabilities

- `secret-scanning`: 敏感信息扫描能力——集成 gitleaks 规则引擎，支持工作区+git 历史双模式扫描，提供终端表格输出、白名单机制、.gitignore 感知和并行扫描

### Modified Capabilities

（无已有能力需要修改）

## Impact

- **新增依赖**：`github.com/zricethezav/gitleaks/v8` 及其传递依赖（aho-corasick、semgroup、zerolog、viper、lipgloss 等），go.mod 依赖树将显著膨胀
- **新增代码**：`cmd/scan.go`（CLI 入口）、`scanner/` 包（扫描引擎封装）
- **已有代码**：`repo/resolver.go`（提供 repo 列表）、`git/parallel.go`（并行框架）将被复用
- **无 breaking change**：纯新增功能，不影响已有命令行为
