## Why

当前 `grepom scan` 命令必须依赖配置文件才能运行（需要通过 `loadConfig()` 解析仓库列表），这在非 grepom 管理的项目目录中使用不便。此外，用户在推送代码前缺少一个自动化的安全检查关口——如果代码中包含泄露的敏感信息（API Key、密码等），很容易被直接推送到远程仓库。需要一个 `push` 命令在推送前自动拦截含有敏感信息的推送，防止密钥泄露。

## What Changes

- **`grepom scan` 命令增强**：当没有配置文件时，不再报错退出，而是扫描当前工作目录下的文件，使用与现有 scan 相同的 gitleaks 引擎和输出格式
- **新增 `grepom push` 命令**：与配置无关的独立命令，执行以下流程：
  1. 先在当前目录自动执行 scan（扫描工作区文件）
  2. 如果发现敏感信息，拒绝执行 push，并打印发现的敏感信息详情
  3. 支持 `-f` / `--force` 标志强制推送，但即使强制推送也要打印发现的敏感信息作为警告

## Capabilities

### New Capabilities
- `scan-without-config`: scan 命令在无配置文件时扫描当前目录的能力
- `push-guard`: push 命令在推送前自动扫描敏感信息并拦截/警告的能力

### Modified Capabilities
- `secret-scanning`: scan 命令行为扩展——无配置文件时自动扫描当前目录而非报错

## Impact

- **cmd/scan.go**: 修改 `runScan` 函数，当配置文件不存在时回退到扫描当前目录模式
- **cmd/push.go**: 新增文件，实现 `push` 子命令
- **cmd/root.go**: 无需修改（push 命令自行注册）
- **scanner 包**: 无需修改，复用现有的 `ScanDir` 方法
- **config 包**: 可能需要提供一种"无需配置"的判断方式（如 `FindConfig` 返回特定错误类型）
- **依赖**: 无新外部依赖，复用现有 scanner 和 git 包
