## 1. 新增 `-p/--path` 标志

- [x] 1.1 在 `cmd/scan.go` 的 `init()` 中注册 `-p/--path` flag（`StringVarP`，默认空字符串）
- [x] 1.2 更新 `scanCmd` 的 `Use`、`Short`、`Long`、`Example` 字段，包含 `-p` 用法说明
- [x] 1.3 在 `scanCmd` 的 `Args` 中调整为 `cobra.MaximumNArgs(1)`（保持现有行为）

## 2. 重构 `runScan` 分支逻辑

- [x] 2.1 新增全局变量 `scanPath string` 存储 `-p` flag 值
- [x] 2.2 重写 `runScan` 函数，实现三分支逻辑：
  - `-p` 非空 → 调用 `runScanPath(scanPath)` 扫描指定路径（忽略配置）
  - 当前目录存在 `.grepom.yml` → 加载配置，走 resolver 路径（保留现有 `[name]` 过滤）
  - 其他 → 调用 `runScanCurrentDir()` 扫描当前目录
- [x] 2.3 实现 `runScanPath(path string) error`：验证路径存在，调用 `scanner.ScanDir`，设置 repo 名称为路径
- [x] 2.4 移除原 `runScan` 中对 `config.IsConfigNotFound(err)` 的回退逻辑（不再需要，因为配置查找逻辑已改变）

## 3. 扫描目标摘要打印

- [x] 3.1 在 `runScanPath` 中：扫描前在 stderr 打印 `Scanning <path>...`
- [x] 3.2 在 `runScanCurrentDir` 中：扫描前在 stderr 打印 `Scanning current directory (no config file found)...`（已有，确认保留）
- [x] 3.3 在配置模式分支中：收集已克隆仓库列表后，打印摘要：
  - ≤5 个仓库：`Scanning repo1, repo2, repo3...`
  - \>5 个仓库：`Scanning repo1, repo2, ..., repo5, ...and N more`
- [x] 3.4 实现辅助函数 `printScanSummary(names []string)` 封装摘要打印逻辑

## 4. 测试与验证

- [x] 4.1 验证 `grepom scan -p /tmp` 正常扫描指定目录
- [x] 4.2 验证 `grepom scan -p .` 在无配置目录下扫描当前目录
- [x] 4.3 验证 `grepom scan` 在无 `.grepom.yml` 时扫描当前目录（不向上查找）
- [x] 4.4 验证 `grepom scan` 在有 `.grepom.yml` 时正常扫描配置中的仓库
- [x] 4.5 验证 `-p` 指定的路径不存在时报错
- [x] 4.6 验证 `grepom scan --help` 显示 `-p/--path` 帮助信息

## 5. 文档更新

- [x] 5.1 更新 `README.md` 中 scan 命令的用法说明，增加 `-p` 示例
- [x] 5.2 更新 `README_en.md` 中 scan 命令的用法说明，增加 `-p` 示例
