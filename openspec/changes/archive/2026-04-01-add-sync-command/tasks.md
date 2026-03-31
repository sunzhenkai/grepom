## 1. Provider 接口扩展

- [x] 1.1 在 `provider/provider.go` 中定义 `SubGroupLister` 接口，包含 `ListSubGroups(ctx, source, groupPath) ([]string, error)` 方法
- [x] 1.2 在 `provider/gitlab.go` 中为 `GitLabProvider` 实现 `ListSubGroups` 方法，调用 `getSubgroups` API 获取指定 group 下的子 group 路径列表

## 2. 配置层支持

- [x] 2.1 在 `config/config.go` 中新增 `SyncGroups(configPath string, sourceIndex int, newGroups []GroupSource) error` 函数，在文件锁保护下读取配置、对比已有 groups、追加新条目、写回文件
- [x] 2.2 在 `config/config.go` 中新增文件锁工具函数 `WithFileLock(path string, timeout time.Duration, fn func() error) error`，使用 `syscall.Flock` 实现排他锁
- [x] 2.3 为 `SyncGroups` 编写单元测试：验证只增不删、重复不追加、锁竞争安全

## 3. sync 命令实现

- [x] 3.1 新建 `cmd/sync.go`，定义 `syncCmd` cobra 命令，支持 `--source`（int 索引）、`--group`（string）、`--org`（string）flag
- [x] 3.2 实现 sync 核心逻辑：遍历匹配的 source/group/org，调用 provider 获取仓库列表，对未 clone 的执行 clone、已 clone 的执行 pull
- [x] 3.3 实现子 group 发现逻辑：对 recursive=true 的 GitLab group，通过 `SubGroupLister` 接口获取子 group 列表，调用 `SyncGroups` 追加新条目
- [x] 3.4 sync 命令添加 verbose 输出：显示发现的新仓库数、clone 数、pull 数、新增 group 数

## 4. 集成验证

- [x] 4.1 运行 `make test` 确保所有现有测试通过
- [x] 4.2 运行 `make build` 确保编译通过
