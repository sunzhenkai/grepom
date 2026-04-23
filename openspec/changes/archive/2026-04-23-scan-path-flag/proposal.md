## Why

`grepom scan` 目前有两个体验问题：

1. **`-p/--path` 缺失**：无配置文件时只能扫描当前目录，无法指定其他路径。用户想快速扫描某个目录时必须先 `cd` 过去，不够灵活。
2. **配置文件向上查找导致误匹配**：当用户在子目录（如 `senv/`）执行 `grepom scan` 时，`FindConfig` 会沿父目录链向上查找 `.grepom.yml`，可能找到一个属于上级工作区的配置文件，导致路径推算全部错误（所有仓库报 "not cloned"）。scan 命令不应依赖向上查找的配置——要么用当前目录的配置，要么直接扫描路径。

## What Changes

- **新增 `-p/--path` 标志**：`grepom scan -p <路径>` 直接扫描指定目录，完全忽略配置文件的存在。不与 `[name]` 位置参数同时使用。
- **简化配置查找逻辑**：`grepom scan`（无 `-p`）仅检查**当前目录**是否存在 `.grepom.yml`：
  - 存在 → 加载配置，扫描所有已克隆的仓库（支持 `[name]` 过滤）
  - 不存在 → 扫描当前目录（等价于 `grepom scan -p .`）
- **扫描开始时打印扫描目标摘要**：显示正在扫描的路径或仓库名称列表。仓库数量较多时（>5），显示前几个 + "...及 N 个仓库"。

## Capabilities

### New Capabilities
- `scan-path-flag`: scan 命令新增 `-p/--path` 标志，支持指定任意目录进行扫描；扫描开始时打印目标摘要信息

### Modified Capabilities
- `scan-without-config`: 配置查找逻辑从"向上遍历查找"改为"仅当前目录"；无配置时默认扫描当前目录的语义不变
- `secret-scanning`: scan 命令的配置查找行为变更（不再向上查找），新增扫描目标摘要输出

## Impact

- `cmd/scan.go`：主要改动文件——新增 `-p` flag、调整 `runScan` 分支逻辑、新增扫描目标摘要打印
- `README.md` / `README_en.md`：更新 scan 命令的用法说明
