package config

import (
	"strings"
)

// NormalizeRepoURL 将 Git 仓库 URL 统一为 host/path 格式用于比较。
// 规则：
//  1. 去掉 https://、http:// 前缀
//  2. 去掉 .git 后缀
//  3. SSH 格式 git@host:path → host/path
//  4. 去掉末尾 /
//  5. host 部分转小写
//  6. path 部分保留原样（大小写敏感）
//  7. 端口号保留
func NormalizeRepoURL(url string) string {
	if url == "" {
		return ""
	}

	s := url

	// 去掉末尾 /（在去掉 .git 之前，避免 .git/ 遗留）
	s = strings.TrimRight(s, "/")

	// 去掉 .git 后缀
	s = strings.TrimSuffix(s, ".git")

	// 处理 https:// 和 http:// 前缀
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(s, prefix) {
			s = strings.TrimPrefix(s, prefix)
			// 分离 host 和 path，将 host 转小写
			return lowerHost(s)
		}
	}

	// 处理 SSH 格式 git@host:path
	if strings.HasPrefix(s, "git@") {
		s = strings.TrimPrefix(s, "git@")
		// 第一个 : 分隔 host 和 path
		idx := strings.Index(s, ":")
		if idx >= 0 {
			host := strings.ToLower(s[:idx])
			path := s[idx+1:]
			return host + "/" + path
		}
		return strings.ToLower(s)
	}

	// 已经是 host/path 格式（无协议前缀）
	return lowerHost(s)
}

// lowerHost 将 host 部分转小写，保留 path 不变。
// host 可能包含端口号（如 gitlab.com:8443）。
func lowerHost(s string) string {
	// 找第一个 / 分隔 host 和 path
	idx := strings.Index(s, "/")
	if idx < 0 {
		return strings.ToLower(s)
	}
	host := strings.ToLower(s[:idx])
	path := s[idx:]
	return host + path
}
