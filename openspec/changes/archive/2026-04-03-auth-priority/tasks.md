## 1. 修改认证策略构建逻辑

- [x] 1.1 重写 `git/git.go` 中 `buildAuthStrategies()` 函数，调整策略构建顺序为新优先级：group/repo SSH → group/repo token → resource SSH → default SSH → resource token → bare HTTP
- [x] 1.2 调整步骤 4 (default SSH) 的条件：只需 `sshURL != ""` 即可加入，不与 resource SSH 做互斥
- [x] 1.3 验证步骤 5 (resource token) 的条件保持 `!HasGroupToken && Token != ""` 不变

## 2. 更新日志标签

- [x] 2.1 确认 `buildAuthStrategies()` 中各策略的 `label` 字段与现有风格一致，特别是新增的 `"SSH 认证 (默认)"` 在 resource token 之前

## 3. 更新测试

- [x] 3.1 修改或新增 `buildAuthStrategies()` 的单元测试，覆盖新优先级场景：
  - group/repo SSH + group/repo token + resource SSH + default SSH + resource token 全链路
  - group 有 SSH key + resource 有 token 时：group SSH → default SSH → resource token
  - group 有 token（无 SSH）+ resource 有 SSH + token 时：group token → resource SSH → default SSH → resource token
  - 仅 resource token（无 SSH key）时：default SSH → resource token
- [x] 3.2 运行现有测试确认无回归

## 4. 更新现有 spec

- [x] 4.1 更新 `openspec/specs/clone-auth-priority/spec.md` 中"克隆认证优先级链" requirement 的优先级描述和场景
