## 1. 版本号注入与 version 命令

- [x] 1.1 在 `cmd` 包新增 `var Version = "dev"`，更新 Makefile LDFLAGS 为 `-X github.com/wii/grepom/cmd.Version=$(VERSION)`
- [x] 1.2 新增 `cmd/version.go`，实现 `grepom version` 子命令并注册到 rootCmd
- [x] 1.3 为 version 命令添加单元测试

## 2. CI workflow

- [x] 2.1 新增 `.github/workflows/ci.yml`：push/PR 到 main 时运行 `go test ./...`

## 3. Release workflow

- [x] 3.1 新增 `.github/workflows/release.yml`：push `v*` tag 触发，先运行测试
- [x] 3.2 配置 matrix 构建 4 平台（linux/darwin × amd64/arm64），使用 `-trimpath` 和版本 ldflags
- [x] 3.3 打包 tar.gz（内含单个 `grepom` 二进制），命名格式 `grepom_{version}_{os}_{arch}.tar.gz`
- [x] 3.4 生成 `checksums.txt`（sha256）并上传所有资产到 GitHub Release
- [x] 3.5 实现 prerelease 判断：tag 含 `-rc` 或 `-beta` 时标记 `prerelease: true`

## 4. 安装脚本

- [x] 4.1 新增 `scripts/install.sh`：检测 OS/ARCH，支持 `VERSION` 和 `INSTALL_DIR` 环境变量
- [x] 4.2 实现从 GitHub Release 下载、sha256 校验、安装到目标目录
- [x] 4.3 处理权限不足（提示 sudo）和 PATH 不在安装目录时的提示

## 5. 文档更新

- [x] 5.1 更新 `README.md` 安装章节：curl 一键安装、INSTALL_DIR 示例、预发布版本安装说明
- [x] 5.2 更新 `README_en.md` 安装章节（与中文版对应）
