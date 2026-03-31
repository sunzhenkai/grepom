## 1. Token 环境变量占位符支持

- [x] 1.1 修改 `config/config.go` 的 `Load()` 函数，移除全局 `os.ExpandEnv`，改为对整个文件 raw 读取并用 yaml 解析保留原始值
- [x] 1.2 在 `config/config.go` 中新增 `resolveToken(token string) string` 函数，检测 `${...}` 格式并从环境变量解析，非占位符直接返回原值
- [x] 1.3 修改 `Source` 结构体的加载逻辑：加载后对 `Token` 字段调用 `resolveToken` 获取运行时值，但保留原始值用于回写
- [x] 1.4 引入 `RawConfig` 或在 `Config` 中增加 `rawSources` 映射，在 `writeConfig` 时使用原始 token 值替换展开后的值
- [x] 1.5 编写单元测试覆盖：占位符解析成功、环境变量不存在时报错、明文 token 直接使用、回写保留占位符

## 2. Source Name 字段支持

- [x] 2.1 在 `config/config.go` 的 `Source` 结构体中新增 `Name string yaml:"name,omitempty"` 字段
- [x] 2.2 在 `config.go` 的 `validate()` 方法中新增 name 唯一性校验（如果指定了 name 则不能重复）
- [x] 2.3 新增 `Config.FindSource(nameOrIndex string) (int, Source, error)` 方法，优先按 name 查找，回退按索引查找
- [x] 2.4 修改 `cmd/root.go` 或相关命令中 `--source` 参数的解析逻辑，支持字符串（name）和整数（索引）
- [x] 2.5 修改 `cmd/add.go` 的 `addSourceCmd`，新增 `--name` flag
- [x] 2.6 编写单元测试：name 匹配、索引回退、name 重复报错、无 name 时省略字段

## 3. Sync 命令改为仅元数据同步

- [x] 3.1 修改 `cmd/sync.go`：移除所有 clone 和 pull 相关代码（`gitpkg.IsCloned`、`gitpkg.Clone`、`gitpkg.Pull` 调用）
- [x] 3.2 修改 `cmd/sync.go`：将发现的仓库信息转为 `RepoEntry` 列表，追加到配置的 `repos` 中（去重：按 URL 匹配）
- [x] 3.3 修改 `cmd/sync.go` 的输出：显示发现的仓库数量和新增的 group/org 数量，提示用户运行 `grepom clone` 克隆新仓库
- [x] 3.4 修改 `--source` 参数解析，支持通过 name 引用 source
- [x] 3.5 确保子 group 发现和追加逻辑不变，仅移除 clone/pull 部分
- [x] 3.6 编写/更新集成测试：验证 sync 不触发 clone/pull，验证新仓库写入 repos 配置

## 4. 配置文件回写兼容性

- [x] 4.1 确保 `writeConfig()` 正确保留 token 的原始占位符值（不被展开后的值覆盖）
- [x] 4.2 确保 `writeConfig()` 新增的 name 字段正确写入（有值时写入，无值时省略）
- [x] 4.3 确保 `SyncGroups()` 和 `AddSource()` 等函数在回写时正确处理新字段
- [x] 4.4 验证现有配置文件（无 name 字段、明文 token）加载和操作正常

## 5. 文档和输出更新

- [x] 5.1 更新 `sync` 命令的 Long 描述和 Example，说明仅同步元数据
- [x] 5.2 更新 `add source` 命令的帮助文本，说明 `--name` 参数
- [x] 5.3 更新 README.md（如有需要）说明 token 占位符用法
