## MODIFIED Requirements

### Requirement: 扫描结果以终端表格形式展示
系统 SHALL 将扫描结果按仓库名称分组展示。每个仓库作为分组标题，其下的发现项以缩进列表形式展示，每行包含文件路径（过长时自动截断）、行号、规则 ID 和严重程度。表格后 SHALL 输出汇总统计（按严重程度计数）。敏感信息 SHALL 脱敏显示（前 8 字符 + `...`）。

#### Scenario: 发现敏感信息时的输出
- **WHEN** 扫描发现敏感信息
- **THEN** 系统按仓库分组输出，每行一个发现项，包含文件路径（过长截断）、行号、规则 ID、严重程度和脱敏后的 secret 片段

#### Scenario: 无发现时的输出
- **WHEN** 扫描完成且未发现任何敏感信息
- **THEN** 系统输出 "No secrets found."

#### Scenario: 汇总统计
- **WHEN** 扫描完成后存在发现项
- **THEN** 系统在所有分组后输出总计行，格式为 "Found N findings in M repos." 并附按严重程度的统计
