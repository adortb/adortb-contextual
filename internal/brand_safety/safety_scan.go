// Package brand_safety 提供广告品牌安全扫描能力，检测敏感内容并返回安全评分。
package brand_safety

import (
	"strings"
)

// SafetyResult 安全扫描结果。
type SafetyResult struct {
	Score      float64  `json:"score"`       // 0.0(不安全) ~ 1.0(完全安全)
	Flags      []string `json:"flags"`       // 触发的安全标志
	BlockLevel string   `json:"block_level"` // "none" | "warn" | "block"
}

// SensitiveCategory 敏感类别。
type SensitiveCategory struct {
	name     string
	weight   float64 // 扣分权重（0~1）
	keywords []string
}

// Scanner 品牌安全扫描器。
type Scanner struct {
	categories []SensitiveCategory
}

// NewScanner 创建内置规则的安全扫描器。
func NewScanner() *Scanner {
	s := &Scanner{}
	s.loadCategories()
	return s
}

func (s *Scanner) loadCategories() {
	s.categories = []SensitiveCategory{
		{
			name:   "adult_content",
			weight: 0.6,
			keywords: []string{
				"porn", "pornography", "xxx", "nude", "nudity", "explicit", "adult content",
				"色情", "成人", "裸体",
			},
		},
		{
			name:   "violence",
			weight: 0.5,
			keywords: []string{
				"violence", "violent", "murder", "kill", "killing", "blood", "gore",
				"torture", "massacre", "terrorist", "terrorism", "bomb", "weapon",
				"暴力", "杀人", "恐怖", "爆炸", "枪击",
			},
		},
		{
			name:   "hate_speech",
			weight: 0.7,
			keywords: []string{
				"hate speech", "racism", "racist", "sexist", "discrimination", "bigotry",
				"仇恨", "种族歧视", "歧视",
			},
		},
		{
			name:   "gambling",
			weight: 0.3,
			keywords: []string{
				"gambling", "casino", "poker", "slot machine", "betting", "wager",
				"赌博", "赌场", "彩票", "博彩",
			},
		},
		{
			name:   "drugs",
			weight: 0.5,
			keywords: []string{
				"drug", "cocaine", "heroin", "marijuana", "cannabis", "narcotics",
				"毒品", "可卡因", "大麻", "吸毒",
			},
		},
		{
			name:   "fake_news",
			weight: 0.2,
			keywords: []string{
				"fake news", "misinformation", "disinformation", "conspiracy",
				"假新闻", "谣言", "虚假信息",
			},
		},
	}
}

// Scan 扫描文本，返回安全评分结果。
func (s *Scanner) Scan(text string) SafetyResult {
	if text == "" {
		return SafetyResult{Score: 1.0, BlockLevel: "none"}
	}

	lower := strings.ToLower(text)
	var flags []string
	totalPenalty := 0.0

	for _, cat := range s.categories {
		hit := false
		for _, kw := range cat.keywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				hit = true
				break
			}
		}
		if hit {
			flags = append(flags, cat.name)
			totalPenalty += cat.weight
		}
	}

	score := 1.0 - totalPenalty
	if score < 0 {
		score = 0
	}
	score = roundFloat(score, 2)

	blockLevel := "none"
	switch {
	case score < 0.4:
		blockLevel = "block"
	case score < 0.7:
		blockLevel = "warn"
	}

	return SafetyResult{
		Score:      score,
		Flags:      flags,
		BlockLevel: blockLevel,
	}
}

func roundFloat(f float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(f*pow+0.5)) / pow
}
