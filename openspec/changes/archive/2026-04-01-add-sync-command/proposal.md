## Why

当前 grepom 的 `clone` 和 `pull` 命令分别处理仓库的首次拉取和更新，但缺少一个统一的"同步"操作：当 GitLab group 或 GitHub org 中新增了仓库时，用户需要手动发现并 clone，无法一步完成"发现新仓库 + clone 新仓库 + pull 已有仓库"。此外，sync 还应能将远程新发现的 group/org 信息（如子 group）更新到配置文件中，保持配置与远程的一致性。同步时只应新增配置，不应删除用户已有的配置。

## What Changes

- 新增 `grepom sync` 命令，对指定 group/org 执行同步操作：
  - 从远程 API 发现仓库列表
  - 对比本地已有仓库，clone 新仓库、pull 已有仓库
  - 对比配置文件中的 group/org 定义，将远程新发现的子 group 追加到配置（仅新增，不删除）
- 新增 `--source` flag 指定同步的 source（provider 实例）
- 新增 `--group` / `--org` flag 指定同步的 group/org（可选，默认同步该 source 下所有 group/org）
- 配置文件更新采用"只增不删"策略：新增的子 group/org 追加到对应 source 条目，已有的条目保持不变
- 使用文件锁（`os.File` + `flock`）防止并发写入配置文件冲突

## Capabilities

### New Capabilities

- `sync-command`: `grepom sync` 命令实现，包括仓库同步（clone + pull）和配置文件更新（新增 group/org）

### Modified Capabilities

- `cli-commands`: 新增 `sync` 子命令注册

## Impact

- **命令行接口**: 新增 `grepom sync` 命令，无 breaking change
- **代码**: 新增 `cmd/sync.go`，可能需要扩展 `config/config.go`（追加 group/org 的函数）和 `provider/` 接口（发现子 group/org 的能力）
- **配置文件**: sync 操作会修改 `.grepom.yml`，追加新发现的 group/org 条目
- **并发安全**: 需要文件锁机制防止多个 sync 实例同时写入配置文件
