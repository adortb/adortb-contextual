// Package metrics 提供 Prometheus 兼容的指标收集。
package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Metrics 轻量级指标收集器（无外部依赖）。
type Metrics struct {
	analyzeTotal    atomic.Int64
	analyzeErrors   atomic.Int64
	fetchTotal      atomic.Int64
	fetchErrors     atomic.Int64
	latencyBuckets  [5]atomic.Int64 // <50ms, <100ms, <200ms, <500ms, >=500ms
	cacheHits       atomic.Int64
	cacheMisses     atomic.Int64
}

// New 创建指标收集器。
func New() *Metrics {
	return &Metrics{}
}

// IncAnalyze 分析请求计数 +1。
func (m *Metrics) IncAnalyze() {
	m.analyzeTotal.Add(1)
}

// IncAnalyzeError 分析错误计数 +1。
func (m *Metrics) IncAnalyzeError() {
	m.analyzeErrors.Add(1)
}

// IncFetch 抓取请求计数 +1。
func (m *Metrics) IncFetch() {
	m.fetchTotal.Add(1)
}

// IncFetchError 抓取错误计数 +1。
func (m *Metrics) IncFetchError() {
	m.fetchErrors.Add(1)
}

// IncCacheHit 缓存命中 +1。
func (m *Metrics) IncCacheHit() {
	m.cacheHits.Add(1)
}

// IncCacheMiss 缓存未命中 +1。
func (m *Metrics) IncCacheMiss() {
	m.cacheMisses.Add(1)
}

// ObserveLatency 记录延迟分桶。
func (m *Metrics) ObserveLatency(d time.Duration) {
	ms := d.Milliseconds()
	switch {
	case ms < 50:
		m.latencyBuckets[0].Add(1)
	case ms < 100:
		m.latencyBuckets[1].Add(1)
	case ms < 200:
		m.latencyBuckets[2].Add(1)
	case ms < 500:
		m.latencyBuckets[3].Add(1)
	default:
		m.latencyBuckets[4].Add(1)
	}
}

// Handler 返回 /metrics 端点的 HTTP 处理器。
func (m *Metrics) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP contextual_analyze_total Total analyze requests\n")
		fmt.Fprintf(w, "# TYPE contextual_analyze_total counter\n")
		fmt.Fprintf(w, "contextual_analyze_total %d\n", m.analyzeTotal.Load())

		fmt.Fprintf(w, "# HELP contextual_analyze_errors_total Total analyze errors\n")
		fmt.Fprintf(w, "# TYPE contextual_analyze_errors_total counter\n")
		fmt.Fprintf(w, "contextual_analyze_errors_total %d\n", m.analyzeErrors.Load())

		fmt.Fprintf(w, "# HELP contextual_fetch_total Total fetch requests\n")
		fmt.Fprintf(w, "# TYPE contextual_fetch_total counter\n")
		fmt.Fprintf(w, "contextual_fetch_total %d\n", m.fetchTotal.Load())

		fmt.Fprintf(w, "# HELP contextual_fetch_errors_total Total fetch errors\n")
		fmt.Fprintf(w, "# TYPE contextual_fetch_errors_total counter\n")
		fmt.Fprintf(w, "contextual_fetch_errors_total %d\n", m.fetchErrors.Load())

		fmt.Fprintf(w, "# HELP contextual_cache_hits_total Cache hits\n")
		fmt.Fprintf(w, "# TYPE contextual_cache_hits_total counter\n")
		fmt.Fprintf(w, "contextual_cache_hits_total %d\n", m.cacheHits.Load())

		fmt.Fprintf(w, "# HELP contextual_cache_misses_total Cache misses\n")
		fmt.Fprintf(w, "# TYPE contextual_cache_misses_total counter\n")
		fmt.Fprintf(w, "contextual_cache_misses_total %d\n", m.cacheMisses.Load())

		bucketLabels := []string{"<50ms", "<100ms", "<200ms", "<500ms", ">=500ms"}
		fmt.Fprintf(w, "# HELP contextual_latency_bucket Analyze latency buckets\n")
		fmt.Fprintf(w, "# TYPE contextual_latency_bucket gauge\n")
		for i, label := range bucketLabels {
			fmt.Fprintf(w, "contextual_latency_bucket{le=%q} %d\n", label, m.latencyBuckets[i].Load())
		}
	}
}
