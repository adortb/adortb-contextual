// Package classifier 实现 IAB Content Taxonomy 3.0 分类、命名实体识别和情感分析。
package classifier

import (
	"sort"
	"strings"

	"github.com/adortb/adortb-contextual/internal/nlp"
)

// Category 表示一个 IAB 内容类别。
type Category struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// iabEntry 内部分类词典条目。
type iabEntry struct {
	id       string
	name     string
	keywords []string
}

// iabTaxonomy IAB Content Taxonomy 3.0 核心类别词典。
var iabTaxonomy = []iabEntry{
	{
		id:   "IAB1",
		name: "Arts & Entertainment",
		keywords: []string{
			"movie", "film", "music", "song", "album", "concert", "theater", "art",
			"painting", "sculpture", "gallery", "celebrity", "actor", "actress",
			"entertainment", "show", "performance", "dance", "comedy", "drama",
			"电影", "音乐", "艺术", "表演", "歌曲", "演员", "娱乐", "综艺", "明星",
		},
	},
	{
		id:   "IAB2",
		name: "Automotive",
		keywords: []string{
			"car", "vehicle", "automobile", "truck", "suv", "sedan", "engine", "motor",
			"driving", "highway", "fuel", "electric vehicle", "ev", "hybrid", "tesla",
			"bmw", "mercedes", "toyota", "honda", "ford", "volkswagen",
			"汽车", "驾驶", "车辆", "发动机", "新能源", "电动车",
		},
	},
	{
		id:   "IAB3",
		name: "Business",
		keywords: []string{
			"business", "company", "corporation", "enterprise", "startup", "revenue",
			"profit", "market", "stock", "investment", "finance", "economy", "trade",
			"management", "executive", "ceo", "strategy", "merger", "acquisition",
			"商业", "企业", "公司", "投资", "市场", "经济", "管理", "战略",
		},
	},
	{
		id:   "IAB5",
		name: "Education",
		keywords: []string{
			"education", "school", "university", "college", "student", "teacher",
			"learning", "course", "curriculum", "classroom", "degree", "academic",
			"research", "study", "scholarship", "training", "skill",
			"教育", "学校", "大学", "学生", "学习", "课程", "培训", "研究",
		},
	},
	{
		id:   "IAB7",
		name: "Health & Fitness",
		keywords: []string{
			"health", "fitness", "exercise", "diet", "nutrition", "medical", "doctor",
			"hospital", "disease", "treatment", "therapy", "wellness", "yoga", "gym",
			"mental health", "healthcare", "medicine", "vaccine", "drug",
			"健康", "医疗", "疾病", "治疗", "运动", "健身", "营养", "医院",
		},
	},
	{
		id:   "IAB8",
		name: "Food & Drink",
		keywords: []string{
			"food", "recipe", "cooking", "restaurant", "cuisine", "meal", "drink",
			"wine", "coffee", "tea", "beer", "dessert", "vegetarian", "vegan",
			"ingredient", "chef", "kitchen", "dining",
			"美食", "食物", "餐厅", "烹饪", "食谱", "饮料", "厨师",
		},
	},
	{
		id:   "IAB11",
		name: "Law, Government & Politics",
		keywords: []string{
			"law", "government", "politics", "policy", "election", "congress", "senate",
			"president", "legislation", "regulation", "court", "legal", "justice",
			"democracy", "republican", "democrat", "vote", "political", "parliament",
			"法律", "政府", "政治", "选举", "政策", "法规", "立法", "司法",
		},
	},
	{
		id:   "IAB12",
		name: "News & Weather",
		keywords: []string{
			"news", "breaking", "report", "journalist", "media", "press", "weather",
			"forecast", "temperature", "storm", "earthquake", "disaster", "headline",
			"新闻", "报道", "媒体", "天气", "预报", "地震", "灾害",
		},
	},
	{
		id:   "IAB13",
		name: "Personal Finance",
		keywords: []string{
			"finance", "money", "investment", "savings", "bank", "loan", "mortgage",
			"insurance", "tax", "budget", "credit", "debt", "retirement", "fund",
			"cryptocurrency", "bitcoin", "blockchain",
			"理财", "金融", "储蓄", "贷款", "保险", "税务", "加密货币", "比特币",
		},
	},
	{
		id:   "IAB15",
		name: "Science",
		keywords: []string{
			"science", "research", "discovery", "experiment", "physics", "chemistry",
			"biology", "astronomy", "space", "nasa", "climate", "environment",
			"evolution", "genetics", "artificial intelligence", "quantum",
			"科学", "研究", "发现", "实验", "物理", "化学", "生物", "太空", "环境",
		},
	},
	{
		id:   "IAB17",
		name: "Sports",
		keywords: []string{
			"sport", "football", "soccer", "basketball", "baseball", "tennis",
			"golf", "swimming", "athletics", "olympic", "championship", "league",
			"player", "team", "coach", "tournament", "match", "score", "goal",
			"体育", "足球", "篮球", "网球", "奥运", "冠军", "联赛", "比赛",
		},
	},
	{
		id:   "IAB18",
		name: "Style & Fashion",
		keywords: []string{
			"fashion", "style", "clothing", "dress", "outfit", "brand", "luxury",
			"designer", "trend", "beauty", "makeup", "skincare", "jewelry",
			"时尚", "服装", "品牌", "美容", "护肤", "珠宝", "奢侈品",
		},
	},
	{
		id:   "IAB19",
		name: "Technology & Computing",
		keywords: []string{
			"technology", "tech", "software", "hardware", "computer", "programming",
			"internet", "app", "mobile", "smartphone", "ai", "machine learning",
			"cloud", "cybersecurity", "data", "startup", "silicon valley", "api",
			"技术", "科技", "软件", "硬件", "计算机", "互联网", "人工智能", "云计算",
		},
	},
	{
		id:   "IAB20",
		name: "Travel",
		keywords: []string{
			"travel", "tourism", "hotel", "flight", "airline", "destination",
			"vacation", "holiday", "adventure", "beach", "mountain", "culture",
			"旅游", "旅行", "酒店", "航班", "目的地", "度假", "景点",
		},
	},
	{
		id:   "IAB22",
		name: "Shopping",
		keywords: []string{
			"shopping", "sale", "discount", "product", "buy", "purchase", "ecommerce",
			"retail", "store", "amazon", "price", "review", "deal", "coupon",
			"购物", "商品", "折扣", "电商", "零售", "价格", "优惠",
		},
	},
	{
		id:   "IAB25-3",
		name: "Gambling",
		keywords: []string{
			"gambling", "casino", "poker", "slot", "betting", "lottery", "jackpot",
			"赌博", "赌场", "彩票", "下注", "博彩",
		},
	},
}

// CategoryClassifier 基于关键词词典的 IAB 内容分类器。
type CategoryClassifier struct {
	taxonomy []iabEntry
	tfidf    *nlp.TFIDF
}

// NewCategoryClassifier 创建分类器实例。
func NewCategoryClassifier() *CategoryClassifier {
	return &CategoryClassifier{
		taxonomy: iabTaxonomy,
		tfidf:    nlp.NewTFIDF(),
	}
}

// Classify 对文本进行 IAB 分类，返回置信度 >= minConf 的类别列表。
func (c *CategoryClassifier) Classify(text string, minConf float64) []Category {
	if text == "" {
		return nil
	}

	tokens := nlp.Tokenize(text)
	words := nlp.FilterTokens(tokens)
	freq := nlp.WordFrequency(words)

	type scored struct {
		entry iabEntry
		score float64
	}
	scores := make([]scored, 0, len(c.taxonomy))

	totalWords := float64(len(words))
	if totalWords == 0 {
		return nil
	}

	for _, entry := range c.taxonomy {
		s := c.scoreEntry(entry, freq, totalWords)
		if s > 0 {
			scores = append(scores, scored{entry: entry, score: s})
		}
	}

	if len(scores) == 0 {
		return []Category{{ID: "IAB24", Name: "Uncategorized", Confidence: 1.0}}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// 归一化
	maxScore := scores[0].score
	categories := make([]Category, 0, len(scores))
	for _, s := range scores {
		conf := s.score / maxScore
		if conf < minConf {
			break
		}
		categories = append(categories, Category{
			ID:         s.entry.id,
			Name:       s.entry.name,
			Confidence: roundFloat(conf, 2),
		})
	}

	if len(categories) == 0 {
		return []Category{{ID: "IAB24", Name: "Uncategorized", Confidence: 1.0}}
	}
	return categories
}

// scoreEntry 计算文本与分类词典的匹配分数。
func (c *CategoryClassifier) scoreEntry(entry iabEntry, freq map[string]int, total float64) float64 {
	score := 0.0
	for _, kw := range entry.keywords {
		kw = strings.ToLower(kw)
		// 精确匹配
		if cnt, ok := freq[kw]; ok {
			score += float64(cnt) / total * weightByLength(kw)
		}
		// 词组匹配：检查词组中的关键词是否都出现
		parts := strings.Fields(kw)
		if len(parts) > 1 {
			allMatch := true
			for _, p := range parts {
				if _, ok := freq[p]; !ok {
					allMatch = false
					break
				}
			}
			if allMatch {
				score += 0.5 / total * weightByLength(kw)
			}
		}
	}
	return score
}

// AllCategories 返回所有 IAB 类别列表（不含分数）。
func AllCategories() []Category {
	cats := make([]Category, 0, len(iabTaxonomy)+1)
	for _, e := range iabTaxonomy {
		cats = append(cats, Category{ID: e.id, Name: e.name})
	}
	cats = append(cats, Category{ID: "IAB24", Name: "Uncategorized"})
	return cats
}

func weightByLength(kw string) float64 {
	l := len([]rune(kw))
	if l <= 3 {
		return 1.0
	}
	if l <= 6 {
		return 1.5
	}
	return 2.0
}

func roundFloat(f float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(f*pow+0.5)) / pow
}
