package cmd

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

// isConnectionError 判断错误是否为 TCP 连接错误（应触发 protocol fallback）。
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	// 兜底：检查常见连接错误子串
	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "dial tcp")
}

// buildHTTPURL 将 https URL 替换为 http URL。
func buildHTTPURL(httpsURL string) string {
	return strings.Replace(httpsURL, "https://", "http://", 1)
}

// warnHTTPFallback 输出 auto fallback 到 HTTP 的警告。
func warnHTTPFallback(resourceName string) {
	fmt.Fprintf(os.Stderr, "warning: resource %q: HTTPS unavailable, fell back to HTTP. To skip the retry, set url to \"http://...\"\n", resourceName)
}
