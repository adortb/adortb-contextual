package nlp

import (
	"math"
	"sort"
)

// Keyword 表示一个关键词及其 TF-IDF 权重。
type Keyword struct {
	Word   string
	Score  float64
}

// TFIDF 基于单文档计算 TF-IDF，使用内置 IDF 估算。
// 无文档库情况下，IDF 使用词长度和词频的启发式估算。
type TFIDF struct {
	// corpusSize 用于 IDF 计算的虚拟语料库大小（常量）
	corpusSize float64
}

// NewTFIDF 创建 TF-IDF 计算器。
func NewTFIDF() *TFIDF {
	return &TFIDF{corpusSize: 1000000}
}

// TopKeywords 从词列表中提取 topN 个关键词。
func (t *TFIDF) TopKeywords(words []string, topN int) []Keyword {
	if len(words) == 0 {
		return nil
	}

	// 计算词频
	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}

	total := float64(len(words))
	keywords := make([]Keyword, 0, len(freq))

	for word, count := range freq {
		tf := float64(count) / total
		idf := t.estimateIDF(word)
		score := tf * idf
		if score > 0 {
			keywords = append(keywords, Keyword{Word: word, Score: score})
		}
	}

	// 次排序键用词本身，保证结果确定性
	sort.Slice(keywords, func(i, j int) bool {
		if keywords[i].Score != keywords[j].Score {
			return keywords[i].Score > keywords[j].Score
		}
		return keywords[i].Word < keywords[j].Word
	})

	if topN > 0 && topN < len(keywords) {
		return keywords[:topN]
	}
	return keywords
}

// estimateIDF 启发式估算 IDF：词越长、越罕见，IDF 越高。
func (t *TFIDF) estimateIDF(word string) float64 {
	runes := []rune(word)
	length := len(runes)
	// 基础 IDF：ln(N/df)，df 根据词长启发式估算
	var df float64
	switch {
	case length <= 2:
		df = t.corpusSize * 0.1 // 短词常见
	case length <= 4:
		df = t.corpusSize * 0.01
	case length <= 6:
		df = t.corpusSize * 0.001
	default:
		df = t.corpusSize * 0.0001 // 长词罕见
	}
	if df < 1 {
		df = 1
	}
	return math.Log(t.corpusSize/df) + 1
}

// WordFrequency 返回词频 map（供分类器复用）。
func WordFrequency(words []string) map[string]int {
	freq := make(map[string]int, len(words))
	for _, w := range words {
		freq[w]++
	}
	return freq
}
