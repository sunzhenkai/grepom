## 1. 重构 status 命令输出格式

- [x] 1.1 重写 `cmd/status.go` 的 `RunE` 函数：先遍历所有 repo 收集状态数据，再统一输出
- [x] 1.2 实现概要行输出：统计 clean、dirty、ahead、behind、not cloned 数量，格式如 `12 repos: 8 clean, 2 dirty, 1 ahead, 1 behind · 3 not cloned`
- [x] 1.3 实现仓库列表输出：每个 repo 一行，包含名称（`r.Name`）、状态标记、本地路径（`fullPath`），三列对齐
- [x] 1.4 实现状态标记优先级逻辑：not cloned > dirty (N) > ahead N > behind N > clean
- [x] 1.5 处理边界情况：所有 repo clean 时概要行不显示 behind/ahead；过滤后无仓库时输出 `No repositories found.`
