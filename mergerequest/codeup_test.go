package mergerequest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testOrgID = "org-123"

// codeupTestServer 构造一个模拟 Codeup OAPI v1 的测试服务器。
type codeupTestServer struct {
	server      *httptest.Server
	createdMR   *codeupCreateMRRequest
	createCalls int
}

func newCodeupTestServer(t *testing.T, repos []codeupRepoEntry, openMRs []codeupMRResponse, createStatus int) *codeupTestServer {
	ts := &codeupTestServer{}
	ts.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		base := fmt.Sprintf("/oapi/v1/codeup/organizations/%s", testOrgID)

		switch {
		case r.URL.Path == base+"/repositories":
			// 仓库列表：单页返回（无 x-next-page）
			w.Header().Set("x-total", fmt.Sprint(len(repos)))
			json.NewEncoder(w).Encode(repos)

		case r.URL.Path == base+"/changeRequests" && r.Method == http.MethodGet:
			// ListChangeRequests：服务端按 state 过滤（mock 不模拟，直接返回）
			json.NewEncoder(w).Encode(openMRs)

		case strings.HasSuffix(r.URL.Path, "/changeRequests") && r.Method == http.MethodPost:
			// CreateChangeRequest：POST /repositories/{id}/changeRequests
			ts.createCalls++
			if createStatus >= 400 {
				w.WriteHeader(createStatus)
				w.Write([]byte("branch has no diff"))
				return
			}
			var body codeupCreateMRRequest
			json.NewDecoder(r.Body).Decode(&body)
			ts.createdMR = &body
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(codeupMRResponse{
				LocalID:      7,
				Title:        body.Title,
				Description:  body.Description,
				State:        "UNDER_REVIEW",
				WebURL:       "https://codeup.aliyun.com/wii/solo/grepom/change/7",
				SourceBranch: body.SourceBranch,
				TargetBranch: body.TargetBranch,
				ProjectID:    body.TargetProjectID,
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found: " + r.URL.Path))
		}
	}))
	t.Cleanup(ts.server.Close)
	return ts
}

func testCodeupParams(serverURL string) CreateMergeRequestParams {
	return CreateMergeRequestParams{
		ServerURL:      serverURL,
		Token:          "tok",
		RepoPath:       "wii/solo/grepom",
		OrganizationID: testOrgID,
		Title:          "feat: add x",
		Description:    "desc",
		SourceBranch:   "feature-x",
		TargetBranch:   "master",
	}
}

func TestCodeupMRProvider_CreateSuccess(t *testing.T) {
	repos := []codeupRepoEntry{{ID: 42, PathWithNamespace: "wii/solo/grepom"}}
	ts := newCodeupTestServer(t, repos, nil, http.StatusOK)

	p := &CodeupMRProvider{}
	mr, err := p.CreateMergeRequest(context.Background(), testCodeupParams(ts.server.URL))
	if err != nil {
		t.Fatalf("CreateMergeRequest: %v", err)
	}

	if mr.Number != 7 {
		t.Errorf("Number = %d, want 7 (localId)", mr.Number)
	}
	if mr.URL != "https://codeup.aliyun.com/wii/solo/grepom/change/7" {
		t.Errorf("URL = %q", mr.URL)
	}
	if mr.State != "open" {
		t.Errorf("State = %q, want open (mapped from UNDER_REVIEW)", mr.State)
	}
	if mr.AlreadyExists {
		t.Error("AlreadyExists should be false for new MR")
	}
	if ts.createdMR == nil {
		t.Fatal("create request not captured")
	}
	if ts.createdMR.SourceBranch != "feature-x" || ts.createdMR.TargetBranch != "master" {
		t.Errorf("branches = %q -> %q", ts.createdMR.SourceBranch, ts.createdMR.TargetBranch)
	}
	if ts.createdMR.SourceProjectID != 42 || ts.createdMR.TargetProjectID != 42 {
		t.Errorf("projectIDs = %d/%d, want 42/42 (同库 MR)",
			ts.createdMR.SourceProjectID, ts.createdMR.TargetProjectID)
	}
	if ts.createdMR.Title != "feat: add x" {
		t.Errorf("title = %q", ts.createdMR.Title)
	}
}

func TestCodeupMRProvider_CreateDraftPrefix(t *testing.T) {
	repos := []codeupRepoEntry{{ID: 42, PathWithNamespace: "wii/solo/grepom"}}
	ts := newCodeupTestServer(t, repos, nil, http.StatusOK)

	params := testCodeupParams(ts.server.URL)
	params.Draft = true

	p := &CodeupMRProvider{}
	if _, err := p.CreateMergeRequest(context.Background(), params); err != nil {
		t.Fatalf("CreateMergeRequest: %v", err)
	}
	if ts.createdMR.Title != "Draft: feat: add x" {
		t.Errorf("draft title = %q", ts.createdMR.Title)
	}
}

func TestCodeupMRProvider_AlreadyExists(t *testing.T) {
	repos := []codeupRepoEntry{{ID: 42, PathWithNamespace: "wii/solo/grepom"}}
	openMRs := []codeupMRResponse{
		{ // 其他源分支的 MR 不应命中
			LocalID: 4, Title: "other", State: "UNDER_REVIEW",
			SourceBranch: "other-branch", TargetBranch: "master",
		},
		{
			LocalID: 5, Title: "old mr", State: "UNDER_REVIEW",
			WebURL: "https://codeup.aliyun.com/wii/solo/grepom/change/5",
			SourceBranch: "feature-x", TargetBranch: "master",
		},
	}
	ts := newCodeupTestServer(t, repos, openMRs, http.StatusOK)

	p := &CodeupMRProvider{}
	mr, err := p.CreateMergeRequest(context.Background(), testCodeupParams(ts.server.URL))
	if err != nil {
		t.Fatalf("CreateMergeRequest: %v", err)
	}
	if !mr.AlreadyExists {
		t.Error("AlreadyExists should be true")
	}
	if mr.Number != 5 {
		t.Errorf("Number = %d, want 5 (客户端按源分支过滤)", mr.Number)
	}
	if ts.createCalls != 0 {
		t.Errorf("create should not be called when open MR exists, got %d calls", ts.createCalls)
	}
}

func TestCodeupMRProvider_RepoNotFound(t *testing.T) {
	repos := []codeupRepoEntry{{ID: 42, PathWithNamespace: "other/repo"}}
	ts := newCodeupTestServer(t, repos, nil, http.StatusOK)

	p := &CodeupMRProvider{}
	_, err := p.CreateMergeRequest(context.Background(), testCodeupParams(ts.server.URL))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `repository not found for path "wii/solo/grepom"`) {
		t.Errorf("err = %v", err)
	}
}

func TestCodeupMRProvider_MissingOrganizationID(t *testing.T) {
	params := testCodeupParams("https://codeup.aliyun.com")
	params.OrganizationID = ""

	p := &CodeupMRProvider{}
	_, err := p.CreateMergeRequest(context.Background(), params)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "organization_id") {
		t.Errorf("err = %v", err)
	}
}

func TestCodeupMRProvider_CreateAPIError(t *testing.T) {
	repos := []codeupRepoEntry{{ID: 42, PathWithNamespace: "wii/solo/grepom"}}
	ts := newCodeupTestServer(t, repos, nil, http.StatusBadRequest)

	p := &CodeupMRProvider{}
	_, err := p.CreateMergeRequest(context.Background(), testCodeupParams(ts.server.URL))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("err = %v", err)
	}
}

func TestCodeupMRProvider_BuildWebURL(t *testing.T) {
	p := &CodeupMRProvider{}
	u := p.BuildWebURL(WebURLParams{
		ServerURL:    "https://codeup.aliyun.com",
		RepoPath:     "wii/solo/grepom",
		SourceBranch: "feature-x",
		TargetBranch: "master",
	})
	want := "https://codeup.aliyun.com/wii/solo/grepom/merge_requests/new?source=feature-x&target=master"
	if u != want {
		t.Errorf("BuildWebURL = %q, want %q", u, want)
	}
}

func TestMapCodeupMRState(t *testing.T) {
	cases := map[string]string{
		"UNDER_DEV":     "open",
		"UNDER_REVIEW":  "open",
		"TO_BE_MERGED":  "open",
		"MERGED":        "merged",
		"CLOSED":        "closed",
	}
	for in, want := range cases {
		if got := mapCodeupMRState(in); got != want {
			t.Errorf("mapCodeupMRState(%q) = %q, want %q", in, got, want)
		}
	}
}
