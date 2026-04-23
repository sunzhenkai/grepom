## MODIFIED Requirements

### Requirement: push 命令在推送前自动扫描当前目录
系统 SHALL 提供 `grepom push` 命令。命令的 cobra 描述（Short、Long、Example）、flag 帮助文本、错误信息和用户可见输出 SHALL 全部使用英文。

#### Scenario: push 命令帮助信息为英文
- **WHEN** 用户运行 `grepom push --help`
- **THEN** 系统显示英文的命令描述、flag 说明和示例

#### Scenario: push 错误信息为英文
- **WHEN** push 过程中发生错误（如当前目录不是 git 仓库、扫描失败）
- **THEN** 系统输出英文错误信息（如 "not a git repository"、"scan failed: ..."）

#### Scenario: 当前目录不是 git 仓库时报错
- **WHEN** 用户在非 git 仓库目录中执行 `grepom push`
- **THEN** 系统输出英文错误信息 "not a git repository" 并以非零退出码退出
