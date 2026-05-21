## 1. git/tag.go — tag 底层操作函数

- [x] 1.1 实现 `FetchTags(path string) error`：执行 `git -C <path> fetch --tags --all`，封装错误信息
- [x] 1.2 实现 `ListTagsByTime(path, pattern string) ([]string, error)`：执行 `git -C <path> tag --sort=-creatordate --list "<pattern>"`，返回按时间降序排列的 tag 列表
- [x] 1.3 实现 `TagExists(path, name string) (bool, error)`：执行 `git -C <path> tag -l "<name>"`，判断 tag 是否存在
- [x] 1.4 实现 `CreateTag(path, name string) error`：执行 `git -C <path> tag <name>` 创建轻量 tag
- [x] 1.5 实现 `CreateAnnotatedTag(path, name, message string) error`：执行 `git -C <path> tag -a <name> -m <message>` 创建附注 tag
- [x] 1.6 实现 `PushTag(path, remote, tag string) error`：执行 `git -C <path> push <remote> <tag>` 推送单个 tag 到指定 remote
- [x] 1.7 实现 `ListRemotes(path string) ([]string, error)`：执行 `git -C <path> remote` 返回所有 remote 名称列表

## 2. git/tag.go — 版本号解析与计算

- [x] 2.1 实现 `ParseVersion(tag string) (prefix string, digits []int, err error)`：解析 tag 字符串，提取前缀（v 或 t）和数字部分。按 `.` 分割，每段转为 int，忽略非数字段
- [x] 2.2 实现 `NormalizeDigits(digits []int, size int) []int`：将数字数组规范化到指定长度。多于 size 截取前 size 位，少于 size 在末尾补 0
- [x] 2.3 实现 `NextVPatch(major, minor, patch int) (int, int, int)`：计算下一个 v 版本号。PATCH + 1，若 PATCH > 99 则 MINOR + 1、PATCH = 0
- [x] 2.4 实现 `FormatVTag(major, minor, patch int) string`：格式化为 `v{MAJOR}.{MINOR}.{PATCH}`
- [x] 2.5 实现 `FormatTTag(major, minor, patch, iter int) string`：格式化为 `t{MAJOR}.{MINOR}.{PATCH}.{ITER}`
- [x] 2.6 编写版本号解析和计算的单元测试，覆盖：正常 3 位、补齐（1 位、2 位）、截取（4 位、5 位）、非数字段、溢出进位、无 tag 默认值等场景

## 3. cmd/tag.go — CLI 命令注册

- [x] 3.1 创建 `cmd/tag.go`，定义 `tagCmd` cobra.Command（`Use: "tag"`，`Short: "Create version tags"`），注册 `-t`、`-p`、`--dry-run`、`-m` 标志
- [x] 3.2 在 `cmd/root.go` 的 `init()` 中注册 `tagCmd`（`rootCmd.AddCommand(tagCmd)`）

## 4. cmd/tag.go — 核心业务逻辑

- [x] 4.1 实现 `runTag(cmd *cobra.Command, args []string) error` 主函数：
  - 前置检查：`gitpkg.IsCloned(".")` 判断当前目录是否为 git 仓库
  - 调用 `gitpkg.FetchTags(".")` 从所有 remote 拉取最新 tags
  - 调用 `gitpkg.ListTagsByTime(".", "v*")` 获取按时间排序的 v tag 列表
  - 解析最新 v tag 或使用默认值 (0, 0, 0)
  - 计算下一个版本号
  - 根据 `-t` 标志分流到 v 或 t 版本计算
- [x] 4.2 实现 `computeNextVTag(path string) (string, error)`：
  - 获取 v* tag 列表，解析最新 tag 的版本号
  - 调用 `NextVPatch` 计算下一版本
  - 冲突检测：循环调用 `TagExists`，已存在则继续递增，直到找到空位（上限 10000 次）
  - 无 v tag 时返回 `v0.0.1`
- [x] 4.3 实现 `computeNextTTag(path string, major, minor, patch int) (string, error)`：
  - 格式化前 3 位匹配模式 `t{M}.{m}.{p}.*`
  - 调用 `ListTagsByTime` 查找匹配的 t tag
  - 解析第 4 位，无匹配则从 0 开始
  - 冲突检测：循环递增第 4 位直到找到空位（上限 10000 次）
- [x] 4.4 实现 tag 创建逻辑：
  - `--dry-run` 时仅输出 `[dry-run]` 前缀的预览信息，不执行创建
  - 有 `-m` 消息时调用 `CreateAnnotatedTag`，否则调用 `CreateTag`
  - 创建成功后输出确认信息
- [x] 4.5 实现推送逻辑 `pushToAllRemotes(path, tag string) error`：
  - 调用 `ListRemotes` 获取所有 remote
  - 逐一调用 `PushTag` 推送，每个 remote 输出结果（✓ / ✗）
  - 无 remote 时给出提示
- [x] 4.6 实现交互式推送提示：
  - 有 `-p` 标志 → 直接调用 `pushToAllRemotes`
  - 无 `-p` + 有 TTY → 使用 `survey.Confirm` 询问用户，确认后推送
  - 无 `-p` + 无 TTY → 输出提示 "Use -p to push."

## 5. 测试与验证

- [x] 5.1 为 `git/tag.go` 中的版本解析函数编写单元测试（`git/tag_test.go`）：`TestParseVersion`、`TestNormalizeDigits`、`TestNextVPatch`、`TestFormatVTag`、`TestFormatTTag`
- [x] 5.2 为 `git/tag.go` 中的 git 操作函数编写集成测试：创建临时 git 仓库，测试 `FetchTags`、`ListTagsByTime`、`TagExists`、`CreateTag`、`CreateAnnotatedTag`、`PushTag`、`ListRemotes`
- [x] 5.3 端到端验证：在测试仓库中创建多个 v tag，验证 `grepom tag` 正确递增；创建 t tag，验证第 4 位独立递增；验证冲突自动解决
- [x] 5.4 编译验证：`go build ./...` 确认无编译错误，`go vet ./...` 无警告
