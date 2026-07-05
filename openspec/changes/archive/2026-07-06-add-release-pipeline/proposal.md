## Why

grepom 目前仅支持 `go install` 或本地 `make install` 安装，没有预编译二进制分发渠道。用户需要自行安装 Go 工具链才能使用，且无法便捷获取特定版本或预发布版本。随着项目已有 v0.1.x 标签但无 GitHub Release，需要建立自动化的 tag 触发发布管线与一键安装脚本，降低安装门槛。

## What Changes

- 新增 GitHub Actions workflow：push `v*` tag 时自动交叉编译并发布到 GitHub Release
- 支持 4 个平台：`linux/amd64`、`linux/arm64`、`darwin/amd64`、`darwin/arm64`
- tag 含 `-rc` 或 `-beta` 时自动标记为 GitHub Pre-release
- 发布资产包含各平台 tar.gz 及 `checksums.txt`（sha256）
- 新增 `scripts/install.sh` 一键安装脚本，默认安装到 `~/.local/bin`，支持 `INSTALL_DIR=/usr/local/bin`
- 修复 `main.version` 版本号注入（Makefile 已有 ldflags 但 main 包缺少变量）
- 新增 `grepom version` 子命令，输出当前二进制版本
- 新增 CI workflow：push/PR 时运行 `go test ./...`
- 更新 README（中英文）安装章节，补充 curl 安装说明

## Capabilities

### New Capabilities

- `release-distribution`: tag 触发的多平台二进制构建、GitHub Release 发布、checksums 生成与一键安装脚本行为

### Modified Capabilities

（无）

## Impact

- 新增文件：`.github/workflows/release.yml`、`.github/workflows/ci.yml`、`scripts/install.sh`
- 修改文件：`main.go`（版本变量）、`cmd/`（version 子命令）、`README.md`、`README_en.md`
- 无破坏性变更，不影响现有 CLI 命令行为
- 依赖 GitHub Actions 免费额度；Release 资产托管于 GitHub Releases
