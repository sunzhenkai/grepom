## Context

当前 `Config.Groups` 表示真实操作单元：它既包含远端 `resource/path`，也包含本地 `local_path`、发现后的 `repos`、`exclude_repos`、`enabled` 等运行时字段。多数命令通过单值 `--group` 过滤真实 group，部分命令直接遍历 `cfg.Groups`，部分命令通过 `repo.Resolver` 展平仓库。

虚拟分组的核心需求是“把多个真实 group 命名为一个场景集合”，例如 `personal`、`work`，并让命令可以批量作用于这些真实 group。它不应该承载远端路径、repo 列表或启停状态，否则会和现有真实 group 的职责混淆。

## Goals / Non-Goals

**Goals:**

- 增加 `virtual_groups` 配置，表达虚拟分组到真实 group 的一对多映射。
- 支持 `--vgroup` 参数选择虚拟分组包含的真实 groups。
- `--group` 与 `--vgroup` 同时指定时按并集执行。
- 允许虚拟分组和真实 group 同名。
- 在 `grepom list groups` 中同时展示真实 group 与虚拟分组，并用类型列区分。
- 让 clone/list/status/pull/search/scan/prune/sync/dedup/list --remote 等现有 group 过滤命令复用同一套展开逻辑。

**Non-Goals:**

- 不支持虚拟分组嵌套虚拟分组。
- 不为虚拟分组引入 `enabled`、`exclude_repos`、`resource`、`path`、`local_path`、`repos` 等真实 group 字段。
- 不改变真实 group 的本地路径推导、远程同步、排除规则或认证优先级。
- 不改变 `--resource` 的语义；它继续作为资源过滤器，并与 group 集合过滤取交集。

## Decisions

### 使用独立顶层 `virtual_groups`

配置新增顶层字段：

```yaml
virtual_groups:
  work:
    groups:
      - work-frontend
      - work-backend
  personal:
    groups:
      - github-oss
      - home-lab
```

选择 map 结构而不是数组结构，是因为虚拟分组只需要名称到成员列表的映射，且不需要维护声明顺序。这样读取某个 `--vgroup` 时也更直接。真实 group 仍保留数组结构，因为它们已有顺序、repo 列表和写回行为。

备选方案是把虚拟分组也放进 `groups` 数组，使用某个字段区分类型。这个方案会让真实 group 的校验和写回逻辑变复杂，也容易误用 `resource/path/repos` 字段，因此不采用。

### 只引用真实 group

虚拟分组的 `groups` 成员只解析为真实 group 名称，不解析其他虚拟分组。加载配置时需要校验每个成员都存在于 `groups` 中；虚拟分组名称允许与真实 group 名称相同，因为 lookup 使用 `--vgroup` 的命名空间。

备选方案是支持嵌套虚拟分组，但这会引入循环检测、展开顺序和错误定位问题。当前痛点是按场景批量选择真实 group，直接列表足够。

### 公共展开层

在配置或命令公共层提供一个函数，将 `groupName` 与 `vgroupName` 展开成去重后的真实 group 名称集合：

```text
ResolveGroupSelection(group, vgroup):
  selected = set()
  if group != "":
    require real group exists
    selected.add(group)
  if vgroup != "":
    require virtual group exists
    for each member:
      require real group exists
      selected.add(member)
  return selected
```

如果两者都为空，表示不按 group 限定，保持现有“全部 group”行为。命令随后再按 `--resource`、repo name 等现有条件继续过滤。

### `repo.Filter` 支持多 group

`repo.Filter` 目前只有单值 `Group`。本变更应增加多值 group 集合，例如 `Groups []string` 或内部 `GroupSet map[string]bool`。为了保持调用点简单，命令层可以继续接受字符串参数，由公共展开函数填充多 group 字段。

兼容策略：保留 `Group` 字段给未迁移代码或测试使用；`ApplyFilter` 同时支持单值 `Group` 与多值 `Groups`，两者按并集处理。实现完成后可逐步让命令统一使用多值字段。

### `list groups` 使用类型列区分

`grepom list groups` 输出新增 `TYPE` 列：

- `group`: 真实 group，显示 RESOURCE、PATH、LOCAL_PATH、RECURSIVE、REPOS。
- `vgroup`: 虚拟分组，显示成员 group 列表和成员 repo 总数；真实 group 专属列使用 `-`。

输出可读性优先于完全复用旧表头。旧脚本如果依赖列顺序可能受影响，但这是该命令展示语义变化带来的合理代价。

## Risks / Trade-offs

- [Risk] 多个命令各自处理 group 过滤，容易漏掉 `--vgroup`。→ Mitigation：先实现公共展开函数和 `repo.Filter` 多 group 能力，再逐个命令接入并补测试。
- [Risk] `list groups` 表头变化可能影响依赖文本输出的脚本。→ Mitigation：在 README 中说明新增 TYPE；真实 group 行继续包含原有核心信息。
- [Risk] 虚拟分组成员引用不存在的真实 group 时，错误可能只在命令运行时暴露。→ Mitigation：配置加载阶段校验所有成员引用，尽早失败。
- [Risk] `--group` 与 `--vgroup` 并集后再叠加 `--resource`，用户可能以为是并集。→ Mitigation：文档明确 group 选择先并集，resource 过滤再取交集。

## Migration Plan

现有配置文件无需修改；未配置 `virtual_groups` 时行为保持不变。新增字段默认为空 map。实现完成后更新 `grepom example`、`README.md`、`README_en.md`，为用户提供渐进迁移示例。
