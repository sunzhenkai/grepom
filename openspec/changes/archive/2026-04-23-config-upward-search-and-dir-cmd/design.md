## Context

当前 grepom 的配置文件查找逻辑（`FindConfig`）仅检查当前工作目录下的 `.grepom.yml`。用户若在项目子目录执行命令，必须通过 `-c` 显式指定配置路径。与此同时，grepom 缺少快速查询项目本地路径的能力，用户需要手动回忆或查找仓库所在目录。

现有代码结构：
- `config/config.go`：`FindConfig(explicitPath)` 负责配置查找
- `cmd/root.go`：`loadConfig()` 调用 `FindConfig`，所有子命令依赖此函数
- `repo/resolver.go`：`Resolver` 负责从配置解析所有仓库及其本地路径

## Goals / Non-Goals

**Goals:**
- 用户在项目任意子目录执行 grepom 命令时，自动沿父目录向上查找 `.grepom.yml`，无需手动 `-c`
- 提供 `grepom dir` 命令，输出 base 目录或指定仓库的本地路径
- `grepom dir <name>` 支持模糊搜索（大小写不敏感子串匹配）
- 完全向后兼容

**Non-Goals:**
- 不实现 shell 集成（如自动 cd、shell function 生成）
- 不实现交互式选择（多匹配结果时列出但不支持序号选择）
- 不验证 cwd 是否在 `base` 目录范围内

## Decisions

### Decision 1: 向上探测采用"纯遍历"策略，不做范围验证

**选择**：沿父目录链向上查找第一个 `.grepom.yml`，不检查 cwd 是否在 `base` 范围内。

**理由**：与 git 查找 `.git` 的行为一致，用户直觉熟悉。嵌套配置天然支持"最近的优先"。

**备选方案**：
- 找到后验证 cwd 是否在 base 下：更安全但限制灵活性（配置文件和 base 可能在不同位置）
- 找到后仅警告不阻止：折中但增加复杂度，收益不明显

### Decision 2: `dir` 命令仅输出路径到 stdout

**选择**：`grepom dir` 只打印路径到 stdout，错误/诊断信息输出到 stderr。

**理由**：最大灵活性，用户可自由组合 `cd "$(grepom dir web-app)"` 或其他 shell 操作。避免 shell 集成的复杂度和跨平台问题。

**备选方案**：
- 提供 `init-shell` 生成 shell function：额外维护成本，用户配置负担
- 直接 `os.Chdir`：子进程无法影响父进程 cwd，技术上不可行

### Decision 3: 模糊搜索策略——子串匹配，多结果时列清单

**选择**：`grepom dir <name>` 使用大小写不敏感子串匹配（复用现有 `ApplySearchFilter`）。多个结果时以表格列出并报错，提示用户精确指定。

**理由**：复用现有 `ApplySearchFilter` 逻辑，零额外代码。精确匹配的场景由 `list` 命令覆盖。

### Decision 4: 向上探测的实现位置

**选择**：在 `config.FindConfig` 内部实现向上遍历，不改变函数签名。

**理由**：所有命令通过 `loadConfig()` → `FindConfig()` 获取配置路径，在此处改造，所有命令自动受益，无需逐个修改。

## Risks / Trade-offs

- **[误用无关配置]** → 向上遍历可能在意外位置找到不相关的 `.grepom.yml`。但由于和 git 行为一致，用户已有此心智模型，风险可控。
- **[路径含空格]** → `cd "$(grepom dir ...)"` 的双引号是必须的。在帮助文档中明确说明。
- **[性能]** → 向上遍历最多访问 O(depth) 个目录，实际深度通常 < 20，可忽略。
