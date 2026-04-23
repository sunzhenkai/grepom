## 1. CLI 命令描述英文化（cobra Short/Long/Example + flag 帮助）

- [x] 1.1 将 `cmd/scan.go` 的 cobra 描述（Short、Long、Example）和所有 flag 帮助文本从中文改为英文
- [x] 1.2 将 `cmd/push.go` 的 cobra 描述（Short、Long、Example）和所有 flag 帮助文本从中文改为英文
- [x] 1.3 将 `cmd/example.go` 的 cobra 描述（Short、Long）和 flag 帮助文本从中文改为英文
- [x] 1.4 将 `cmd/example.go` 中 `exampleConfig` 常量的 YAML 注释从中文改为英文

## 2. CLI 输出英文化（fmt.* 输出、错误信息）

- [x] 2.1 将 `cmd/scan.go` 中的中文错误信息（fmt.Errorf）改为英文
- [x] 2.2 将 `cmd/push.go` 中的中文错误信息（fmt.Errorf）改为英文
- [x] 2.3 将 `cmd/sync.go` 中混合中英文的输出改为纯英文
- [x] 2.4 将 `cmd/list.go` 中混合中英文的输出改为纯英文

## 3. 交互式模式英文化

- [x] 3.1 将 `cmd/interactive.go` 中所有菜单选项字符串从中文改为英文（初始化配置→Initialize config、添加资源→Add resource、添加组→Add group、添加仓库→Add repo、同步远程仓库→Sync remote repos、克隆仓库→Clone repos、拉取更新→Pull updates、查看状态→Check status、退出→Exit）
- [x] 3.2 将 `cmd/interactive.go` 中所有提示信息（survey Message）从中文改为英文
- [x] 3.3 将 `cmd/interactive.go` 中所有错误信息（fmt.Fprintf stderr）从中文改为英文
- [x] 3.4 将 `cmd/interactive.go` 中所有进度/状态输出（fmt.Printf/Println）从中文改为英文
- [x] 3.5 将 `cmd/interactive.go` 中所有确认和取消消息改为英文
- [x] 3.6 将 `cmd/interactive.go` 中范围选择选项（全部/按组/按资源）改为英文（All/By group/By resource）
- [x] 3.7 将 `cmd/interactive.go` 中并行度选项（1 (顺序)、4 (默认)）改为英文（1 (sequential)、4 (default)）

## 4. Git 包输出英文化

- [x] 4.1 将 `git/git.go` 中所有认证策略标签（label）从中文改为英文（如 "SSH key auth (group/repo)"、"Token auth (group/repo)" 等）
- [x] 4.2 将 `git/git.go` 中进度输出字符串（尝试/成功/失败）改为英文（trying/ok/failed）
- [x] 4.3 将 `git/git.go` 中错误信息（所有认证方式均失败、无法检测默认分支）改为英文

## 5. Scanner 包输出英文化

- [x] 5.1 将 `scanner/scanner.go` 中所有中文错误信息（fmt.Errorf）改为英文
- [x] 5.2 将 `scanner/scanner.go` 中的中文警告信息（fmt.Fprintf stderr）改为英文

## 6. 测试文件同步更新

- [x] 6.1 更新 `git/git_test.go` 中引用中文认证策略标签的断言为英文
- [x] 6.2 运行全量测试确认所有测试通过：`go test ./...`

## 7. 文档更新

- [x] 7.1 更新 README.md（中文），补充缺失的命令文档：scan、push、search、prune、dedup、pipeline、init、example；补充缺失的 flags 说明
- [x] 7.2 更新 README_en.md（英文），与中文版内容同步，补充所有缺失命令和 flags
- [x] 7.3 确认 README.md 和 README_en.md 内容覆盖一致
