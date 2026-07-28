## 1. 配置加载：resources list 兼容

- [x] 1.1 在 `config` 中实现 `resources` 自定义 YAML 解码（map 直解 / list 按 `name` 归一为 map）
- [x] 1.2 校验 list 缺 `name`、重复 `name`、非法节点类型并返回明确错误
- [x] 1.3 补充 config 单测覆盖 map、list、缺名、重名场景

## 2. URL 分类解析与 Resolver

- [x] 2.1 实现 `repo.url` 分类（相对路径 / AbsoluteHTTP / AbsoluteSSH），禁止完整 URL 二次拼接
- [x] 2.2 更新 `repo/resolver.go`：绑定 resource 的独立仓按分类设置 CloneURL/SSHURL，并标记首选协议
- [x] 2.3 补充 resolver 单测：相对路径拼接、完整 HTTPS、`git@`、`ssh://` 均不拼坏

## 3. 克隆认证策略

- [x] 3.1 扩展 `buildAuthStrategies`：按首选协议分支；绝对 HTTPS 支持匿名 HTTPS；相对路径保持 SSH 优先
- [x] 3.2 实现未配置 `ssh_key` 时的默认私钥文件探测（id_ed25519 / id_rsa / id_ecdsa / id_ed25519_sk）
- [x] 3.3 更新 `git/git_test.go` 覆盖 HTTPS 优先、匿名 HTTPS、默认私钥跳过/尝试顺序

## 4. 克隆错误诊断

- [x] 4.1 修改 `tryClone`/`Pull`：捕获并返回脱敏后的 git stderr
- [x] 4.2 全部策略失败时汇总各步原因到最终错误；`-v` 打印完整脱敏 stderr
- [x] 4.3 补充诊断相关单测（含 token URL 脱敏）

## 5. virtual_groups 支持独立 repos

- [x] 5.1 `VirtualGroup` 增加 `Repos []string`，加载时校验引用存在于顶层 `repos`
- [x] 5.2 扩展 scope 解析与 `repo.Filter`（`RepoNames`），使 `--vgroup` 同时匹配成员 groups 与独立仓
- [x] 5.3 更新 list/status/clone/pull/search/prune 等过滤路径；`sync --vgroup` 仅 sync 成员 groups
- [x] 5.4 补充 virtual_groups 与 filter 单测（仅 repos / 混合 / 引用缺失 / 与 --group 并集）

## 6. 文档与 example

- [x] 6.1 更新 `README.md`：resources 改为 map；增加 URL 写法、协议/认证、ssh_key、vgroup repos 说明（虚构示例，无真实私密仓信息）
- [x] 6.2 同步更新 `README_en.md` 对应章节
- [x] 6.3 更新 `cmd/example.go` 注释与示例（相对路径 + 完整 HTTPS、`virtual_groups.repos`、ssh_key 建议）
- [x] 6.4 跑通相关单测与 `go test ./...` 回归
