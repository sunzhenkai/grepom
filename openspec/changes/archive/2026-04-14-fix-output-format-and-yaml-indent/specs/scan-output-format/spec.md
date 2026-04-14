## ADDED Requirements

### Requirement: 扫描结果按仓库分组展示
系统 SHALL 将扫描结果按仓库名称分组展示，每个仓库作为一个分组标题，其下的发现项以缩进列表形式展示。每个发现项 SHALL 包含文件路径、行号、规则 ID 和严重程度。

#### Scenario: 多仓库发现项分组展示
- **WHEN** 扫描发现 repo-a 和 repo-b 两个仓库中的敏感信息
- **THEN** 系统先展示 repo-a 的分组标题，其后列出该仓库的所有发现项，再展示 repo-b 的分组标题和发现项

#### Scenario: 单仓库展示
- **WHEN** 仅在一个仓库中发现敏感信息
- **THEN** 系统展示该仓库的分组标题及其下发现项

### Requirement: 过长文件路径自动截断
系统 SHALL 在表格输出中对过长的文件路径进行截断，保留路径前部分和文件名，中间用 `...` 替代。截断后的路径 SHALL 保持在合理的终端宽度内。

#### Scenario: 长路径截断
- **WHEN** 发现项的文件路径为 `repos/github/sunzhenkai/notes/computer science/cryptography/rsa.md`（超过 40 字符）
- **THEN** 输出中该路径被截断为 `repos/github/sun.../rsa.md` 形式

#### Scenario: 短路径不截断
- **WHEN** 发现项的文件路径为 `config/settings.json`（未超过 40 字符）
- **THEN** 输出中该路径保持原样

### Requirement: 汇总统计按仓库分组后展示
系统 SHALL 在所有分组展示完毕后输出汇总统计信息，包含总发现数和涉及仓库数，以及按严重程度的统计。

#### Scenario: 分组后的汇总输出
- **WHEN** 扫描完成后存在多个仓库的发现项
- **THEN** 系统在所有分组后输出 "Found N findings in M repos." 并附按严重程度的统计（如 "2 CRITICAL, 5 HIGH, 3 MEDIUM"）
