## 1. 配置与数据模型

- [x] 1.1 扩展配置结构，新增可选 `services` 字段，支持服务名、`cwd` 和字符串/数组两种 `command` 形式
- [x] 1.2 实现服务配置解析与路径归一化，确保相对 `cwd` 按配置文件目录解析
- [x] 1.3 设计并实现服务运行态记录模型，包含 name、pid、pgid、cwd、command、log_path、started_at、last_status、exit_status 和 config_path
- [x] 1.4 实现按配置路径或当前目录隔离的 registry/log 状态目录解析，避免污染 `.grepom.yml`
- [x] 1.5 为配置解析、路径解析和 registry scope 生成添加单元测试

## 2. 服务管理核心

- [x] 2.1 新增服务 manager 包或模块，封装 Run、List、Status、Logs、Kill、Clean、Dir 等核心操作
- [x] 2.2 实现 registry 读写与文件锁，支持并发命令下的安全更新
- [x] 2.3 实现服务启动逻辑，支持直接命令、配置命令、默认服务名和日志文件追加写入
- [x] 2.4 实现 Unix 进程组启动与记录，无法使用进程组时提供降级行为
- [x] 2.5 实现进程存活探测和状态归类，至少支持 running、exited、stale
- [x] 2.6 实现同一 registry scope 内同名运行服务的重复启动保护
- [x] 2.7 为启动、记录、状态探测和重复启动保护添加单元测试或集成测试

## 3. CLI 子命令

- [x] 3.1 新增 `grepom svc` 命令和 `grepom service` 别名，并接入 root command
- [x] 3.2 实现 `svc run [name] -- <command> [args...]` 和 `svc run [name]` 配置启动
- [x] 3.3 实现 `svc list` 和 `svc status [name]`，使用表格展示名称、状态、PID、路径、命令和日志路径
- [x] 3.4 实现 `svc logs [name]`、`svc logs -n <count> [name]`、`svc logs -f [name]` 和 `svc logs --open [name]`
- [x] 3.5 实现 `svc kill [name]` 和 `svc kill -9 [name]`，优先向进程组发送信号
- [x] 3.6 实现 `svc clean`，默认仅清理 exited/stale 记录并保留日志，显式参数才删除日志
- [x] 3.7 实现 `svc dir [name]` 和 `svc --shell`，支持 shell 中 cd 到服务目录
- [x] 3.8 为各 CLI 命令补充参数校验、错误信息和命令帮助

## 4. 日志与编辑器体验

- [x] 4.1 实现高效读取日志末尾 N 行的逻辑，避免大日志文件一次性全部加载
- [x] 4.2 实现日志 follow 能力，持续输出新增内容并正确响应中断信号
- [x] 4.3 实现 `--open` 的打开逻辑，按 `$VISUAL`、`$EDITOR`、平台默认 opener、打印路径的顺序降级
- [x] 4.4 为日志读取、follow 取消和 editor/open 降级行为添加测试

## 5. TUI 管理界面

- [x] 5.1 评估并引入 TUI 依赖，优先保持核心 service manager 不依赖 TUI 库
- [x] 5.2 实现 `grepom svc tui` 入口，启动服务列表界面
- [x] 5.3 在 TUI 中展示服务名称、状态、PID、路径、命令和日志路径或等价详情视图
- [x] 5.4 实现 TUI 状态刷新、选择服务、查看日志尾部、停止服务、强制停止服务和清理退出服务
- [x] 5.5 实现 TUI 中显示或复制服务路径的交互
- [x] 5.6 为 TUI model/update 逻辑添加可测试覆盖，避免只依赖人工测试

## 6. 文档与示例

- [x] 6.1 更新 `README.md`，加入 `services` 配置示例、`svc` 命令示例、list 表格字段、日志、清理和 TUI 说明
- [x] 6.2 更新 `README_en.md`，保持英文文档与中文文档一致
- [x] 6.3 更新 `grepom example` 输出，展示可选 `services` 配置字段
- [x] 6.4 在命令 help 示例中覆盖直接命令启动、配置启动、日志 follow、kill、clean、dir 和 TUI

## 7. 验证

- [x] 7.1 运行服务管理相关单元测试和 CLI 集成测试
- [x] 7.2 运行现有全量 Go 测试，确保配置扩展不破坏已有能力
- [x] 7.3 手动验证 `svc run/list/logs/logs -f/kill/kill -9/clean/dir/tui` 的主要流程
- [x] 7.4 验证 list/status 表格包含路径和状态，并能正确反映已退出服务
- [x] 7.5 验证 README 中的示例命令与实际 CLI 行为一致
