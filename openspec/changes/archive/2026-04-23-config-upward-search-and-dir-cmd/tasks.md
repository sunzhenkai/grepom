## 1. 配置文件向上探测

- [x] 1.1 在 `config/config.go` 中新增 `findConfigUpward(startDir string) (string, error)` 函数，从 startDir 开始沿父目录链向上查找 `.grepom.yml`，直到找到或到达文件系统根目录
- [x] 1.2 修改 `FindConfig(explicitPath string)` 函数：当 explicitPath 为空且当前目录无 `.grepom.yml` 时，调用 `findConfigUpward` 继续查找
- [x] 1.3 更新 `ErrConfigNotFound` 的错误提示信息，建议用户使用 `-c` 或在任意父目录创建配置文件
- [x] 1.4 为 `findConfigUpward` 编写单元测试，覆盖：当前目录找到、父目录找到、多级向上找到、根目录仍未找到、嵌套配置就近原则

## 2. `dir` 命令实现

- [x] 2.1 创建 `cmd/dir.go`，定义 `dirCmd`（cobra.Command），注册到 rootCmd
- [x] 2.2 实现无参数模式：加载配置后输出 `cfg.Base` 绝对路径到 stdout
- [x] 2.3 实现有参数模式：通过 `repo.NewResolver(cfg).Resolve()` 解析所有仓库，使用 `repo.ApplySearchFilter` 进行大小写不敏感子串匹配
- [x] 2.4 实现结果分支逻辑：0 结果输出错误到 stderr（退出码 1）；1 结果输出路径到 stdout；多结果以表格列出到 stderr 并退出码 1
- [x] 2.5 确保 stdout 仅包含路径、所有错误和表格输出到 stderr

## 3. 测试

- [x] 3.1 编写 `dir` 命令测试：覆盖无参数输出 base、精确匹配、模糊匹配单个结果、模糊匹配多个结果、无匹配结果
- [x] 3.2 编写集成场景测试：在子目录中执行命令验证向上探测生效

## 4. 文档更新

- [x] 4.1 更新 `README.md`（中文）：新增 `dir` 命令说明和用法示例，更新配置文件查找逻辑描述
- [x] 4.2 更新 `README_en.md`（英文）：同步中文文档的变更
