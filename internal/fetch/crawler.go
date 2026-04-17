package fetch

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// userAgents 用于轮转的 User-Agent 池。
var userAgents = []string{
	"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (compatible; adortb-contextual/1.0; +https://adortb.com/bot)",
}

// uaCounter 原子计数器，用于 User-Agent 轮转。
var uaCounter atomic.Uint64

// Crawler HTTP 抓取器。
type Crawler struct {
	client       *http.Client
	maxBodyBytes int64
}

// NewCrawler 创建抓取器，timeout 控制请求超时。
func NewCrawler(timeout time.Duration) *Crawler {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Crawler{
		client: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		maxBodyBytes: 2 * 1024 * 1024, // 2MB
	}
}

// Fetch 抓取指定 URL 的 HTML 内容。
// 会先检查 robots.txt，若不允许则返回错误。
func (c *Crawler) Fetch(ctx context.Context, rawURL string) (string, error) {
	if err := c.checkRobots(ctx, rawURL); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	idx := uaCounter.Add(1) % uint64(len(userAgents))
	req.Header.Set("User-Agent", userAgents[idx])
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status %d for %s", resp.StatusCode, rawURL)
	}

	limited := io.LimitReader(resp.Body, c.maxBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	return string(body), nil
}

// checkRobots 简单的 robots.txt 检查：仅拒绝明确 Disallow: / 的情况。
func (c *Crawler) checkRobots(ctx context.Context, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, robotsURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "adortb-contextual/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil // 无法访问 robots.txt，继续
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	limited := io.LimitReader(resp.Body, 64*1024) // 最多读 64KB
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil
	}

	if isDisallowedByRobots(string(data), u.Path) {
		return fmt.Errorf("disallowed by robots.txt: %s", rawURL)
	}
	return nil
}

// isDisallowedByRobots 检查 robots.txt 是否禁止当前 path。
func isDisallowedByRobots(content, path string) bool {
	lines := strings.Split(content, "\n")
	inRelevantAgent := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "user-agent:") {
			agent := strings.TrimSpace(line[len("user-agent:"):])
			inRelevantAgent = agent == "*" ||
				strings.Contains(strings.ToLower(agent), "adortb")
		}
		if inRelevantAgent && strings.HasPrefix(lower, "disallow:") {
			disallowed := strings.TrimSpace(line[len("disallow:"):])
			if disallowed == "/" || (disallowed != "" && strings.HasPrefix(path, disallowed)) {
				return true
			}
		}
	}
	return false
}

// randomSleep 用于测试时的随机延迟（避免过于频繁请求）。
func randomSleep(min, max time.Duration) {
	d := min + time.Duration(rand.Int63n(int64(max-min)))
	time.Sleep(d)
}
