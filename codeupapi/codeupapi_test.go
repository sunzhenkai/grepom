package codeupapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIBaseURL(t *testing.T) {
	orgID := "60de7a6852743a5162b5f957"
	tests := []struct {
		name     string
		input    string
		product  string
		expected string
	}{
		{"codeup 官方域名", "codeup.aliyun.com", "codeup", "https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/" + orgID},
		{"codeup 官方域名带 scheme", "https://codeup.aliyun.com", "codeup", "https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/" + orgID},
		{"flow 产品", "codeup.aliyun.com", "flow", "https://openapi-rdc.aliyuncs.com/oapi/v1/flow/organizations/" + orgID},
		{"自定义域名", "codeup.example.com", "codeup", "https://codeup.example.com/oapi/v1/codeup/organizations/" + orgID},
		{"http 保留 scheme", "http://codeup.example.com", "codeup", "http://codeup.example.com/oapi/v1/codeup/organizations/" + orgID},
		{"末尾斜杠", "https://codeup.aliyun.com/", "codeup", "https://openapi-rdc.aliyuncs.com/oapi/v1/codeup/organizations/" + orgID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := APIBaseURL(tt.input, tt.product, orgID)
			if result != tt.expected {
				t.Errorf("APIBaseURL(%q, %q, %q) = %q, want %q", tt.input, tt.product, orgID, result, tt.expected)
			}
		})
	}
}

func TestCloneHost(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"codeup.aliyun.com", "codeup.aliyun.com"},
		{"https://codeup.aliyun.com", "codeup.aliyun.com"},
		{"http://codeup.example.com/", "codeup.example.com"},
	}

	for _, tt := range tests {
		if got := CloneHost(tt.input); got != tt.expected {
			t.Errorf("CloneHost(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParsePagination(t *testing.T) {
	resp := &http.Response{Header: http.Header{}}
	resp.Header.Set("x-total", "250")
	resp.Header.Set("x-total-pages", "3")
	resp.Header.Set("x-page", "1")
	resp.Header.Set("x-per-page", "100")
	resp.Header.Set("x-next-page", "2")

	pg := ParsePagination(resp)
	if pg.Total != 250 || pg.TotalPages != 3 || pg.Page != 1 || pg.PerPage != 100 || pg.NextPage != 2 {
		t.Errorf("ParsePagination = %+v", pg)
	}
}

func TestGet_SetsAuthHeader(t *testing.T) {
	var gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("x-yunxiao-token")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"id":1}]`))
	}))
	defer server.Close()

	c := NewClient()
	var out []map[string]int
	if err := c.Get(context.Background(), "tok123", server.URL+"/x", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if gotToken != "tok123" {
		t.Errorf("x-yunxiao-token = %q, want %q", gotToken, "tok123")
	}
	if len(out) != 1 || out[0]["id"] != 1 {
		t.Errorf("decoded = %v", out)
	}
}

func TestGet_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := NewClient()
	var out []map[string]int
	err := c.Get(context.Background(), "bad", server.URL, &out)
	if err == nil || err.Error() != "codeup: authentication failed (invalid token)" {
		t.Errorf("err = %v", err)
	}
}

func TestGet_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("boom"))
	}))
	defer server.Close()

	c := NewClient()
	var out []map[string]int
	err := c.Get(context.Background(), "tok", server.URL, &out)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "codeup API error 500: boom" {
		t.Errorf("err = %q", got)
	}
}

func TestGetWithPagination_ReadsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-total", "15")
		w.Header().Set("x-per-page", "100")
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := NewClient()
	var out []map[string]int
	pg, err := c.GetWithPagination(context.Background(), "tok", server.URL, &out)
	if err != nil {
		t.Fatalf("GetWithPagination: %v", err)
	}
	if pg.Total != 15 || pg.PerPage != 100 {
		t.Errorf("pagination = %+v", pg)
	}
}

func TestPost_SendsJSONBody(t *testing.T) {
	var gotMethod, gotToken, gotCT string
	var gotBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotToken = r.Header.Get("x-yunxiao-token")
		gotCT = r.Header.Get("Content-Type")
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":42}`))
	}))
	defer server.Close()

	c := NewClient()
	var out struct {
		ID int `json:"id"`
	}
	err := c.Post(context.Background(), "tok123", server.URL+"/mr",
		map[string]string{"title": "hello"}, &out)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
	if gotToken != "tok123" {
		t.Errorf("token = %q", gotToken)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q", gotCT)
	}
	if gotBody["title"] != "hello" {
		t.Errorf("body = %v", gotBody)
	}
	if out.ID != 42 {
		t.Errorf("decoded id = %d", out.ID)
	}
}

func TestPost_Accepts200And201(t *testing.T) {
	for _, code := range []int{http.StatusOK, http.StatusCreated} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
			w.Write([]byte(`{}`))
		}))
		c := NewClient()
		if err := c.Post(context.Background(), "t", server.URL, map[string]string{}, nil); err != nil {
			t.Errorf("status %d: %v", code, err)
		}
		server.Close()
	}
}

func TestPost_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid branch"))
	}))
	defer server.Close()

	c := NewClient()
	err := c.Post(context.Background(), "tok", server.URL, map[string]string{}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "codeup API error 400: invalid branch" {
		t.Errorf("err = %q", got)
	}
}

func TestPost_AuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c := NewClient()
	err := c.Post(context.Background(), "bad", server.URL, map[string]string{}, nil)
	if err == nil || err.Error() != "codeup: authentication failed (invalid token)" {
		t.Errorf("err = %v", err)
	}
}
