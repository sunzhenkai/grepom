## Why

`grepom dir` 命令存在三个行为缺陷，导致日常使用体验不佳：
1. 无参数时输出 `cfg.Base`（仓库存储根目录），而非配置文件所在目录——用户 cd 到的应该是工作区根目录，不是 repos 子目录
2. 多个匹配时直接报错退出，无法配合 fzf 等工具进行交互式选择
3. 搜索策略为纯子串匹配，缺乏精确匹配优先级——当存在精确匹配的仓库名时仍可能被子串匹配干扰

## What Changes

- **BREAKING**: `grepom dir` 无参数行为变更——输出配置文件所在目录（`.grepom.yml` 的父目录），而非 `cfg.Base`
- **BREAKING**: `grepom dir <keyword>` 多匹配时输出所有匹配路径到 stdout（每行一个），退出码 0；当前行为是 stderr 列表 + 退出码 1
- 搜索策略改为**先精确匹配、后子串匹配**（均为 case-insensitive）
- `--shell` 输出固定的单一 shell function，由 shell 端运行时检测 fzf，移除 Go 端的 `fzfAvailable()` 编译时检测

## Capabilities

### New Capabilities

- `dir-search-priority`: `grepom dir` 的搜索策略定义——精确匹配优先、子串匹配兜底、多匹配输出行为

### Modified Capabilities

（无既有 spec 需要修改）

## Impact

- **代码**: `cmd/dir.go`（主要改动）、`repo/resolver.go`（`ApplySearchFilter` 函数）
- **测试**: `cmd/dir_test.go`（多匹配用例需更新、新增精确优先用例）
- **Shell 集成**: `--shell` 输出的 gcd() 函数签名变更，已 eval 的用户需重新 eval
- **下游影响**: 依赖 `grepom dir` 多匹配时报错行为的脚本会受影响（BREAKING）
