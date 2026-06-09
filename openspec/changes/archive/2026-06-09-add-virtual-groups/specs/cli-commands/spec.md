## ADDED Requirements

### Requirement: 常用仓库命令支持 --vgroup
支持 `--group` 过滤真实 group 的常用仓库命令 SHALL 同时支持 `--vgroup` 过滤虚拟分组。适用命令包括 `clone`、`list`、`status`、`pull`、`scan`。`--vgroup` SHALL 展开为虚拟分组包含的真实 groups，并与 `--group` 取并集后参与仓库过滤。

#### Scenario: clone 使用 --vgroup
- **WHEN** 用户运行 `grepom clone --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 仅 clone `frontend` 和 `backend` 下符合条件的仓库

#### Scenario: list 使用 --vgroup
- **WHEN** 用户运行 `grepom list --vgroup work`
- **THEN** 系统 SHALL 仅列出虚拟分组 `work` 包含的真实 groups 下的仓库

#### Scenario: status 使用 --vgroup
- **WHEN** 用户运行 `grepom status --vgroup work`
- **THEN** 系统 SHALL 仅统计并展示虚拟分组 `work` 包含的真实 groups 下的仓库状态

#### Scenario: pull 使用 --vgroup
- **WHEN** 用户运行 `grepom pull --vgroup work`
- **THEN** 系统 SHALL 仅对虚拟分组 `work` 包含的真实 groups 下符合安全检查的仓库执行 pull

#### Scenario: scan 使用 --vgroup
- **WHEN** 用户运行 `grepom scan --vgroup work`
- **THEN** 系统 SHALL 仅扫描虚拟分组 `work` 包含的真实 groups 下的仓库

#### Scenario: --group 与 --vgroup 并集
- **WHEN** 用户运行 `grepom list --group infra --vgroup work`，虚拟分组 `work` 包含真实 groups `frontend` 和 `backend`
- **THEN** 系统 SHALL 列出真实 groups `infra`、`frontend` 和 `backend` 下的仓库

#### Scenario: --vgroup 短参数不存在
- **WHEN** 用户查看任一支持 `--vgroup` 的命令帮助
- **THEN** 系统 SHALL 展示长参数 `--vgroup`，不要求提供短参数别名
