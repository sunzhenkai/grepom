## 1. list --remote 远程仓库查询

- [x] 1.1 在 `cmd/list.go` 中新增 `--remote` 布尔标志（默认 false），注册到 `listCmd`
- [x] 1.2 在 `listCmd.RunE` 中添加 `--remote` 与 `--type resources`/`--type groups` 的互斥校验，不兼容时返回错误
- [x] 1.3 实现 `runListRemoteRepos` 函数：遍历配置中的 groups，根据 `--group`/`--resource` 过滤，调用 `provider.ListRepos` 获取远程仓库列表
- [x] 1.4 远程仓库以表格形式输出，列：`NAME`、`PATH`、`GROUP`、`RESOURCE`、`CLONE_URL`
- [x] 1.5 处理 API 错误：单个 group 查询失败时输出 stderr 错误并继续查询其他 groups
- [x] 1.6 处理无结果情况：远程查询无仓库时输出 `No remote repositories found.`

## 2. status 头部摘要表格化

- [x] 2.1 修改 `cmd/status.go` 中头部摘要的输出逻辑，将单行文本改为使用 `text/tabwriter` 输出两列表格（STATUS / COUNT）
- [x] 2.2 表格仅显示数量 > 0 的状态行（clean、dirty、ahead、behind、not cloned），表格下方显示总 repo 数
- [x] 2.3 验证 `--group`/`--resource` 过滤时概要表格仅统计过滤后的 repo
