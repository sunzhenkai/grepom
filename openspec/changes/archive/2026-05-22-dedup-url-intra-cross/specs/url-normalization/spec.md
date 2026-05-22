## ADDED Requirements

### Requirement: URL 规范化函数

系统 SHALL 提供 `NormalizeRepoURL(url string) string` 函数，将 Git 仓库 URL 统一为 `host/path` 格式用于比较。

规范化规则 SHALL 如下：
1. 去掉 `https://`、`http://` 前缀
2. 去掉 `.git` 后缀
3. SSH 格式 `git@host:path` 转换为 `host/path`
4. 去掉末尾 `/`
5. host 部分 SHALL 转为小写
6. path 部分 SHALL 保留原样（大小写敏感）
7. 端口号 SHALL 保留

#### Scenario: HTTPS URL 规范化
- **WHEN** 输入为 `https://gitlab.com/my-org/infra/api-lib.git`
- **THEN** 输出 SHALL 为 `gitlab.com/my-org/infra/api-lib`

#### Scenario: HTTP URL 规范化
- **WHEN** 输入为 `http://gitlab.com/my-org/infra/api-lib.git`
- **THEN** 输出 SHALL 为 `gitlab.com/my-org/infra/api-lib`

#### Scenario: SSH URL 规范化
- **WHEN** 输入为 `git@gitlab.com:my-org/infra/api-lib.git`
- **THEN** 输出 SHALL 为 `gitlab.com/my-org/infra/api-lib`

#### Scenario: 无 .git 后缀的 URL
- **WHEN** 输入为 `https://gitlab.com/my-org/infra/api-lib`
- **THEN** 输出 SHALL 为 `gitlab.com/my-org/infra/api-lib`

#### Scenario: 带端口的 URL
- **WHEN** 输入为 `https://gitlab.com:8443/my-org/infra/api-lib.git`
- **THEN** 输出 SHALL 为 `gitlab.com:8443/my-org/infra/api-lib`

#### Scenario: 末尾带斜杠的 URL
- **WHEN** 输入为 `https://gitlab.com/my-org/infra/api-lib/`
- **THEN** 输出 SHALL 为 `gitlab.com/my-org/infra/api-lib`

#### Scenario: host 大小写不敏感
- **WHEN** 输入为 `https://GitLab.com/my-org/infra/api-lib.git`
- **THEN** 输出 SHALL 为 `gitlab.com/my-org/infra/api-lib`

#### Scenario: path 大小写敏感
- **WHEN** 输入为 `https://gitlab.com/My-Org/Infra/api-lib.git`
- **THEN** 输出 SHALL 为 `gitlab.com/My-Org/Infra/api-lib`

#### Scenario: 空 URL
- **WHEN** 输入为空字符串
- **THEN** 输出 SHALL 为空字符串
