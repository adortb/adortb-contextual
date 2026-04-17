// Package main 是 adortb-contextual 服务入口。
// 提供网页内容语义分析能力：IAB 分类、命名实体识别、情感分析、品牌安全检测。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adortb/adortb-contextual/internal/api"
	"github.com/adortb/adortb-contextual/internal/brand_safety"
	"github.com/adortb/adortb-contextual/internal/classifier"
	"github.com/adortb/adortb-contextual/internal/fetch"
	"github.com/adortb/adortb-contextual/internal/metrics"
	"github.com/go-redis/redis/v8"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	port := getEnv("PORT", "8097")
	redisAddr := getEnv("REDIS_ADDR", "")

	// 初始化各组件
	catCls := classifier.NewCategoryClassifier()
	entRec := classifier.NewEntityRecognizer()
	sentAna := classifier.NewSentimentAnalyzer()
	scanner := brand_safety.NewScanner()
	crawler := fetch.NewCrawler(10 * time.Second)
	m := metrics.New()

	// Redis 缓存（可选）
	var cache *fetch.Cache
	if redisAddr != "" {
		rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
		cache = fetch.NewCache(rdb)
		slog.Info("redis cache enabled", "addr", redisAddr)
	} else {
		cache = fetch.NewCache(nil)
		slog.Info("redis cache disabled, running without cache")
	}

	h := api.New(catCls, entRec, sentAna, scanner, crawler, cache, m)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.HandleFunc("GET /metrics", m.Handler())

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("adortb-contextual starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
