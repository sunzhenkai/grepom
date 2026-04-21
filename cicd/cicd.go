package cicd

import (
	"context"
	"fmt"
	"time"
)

// PipelineStatus 表示 CI/CD pipeline 的运行状态。
type PipelineStatus string

const (
	StatusRunning  PipelineStatus = "running"
	StatusPending  PipelineStatus = "pending"
	StatusSuccess  PipelineStatus = "success"
	StatusFailed   PipelineStatus = "failed"
	StatusCanceled PipelineStatus = "canceled"
)

// IsTerminal 返回 pipeline 是否处于终态。
func (s PipelineStatus) IsTerminal() bool {
	return s == StatusSuccess || s == StatusFailed || s == StatusCanceled
}

// Pipeline 表示一次 CI/CD pipeline 运行。
type Pipeline struct {
	ID        int
	Status    PipelineStatus
	Branch    string
	SHA       string        // 短 commit hash（前 7 位）
	StartedAt time.Time     // zero value 表示未开始
	Duration  time.Duration // 0 表示未开始或无数据
	URL       string        // Web URL
}

// PipelineProvider 是 CI/CD pipeline 的数据查询接口。
type PipelineProvider interface {
	ListPipelines(ctx context.Context, params ListPipelinesParams) ([]Pipeline, error)
	GetPipeline(ctx context.Context, params GetPipelineParams) (*Pipeline, error)
}

// ListPipelinesParams 包含列出 pipeline 的参数。
type ListPipelinesParams struct {
	ServerURL string
	Token     string
	RepoPath  string // "org/fe/web-app" 或 "owner/repo"
	Limit     int
}

// GetPipelineParams 包含获取单个 pipeline 的参数。
type GetPipelineParams struct {
	ServerURL  string
	Token      string
	RepoPath   string
	PipelineID int
}

// --- 注册表 ---

var registry = map[string]func() PipelineProvider{}

// Register 注册一个 PipelineProvider 工厂函数。
func Register(name string, factory func() PipelineProvider) {
	registry[name] = factory
}

// Get 通过名称获取 PipelineProvider 实例。
func Get(name string) (PipelineProvider, error) {
	factory, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unsupported cicd provider: %s", name)
	}
	return factory(), nil
}

// --- 格式化辅助函数 ---

// FormatStatus 返回带图标的状态文本。
func FormatStatus(status PipelineStatus) string {
	switch status {
	case StatusRunning:
		return "🔄 running"
	case StatusPending:
		return "⏳ pending"
	case StatusSuccess:
		return "✅ success"
	case StatusFailed:
		return "❌ failed"
	case StatusCanceled:
		return "🚫 canceled"
	default:
		return string(status)
	}
}

// FormatDuration 返回人可读的持续时间文本。
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	seconds := int(d.Seconds())
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	minutes := seconds / 60
	secs := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%dm%02ds", minutes, secs)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%dh%02dm", hours, mins)
}
