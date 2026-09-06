// Package codeupapi 封装阿里云 Codeup / 云效 Flow OAPI v1 的公共基础设施：
// API 基础地址映射、x-yunxiao-token 认证、响应头分页解析、带认证的 GET/POST 请求。
// 供 provider、mergerequest、cicd 三个包共用。
package codeupapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client 是 Codeup OAPI v1 的 HTTP 客户端。
type Client struct {
	httpClient *http.Client
}

// NewClient 创建一个新的 Client。
func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) getHTTPClient() *http.Client {
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return c.httpClient
}

// APIBaseURL 将用户可见的 server URL 映射为指定产品的 OAPI v1 API 基础地址。
// product 为 OAPI v1 路径中的产品段，如 "codeup"、"flow"。
// 对于 codeup.aliyun.com → https://openapi-rdc.aliyuncs.com/oapi/v1/{product}/organizations/{orgId}
// 对于自定义/测试 URL，保留原有 scheme 和 host。
// 原始 serverURL 仅用于 clone URL 构造（见 CloneHost）。
func APIBaseURL(serverURL, product, orgID string) string {
	h := strings.TrimRight(serverURL, "/")
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")

	scheme := "https"
	if strings.HasPrefix(serverURL, "http://") {
		scheme = "http"
	}

	apiHost := h
	if h == "codeup.aliyun.com" {
		apiHost = "openapi-rdc.aliyuncs.com"
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/oapi/v1/%s/organizations/%s", scheme, apiHost, product, url.PathEscape(orgID))
}

// CloneHost 从用户可见的 server URL 中提取 clone 域名。
func CloneHost(serverURL string) string {
	h := strings.TrimRight(serverURL, "/")
	h = strings.TrimPrefix(h, "https://")
	h = strings.TrimPrefix(h, "http://")
	return h
}

// Pagination 保存从响应头提取的分页信息。
type Pagination struct {
	Total      int
	TotalPages int
	Page       int
	PerPage    int
	NextPage   int // 0 表示无下一页
}

// ParsePagination 从响应头提取分页信息。
func ParsePagination(resp *http.Response) Pagination {
	var pg Pagination
	pg.Total, _ = strconv.Atoi(resp.Header.Get("x-total"))
	pg.TotalPages, _ = strconv.Atoi(resp.Header.Get("x-total-pages"))
	pg.Page, _ = strconv.Atoi(resp.Header.Get("x-page"))
	pg.PerPage, _ = strconv.Atoi(resp.Header.Get("x-per-page"))
	pg.NextPage, _ = strconv.Atoi(resp.Header.Get("x-next-page"))
	return pg
}

// Get 发起带 x-yunxiao-token 认证的 GET 请求，将 JSON 响应体直接解码到 v（无 wrapper）。
func (c *Client) Get(ctx context.Context, token, apiURL string, v interface{}) error {
	resp, err := c.do(ctx, http.MethodGet, apiURL, token, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return err
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// GetWithPagination 发起 GET 请求，返回解析结果和分页信息。
func (c *Client) GetWithPagination(ctx context.Context, token, apiURL string, v interface{}) (Pagination, error) {
	resp, err := c.do(ctx, http.MethodGet, apiURL, token, nil)
	if err != nil {
		return Pagination{}, err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp); err != nil {
		return Pagination{}, err
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return Pagination{}, fmt.Errorf("decode response: %w", err)
	}

	return ParsePagination(resp), nil
}

// Post 发起带 x-yunxiao-token 认证的 POST 请求（JSON body），将 JSON 响应体解码到 v（v 可为 nil）。
func (c *Client) Post(ctx context.Context, token, apiURL string, reqBody, v interface{}) error {
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.do(ctx, http.MethodPost, apiURL, token, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkStatus(resp, http.StatusOK, http.StatusCreated); err != nil {
		return err
	}

	if v == nil {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// do 构造并执行 HTTP 请求。
func (c *Client) do(ctx context.Context, method, apiURL, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if token != "" {
		req.Header.Set("x-yunxiao-token", token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.getHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// checkStatus 检查响应状态码。okCodes 为空时默认仅接受 200。
// 401 返回认证失败错误；其他非预期状态码返回包含状态码和响应体的错误。
func checkStatus(resp *http.Response, okCodes ...int) error {
	if len(okCodes) == 0 {
		okCodes = []int{http.StatusOK}
	}

	for _, code := range okCodes {
		if resp.StatusCode == code {
			return nil
		}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("codeup: authentication failed (invalid token)")
	}

	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("codeup API error %d: %s", resp.StatusCode, string(body))
}
