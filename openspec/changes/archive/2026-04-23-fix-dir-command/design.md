## Context

`grepom dir` 是用户最频繁使用的命令之一，主要用于 `cd "$(grepom dir xxx)"` 跳转到仓库目录。当前实现有三个问题：

1. **无参数语义错误**：输出 `cfg.Base`（配置中的 `base` 字段，如 `./repos`），但用户期望的是配置文件所在目录
2. **多匹配阻断**：多匹配时 stderr 列表 + exit 1，无法配合 fzf 交互选择
3. **搜索无优先级**：纯子串匹配，当有精确同名仓库时仍被子串匹配淹没

当前代码结构：
- `cmd/dir.go`：命令定义，shell function 模板
- `repo/resolver.go`：`ApplySearchFilter` 实现子串匹配
- Go 端 `fzfAvailable()` 在 `--shell` 时编译时检测 fzf

## Goals / Non-Goals

**Goals:**

- `grepom dir` 无参数输出配置文件所在目录
- 搜索策略：先精确匹配、后子串匹配（case-insensitive）
- 多匹配时输出所有路径到 stdout（每行一个），退出码 0
- `--shell` 输出固定单一 gcd() 函数，shell 端运行时检测 fzf

**Non-Goals:**

- 不修改 `cfg.Base` 的语义或用途
- 不修改其他命令（list、clone 等）的匹配行为
- 不引入新的依赖

## Decisions

### D1: 无参数时返回配置文件所在目录

`loadConfig()` 已经通过 `config.FindConfig()` 获取了配置文件路径。只需在 `dirCmd.RunE` 中保存该路径，无参数时输出 `filepath.Dir(configPath)` 即可。

需要修改 `loadConfig` 使其同时返回配置文件路径，或在 `dirCmd` 中单独调用 `FindConfig`。

**选择**：让 `loadConfig` 返回 `(path, config, error)` 三元组，因为其他命令未来也可能需要配置文件路径。

### D2: 多匹配输出到 stdout

当前多匹配时写到 stderr 并返回 error。改为：所有匹配路径写到 stdout（每行一个），退出码 0。无匹配时仍然 stderr + exit 1。

这让 `fzf` 和 `head -n 1` 能自然接住输出。调用方通过行数判断是否唯一匹配。

### D3: 搜索优先级——先精确后子串

修改 `ApplySearchFilter` 或在 `dirCmd` 中增加两阶段搜索：
1. 先用 case-insensitive **精确匹配**（`strings.ToLower(name) == keyword`）
2. 如果精确匹配有结果，直接使用（多个精确匹配也输出多行）
3. 如果精确匹配无结果，退回子串匹配

**选择**：新增 `ApplyExactFirstSearch` 函数（或在 `ApplySearchFilter` 中增加优先级参数），保持 `ApplySearchFilter` 原有行为不变，避免影响其他命令。

### D4: Shell function 运行时检测 fzf

```bash
gcd() {
  local dir
  if [ $# -eq 0 ]; then
    dir=$(grepom dir)
  else
    if command -v fzf >/dev/null 2>&1; then
      dir=$(grepom dir "$@" | fzf --select-1)
    else
      dir=$(grepom dir "$@" | head -n 1)
    fi
  fi || return
  cd "$dir"
}
```

移除 Go 端 `fzfAvailable()`、`shellHelperWithoutFzf`，只保留一个 `shellHelper` 模板。

`--select-1` 在 fzf 中的作用：stdin 只有 1 行时自动选择，无需交互。当 `grepom dir` 输出多行时 fzf 弹出选择 UI。

## Risks / Trade-offs

- **[BREAKING] 多匹配行为变更**：依赖 `grepom dir` 多匹配时报错的脚本会中断 → 影响范围小，`grepom dir` 主要用于交互式 cd
- **[BREAKING] 无参数行为变更**：脚本中 `cd "$(grepom dir)"` 的目标目录会变 → 变更后更符合直觉，且迁移成本低（只是 cd 目标变了）
- **无 fzf 环境下 `head -n 1` 静默选择第一个**：可能不是用户想要的 → 可接受，用户可自行安装 fzf 或用更精确的关键字
