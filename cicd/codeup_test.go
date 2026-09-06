package cicd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"
)

const flowTestOrgID = "org-123"

// flowTestServer 模拟云效 Flow OAPI v1。
type flowTestServer struct {
	server *httptest.Server
}

// pipelinesJSON 构造流水线条目。sources 为代码源地址列表。
func pipelineJSON(id int64, name string, sources ...string) map[string]interface{} {
	srcs := make([]map[string]string, 0, len(sources))
	for _, s := range sources {
		srcs = append(srcs, map[string]string{"type": "CODEUP", "repo": s})
	}
	return map[string]interface{}{"id": id, "name": name, "sources": srcs}
}

func runJSON(id int64, status string, startMs, endMs int64, branch, sha string) map[string]interface{} {
	return map[string]interface{}{
		"id":        id,
		"status":    status,
		"startTime": startMs,
		"endTime":   endMs,
		"sources":   []map[string]string{{"branch": branch, "sha": sha}},
	}
}

func newFlowTestServer(t *testing.T, pipelines []map[string]interface{}, runsByPipeline map[int64][]map[string]interface{}) *flowTestServer {
	ts := &flowTestServer{}
	runsRe := regexp.MustCompile(`^/oapi/v1/flow/organizations/[^/]+/pipelines/(\d+)/runs$`)
	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		base := fmt.Sprintf("/oapi/v1/flow/organizations/%s", flowTestOrgID)

		if r.URL.Path == base+"/pipelines" {
			w.Header().Set("x-total", fmt.Sprint(len(pipelines)))
			json.NewEncoder(w).Encode(pipelines)
			return
		}

		if m := runsRe.FindStringSubmatch(r.URL.Path); m != nil {
			var pid int64
			fmt.Sscanf(m[1], "%d", &pid)
			json.NewEncoder(w).Encode(runsByPipeline[pid])
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found: " + r.URL.Path))
	}))
	t.Cleanup(ts.server.Close)
	return ts
}

func flowParams(serverURL string) ListPipelinesParams {
	return ListPipelinesParams{
		ServerURL:      serverURL,
		Token:          "tok",
		RepoPath:       "wii/solo/grepom",
		Limit:          5,
		OrganizationID: flowTestOrgID,
	}
}

func TestCodeupPipeline_ListMatchedSingle(t *testing.T) {
	pipelines := []map[string]interface{}{
		pipelineJSON(100, "ci", "https://codeup.aliyun.com/wii/solo/grepom.git"),
		pipelineJSON(200, "other", "https://codeup.aliyun.com/other/repo.git"),
	}
	runs := map[int64][]map[string]interface{}{
		100: {runJSON(5001, "SUCCESS", 1690000000000, 1690000154000, "master", "abc1234567890")},
	}
	ts := newFlowTestServer(t, pipelines, runs)

	p := &CodeupPipelineProvider{}
	result, err := p.ListPipelines(context.Background(), flowParams(ts.server.URL))
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1 (只匹配代码源指向目标仓库的流水线)", len(result))
	}
	pl := result[0]
	if pl.ID != 5001 {
		t.Errorf("ID = %d", pl.ID)
	}
	if pl.Status != StatusSuccess {
		t.Errorf("Status = %q", pl.Status)
	}
	if pl.Branch != "master" {
		t.Errorf("Branch = %q", pl.Branch)
	}
	if pl.SHA != "abc1234" {
		t.Errorf("SHA = %q, want 短 7 位", pl.SHA)
	}
	if pl.Duration != 154*time.Second {
		t.Errorf("Duration = %v", pl.Duration)
	}
	if !strings.Contains(pl.URL, "/pipelines/100") {
		t.Errorf("URL = %q, 应包含流水线 ID", pl.URL)
	}
}

func TestCodeupPipeline_MatchSSHSource(t *testing.T) {
	pipelines := []map[string]interface{}{
		pipelineJSON(100, "ci", "git@codeup.aliyun.com:wii/solo/grepom.git"),
	}
	runs := map[int64][]map[string]interface{}{
		100: {runJSON(1, "RUNNING", 1690000000000, 0, "main", "deadbeef")},
	}
	ts := newFlowTestServer(t, pipelines, runs)

	p := &CodeupPipelineProvider{}
	result, err := p.ListPipelines(context.Background(), flowParams(ts.server.URL))
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ssh 代码源未匹配，len = %d", len(result))
	}
	if result[0].Status != StatusRunning {
		t.Errorf("Status = %q", result[0].Status)
	}
}

func TestCodeupPipeline_MergeMultiplePipelinesSortedDesc(t *testing.T) {
	pipelines := []map[string]interface{}{
		pipelineJSON(100, "ci", "https://codeup.aliyun.com/wii/solo/grepom.git"),
		pipelineJSON(200, "release", "https://codeup.aliyun.com/wii/solo/grepom.git"),
	}
	runs := map[int64][]map[string]interface{}{
		100: {
			runJSON(1, "SUCCESS", 1690000000000, 1690000060000, "master", "aaa1111"),
			runJSON(3, "RUNNING", 1690000200000, 0, "master", "ccc3333"),
		},
		200: {
			runJSON(2, "FAIL", 1690000100000, 1690000400000, "master", "bbb2222"),
		},
	}
	ts := newFlowTestServer(t, pipelines, runs)

	params := flowParams(ts.server.URL)
	params.Limit = 2
	p := &CodeupPipelineProvider{}
	result, err := p.ListPipelines(context.Background(), params)
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2 (Limit)", len(result))
	}
	// 按开始时间倒序：run 3 (200s) > run 2 (100s)
	if result[0].ID != 3 || result[1].ID != 2 {
		t.Errorf("order = [%d %d], want [3 2]", result[0].ID, result[1].ID)
	}
	if result[1].Status != StatusFailed {
		t.Errorf("FAIL 状态映射 = %q, want failed", result[1].Status)
	}
}

func TestCodeupPipeline_NoMatchReturnsEmpty(t *testing.T) {
	pipelines := []map[string]interface{}{
		pipelineJSON(100, "ci", "https://codeup.aliyun.com/other/repo.git"),
	}
	ts := newFlowTestServer(t, pipelines, nil)

	p := &CodeupPipelineProvider{}
	result, err := p.ListPipelines(context.Background(), flowParams(ts.server.URL))
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if result != nil {
		t.Errorf("result = %v, want nil", result)
	}
}

func TestCodeupPipeline_MissingOrganizationID(t *testing.T) {
	params := flowParams("https://codeup.aliyun.com")
	params.OrganizationID = ""

	p := &CodeupPipelineProvider{}
	_, err := p.ListPipelines(context.Background(), params)
	if err == nil || !strings.Contains(err.Error(), "organization_id") {
		t.Errorf("ListPipelines err = %v", err)
	}

	_, err = p.GetPipeline(context.Background(), GetPipelineParams{
		ServerURL: params.ServerURL, Token: "t", RepoPath: "wii/solo/grepom", PipelineID: 1,
	})
	if err == nil || !strings.Contains(err.Error(), "organization_id") {
		t.Errorf("GetPipeline err = %v", err)
	}
}

func TestCodeupPipeline_GetPipelineFindsRun(t *testing.T) {
	pipelines := []map[string]interface{}{
		pipelineJSON(100, "ci", "https://codeup.aliyun.com/wii/solo/grepom.git"),
		pipelineJSON(200, "release", "https://codeup.aliyun.com/wii/solo/grepom.git"),
	}
	runs := map[int64][]map[string]interface{}{
		100: {runJSON(1, "SUCCESS", 1690000000000, 1690000060000, "master", "aaa1111")},
		200: {runJSON(2, "CANCELED", 1690000100000, 1690000200000, "master", "bbb2222")},
	}
	ts := newFlowTestServer(t, pipelines, runs)

	p := &CodeupPipelineProvider{}
	pl, err := p.GetPipeline(context.Background(), GetPipelineParams{
		ServerURL: ts.server.URL, Token: "tok", RepoPath: "wii/solo/grepom",
		PipelineID: 2, OrganizationID: flowTestOrgID,
	})
	if err != nil {
		t.Fatalf("GetPipeline: %v", err)
	}
	if pl.ID != 2 || pl.Status != StatusCanceled {
		t.Errorf("pipeline = %+v", pl)
	}

	_, err = p.GetPipeline(context.Background(), GetPipelineParams{
		ServerURL: ts.server.URL, Token: "tok", RepoPath: "wii/solo/grepom",
		PipelineID: 999, OrganizationID: flowTestOrgID,
	})
	if err == nil || !strings.Contains(err.Error(), "#999 not found") {
		t.Errorf("not found err = %v", err)
	}
}

func TestCodeupPipeline_ParseFlowTimeFormats(t *testing.T) {
	pipelines := []map[string]interface{}{
		pipelineJSON(100, "ci", "https://codeup.aliyun.com/wii/solo/grepom.git"),
	}
	// RFC3339 字符串时间
	rfcRun := map[string]interface{}{
		"id":        1,
		"status":    "SUCCESS",
		"startTime": "2023-07-22T02:26:40Z",
		"endTime":   "2023-07-22T02:29:14Z",
		"sources":   []map[string]string{{"branch": "master", "sha": "abc1234"}},
	}
	ts := newFlowTestServer(t, pipelines, map[int64][]map[string]interface{}{100: {rfcRun}})

	p := &CodeupPipelineProvider{}
	result, err := p.ListPipelines(context.Background(), flowParams(ts.server.URL))
	if err != nil {
		t.Fatalf("ListPipelines: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("len = %d", len(result))
	}
	if result[0].Duration != 154*time.Second {
		t.Errorf("RFC3339 Duration = %v", result[0].Duration)
	}
}

func TestMapFlowStatus(t *testing.T) {
	cases := map[string]PipelineStatus{
		"RUNNING":   StatusRunning,
		"QUEUING":   StatusPending,
		"SUCCESS":   StatusSuccess,
		"FAIL":      StatusFailed,
		"CANCELED":  StatusCanceled,
		"CANCELLED": StatusCanceled,
	}
	for in, want := range cases {
		if got := mapFlowStatus(in); got != want {
			t.Errorf("mapFlowStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeGitPath(t *testing.T) {
	cases := map[string]string{
		"https://codeup.aliyun.com/wii/solo/grepom.git": "wii/solo/grepom",
		"git@codeup.aliyun.com:wii/solo/grepom.git":     "wii/solo/grepom",
		"http://codeup.example.com/a/b.git":             "a/b",
		"ssh://git@codeup.aliyun.com/wii/solo/grepom":   "wii/solo/grepom",
	}
	for in, want := range cases {
		if got := normalizeGitPath(in); got != want {
			t.Errorf("normalizeGitPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFlowRunWebURL(t *testing.T) {
	if got := flowRunWebURL("https://codeup.aliyun.com", 100); got != "https://flow.aliyun.com/pipelines/100" {
		t.Errorf("official URL = %q", got)
	}
	if got := flowRunWebURL("http://flow.example.com", 100); got != "http://flow.example.com/flow/pipelines/100" {
		t.Errorf("custom URL = %q", got)
	}
}
