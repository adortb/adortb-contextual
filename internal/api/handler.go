// Package api 提供 contextual 服务的 HTTP API 处理器。
package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/adortb/adortb-contextual/internal/brand_safety"
	"github.com/adortb/adortb-contextual/internal/classifier"
	"github.com/adortb/adortb-contextual/internal/fetch"
	"github.com/adortb/adortb-contextual/internal/metrics"
	"github.com/adortb/adortb-contextual/internal/nlp"
)

// AnalyzeRequest 分析请求体。
type AnalyzeRequest struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// AnalyzeResponse 分析响应体。
type AnalyzeResponse struct {
	Categories     []classifier.Category   `json:"categories"`
	Entities       []classifier.Entity     `json:"entities"`
	Sentiment      classifier.Sentiment    `json:"sentiment"`
	Keywords       []string                `json:"keywords"`
	Language       string                  `json:"language"`
	SafetyScore    float64                 `json:"safety_score"`
	FetchLatencyMs int64                   `json:"fetch_latency_ms,omitempty"`
}

// BatchRequest 批量分析请求。
type BatchRequest struct {
	Items []AnalyzeRequest `json:"items"`
}

// BatchResponse 批量分析响应。
type BatchResponse struct {
	Results []batchItem `json:"results"`
}

type batchItem struct {
	URL    string           `json:"url,omitempty"`
	Result *AnalyzeResponse `json:"result,omitempty"`
	Error  string           `json:"error,omitempty"`
}

// Handler HTTP API 处理器。
type Handler struct {
	catCls  *classifier.CategoryClassifier
	entRec  *classifier.EntityRecognizer
	sentAna *classifier.SentimentAnalyzer
	scanner *brand_safety.Scanner
	crawler *fetch.Crawler
	cache   *fetch.Cache
	tfidf   *nlp.TFIDF
	metrics *metrics.Metrics
}

// New 创建 API 处理器。
func New(
	catCls *classifier.CategoryClassifier,
	entRec *classifier.EntityRecognizer,
	sentAna *classifier.SentimentAnalyzer,
	scanner *brand_safety.Scanner,
	crawler *fetch.Crawler,
	cache *fetch.Cache,
	m *metrics.Metrics,
) *Handler {
	return &Handler{
		catCls:  catCls,
		entRec:  entRec,
		sentAna: sentAna,
		scanner: scanner,
		crawler: crawler,
		cache:   cache,
		tfidf:   nlp.NewTFIDF(),
		metrics: m,
	}
}

// RegisterRoutes 注册所有路由。
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/analyze", h.handleAnalyze)
	mux.HandleFunc("POST /v1/analyze/batch", h.handleAnalyzeBatch)
	mux.HandleFunc("GET /v1/categories", h.handleCategories)
	mux.HandleFunc("GET /health", h.handleHealth)
}

// handleAnalyze 处理单个 URL 的分析请求。
func (h *Handler) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	h.metrics.IncAnalyze()
	start := time.Now()

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		h.metrics.IncAnalyzeError()
		return
	}

	resp, err := h.analyze(r.Context(), req)
	if err != nil {
		slog.Error("analyze failed", "error", err, "url", req.URL)
		h.writeError(w, http.StatusInternalServerError, "analysis failed")
		h.metrics.IncAnalyzeError()
		return
	}

	h.metrics.ObserveLatency(time.Since(start))
	h.writeJSON(w, http.StatusOK, resp)
}

// handleAnalyzeBatch 批量分析（最多 10 个，并发执行）。
func (h *Handler) handleAnalyzeBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	const maxBatch = 10
	if len(req.Items) > maxBatch {
		req.Items = req.Items[:maxBatch]
	}

	results := make([]batchItem, len(req.Items))
	var wg sync.WaitGroup
	for i, item := range req.Items {
		wg.Add(1)
		go func(idx int, it AnalyzeRequest) {
			defer wg.Done()
			resp, err := h.analyze(r.Context(), it)
			if err != nil {
				results[idx] = batchItem{URL: it.URL, Error: err.Error()}
			} else {
				results[idx] = batchItem{URL: it.URL, Result: resp}
			}
		}(i, item)
	}
	wg.Wait()

	h.writeJSON(w, http.StatusOK, BatchResponse{Results: results})
}

// handleCategories 返回 IAB 分类列表。
func (h *Handler) handleCategories(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"categories": classifier.AllCategories(),
	})
}

// handleHealth 健康检查端点。
func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// analyze 核心分析逻辑。
func (h *Handler) analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	text, lang, fetchLatencyMs := req.Content, "en", int64(0)

	// 若无内容，尝试抓取
	if strings.TrimSpace(text) == "" && req.URL != "" {
		fetchStart := time.Now()

		// 先查缓存
		if cached, err := h.cache.Get(ctx, req.URL); err == nil && cached != nil {
			h.metrics.IncCacheHit()
			text = cached.Title + " " + cached.Description + " " + cached.Body
			lang = cached.Language
			if req.Title != "" {
				text = req.Title + " " + text
			}
		} else {
			h.metrics.IncCacheMiss()
			h.metrics.IncFetch()
			rawHTML, err := h.crawler.Fetch(ctx, req.URL)
			fetchLatencyMs = time.Since(fetchStart).Milliseconds()
			if err != nil {
				h.metrics.IncFetchError()
				// 降级：仅用标题和请求中的内容
				if req.Title != "" {
					text = req.Title
				}
			} else {
				extracted := fetch.ExtractHTML(rawHTML)
				lang = extracted.Language
				text = extracted.Title + " " + extracted.Description + " " + extracted.Body
				if req.Title != "" {
					text = req.Title + " " + text
				}
				// 写入缓存（异步，不阻塞响应）
				page := fetch.NewPageContent(req.URL, extracted.Title, extracted.Description, extracted.Body, lang)
				go func() {
					bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
					defer cancel()
					_ = h.cache.Set(bgCtx, page)
				}()
			}
		}
	} else {
		if req.Title != "" {
			text = req.Title + " " + text
		}
	}

	// 提取关键词
	tokens := nlp.Tokenize(text)
	words := nlp.FilterTokens(tokens)
	kwList := h.tfidf.TopKeywords(words, 10)
	keywords := make([]string, 0, len(kwList))
	for _, kw := range kwList {
		keywords = append(keywords, kw.Word)
	}

	// 分类
	categories := h.catCls.Classify(text, 0.3)

	// 实体识别
	entities := h.entRec.Recognize(text)

	// 情感分析
	sentiment := h.sentAna.Analyze(text)

	// 品牌安全
	safetyResult := h.scanner.Scan(text)

	return &AnalyzeResponse{
		Categories:     categories,
		Entities:       entities,
		Sentiment:      sentiment,
		Keywords:       keywords,
		Language:       lang,
		SafetyScore:    safetyResult.Score,
		FetchLatencyMs: fetchLatencyMs,
	}, nil
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}
