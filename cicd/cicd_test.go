package cicd

import (
	"testing"
	"time"
)

func TestRegisterAndGet(t *testing.T) {
	// gitlab and github are auto-registered via init()
	providers := []string{"gitlab", "github"}
	for _, name := range providers {
		p, err := Get(name)
		if err != nil {
			t.Errorf("Get(%q) returned error: %v", name, err)
		}
		if p == nil {
			t.Errorf("Get(%q) returned nil", name)
		}
	}
}

func TestGet_UnsupportedProvider(t *testing.T) {
	_, err := Get("unsupported")
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
}

func TestPipelineStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status   PipelineStatus
		terminal bool
	}{
		{StatusRunning, false},
		{StatusPending, false},
		{StatusSuccess, true},
		{StatusFailed, true},
		{StatusCanceled, true},
		{PipelineStatus("unknown"), false},
	}
	for _, tt := range tests {
		if got := tt.status.IsTerminal(); got != tt.terminal {
			t.Errorf("%s.IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
		}
	}
}

func TestFormatStatus(t *testing.T) {
	tests := []struct {
		status PipelineStatus
		want   string
	}{
		{StatusRunning, "🔄 running"},
		{StatusPending, "⏳ pending"},
		{StatusSuccess, "✅ success"},
		{StatusFailed, "❌ failed"},
		{StatusCanceled, "🚫 canceled"},
		{PipelineStatus("unknown"), "unknown"},
	}
	for _, tt := range tests {
		if got := FormatStatus(tt.status); got != tt.want {
			t.Errorf("FormatStatus(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "-"},
		{5 * time.Second, "5s"},
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m30s"},
		{2*time.Minute + 34*time.Second, "2m34s"},
		{61 * time.Minute, "1h01m"},
	}
	for _, tt := range tests {
		if got := FormatDuration(tt.d); got != tt.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
