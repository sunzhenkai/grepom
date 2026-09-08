package cmd

// 正式回归测试：`tag -w` 必须按 SHA 等待新 tag 的 pipeline 出现，
// 不得在竞态窗口内退而监控旧 pipeline（GitHub Actions 异步创建 run）。
// 场景对应 specs/tag-watch "Requirement: -w 监控新 tag 对应的 pipeline"。

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/wii/grepom/cicd"
)

type fakeAsyncProvider struct {
	mu           sync.Mutex
	newVisible   time.Time // 新 run 的可见时刻；zero 表示永不出现
	oldRun       cicd.Pipeline
	newRun       cicd.Pipeline
	watchedNew   bool
	finishedOld  bool
	seenSHAFiter bool // 收到过带 SHA 过滤的请求
}

func newFakeAsyncProvider(newVisible time.Time) *fakeAsyncProvider {
	return &fakeAsyncProvider{
		newVisible: newVisible,
		oldRun: cicd.Pipeline{
			ID: 1000, Status: cicd.StatusSuccess, Branch: "v0.1.5",
			SHA: "aaa1111", Duration: 90 * time.Second,
			StartedAt: time.Now().Add(-10 * time.Minute),
		},
		newRun: cicd.Pipeline{
			ID: 2001, Status: cicd.StatusRunning, Branch: "v0.1.6",
			SHA: "bbb2222", StartedAt: time.Now(),
		},
	}
}

const newRunFullSHA = "bbb2222ccc3333ddd4444eee5555fff6666aaa77"

func (p *fakeAsyncProvider) ListPipelines(_ context.Context, params cicd.ListPipelinesParams) ([]cicd.Pipeline, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if params.SHA != "" {
		p.seenSHAFiter = true
	}
	visible := []cicd.Pipeline{p.oldRun}
	if !p.newVisible.IsZero() && !time.Now().Before(p.newVisible) {
		visible = []cicd.Pipeline{p.newRun, p.oldRun}
	}
	if params.SHA == "" {
		return visible, nil
	}
	// 模拟 provider 侧 SHA 过滤。
	filtered := visible[:0:0]
	for _, pl := range visible {
		if strings.HasPrefix(newRunFullSHA, pl.SHA) && params.SHA == newRunFullSHA {
			filtered = append(filtered, pl)
		}
	}
	return filtered, nil
}

func (p *fakeAsyncProvider) GetPipeline(_ context.Context, params cicd.GetPipelineParams) (*cicd.Pipeline, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if params.PipelineID == 2001 {
		p.watchedNew = true
		// 新 run 运行 2 秒后转为 success，让 watch 循环能到达终态。
		pl := p.newRun
		if !p.newVisible.IsZero() && time.Now().After(p.newVisible.Add(2*time.Second)) {
			pl.Status = cicd.StatusSuccess
			pl.Duration = 2 * time.Second
		}
		return &pl, nil
	}
	pl := p.oldRun
	if params.PipelineID == 1000 {
		p.finishedOld = true
	}
	return &pl, nil
}

func regressionWatchTarget(prov cicd.PipelineProvider) WatchTarget {
	return WatchTarget{
		Provider:  prov,
		ServerURL: "https://github.com",
		RepoPath:  "owner/repo",
		Token:     "x",
		RepoName:  "repo",
		WatchTag:  "v0.1.6",
		WatchSHA:  newRunFullSHA,
	}
}

// 场景：新 pipeline 未立即出现时持续等待，且绝不监控旧 pipeline。
func TestTagWatchWaitsForNewPipelineBySHA(t *testing.T) {
	oldTimeout, oldPoll := watchSHAWaitTimeout, watchSHAPollEvery
	watchSHAWaitTimeout, watchSHAPollEvery = 10*time.Second, 200*time.Millisecond
	defer func() { watchSHAWaitTimeout, watchSHAPollEvery = oldTimeout, oldPoll }()

	prov := newFakeAsyncProvider(time.Now().Add(1500 * time.Millisecond))
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	start := time.Now()
	err := runWatchLoop(regressionWatchTarget(prov), 0, cmd)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("runWatchLoop: %v", err)
	}

	prov.mu.Lock()
	defer prov.mu.Unlock()
	if prov.finishedOld {
		t.Errorf("BUG: watched/finished the OLD run #1000")
	}
	if !prov.watchedNew {
		t.Errorf("BUG: new run #2001 was never watched")
	}
	if !prov.seenSHAFiter {
		t.Errorf("BUG: wait phase did not pass SHA filter to ListPipelines")
	}
	if elapsed < 1500*time.Millisecond {
		t.Errorf("BUG: returned after %v, before the new run even existed", elapsed)
	}
}

// 场景：超时未出现——必须报错，不得监控任何其他 pipeline。
func TestTagWatchTimeoutWhenPipelineNeverAppears(t *testing.T) {
	oldTimeout, oldPoll := watchSHAWaitTimeout, watchSHAPollEvery
	watchSHAWaitTimeout, watchSHAPollEvery = 600*time.Millisecond, 100*time.Millisecond
	defer func() { watchSHAWaitTimeout, watchSHAPollEvery = oldTimeout, oldPoll }()

	prov := newFakeAsyncProvider(time.Time{}) // 永不出现
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())

	err := runWatchLoop(regressionWatchTarget(prov), 0, cmd)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	msg := err.Error()
	for _, want := range []string{"v0.1.6", "bbb2222", "not pushed"} {
		if !strings.Contains(msg, want) {
			t.Errorf("timeout error %q missing %q", msg, want)
		}
	}

	prov.mu.Lock()
	defer prov.mu.Unlock()
	if prov.finishedOld || prov.watchedNew {
		t.Errorf("BUG: monitored a pipeline during wait phase (old=%v new=%v)", prov.finishedOld, prov.watchedNew)
	}
}

// 场景：等待阶段 Ctrl+C（ctx 取消）——优雅退出，无错误。
func TestTagWatchCancelWhileWaiting(t *testing.T) {
	oldTimeout, oldPoll := watchSHAWaitTimeout, watchSHAPollEvery
	watchSHAWaitTimeout, watchSHAPollEvery = 30*time.Second, 200*time.Millisecond
	defer func() { watchSHAWaitTimeout, watchSHAPollEvery = oldTimeout, oldPoll }()

	prov := newFakeAsyncProvider(time.Time{})
	ctx, cancel := context.WithCancel(context.Background())
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()

	err := runWatchLoop(regressionWatchTarget(prov), 0, cmd)
	if err != nil {
		t.Fatalf("expected graceful stop, got error: %v", err)
	}

	prov.mu.Lock()
	defer prov.mu.Unlock()
	if prov.finishedOld || prov.watchedNew {
		t.Errorf("BUG: monitored a pipeline after cancel (old=%v new=%v)", prov.finishedOld, prov.watchedNew)
	}
}
