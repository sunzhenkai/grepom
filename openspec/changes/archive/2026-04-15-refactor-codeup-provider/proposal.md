## Why

Codeup provider 当前使用旧版 API（devops.aliyun.com 域名、accessToken query param 认证、`/repository/list` 路径），该域名实际是云效 Web 前端而非 API 端点，导致所有 API 调用返回 HTML 页面，sync 和 list 命令完全不可用。云效官方已推出新版 OAPI v1，需要全面迁移。

## What Changes

- **认证方式**: `accessToken` query parameter → `x-yunxiao-token` 请求头
- **API 域名**: `devops.aliyun.com`（Web 前端）→ `openapi-rdc.aliyuncs.com`（API 端点），配置文件中 `url` 字段保持 `codeup.aliyun.com` 不变，仅用于 clone URL 推导
- **响应格式**: 旧版 wrapper `{ requestId, success, errorCode, errorMessage, total, result }` → 直接返回数组，分页信息移至响应头 `x-total`/`x-total-pages`/`x-next-page`
- **ListRepos 实现**: 旧版拉全量 org repos 客户端过滤 → 新版先通过 `ListNamespaces` 解析 group path 为 groupId，再通过 `ListGroupRepositories` 精确按组拉取（支持 `includeSubgroups`）
- **ListGroups 实现**: 旧版两步（`find_by_path` + `groups/get/all`）→ 新版 `ListNamespaces` 一步到位
- **orgId 位置**: query parameter → path parameter

## Capabilities

### New Capabilities

（无新增能力）

### Modified Capabilities

- `codeup-provider`: API 端点域名、认证方式、响应解析、分页机制、ListRepos 数据流、ListGroups 数据流全面迁移至新版 OAPI v1

## Impact

- **provider/codeup.go**: 完全重写（API URL 映射、认证、请求/响应结构体、ListRepos、ListGroups）
- **provider/codeup_test.go**: 完全重写（mock server 须匹配新版响应格式——直接数组 + 响应头分页 + Header 认证）
- **config/config.go**: 无改动（`url` 字段含义不变，仍然是 clone host）
- **cmd/sync.go**: 无改动（调用路径不变）
- **cmd/list.go**: 无改动（调用路径不变）
- **配置文件兼容**: 用户配置无需任何修改，`url: codeup.aliyun.com` 保持不变
