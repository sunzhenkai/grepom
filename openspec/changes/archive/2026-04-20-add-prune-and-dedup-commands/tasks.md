## 1. Prune 命令

- [x] 1.1 创建 `cmd/prune.go`，注册 `pruneCmd` cobra command，定义 flags：`--group`、`--resource`、`--force`、`--dry-run`（默认 true）、`--apply`（设置 dry-run=false）
- [x] 1.2 实现 prune 主逻辑：加载 config → 用 `IncludeDisabled=true` resolve repos → 过滤 `DisabledReason=="excluded"` → 按 group/resource 过滤
- [x] 1.3 实现磁盘状态检查：对每个 excluded repo 调用 `git.IsCloned()` 和 `git.GetStatus()`，分类为 safe_delete / unsafe_delete / not_cloned
- [x] 1.4 实现删除逻辑：`--apply` 时对 safe_delete repos 执行 `os.RemoveAll()`，`--force` 时跳过安全检查；`--dry-run`（默认）时只输出计划不删除
- [x] 1.5 实现输出格式：逐行输出每个 repo 的处理结果（deleted/skipped/not cloned + 原因），末尾输出摘要

## 2. Dedup 命令

- [x] 2.1 创建 `cmd/dedup.go`，注册 `dedupCmd` cobra command，定义 flags：`--group`（必需）、`--reference`（可选，逗号分隔）、`--dry-run`（默认 true）、`--apply`
- [x] 2.2 实现 reference group 解析：`--reference` 指定时按逗号分隔查找对应 group；不指定时收集除 target 外的所有 group
- [x] 2.3 实现同名检测：收集所有 reference groups 的 repo names 到 Set，遍历 target group 的 repos 检查 name 是否在 Set 中
- [x] 2.4 实现 exclude 处理：对每个同名 repo 检查现有 `exclude_repos` 是否已覆盖（用 `repo.IsExcluded()`），未覆盖则追加到 exclude_repos
- [x] 2.5 实现 repos 列表清理：从 target group 的 `repos` 中移除被 exclude 的条目
- [x] 2.6 实现写入逻辑：`--apply` 时用 `config.WithFileLock` 写入修改后的 config；`--dry-run`（默认）时只输出计划
- [x] 2.7 实现输出格式：逐行输出冲突处理结果 + 摘要 + prune 提示

## 3. Config 层支持

- [x] 3.1 在 `config/config.go` 中新增 `DedupGroupRepos(configPath, targetGroupName string, repoNames []string) error` 方法：加载 config → 从 target group 的 repos 中移除匹配的条目 → 追加到 exclude_repos（去重） → 写入 config。使用 WithFileLock 保证并发安全。

## 4. 测试

- [x] 4.1 为 `cmd/prune.go` 编写测试：dry-run 模式、--apply 模式、dirty repo 跳过、ahead repo 跳过、--force 跳过安全检查、group/resource 过滤
- [x] 4.2 为 `cmd/dedup.go` 编写测试：同名检测、多 reference 合并、reference 不指定时对比所有、exclude_repos 去重、repos 列表清理、dry-run 不写入
- [x] 4.3 为 `config.DedupGroupRepos` 编写测试：基本功能、exclude_repos 去重、repos 移除、文件锁
