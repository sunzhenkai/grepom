## 1. Resource URL 解析与 APIURL 适配

- [x] 1.1 将 `stripScheme()` 改为 `parseResourceURL()`，返回 `(host, scheme)`，scheme 为 `""`/`"http"`/`"https"`
- [x] 1.2 Resource 新增内部 `scheme` 字段（不导出到 YAML），在 validate 时通过 `parseResourceURL()` 设置
- [x] 1.3 修改 `Resource.APIURL()` 根据 scheme 返回对应协议前缀的 URL

## 2. Fallback 逻辑

- [x] 2.1 在 `cmd` 层实现 fallback 包装函数：scheme 为空（auto）时先用 HTTPS 尝试，`net.OpError` 后回退 HTTP
- [x] 2.2 fallback 成功时输出警告，建议用户在 URL 中显式指定 `http://` 前缀
- [x] 2.3 scheme 非空（`http`/`https`）时不触发 fallback

## 3. 集成到命令

- [x] 3.1 修改 `cmd/sync.go` 中 `ListRepos` 调用，接入 fallback 逻辑
- [x] 3.2 修改 `cmd/group-list.go` 中 `ListGroups` 调用，接入 fallback 逻辑
- [x] 3.3 检查其他使用 `APIURL()` 的命令，确认是否需要适配

## 4. Clone 结果摘要

- [x] 4.1 `cmd/clone.go` 顺序克隆模式增加最终结果摘要（成功/失败计数）

## 5. 测试

- [x] 5.1 为 `parseResourceURL()` 新增单元测试（有前缀、无前缀、带端口）
- [x] 5.2 为 `Resource.APIURL()` 新增 scheme 相关单元测试
- [x] 5.3 为 fallback 逻辑新增单元测试
