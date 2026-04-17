package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adortb/adortb-contextual/internal/brand_safety"
	"github.com/adortb/adortb-contextual/internal/classifier"
	"github.com/adortb/adortb-contextual/internal/fetch"
	"github.com/adortb/adortb-contextual/internal/metrics"
)

func newTestHandler() *Handler {
	return New(
		classifier.NewCategoryClassifier(),
		classifier.NewEntityRecognizer(),
		classifier.NewSentimentAnalyzer(),
		brand_safety.NewScanner(),
		fetch.NewCrawler(0),
		fetch.NewCache(nil),
		metrics.New(),
	)
}

func TestHandleAnalyze_WithContent(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	reqBody := AnalyzeRequest{
		Title:   "Tesla announces new electric car model",
		Content: "Tesla has released its latest electric vehicle with improved battery range and autonomous driving features.",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AnalyzeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Categories) == 0 {
		t.Error("expected at least one category")
	}
	if resp.SafetyScore < 0 || resp.SafetyScore > 1 {
		t.Errorf("safety score out of range: %f", resp.SafetyScore)
	}
	if len(resp.Keywords) == 0 {
		t.Error("expected at least one keyword")
	}
}

func TestHandleAnalyze_EmptyRequest(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	reqBody := AnalyzeRequest{}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", bytes.NewReader(body))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// 空请求应正常返回
	if w.Code != http.StatusOK {
		t.Logf("empty request response: %s", w.Body.String())
	}
}

func TestHandleAnalyze_InvalidJSON(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/v1/analyze", bytes.NewReader([]byte("not-json")))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleAnalyzeBatch(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	batch := BatchRequest{
		Items: []AnalyzeRequest{
			{Title: "Sports game basketball", Content: "The NBA game was exciting."},
			{Title: "Tech news AI", Content: "Artificial intelligence transforms the industry."},
		},
	}
	body, _ := json.Marshal(batch)

	req := httptest.NewRequest(http.MethodPost, "/v1/analyze/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp BatchResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(resp.Results))
	}
}

func TestHandleCategories(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/v1/categories", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal("decode error")
	}
	cats, ok := resp["categories"].([]interface{})
	if !ok || len(cats) == 0 {
		t.Error("expected non-empty categories list")
	}
}

func TestHandleHealth(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
