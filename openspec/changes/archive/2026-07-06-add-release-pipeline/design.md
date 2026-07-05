## Context

grepom 是纯 Go CLI 工具（module `github.com/wii/grepom`，仓库 `sunzhenkai/grepom`），当前安装方式仅有 `go install` 和 `make install`。项目已有 v0.1.x 标签但无 GitHub Release 和 CI workflow。Makefile 已定义 `VERSION`（`git describe`）和 `-ldflags "-X main.version=$(VERSION)"`，但 `main.go` 缺少 `version` 变量，版本注入实际无效。

交叉编译实测：`linux` 和 `darwin` 的 `amd64`/`arm64` 均可直接构建；`windows` 因 `syscall.Flock`（`config/config.go`、`service/registry.go`）无法编译，需平台拆分，本次不在范围内。

## Goals / Non-Goals

**Goals:**

- push `v*` tag 时自动构建 4 平台二进制并发布到 GitHub Release
- tag 含 `-rc` 或 `-beta` 时标记为 Pre-release
- 生成 `checksums.txt`（sha256）供安装脚本校验
- 提供 `scripts/install.sh`，默认安装到 `~/.local/bin`，支持 `INSTALL_DIR=/usr/local/bin`
- 修复版本号注入，新增 `grepom version` 子命令
- 新增 CI workflow 在 push/PR 时运行测试
- 更新 README 安装说明

**Non-Goals:**

- Windows 平台构建与 `install.ps1`（需先修复 flock 平台拆分）
- GoReleaser 或 Homebrew formula
- 签名/notarization（macOS 公证）
- `workflow_dispatch` 手动触发（后续可加）
- 修改 `go install` 路径或 module 重定向

## Decisions

### 1. 原生 GitHub Actions matrix，不用 GoReleaser

**选择**：`release.yml` 使用 matrix 并行 `go build`，`softprops/action-gh-release` 上传资产。

**替代方案**：GoReleaser — 功能更全但对当前体量过重。

**理由**：4 平台静态二进制 + tar.gz，原生 matrix 足够简洁，无额外依赖。

### 2. 触发条件：仅 tag push

**选择**：`on.push.tags: ['v*']`。

**理由**：与现有 `grepom tag` 工作流一致；避免 main 每次 push 都产 Release。

### 3. Pre-release 判断

**选择**：tag 去掉 `v` 前缀后，若包含 `-rc` 或 `-beta` 子串，则 `prerelease: true`。

```bash
# 示例
v0.2.0        → 正式 release
v0.2.0-rc.1   → prerelease
v0.2.0-beta.2 → prerelease
```

**理由**：简单可靠，覆盖常见预发布命名；与探索阶段用户决策一致。

### 4. 资产命名与打包格式

**选择**：

```
grepom_{version}_{os}_{arch}.tar.gz
# 例: grepom_v0.2.0_linux_arm64.tar.gz
```

tar.gz 内仅含单个 `grepom` 二进制（无嵌套目录），解压即用。

**理由**：命名包含版本和平台，便于安装脚本按模式匹配；与 ripgrep/fd 等 CLI 惯例一致。

### 5. 版本号注入位置

**选择**：`main.go` 中 `var version = "dev"`，构建时 `-ldflags "-X main.version=..."`；`cmd/version.go` 通过 `main` 包导出或 `cmd` 包内 `var Version` 由 main 设置。

**实现**：在 `main.go` 定义 `version`，`cmd` 包通过 `SetVersion(v string)` 或直接在 `cmd` 包定义 `var Version string` 由 ldflags 注入 `github.com/wii/grepom/cmd.Version`（与 Makefile 对齐需更新 LDFLAGS 目标）。

**更新 Makefile**：`LDFLAGS := -ldflags "-X github.com/wii/grepom/cmd.Version=$(VERSION)"`

### 6. 安装脚本设计

**选择**：`scripts/install.sh`，bash，支持环境变量：

| 变量 | 默认 | 说明 |
|------|------|------|
| `VERSION` | `latest` | 指定 tag（如 `v0.2.0-rc.1`）或 `latest` |
| `INSTALL_DIR` | `$HOME/.local/bin` | 安装目录 |
| `REPO` | `sunzhenkai/grepom` | GitHub 仓库 |

流程：
1. 检测 OS（`linux`/`darwin`）和 ARCH（`amd64`/`arm64`）
2. 解析版本：`latest` → GitHub API `/releases/latest`；否则用指定 tag
3. 下载对应 tar.gz 和 `checksums.txt`
4. sha256 校验
5. 解压并 `chmod +x`，移动到 `INSTALL_DIR`
6. `INSTALL_DIR` 无写权限时提示使用 `sudo`
7. PATH 不含 `INSTALL_DIR` 时打印提示

**`/usr/local/bin` 支持**：用户执行 `sudo INSTALL_DIR=/usr/local/bin bash install.sh`。

### 7. CI workflow 独立

**选择**：`ci.yml` 在 push/PR 到 main 时 `go test ./...`；`release.yml` 在 tag 时额外先跑测试再构建。

**理由**：PR 阶段即发现问题；tag 发布前二次保障。

## Risks / Trade-offs

- **[module path 与仓库不一致]** `go install github.com/wii/grepom` vs 仓库 `sunzhenkai/grepom` → README 中明确两种安装方式，安装脚本基于 GitHub Release 不依赖 module path
- **[latest 不含 prerelease]** 安装 `latest` 不会装 rc/beta → 用户需 `VERSION=v0.2.0-rc.1` 显式指定，README 中说明
- **[无代码签名]** macOS 可能触发 Gatekeeper 警告 → 用户需 `xattr -d com.apple.quarantine` 或右键打开；后续可考虑 notarization
- **[GitHub Actions 额度]** 4 平台并行构建消耗 minutes → 仅 tag 触发，频率低，可接受
- **[Windows 延后]** Windows 用户仍需 `go install` → 文档注明，后续独立 change 修复 flock

## Migration Plan

1. 合并 change 后，对最新稳定 tag（如 `v0.1.4`）重新 push 或创建 `v0.1.5` 触发首次 Release
2. 验证 4 平台资产和 `checksums.txt` 正确
3. 测试 `install.sh` 在 macOS arm64 和 Linux amd64 上安装
4. 更新 README 安装章节

回滚：删除 workflow 文件即可，不影响现有 `go install` 路径。

## Open Questions

（无 — 探索阶段决策已收敛）
