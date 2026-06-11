## ADDED Requirements

### Requirement: Tag-triggered release workflow

系统 SHALL 在 push 匹配 `v*` 模式的 git tag 时，自动触发 GitHub Actions release workflow，构建多平台二进制并发布到 GitHub Release。

#### Scenario: Stable tag triggers release
- **WHEN** 用户 push tag `v0.2.0` 到远程仓库
- **THEN** release workflow SHALL 自动运行，创建 GitHub Release 并上传构建产物

#### Scenario: Pre-release tag marked as prerelease
- **WHEN** 用户 push tag `v0.2.0-rc.1` 或 `v0.2.0-beta.1`
- **THEN** 创建的 GitHub Release SHALL 标记为 prerelease

#### Scenario: Stable tag not marked as prerelease
- **WHEN** 用户 push tag `v0.2.0`（不含 `-rc` 或 `-beta`）
- **THEN** 创建的 GitHub Release SHALL NOT 标记为 prerelease

### Requirement: Multi-platform binary builds

release workflow SHALL 为以下 4 个平台交叉编译 `grepom` 二进制：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

#### Scenario: Linux amd64 build
- **WHEN** release workflow 执行 linux/amd64 job
- **THEN** SHALL 产出可执行的 `grepom` 二进制并打包为 tar.gz

#### Scenario: Darwin arm64 build
- **WHEN** release workflow 执行 darwin/arm64 job
- **THEN** SHALL 产出可执行的 `grepom` 二进制并打包为 tar.gz

### Requirement: Release artifact naming

每个平台的发布资产 SHALL 使用命名格式 `grepom_{version}_{os}_{arch}.tar.gz`，其中 `version` 为完整 tag 名（含 `v` 前缀），`os` 为 `linux` 或 `darwin`，`arch` 为 `amd64` 或 `arm64`。tar.gz 内 SHALL 仅包含单个 `grepom` 二进制文件。

#### Scenario: Asset name for linux arm64
- **WHEN** tag 为 `v0.2.0`，构建 `linux/arm64` 平台
- **THEN** 资产文件名 SHALL 为 `grepom_v0.2.0_linux_arm64.tar.gz`

#### Scenario: Archive contains single binary
- **WHEN** 用户解压任意平台的 tar.gz
- **THEN** SHALL 得到名为 `grepom` 的单个可执行文件

### Requirement: Checksums file

release workflow SHALL 生成 `checksums.txt` 文件，包含所有发布资产的 sha256 校验和，并作为 Release 资产一并上传。

#### Scenario: Checksums included in release
- **WHEN** release workflow 完成所有平台构建
- **THEN** GitHub Release SHALL 包含 `checksums.txt`，且其中列出每个 tar.gz 的 sha256 值

### Requirement: Install script

系统 SHALL 提供 `scripts/install.sh` 一键安装脚本，支持从 GitHub Release 下载并安装 grepom 二进制。

#### Scenario: Default install to user directory
- **WHEN** 用户执行 `curl -fsSL .../install.sh | bash` 且未设置 `INSTALL_DIR`
- **THEN** 脚本 SHALL 将 `grepom` 安装到 `~/.local/bin`

#### Scenario: Install to system directory
- **WHEN** 用户执行 `sudo INSTALL_DIR=/usr/local/bin bash install.sh`
- **THEN** 脚本 SHALL 将 `grepom` 安装到 `/usr/local/bin`

#### Scenario: Install specific version
- **WHEN** 用户设置 `VERSION=v0.2.0-rc.1` 后执行安装脚本
- **THEN** 脚本 SHALL 下载并安装该 tag 对应的二进制

#### Scenario: Install latest stable
- **WHEN** 用户未设置 `VERSION`（默认 `latest`）
- **THEN** 脚本 SHALL 从 GitHub `/releases/latest` 获取最新正式 release 并安装

#### Scenario: Verify checksum on install
- **WHEN** 安装脚本下载二进制资产
- **THEN** SHALL 使用 `checksums.txt` 校验下载文件的 sha256

#### Scenario: Unsupported platform rejected
- **WHEN** 用户在不受支持的操作系统或架构上执行安装脚本
- **THEN** 脚本 SHALL 以非零退出码退出并输出错误信息

#### Scenario: PATH hint after install
- **WHEN** 安装完成且 `INSTALL_DIR` 不在用户 PATH 中
- **THEN** 脚本 SHALL 输出提示信息，告知用户如何将 `INSTALL_DIR` 加入 PATH

### Requirement: Version command

系统 SHALL 提供 `grepom version` 子命令，输出当前二进制的版本号。

#### Scenario: Show embedded version
- **WHEN** 用户执行 `grepom version`
- **THEN** 系统 SHALL 输出版本号（release 构建为 tag 版本，开发构建为 `dev` 或 git describe 结果）

### Requirement: CI test workflow

系统 SHALL 提供 GitHub Actions CI workflow，在 push 和 pull request 到 main 分支时运行 `go test ./...`。

#### Scenario: Tests run on pull request
- **WHEN** 用户创建针对 main 的 pull request
- **THEN** CI workflow SHALL 运行并通过 `go test ./...`

#### Scenario: Tests run before release
- **WHEN** release workflow 被 tag 触发
- **THEN** SHALL 先运行测试，测试通过后才构建并发布
