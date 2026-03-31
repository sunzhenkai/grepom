### Requirement: sync 命令
系统 SHALL 提供 `grepom sync` 命令，对配置中的 group/org 执行同步操作：从远程 API 发现仓库，clone 新仓库、pull 已有仓库，并将远程新发现的子 group/org 追加到配置文件。

#### Scenario: 同步所有 source 的所有 group/org
- **WHEN** 用户运行 `grepom sync`（无参数）
- **THEN** 系统对所有配置中的 source 下所有 group/org 执行同步，clone 新仓库、pull 已有仓库，并将新发现的子 group/org 追加到配置文件

#### Scenario: 同步指定 source
- **WHEN** 用户运行 `grepom sync --source 0`（按索引指定 source）
- **THEN** 系统仅同步该 source 下的 group/org

#### Scenario: 同步指定 group
- **WHEN** 用户运行 `grepom sync --group my-org/frontend`
- **THEN** 系统仅同步匹配该 group 路径的仓库，并检查该 group 下是否有新的子 group 需要追加

#### Scenario: 同步指定 org
- **WHEN** 用户运行 `grepom sync --org my-org`
- **THEN** 系统仅同步该 org 下的仓库

#### Scenario: 同步时无新内容
- **WHEN** 用户运行 `grepom sync` 且没有新仓库也没有新的子 group/org
- **THEN** 系统仅对已有仓库执行 pull，不修改配置文件

### Requirement: sync 配置更新策略（只增不删）
sync 命令在更新配置文件时 SHALL 仅追加新发现的 group/org 条目，不删除或修改已有条目。

#### Scenario: 远程新增子 group
- **WHEN** GitLab group `my-org`（recursive: true）下远程新增了子 group `my-org/backend`
- **THEN** 系统将 `my-org/backend`（recursive: true）追加到配置文件中对应 source 的 groups 列表

#### Scenario: 子 group 已存在于配置
- **WHEN** 远程子 group `my-org/backend` 已存在于配置文件的 groups 列表中
- **THEN** 系统不重复追加

#### Scenario: 非递归 group 不发现子 group
- **WHEN** GitLab group 配置为 `recursive: false`（或未设置）
- **THEN** 系统不检查该 group 下的子 group，不修改配置

#### Scenario: 远程仓库被删除不影响配置
- **WHEN** 远程某仓库已被删除，但配置文件中对应的 group 仍存在
- **THEN** 系统不从配置中删除该 group 条目，仅跳过该仓库

### Requirement: sync 并发写入保护
当多个 sync 实例同时运行时，系统 SHALL 使用文件锁防止配置文件的并发写入冲突。

#### Scenario: 多个 sync 实例同时运行
- **WHEN** 两个 `grepom sync` 进程同时运行并都需要写入配置文件
- **THEN** 第二个进程等待第一个进程完成写入后再执行自己的写入，不会导致配置文件损坏或数据丢失

#### Scenario: 获取锁超时
- **WHEN** sync 进程无法在合理时间内获取配置文件锁
- **THEN** 系统报告错误并退出，提示用户另一个 sync 正在运行
