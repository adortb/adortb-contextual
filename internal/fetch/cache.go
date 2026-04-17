// Package fetch 提供网页内容抓取、解析和缓存能力。
package fetch

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	cacheTTL    = 24 * time.Hour
	cachePrefix = "ctx:page:"
)

// PageContent 缓存的页面内容。
type PageContent struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
	Language    string `json:"language"`
	FetchedAt   int64  `json:"fetched_at"`
}

// Cache 页面内容 Redis 缓存。
type Cache struct {
	rdb *redis.Client
}

// NewCache 创建缓存实例。如果 rdb 为 nil，缓存操作为 no-op。
func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

// Get 从缓存获取页面内容。
func (c *Cache) Get(ctx context.Context, url string) (*PageContent, error) {
	if c.rdb == nil {
		return nil, nil
	}
	key := c.key(url)
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}
	var content PageContent
	if err := json.Unmarshal(val, &content); err != nil {
		return nil, fmt.Errorf("cache unmarshal: %w", err)
	}
	return &content, nil
}

// Set 将页面内容写入缓存（TTL 24h）。
func (c *Cache) Set(ctx context.Context, content *PageContent) error {
	if c.rdb == nil {
		return nil
	}
	data, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("cache marshal: %w", err)
	}
	key := c.key(content.URL)
	return c.rdb.Set(ctx, key, data, cacheTTL).Err()
}

func (c *Cache) key(url string) string {
	h := sha256.Sum256([]byte(url))
	return cachePrefix + hex.EncodeToString(h[:16])
}

// PageContentWithTime 创建带时间戳的 PageContent。
func NewPageContent(url, title, description, body, lang string) *PageContent {
	return &PageContent{
		URL:         url,
		Title:       title,
		Description: description,
		Body:        body,
		Language:    lang,
		FetchedAt:   time.Now().Unix(),
	}
}
