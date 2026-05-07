## Context

当前 `config.Load` 加载 `.grepom.yml` 时，对 `base` 字段仅执行 `expandTilde`（处理 `~/` 前缀）。若 `base` 为相对路径（如 `repos/feg/algogear-bidder`），所有后续路径计算（`ResolveGroupRepoPath`、`ResolveRepoPath`）都产出相对路径，导致 `grepom dir` 输出相对路径。

关键调用链：

```
loadConfig() → config.Load(absPath) → cfg.Base 仍是相对路径
                  ↓
dirCmd → FullPath(cfg.Base, r) → filepath.Clean(r.Path) → 相对路径输出到 stdout
                  ↓
gcd() → cd "$dir" → 从子目录执行时路径解析失败
```

## Goals / Non-Goals

**Goals:**
- 无论 `.grepom.yml` 中 `base` 配置为绝对路径、`~` 路径还是相对路径，`grepom dir` 始终输出绝对路径
- 修复在项目子目录中使用 `gcd` 跳转失败的问题
- 保持向后兼容，不改变配置文件格式

**Non-Goals:**
- 不修改 shell 函数 `gcd()` 的实现
- 不改变配置文件 `.grepom.yml` 的 schema
- 不处理 `base` 路径不存在的情况（这属于配置校验范畴）

## Decisions

### Decision 1: 在 config.Load 中增加 ResolveBasePath 方法

**选择**: 在 `config` 包中新增 `ResolveBasePath(cfg *Config, configDir string)` 方法，在 `loadConfig` 中调用。

**理由**: 
- `config.Load` 目前不接受配置文件路径参数，修改其签名是 breaking change
- 通过独立的导出方法，调用方（`loadConfig`）可以在加载后立即解析
- 保持 `Load` 函数职责单一（只做 YAML 解析和校验）

**备选方案**:
- ~~修改 `Load` 签名增加 configPath 参数~~ — 影响所有调用方
- ~~在 `dirCmd.RunE` 中用 `filepath.Abs` 处理输出~~ — 治标不治本，其他命令（clone、pull 等）也可能受影响

### Decision 2: 相对 base 的参照点为配置文件所在目录

**选择**: 若 `base` 为相对路径，相对于 `.grepom.yml` 所在目录解析。

**理由**: 配置文件通常在项目根目录，`base` 的相对路径含义应是"相对于配置文件位置"。这与 `loadConfig` 已有的 `filepath.Abs(path)` 获取配置文件绝对路径的流程自然衔接。

## Risks / Trade-offs

- **[已有绝对路径或 ~ 路径不受影响]** → `ResolveBasePath` 仅在 `!filepath.IsAbs(cfg.Base)` 时介入，`expandTilde` 已处理 `~/`，两种场景互不干扰
- **[测试未覆盖相对 base]** → 需要新增专门的测试用例，确保从子目录调用和从根目录调用都正确
