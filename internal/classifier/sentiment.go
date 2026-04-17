package classifier

import (
	"strings"

	"github.com/adortb/adortb-contextual/internal/nlp"
)

// Sentiment 情感分析结果。
type Sentiment struct {
	Polarity string  `json:"polarity"` // "positive" | "negative" | "neutral"
	Score    float64 `json:"score"`    // 0.0 ~ 1.0
}

// SentimentAnalyzer 基于情感词典的情感分析器。
type SentimentAnalyzer struct {
	positiveWords map[string]float64
	negativeWords map[string]float64
	negationWords map[string]struct{}
}

// NewSentimentAnalyzer 创建情感分析器。
func NewSentimentAnalyzer() *SentimentAnalyzer {
	sa := &SentimentAnalyzer{
		positiveWords: make(map[string]float64),
		negativeWords: make(map[string]float64),
		negationWords: make(map[string]struct{}),
	}
	sa.loadDicts()
	return sa
}

func (sa *SentimentAnalyzer) loadDicts() {
	// 英文正面词
	positiveEN := map[string]float64{
		"good": 1.0, "great": 1.5, "excellent": 2.0, "amazing": 2.0, "wonderful": 1.8,
		"fantastic": 1.8, "outstanding": 2.0, "superb": 1.8, "brilliant": 1.8,
		"perfect": 2.0, "love": 1.5, "best": 1.5, "awesome": 1.8, "beautiful": 1.5,
		"happy": 1.2, "joy": 1.2, "success": 1.5, "successful": 1.5, "win": 1.2,
		"winner": 1.2, "profit": 1.2, "growth": 1.2, "improve": 1.0, "innovative": 1.5,
		"positive": 1.0, "benefit": 1.0, "efficient": 1.0, "effective": 1.0,
		"reliable": 1.0, "powerful": 1.2, "strong": 1.0, "advanced": 1.2,
		"revolutionary": 1.8, "breakthrough": 1.8, "exciting": 1.5, "impressive": 1.5,
		"leading": 1.0, "top": 1.0, "popular": 1.0, "recommend": 1.0,
		"safe": 1.0, "secure": 1.0, "trusted": 1.0, "quality": 1.0,
		"value": 1.0, "affordable": 1.0, "helpful": 1.0, "useful": 1.0,
		"smart": 1.2, "intelligent": 1.2, "creative": 1.2, "unique": 1.0,
		"reward": 1.0, "opportunity": 1.0, "solution": 1.0,
	}
	// 英文负面词
	negativeEN := map[string]float64{
		"bad": 1.0, "terrible": 2.0, "horrible": 2.0, "awful": 1.8, "dreadful": 1.8,
		"worst": 2.0, "poor": 1.2, "failure": 1.5, "fail": 1.2, "failed": 1.2,
		"problem": 1.0, "issue": 0.8, "bug": 1.0, "error": 1.0, "crash": 1.5,
		"loss": 1.2, "lose": 1.0, "losing": 1.0, "danger": 1.5, "dangerous": 1.5,
		"risk": 1.0, "risky": 1.2, "harm": 1.5, "harmful": 1.5, "toxic": 2.0,
		"violence": 2.0, "violent": 2.0, "abuse": 2.0, "scam": 2.0, "fraud": 2.0,
		"hate": 1.8, "anger": 1.5, "angry": 1.5, "fear": 1.2, "scared": 1.2,
		"sad": 1.0, "unhappy": 1.2, "disappointment": 1.5, "disappointed": 1.5,
		"corrupt": 1.8, "corruption": 1.8, "illegal": 1.8, "crime": 1.5,
		"attack": 1.5, "threat": 1.5, "disaster": 1.8, "crisis": 1.5,
		"controversy": 1.2, "controversial": 1.2, "scandal": 1.8,
		"expensive": 0.8, "overpriced": 1.0, "slow": 0.8, "broken": 1.2,
	}
	// 中文正面词
	positiveZH := map[string]float64{
		"好": 1.0, "很好": 1.5, "非常好": 1.8, "优秀": 1.5, "出色": 1.8,
		"棒": 1.2, "完美": 2.0, "喜欢": 1.2, "爱": 1.5, "最佳": 1.5,
		"成功": 1.5, "胜利": 1.5, "利润": 1.2, "增长": 1.2, "创新": 1.5,
		"突破": 1.8, "领先": 1.2, "高效": 1.0, "安全": 1.0, "可靠": 1.0,
		"强大": 1.2, "智能": 1.2, "优质": 1.2, "推荐": 1.0, "满意": 1.2,
		"快乐": 1.2, "幸福": 1.5, "积极": 1.0, "正面": 1.0, "赞": 1.2,
	}
	// 中文负面词
	negativeZH := map[string]float64{
		"差": 1.0, "糟": 1.5, "糟糕": 1.8, "失败": 1.5, "问题": 1.0,
		"危险": 1.5, "有害": 1.5, "违法": 1.8, "欺诈": 2.0, "骗": 1.8,
		"损失": 1.2, "亏损": 1.5, "暴力": 2.0, "腐败": 1.8, "犯罪": 1.5,
		"丑闻": 1.8, "危机": 1.5, "灾难": 1.8, "争议": 1.2, "愤怒": 1.5,
		"担忧": 1.0, "恐惧": 1.2, "悲伤": 1.0, "失望": 1.5, "不满": 1.2,
	}

	for k, v := range positiveEN {
		sa.positiveWords[k] = v
	}
	for k, v := range negativeEN {
		sa.negativeWords[k] = v
	}
	for k, v := range positiveZH {
		sa.positiveWords[k] = v
	}
	for k, v := range negativeZH {
		sa.negativeWords[k] = v
	}

	// 否定词
	negations := []string{"not", "no", "never", "neither", "nor", "don't", "doesn't",
		"didn't", "won't", "can't", "isn't", "aren't", "wasn't", "without",
		"没有", "不", "非", "未", "无"}
	for _, n := range negations {
		sa.negationWords[n] = struct{}{}
	}
}

// Analyze 对文本进行情感分析。
func (sa *SentimentAnalyzer) Analyze(text string) Sentiment {
	if text == "" {
		return Sentiment{Polarity: "neutral", Score: 0.5}
	}

	tokens := nlp.Tokenize(text)
	words := make([]string, 0, len(tokens))
	for _, t := range tokens {
		words = append(words, strings.ToLower(t.Text))
	}

	var posScore, negScore float64
	for i, word := range words {
		// 检查前置否定词
		negated := false
		for j := max(0, i-3); j < i; j++ {
			if _, ok := sa.negationWords[words[j]]; ok {
				negated = true
				break
			}
		}

		if score, ok := sa.positiveWords[word]; ok {
			if negated {
				negScore += score * 0.5
			} else {
				posScore += score
			}
		}
		if score, ok := sa.negativeWords[word]; ok {
			if negated {
				posScore += score * 0.3
			} else {
				negScore += score
			}
		}
	}

	total := posScore + negScore
	if total < 0.01 {
		return Sentiment{Polarity: "neutral", Score: 0.5}
	}

	diff := posScore - negScore
	normalizedScore := (diff/total + 1) / 2 // 映射到 [0, 1]
	normalizedScore = roundFloat(normalizedScore, 2)

	switch {
	case normalizedScore > 0.6:
		return Sentiment{Polarity: "positive", Score: normalizedScore}
	case normalizedScore < 0.4:
		return Sentiment{Polarity: "negative", Score: roundFloat(1-normalizedScore, 2)}
	default:
		return Sentiment{Polarity: "neutral", Score: 0.5}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
