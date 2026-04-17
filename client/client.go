// Package client 提供 adortb-contextual 服务的 Go 客户端，
// 供 adortb-adx 在 BidRequest 时查询页面上下文信息。
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// PageContext 页面上下文信息，填充到 BidRequest.site.ext.context。
type PageContext struct {
	Categories []string `json:"categories"`
	Entities   []string `json:"entities"`
	Sentiment  string   `json:"sentiment"`
	Keywords   []string `json:"keywords,omitempty"`
	SafeScore  float64  `json:"safety_score,omitempty"`
}

// AnalyzeRequest 分析请求。
type AnalyzeRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

// analyzeResponse 内部响应结构，与服务端 AnalyzeResponse 对应。
type analyzeResponse struct {
	Categories []struct {
		ID         string  `json:"id"`
		Name       string  `json:"name"`
		Confidence float64 `json:"confidence"`
	} `json:"categories"`
	Entities []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"entities"`
	Sentiment struct {
		Polarity string  `json:"polarity"`
		Score    float64 `json:"score"`
	} `json:"sentiment"`
	Keywords    []string `json:"keywords"`
	SafetyScore float64  `json:"safety_score"`
}

// Client contextual 服务客户端。
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// Config 客户端配置。
type Config struct {
	BaseURL string
	Timeout time.Duration
}

// New 创建客户端实例。
func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 500 * time.Millisecond
	}
	return &Client{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Analyze 查询指定 URL 的页面上下文，返回供 DSP 使用的 PageContext。
func (c *Client) Analyze(ctx context.Context, req AnalyzeRequest) (*PageContext, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/analyze", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var ar analyzeResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return toPageContext(&ar), nil
}

// AnalyzeBatch 批量查询多个 URL 的页面上下文。
// 使用服务端 /v1/analyze/batch 端点，单次 RPC。
func (c *Client) AnalyzeBatch(ctx context.Context, reqs []AnalyzeRequest) ([]*PageContext, error) {
	payload := struct {
		Items []AnalyzeRequest `json:"items"`
	}{Items: reqs}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal batch request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/analyze/batch", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var br struct {
		Results []struct {
			URL    string           `json:"url"`
			Result *analyzeResponse `json:"result"`
			Error  string           `json:"error"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	contexts := make([]*PageContext, len(br.Results))
	for i, r := range br.Results {
		if r.Result != nil {
			contexts[i] = toPageContext(r.Result)
		}
	}
	return contexts, nil
}

// toPageContext 将服务端响应转换为 PageContext。
func toPageContext(ar *analyzeResponse) *PageContext {
	cats := make([]string, 0, len(ar.Categories))
	for _, c := range ar.Categories {
		cats = append(cats, c.ID)
	}

	entities := make([]string, 0, len(ar.Entities))
	for _, e := range ar.Entities {
		entities = append(entities, e.Text)
	}

	return &PageContext{
		Categories: cats,
		Entities:   entities,
		Sentiment:  ar.Sentiment.Polarity,
		Keywords:   ar.Keywords,
		SafeScore:  ar.SafetyScore,
	}
}
