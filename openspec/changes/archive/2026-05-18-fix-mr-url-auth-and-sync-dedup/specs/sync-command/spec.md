## MODIFIED Requirements

### Requirement: sync 配置更新策略（只增不删）
sync 命令在更新配置文件时 SHALL 仅追加新发现的 repo 条目到对应 group 的 repos 列表，不删除或修改已有条目。sync 命令 SHALL 在写入配置时保留 group 的 `exclude_repos` 列表。sync 命令 SHALL 在写入前过滤掉匹配 group `exclude_repos` 的新发现仓库。sync 命令 SHALL 对单次批次内的新增仓库进行自去重，确保同一 URL 的仓库在批次内不重复写入。

#### Scenario: 远程新增仓库
- **WHEN** group `frontend` 下远程新增了 repo `new-app`
- **THEN** 系统将 `new-app` 追加到 group `frontend` 的 repos 列表

#### Scenario: Repo 已存在于 group repos
- **WHEN** 远程 repo `shared-utils` 已存在于 group `frontend` 的 repos 列表中（按 URL 匹配）
- **THEN** 系统不重复追加

#### Scenario: 批次内包含重复 URL 的仓库
- **WHEN** 单次 sync 从 provider 获取的仓库列表中，同一 URL 出现多次（如 `https://gitlab.com/org/app.git` 出现两次）
- **THEN** 系统 SHALL 仅保留第一个出现的条目，不将重复条目写入配置文件

#### Scenario: 远程仓库被删除不影响配置
- **WHEN** 远程某仓库已被删除，但配置文件中对应 group 的 repos 仍有该条目
- **THEN** 系统不从配置中删除该 repo 条目

#### Scenario: 非 recursive group 仅发现直接项目
- **WHEN** GitLab group 配置为 `recursive: false`（或未设置）
- **THEN** 系统仅发现该 group 直接包含的项目，不递归子 group

#### Scenario: sync 保留 exclude_repos 配置
- **WHEN** group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`，sync 发现新 repo 并写入配置
- **THEN** 写入后的配置中 `exclude_repos` 仍为 `["deprecated-app"]`，不被清空或覆盖

#### Scenario: sync 不写入被排除的仓库
- **WHEN** group `frontend` 的 `exclude_repos` 为 `["deprecated-app"]`，远程新发现了 `deprecated-app`
- **THEN** `deprecated-app` 不被写入配置文件

#### Scenario: 已在配置中的被排除仓库不受 sync 影响
- **WHEN** group `frontend` 的 repos 中已有 `deprecated-app`，`exclude_repos` 为 `["deprecated-app"]`
- **THEN** sync 不删除已有的 `deprecated-app` 条目（sync 只增不删策略不变）
