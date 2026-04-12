### Requirement: Group 内 repo 路径自动推导
Group 内 repo 的本地完整路径 SHALL 通过公式自动推导：`base + group.local_path + trimPrefix(repo.path, group.path)`。Group 内 repo 不存储 `local_path` 字段。无 resource 的 group，其 repo 的 url 字段 SHALL 由用户手动填写完整 clone URL。

#### Scenario: 直接子项目
- **WHEN** group path 为 `my-org/frontend`，local_path 为 `./frontend`，repo path 为 `my-org/frontend/shared-utils`
- **THEN** 本地路径推导为 `<base>/frontend/shared-utils`

#### Scenario: 多级 subgroup 下的项目
- **WHEN** group path 为 `my-org/frontend`，local_path 为 `./frontend`，repo path 为 `my-org/frontend/ui/design-system`
- **THEN** 本地路径推导为 `<base>/frontend/ui/design-system`

#### Scenario: 三级 subgroup 下的项目
- **WHEN** group path 为 `my-org/frontend`，local_path 为 `./frontend`，repo path 为 `my-org/frontend/ui/components/button-lib`
- **THEN** 本地路径推导为 `<base>/frontend/ui/components/button-lib`

#### Scenario: repo path 恰好等于 group path
- **WHEN** group path 为 `my-org/frontend`，repo path 也为 `my-org/frontend`（项目名和 group 同名）
- **THEN** 本地路径推导为 `<base>/frontend/`（trimPrefix 后为空，路径为 group local_path 本身）

#### Scenario: 无 resource 的 group 内 repo 使用手动 url
- **WHEN** group 未绑定 resource，其 repo 的 url 为 `git@github.com:org/repo.git`
- **THEN** clone/pull 时直接使用该 url，不从 resource 推导

### Requirement: 独立 repo 路径解析
不属于任何 group 的独立 repo，本地完整路径为 `base + repo.local_path`。`local_path` 可省略，默认为 `./<name>`。

#### Scenario: 独立 repo 指定 local_path
- **WHEN** 独立 repo 的 `local_path` 为 `./dotfiles`，base 为 `~/projects`
- **THEN** 本地完整路径为 `~/projects/dotfiles`

#### Scenario: 独立 repo 省略 local_path
- **WHEN** 独立 repo 的 name 为 `dotfiles`，未指定 `local_path`
- **THEN** 本地完整路径为 `~/projects/dotfiles`（默认 `./dotfiles`）

#### Scenario: 独立 repo 的 local_path 包含子目录
- **WHEN** 独立 repo 的 `local_path` 为 `./tools/special`
- **THEN** 本地完整路径为 `<base>/tools/special`

### Requirement: 本地路径冲突检测
系统 SHALL 在写入配置时检测所有 repo 的本地路径是否冲突。不同 repo（无论是否属于同一 group 或同一 resource）不能映射到同一个本地目录。

#### Scenario: 两个不同 group 的 repo 路径冲突
- **WHEN** group A 的某 repo 推导路径为 `./frontend/lib`，group B 的某 repo 推导路径也为 `./frontend/lib`
- **THEN** 系统 SHALL 拒绝写入并报错，提示路径冲突

#### Scenario: group repo 与独立 repo 路径冲突
- **WHEN** group 内某 repo 推导路径为 `./tools`，独立 repo 的 local_path 也为 `./tools`
- **THEN** 系统 SHALL 拒绝写入并报错

#### Scenario: 同一 group 内两个 repo 路径冲突
- **WHEN** 同一 group 内两个 repo 推导出相同的本地路径
- **THEN** 系统 SHALL 报错（这种情况通常意味着远端 path 重复）

#### Scenario: 无冲突时正常写入
- **WHEN** 所有 repo 的本地路径互不相同
- **THEN** 系统正常写入配置文件

### Requirement: 路径规范化
路径推导前 SHALL 对路径进行规范化处理：去除前缀 `./` 用于拼接，处理 `../` 等相对路径标记，统一使用 filepath.Join 拼接。

#### Scenario: local_path 包含 ./
- **WHEN** group 的 local_path 为 `./frontend`
- **THEN** 系统正确解析，等价于 `frontend`

#### Scenario: local_path 包含多级
- **WHEN** group 的 local_path 为 `./work/frontend`
- **THEN** 系统正确解析为 `<base>/work/frontend`
