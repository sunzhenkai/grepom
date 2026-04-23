## Why

当前 `FindConfig` 只在当前目录查找 `.grepom.yml`，如果用户在项目子目录（如 `~/projects/my-org/web-app`）运行 grepom，必须手动 `-c` 指定配置文件路径。这与 git 在父目录查找 `.git` 的直觉不一致，增加了日常使用的心智负担。同时，grepom 缺少一个快速查询项目本地路径的命令，用户无法方便地跳转到目标仓库。

## What Changes

- **配置文件向上探测**：`FindConfig` 在当前目录未找到 `.grepom.yml` 时，自动沿父目录向上遍历查找，直到找到配置文件或到达文件系统根目录。行为与 git 查找 `.git` 完全一致。
- **新增 `dir` 命令**：输出仓库或 base 目录的本地路径到 stdout，支持模糊搜索（子串匹配）。用户可通过 `cd "$(grepom dir web-app)"` 快速跳转。

## Capabilities

### New Capabilities
- `config-upward-search`: 配置文件查找逻辑改为从当前目录沿父目录链向上遍历，直到找到 `.grepom.yml` 或到达文件系统根
- `dir-command`: 新增 `dir` 命令，输出 base 目录或指定仓库的本地路径，支持模糊搜索

### Modified Capabilities
- `config-path-resolution`: 配置查找策略从"仅当前目录"变为"当前目录+向上遍历"，影响所有依赖 `FindConfig` 的命令

## Impact

- **代码影响**：`config/config.go` 的 `FindConfig` 函数逻辑变更；`cmd/` 下新增 `dir.go`
- **命令行影响**：所有现有命令自动受益于向上探测，无需额外参数；新增 `grepom dir` 子命令
- **兼容性**：完全向后兼容，显式 `-c` 参数优先级最高，向上探测仅在未指定 `-c` 且当前目录无配置时生效
- **文档影响**：README（中英文）需更新，新增 `dir` 命令说明，更新配置文件查找逻辑描述
