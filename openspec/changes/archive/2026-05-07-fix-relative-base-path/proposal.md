## Why

当 `.grepom.yml` 中的 `base` 字段配置为相对路径（如 `repos/feg/algogear-bidder`）时，`grepom dir` 输出的仓库路径也是相对路径。用户在项目子目录中执行 `gcd <repo>` 时，shell 的 `cd` 会将相对路径解释为相对于当前工作目录，导致跳转失败。从项目根目录执行时因为路径恰好存在而不暴露问题，但进入子目录后必现。

## What Changes

- `config.Load` 加载配置后，若 `cfg.Base` 为相对路径，将其解析为绝对路径（相对于配置文件所在目录）
- 确保 `grepom dir` 始终输出绝对路径，不受用户 `cwd` 影响
- 新增测试覆盖：使用相对路径 `base` 配置 + 从子目录调用 `grepom dir` 的场景

## Capabilities

### New Capabilities

- `absolute-base-path`: 配置加载后自动将相对 `base` 解析为绝对路径，确保所有路径输出始终为绝对路径

### Modified Capabilities

- `dir-search-priority`: 需补充场景——当 `base` 为相对路径时，`grepom dir` 仍应输出绝对路径

## Impact

- **代码**: `config/config.go`（`Load` 函数或新增导出方法）、`cmd/dir.go`（可能需要传入配置文件路径）
- **测试**: `cmd/dir_test.go` 新增相对路径 base 的测试场景
- **向后兼容**: 完全兼容，仅修复了路径解析行为，不改变任何 API 或配置格式
