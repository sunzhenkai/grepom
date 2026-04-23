## ADDED Requirements

### Requirement: push 命令在推送前自动扫描当前目录
系统 SHALL 提供 `grepom push` 命令，在执行 `git push` 前自动扫描当前工作目录的文件。该命令不依赖 grepom 配置文件，可在任何 git 仓库中使用。命令的 cobra 描述（Short、Long、Example）、flag 帮助文本、错误信息和用户可见输出 SHALL 全部使用英文。

#### Scenario: 当前目录无敏感信息时执行推送
- **WHEN** 用户在 git 仓库中执行 `grepom push`，且扫描未发现敏感信息
- **THEN** 系统执行 `git push`，将本地提交推送到远程仓库

#### Scenario: 当前目录有敏感信息时拒绝推送
- **WHEN** 用户在 git 仓库中执行 `grepom push`，且扫描发现敏感信息
- **THEN** 系统打印所有发现的敏感信息详情（文件路径、行号、规则 ID、严重程度、脱敏 secret），拒绝执行 `git push`，并以非零退出码退出

#### Scenario: 使用 --force 强制推送时打印警告
- **WHEN** 用户在 git 仓库中执行 `grepom push -f`（或 `--force`），且扫描发现敏感信息
- **THEN** 系统打印所有发现的敏感信息详情，并在结果后显示警告信息（如 "⚠ Secrets detected but push forced"），然后继续执行 `git push`

#### Scenario: push 命令帮助信息为英文
- **WHEN** 用户运行 `grepom push --help`
- **THEN** 系统显示英文的命令描述、flag 说明和示例

#### Scenario: push 错误信息为英文
- **WHEN** push 过程中发生错误（如当前目录不是 git 仓库、扫描失败）
- **THEN** 系统输出英文错误信息（如 "not a git repository"、"scan failed: ..."）

#### Scenario: 当前目录不是 git 仓库时报错
- **WHEN** 用户在非 git 仓库目录中执行 `grepom push`
- **THEN** 系统输出英文错误信息 "not a git repository" 并以非零退出码退出

### Requirement: push 命令的扫描范围
系统 SHALL 在 push 命令中仅扫描工作区文件（不包括 git 历史），遵循当前目录的 `.gitignore` 排除规则。

#### Scenario: push 扫描遵循 .gitignore
- **WHEN** 当前目录的 `.gitignore` 包含 `node_modules/` 和 `*.log`
- **THEN** push 命令的扫描不检查 `node_modules/` 目录和 `.log` 文件

### Requirement: push 命令支持自定义 gitleaks 规则
系统 SHALL 支持 `--gitleaks-config` 标志，允许用户在 push 时指定自定义的 gitleaks 规则文件。

#### Scenario: push 使用自定义规则
- **WHEN** 用户执行 `grepom push --gitleaks-config ./strict-rules.toml`
- **THEN** 系统使用指定的自定义规则扫描当前目录

### Requirement: push 命令传递 git push 的额外参数
系统 SHALL 允许用户在 `grepom push` 后传递额外参数给底层的 `git push` 命令，例如 `grepom push -- origin main`。

#### Scenario: 传递额外参数给 git push
- **WHEN** 用户执行 `grepom push --force -- origin main`
- **THEN** 系统在安全检查通过（或强制跳过）后，执行 `git push -- origin main`

#### Scenario: 不传递额外参数
- **WHEN** 用户执行 `grepom push`
- **THEN** 系统在安全检查通过后，执行 `git push`（不带额外参数）

### Requirement: push 命令输出格式与 scan 一致
系统 SHALL 在 push 命令中扫描发现敏感信息时，使用与 `grepom scan` 相同的输出格式展示发现项（按仓库分组、文件路径截断、严重程度、secret 脱敏）。

#### Scenario: push 发现敏感信息的输出格式
- **WHEN** push 命令扫描发现敏感信息
- **THEN** 系统以与 `grepom scan` 相同的表格格式输出发现项，包含汇总统计
