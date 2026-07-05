package provider

import "strings"

// deletionScheduledMarker 是 Codeup（云效）回收站/计划删除代码库的命名标记。
// 被删除的代码库或代码组会在 name 或 pathWithNamespace 中被重命名为包含此标记的形式
// （如 "creative-matching-deletion_scheduled-499"），这类库已不可克隆。
const deletionScheduledMarker = "deletion_scheduled"

// IsDeletionScheduled 判断一个代码库是否处于"计划删除"状态。
// 当 name 或 pathWithNamespace 包含 deletion_scheduled 标记时返回 true。
// 该检测被 provider 发现阶段与 resolver 运行时兜底共用。
func IsDeletionScheduled(name, pathWithNamespace string) bool {
	return strings.Contains(name, deletionScheduledMarker) ||
		strings.Contains(pathWithNamespace, deletionScheduledMarker)
}
