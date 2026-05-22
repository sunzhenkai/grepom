## Why

现有 `dedup` 命令只能按 name 跨组去重，存在两个缺失场景：(1) 同一 group 内 URL 相同的重复仓库无法被检测和清理；(2) 不同 group 之间 URL 相同（指向同一远程仓库）的情况只会在同名时被捕获，遗漏了"同名不同 URL 误判"和"不同名同 URL 漏判"的问题。组内重复通常由手动编辑 YAML 或 URL 格式不一致引入，跨组 URL 重复则可能造成同一仓库被重复克隆到不同本地路径，浪费磁盘且造成管理混乱。

## What Changes

- 新增组内去重能力：按规范化 URL 检测同一 group 内的重复仓库，删除多余条目（保留第一个），不加入 exclude_repos
- 新增跨组 URL 警告能力：按规范化 URL 检测不同 group 之间的相同仓库，只打印警告不删除、不影响退出码
- 扩展 dedup 命令执行流程：Step 1 组内去重 + Step 2 跨组警告始终执行，Step 3 原有按 name 跨组排除仅在同时指定 `--group` + `--reference` 时触发
- `--group` 从必选变为可选：不指定时检查所有 group 的组内去重和跨组警告
- 新增 URL 规范化工具函数（去掉协议前缀、.git 后缀，统一 SSH/HTTPS 格式为 host/path）

## Capabilities

### New Capabilities
- `url-normalization`: URL 规范化工具函数，将多种格式的 Git 仓库 URL 统一为 host/path 格式用于比较
- `dedup-intra-group`: 按 URL 检测并清理同一 group 内的重复仓库条目
- `dedup-cross-warning`: 按 URL 检测不同 group 之间的相同仓库，输出警告信息

### Modified Capabilities
- `dedup-command`: `--group` 从必选变为可选；命令执行流程从单一按 name 跨组排除变为三步流程（组内去重 → 跨组警告 → 按 name 跨组排除）

## Impact

- `config/config.go`: 新增 URL 规范化函数和组内去重写入函数
- `cmd/dedup.go`: 重构命令逻辑，新增 Step 1 和 Step 2
- `cmd/dedup_test.go`: 新增组内去重和跨组警告的测试用例，确保旧测试继续通过
- `README.md` / `README_en.md`: 更新 dedup 命令文档
