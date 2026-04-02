## 1. 数据结构修改

- [x] 1.1 在 `config/config.go` 的 `Resource` 结构体中添加 `Enabled` 字段（`yaml:"enabled,omitempty"`，默认 `true`）
- [x] 1.2 在 `config/config.go` 的 `Group` 结构体中添加 `Enabled` 字段（`yaml:"enabled,omitempty"`，默认 `true`）
- [x] 1.3 在 `config/config.go` 的 `Group` 结构体中添加 `ExcludeRepos` 字段（`yaml:"exclude_repos,omitempty"`，`[]string` 类型）
- [x] 1.4 在 `config/config.go` 的 `Repo` 结构体中添加 `Enabled` 字段（`yaml:"enabled,omitempty"`，默认 `true`）

## 2. 解析器排除逻辑

- [x] 2.1 在 `repo/resolver.go` 的 `Filter` 结构体中添加 `IncludeDisabled bool` 字段
- [x] 2.2 在 `Resolve()` 方法中实现 resource 级别的 enabled 过滤（遍历 group 和独立 repo 时检查引用的 resource 是否启用）
- [x] 2.3 在 `Resolve()` 方法中实现 group 级别的 enabled 过滤（跳过 `enabled: false` 的 group）
- [x] 2.4 在 `Resolve()` 方法中实现 group 级别的 exclude_repos 过滤（按 repo name 精确匹配排除）
- [x] 2.5 在 `Resolve()` 方法中实现独立 repo 级别的 enabled 过滤（跳过 `enabled: false` 的独立 repo）
- [x] 2.6 当 `IncludeDisabled` 为 `true` 时跳过所有排除逻辑，返回完整 repo 列表

## 3. CLI 命令适配

- [x] 3.1 在 `cmd/list.go` 中添加 `--all` 标志，传入 `Filter.IncludeDisabled` 以显示被排除的条目
- [x] 3.2 在 `cmd/list.go` 中为被排除/被禁用的 repo 添加 `[disabled]`/`[excluded]` 标注
- [x] 3.3 在 `cmd/clone.go` 中确保默认排除被禁用条目（已通过 resolver 自动实现）
- [x] 3.4 在 `cmd/pull.go` 中确保默认排除被禁用条目（已通过 resolver 自动实现）
- [x] 3.5 在 `cmd/status.go` 中确保默认排除被禁用条目（已通过 resolver 自动实现）

## 4. Sync 命令适配

- [x] 4.1 在 `cmd/sync.go` 中跳过 `enabled: false` 的 group，不对其执行远程发现
- [x] 4.2 在 `cmd/sync.go` 中跳过 `enabled: false` 的 resource 下的所有 group
- [x] 4.3 验证 sync 写入配置时 `exclude_repos` 列表被正确保留（现有只增不删逻辑应已覆盖）

## 5. 测试

- [x] 5.1 在 `config/config_test.go` 中添加 Resource `enabled` 字段的加载测试（启用、禁用、省略默认值）
- [x] 5.2 在 `config/config_test.go` 中添加 Group `enabled` 和 `exclude_repos` 字段的加载测试
- [x] 5.3 在 `config/config_test.go` 中添加独立 Repo `enabled` 字段的加载测试
- [x] 5.4 在 `repo/resolver_test.go` 中添加 resource 禁用时排除所有关联条目的测试
- [x] 5.5 在 `repo/resolver_test.go` 中添加 group 禁用时排除其下所有 repo 的测试
- [x] 5.6 在 `repo/resolver_test.go` 中添加 exclude_repos 排除特定 repo 的测试
- [x] 5.7 在 `repo/resolver_test.go` 中添加独立 repo 禁用的测试
- [x] 5.8 在 `repo/resolver_test.go` 中添加 `IncludeDisabled: true` 时包含所有条目的测试
- [x] 5.9 在 `repo/resolver_test.go` 中添加排除优先级测试（resource > group > exclude_repos > repo）
- [x] 5.10 运行全部测试确认通过：`make test`
