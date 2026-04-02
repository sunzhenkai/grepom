## ADDED Requirements

### Requirement: 多策略 TTY 检测
`isStdoutTerminal()` SHALL 使用多层级检测策略判断 stdout 是否连接到终端，按优先级依次尝试，任一策略返回 true 即判定为 TTY。

#### Scenario: go-isatty 检测成功时直接返回 true
- **WHEN** `go-isatty.IsTerminal(os.Stdout.Fd())` 返回 true
- **THEN** `isStdoutTerminal()` 返回 true，不进行后续检测

#### Scenario: go-isatty 失败时回退到 TERM 环境变量检查
- **WHEN** `go-isatty.IsTerminal()` 返回 false，且 `TERM` 环境变量非空且不等于 `dumb` 或 `unknown`
- **THEN** `isStdoutTerminal()` 返回 true

#### Scenario: TERM 环境变量为 dumb 时判定为 non-TTY
- **WHEN** `go-isatty` 检测失败，且 `TERM` 环境变量为 `dumb` 或 `unknown`
- **THEN** `isStdoutTerminal()` 返回 false

#### Scenario: TERM 环境变量未设置时回退到 /proc 检查（Linux）
- **WHEN** `go-isatty` 检测失败，且 `TERM` 环境变量为空，系统为 Linux
- **THEN** 系统通过 stat 检查 `/proc/self/fd/1` 的 inode，如果指向 tty 设备（major number = 5），则返回 true

#### Scenario: 所有检测策略均失败时判定为 non-TTY
- **WHEN** `go-isatty` 检测失败，`TERM` 为空，且 `/proc` 检查也不支持或失败
- **THEN** `isStdoutTerminal()` 返回 false

### Requirement: TTY 检测日志（verbose 模式）
在 verbose 模式下，`isStdoutTerminal()` SHALL 记录各层检测的结果，便于排查终端检测问题。

#### Scenario: verbose 模式下输出检测详情
- **WHEN** verbose 模式启用，且 `isStdoutTerminal()` 执行检测
- **THEN** 系统输出每一层检测的结果（`[verbose] TTY check: isatty=false, TERM=xterm-256color, /proc=true → TTY`）
