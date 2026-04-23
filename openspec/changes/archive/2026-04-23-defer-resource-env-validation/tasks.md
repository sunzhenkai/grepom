## 1. 修改 config.Load() 移除 eager token 解析

- [x] 1.1 修改 `config/config.go` 的 `Load()` 函数：移除 resource/group/repo token 的 `resolveToken()` 调用，仅保存原始值到 rawTokens map，Token 字段保留 `${VAR}` 占位符字符串
- [x] 1.2 确保 `writeConfig()` 逻辑仍然正确：当 Token 字段本身就是 `${VAR}` 字符串时，rawTokens map 的 write-back 恢复逻辑不产生重复替换
- [x] 1.3 验证 `ensureConfigFile()` → `Load()` 路径在环境变量未设置时不再报错

## 2. 在 Resolver 中添加延迟 token 解析

- [x] 2.1 在 `repo/resolver.go` 的 `resolveInternal()` 中，获取 token 后调用 `resolveToken()` 进行延迟解析，仅对未被 disabled 过滤的 resource 的 token 解析
- [x] 2.2 确保 disabled resource 的 token 不会被解析（`resolveInternal()` 已经在组装 repo 时判断 disabled，需在 disabled 判断之前不解析 token）
- [x] 2.3 为延迟解析失败添加包含 resource/group/repo 名称的错误上下文信息

## 3. 保持 AddResource 的立即验证行为

- [x] 3.1 确认 `config.AddResource()` 仍然在添加时调用 `resolveToken()` 验证 token 环境变量
- [x] 3.2 确认 `ensureConfigFile()` 加载配置时不会因其他 resource 的环境变量缺失而报错（因为 Load 不再 eager 解析）

## 4. 简化 cmd/add.go 的变通方案

- [x] 4.1 评估 `loadExistingConfig()` 是否可以替换为直接调用 `config.Load()`，移除特殊处理逻辑
- [x] 4.2 如果 `loadExistingConfig()` 不再需要，清理相关代码

## 5. 更新测试

- [x] 5.1 修改 `config/config_test.go` 中依赖 "Load 时解析环境变量" 的测试用例，调整为验证 Load 不再解析环境变量
- [x] 5.2 新增测试：环境变量未设置时 `Load()` 成功返回
- [x] 5.3 新增测试：环境变量未设置但 resource disabled 时操作正常
- [x] 5.4 新增测试：Resolver 延迟解析 token 成功和失败的场景
- [x] 5.5 新增测试：延迟解析失败时错误信息包含 resource 上下文
- [x] 5.6 运行全部测试确保无回归
