## Why

当前 `sync` 命令同时执行"远程发现"和"克隆/拉取"操作，职责过重，用户无法仅同步元数据而不触发实际克隆。同时，配置中的 source 仅靠数组索引引用（如 `--source 0`），操作不便且可读性差。此外，token 字段在配置回写时会被展开为明文值，存在安全隐患且破坏了环境变量占位符的持久性。

需要将 sync 拆分为"轻量级元数据同步"和"实际克隆/拉取"两个独立关注点，赋予 source 可辨识的名称，并确保 token 的环境变量占位符在读写过程中保持不变。

## What Changes

- **sync 命令职责变更**：`sync` 命令仅从远程 API 读取仓库信息和子 group/org 信息，将发现的条目保存到配置文件中。不再执行 clone 或 pull 操作。clone/pull 由已有的 `clone` 和 `pull` 命令单独完成。
- **source 增加 name 字段**：`Source` 结构体新增 `name` 字段（可选），用户可通过 `--source <name>` 引用特定 source，而非仅靠数组索引。命令行参数和行为向后兼容（索引仍可用）。
- **token 环境变量占位符保持**：配置文件中的 token 字段支持 `${ENV_VAR}` 占位符语法。加载时从环境变量解析实际值，但回写配置文件时保留原始占位符字符串，不将明文 token 写入磁盘。

## Capabilities

### New Capabilities
- `sync-metadata-only`: sync 命令改为仅发现远程仓库信息并更新配置文件，不执行 clone/pull
- `source-naming`: source 支持通过 name 字段标识和引用，替代纯数组索引
- `token-env-placeholder`: token 字段的环境变量占位符在配置读写中保持不变，运行时解析

### Modified Capabilities
- `sync-command`: sync 命令的行为从"发现+克隆"变更为"仅发现并保存元数据"
- `cli-commands`: `--source` 参数支持 name 字符串（兼容原有数字索引）；add source 命令支持 `--name` 参数

## Impact

- **配置文件格式**：`.grepom.yml` 的 `sources` 条目新增可选的 `name` 字段；token 值可使用 `${ENV_VAR}` 占位符
- **命令行接口**：`sync` 命令不再执行 clone/pull，输出信息相应调整；`--source` 参数行为扩展
- **代码模块**：`config/config.go`（Source 结构体、加载/回写逻辑）、`cmd/sync.go`（移除 clone/pull 逻辑）、`cmd/add.go`（新增 --name flag）
- **向后兼容**：无 `name` 字段的 source 仍通过索引访问；token 直接写值仍可用；sync 命令参数不变，行为变更需用户注意
