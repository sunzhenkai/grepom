package scanner

import (
	"encoding/json"
	"strings"
)

// Severity 表示发现项的严重程度。
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
)

// Finding 表示一次敏感信息扫描的发现项。
type Finding struct {
	Repo        string   `json:"repo"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	RuleID      string   `json:"rule_id"`
	Description string   `json:"description"`
	Secret      string   `json:"secret"`
	Severity    Severity `json:"severity"`
}

// MaskSecret 对 secret 进行部分脱敏，仅显示前 8 个字符，其余用 "..." 替代。
// 如果 secret 长度 <= 8，则显示前半部分 + "..."。
func MaskSecret(secret string) string {
	const visibleChars = 8
	// 去除首尾空白
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return ""
	}
	if len(secret) <= visibleChars {
		if len(secret) <= 3 {
			return secret[:1] + "..."
		}
		return secret[:len(secret)/2] + "..."
	}
	return secret[:visibleChars] + "..."
}

// TruncatePath 截断过长的文件路径，保留路径前部分和文件名，中间用 "..." 替代。
// 如果路径长度不超过 maxLen，则原样返回。
func TruncatePath(path string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 40
	}
	if len(path) <= maxLen {
		return path
	}

	// 找到最后一个路径分隔符，保留文件名
	sep := "/"
	idx := strings.LastIndex(path, sep)
	if idx < 0 {
		// 没有路径分隔符，直接截断前面部分
		headLen := maxLen - 3 // 3 = len("...")
		if headLen <= 0 {
			headLen = 1
		}
		return path[:headLen] + "..."
	}

	filename := path[idx:]                // 包含前导 "/"
	headLen := maxLen - len(filename) - 3 // 3 = len("...")
	if headLen <= 0 {
		// 文件名本身就超长，截断文件名
		return path[:maxLen-3] + "..."
	}
	return path[:headLen] + "..." + filename
}

// FindingsToJSON 将发现项列表序列化为 JSON 数组。
func FindingsToJSON(findings []Finding) ([]byte, error) {
	return json.MarshalIndent(findings, "", "  ")
}
