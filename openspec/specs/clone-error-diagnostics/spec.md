# clone-error-diagnostics Specification

## Purpose

克隆失败时透出脱敏后的 git 原始错误，便于区分 URL 拼接错误与认证失败。

## Requirements

### Requirement: 克隆失败透出 git 原始错误
系统在执行 `git clone` / `git pull` 失败时 SHALL 捕获命令的 stderr（或等价输出），并在返回错误与 verbose 日志中包含脱敏后的原始错误文本。全部认证策略失败时，最终错误 SHALL 在 `all authentication methods failed` 之外附带各策略失败原因摘要，便于区分 URL 拼接错误与 SSH/HTTPS 认证错误。

#### Scenario: verbose 输出单步失败原因
- **WHEN** 用户以 `-v` 克隆某仓库，某一认证策略的 `git clone` 失败
- **THEN** 日志 SHALL 包含该策略标签、脱敏后的 URL，以及脱敏后的 git stderr 内容（例如 `Permission denied` 或 `Repository not found`）

#### Scenario: 最终错误包含摘要
- **WHEN** 所有认证策略均失败
- **THEN** 返回错误信息 SHALL 包含 `all authentication methods failed`，并追加至少一条脱敏后的底层失败原因，而不是仅有这一句笼统文案

#### Scenario: 错误中的 token 被脱敏
- **WHEN** 失败的 clone URL 或 stderr 中包含 `https://user:token@host/...` 形式
- **THEN** 对外展示的错误与日志 SHALL 将 token 替换为掩码（如 `***`），不泄露明文凭据
