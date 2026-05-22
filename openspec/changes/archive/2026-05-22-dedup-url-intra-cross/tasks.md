## 1. URL 规范化函数

- [x] 1.1 在 `config/normalize.go` 中实现 `NormalizeRepoURL(url string) string` 函数，按设计文档中的规则处理 HTTPS/HTTP/SSH 格式、.git 后缀、末尾斜杠、端口、host 大小写
- [x] 1.2 在 `config/normalize_test.go` 中编写单元测试，覆盖所有 spec 中定义的场景：HTTPS/HTTP/SSH URL、无 .git 后缀、带端口、末尾斜杠、host 大小写不敏感、path 大小写敏感、空 URL

## 2. 组内 URL 去重逻辑

- [x] 2.1 在 `config/config.go` 中实现 `DedupIntraGroupRepos(configPath, groupName string) ([]string, error)` 函数，按规范化 URL 检测组内重复，从 repos 列表删除多余条目（保留第一个），不加入 exclude_repos，使用 WithFileLock 保证并发安全
- [x] 2.2 在 `cmd/dedup.go` 中实现 Step 1 逻辑：遍历指定 group（或所有 group），对每个 group 调用 URL 规范化检测组内重复，收集待删除条目，输出结果

## 3. 跨组 URL 警告逻辑

- [x] 3.1 在 `cmd/dedup.go` 中实现 Step 2 逻辑：构建 url→groups 映射，检测跨组 URL 重复，输出 ⚠️ 警告信息（只打印不删除，不影响退出码）

## 4. 重构 dedup 命令

- [x] 4.1 将 `--group` flag 从必选（`MarkFlagRequired`）改为可选
- [x] 4.2 重构 `dedupCmd.RunE`：按 Step 1 → Step 2 → Step 3 顺序执行，Step 1+2 始终运行，Step 3 仅在 `--group` + `--reference` 同时指定时触发
- [x] 4.3 更新 dry-run 输出格式：分 "Intra-group dedup"、"Cross-group URL warnings"、"Cross-group name dedup"（条件）三部分
- [x] 4.4 更新 `--apply` 逻辑：Step 1 的组内去重和 Step 3 的跨组 name 去重统一在 --apply 时写入，Step 2 跨组警告始终打印

## 5. 测试

- [x] 5.1 添加组内去重测试：同 URL 不同格式的重复、多于两个重复、无重复、--apply 写入验证、exclude_repos 不受影响
- [x] 5.2 添加跨组 URL 警告测试：检测跨组重复并验证只输出警告不修改配置、退出码为 0、无跨组重复时无警告输出
- [x] 5.3 添加 `--group` 可选测试：不指定 --group 时对所有 group 执行组内去重、指定 --group 时只处理指定 group
- [x] 5.4 添加三步流程集成测试：无参数运行（Step 1+2）、仅 --group（Step 1+2）、--group + --reference（Step 1+2+3）
- [x] 5.5 验证所有现有 dedup 测试继续通过

## 6. 文档更新

- [x] 6.1 更新 `README.md` 中 dedup 命令的使用说明和示例
- [x] 6.2 更新 `README_en.md` 中 dedup 命令的使用说明和示例
