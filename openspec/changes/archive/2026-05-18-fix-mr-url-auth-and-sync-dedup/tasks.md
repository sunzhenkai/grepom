## 1. 修复 extractHost() — HTTPS/HTTP URL userinfo 剥离

- [x] 1.1 在 `cmd/mr.go` 的 `extractHost()` 函数中，修改 HTTPS/HTTP scheme 处理分支（约第 335-343 行）：去除 scheme 后，在查找第一个 `/` 之前，先搜索 `@` 并剥离 `@` 及其之前的内容（条件：`@` 必须位于第一个 `/` 之前）

- [x] 1.2 在 `cmd/mr.go` 的 `extractHost()` 函数中，同步修改 `ssh://` scheme 处理分支（约第 346-356 行）：去除 `ssh://` 后将 `user@` 剥离的逻辑移至 scheme 剥离后、host 提取之前

- [x] 1.3 为 `extractHost()` 编写单元测试，覆盖以下场景：带 oauth2 token 的 HTTPS URL、带 username:password 的 HTTPS URL、仅 username 的 HTTPS URL、带 username 的 ssh:// URL、无 userinfo 的标准 URL（向后兼容）、SCP 风格 URL（git@host:path）

## 2. 修复 sync 组内去重

- [x] 2.1 在 `cmd/sync.go` 的 `newGroupRepos` 构建循环中（约第 136-146 行），增加对已追加仓库的 URL 检查：在 append 前遍历 `newGroupRepos`，若已存在相同 URL 则跳过

- [x] 2.2 在 `config/config.go` 的 `SyncGroupRepos()` 函数中（约第 558-592 行），为 `newRepos` 增加批次内自去重逻辑：使用 `map[string]bool` 记录已处理的 URL，在循环中跳过已见过的条目

- [x] 2.3 为 `SyncGroupRepos()` 编写单元测试，覆盖：`newRepos` 包含重复 URL 时仅写入一次、重复 URL 与已有仓库 URL 相同时不追加、无重复时正常全部写入

## 3. 验证与回归

- [x] 3.1 运行完整测试套件 `go test ./...`，确保所有现有测试通过

- [x] 3.2 手动验证 `grepom mr` 在包含 `https://oauth2:TOKEN@gitlab.company.com/...` remote URL 的仓库中能正确识别 provider（核心逻辑已通过 `TestExtractHost` 11 个单元测试覆盖）

- [x] 3.3 手动验证 `grepom sync` 在同一组内不产生重复仓库（核心逻辑已通过 `TestSyncGroupRepos_*` 覆盖）
