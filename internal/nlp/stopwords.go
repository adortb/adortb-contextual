// Package nlp 提供文本自然语言处理能力：分词、TF-IDF、停用词过滤。
package nlp

// stopwordsEN 英文停用词集合。
var stopwordsEN = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "but": {}, "in": {},
	"on": {}, "at": {}, "to": {}, "for": {}, "of": {}, "with": {}, "by": {},
	"from": {}, "is": {}, "was": {}, "are": {}, "were": {}, "be": {}, "been": {},
	"being": {}, "have": {}, "has": {}, "had": {}, "do": {}, "does": {}, "did": {},
	"will": {}, "would": {}, "could": {}, "should": {}, "may": {}, "might": {},
	"this": {}, "that": {}, "these": {}, "those": {}, "it": {}, "its": {},
	"he": {}, "she": {}, "they": {}, "we": {}, "i": {}, "you": {}, "me": {},
	"him": {}, "her": {}, "them": {}, "us": {}, "my": {}, "your": {}, "his": {},
	"their": {}, "our": {}, "not": {}, "no": {}, "nor": {}, "so": {}, "yet": {},
	"as": {}, "if": {}, "when": {}, "while": {}, "than": {}, "then": {}, "too": {},
	"also": {}, "just": {}, "about": {}, "up": {}, "out": {}, "how": {}, "all": {},
	"what": {}, "which": {}, "who": {}, "whom": {}, "there": {}, "here": {},
	"after": {}, "before": {}, "into": {}, "over": {}, "through": {}, "more": {},
	"can": {}, "one": {}, "two": {}, "new": {}, "get": {}, "said": {}, "says": {},
	"like": {}, "well": {}, "back": {}, "see": {}, "now": {}, "know": {},
}

// stopwordsZH 中文停用词集合。
var stopwordsZH = map[string]struct{}{
	"的": {}, "了": {}, "在": {}, "是": {}, "我": {}, "有": {}, "和": {}, "就": {},
	"不": {}, "人": {}, "都": {}, "一": {}, "一个": {}, "上": {}, "也": {}, "很": {},
	"到": {}, "说": {}, "要": {}, "去": {}, "你": {}, "会": {}, "着": {}, "没有": {},
	"看": {}, "好": {}, "自己": {}, "这": {}, "那": {}, "里": {}, "以": {}, "及": {},
	"等": {}, "中": {}, "为": {}, "对": {}, "他": {}, "她": {}, "它": {}, "们": {},
	"而": {}, "被": {}, "把": {}, "与": {}, "或": {}, "所": {}, "因": {}, "此": {},
	"但": {}, "并": {}, "已": {}, "来": {}, "后": {}, "由": {}, "其": {}, "之": {},
	"时": {}, "年": {}, "月": {}, "日": {}, "进行": {}, "通过": {}, "表示": {},
	"可以": {}, "如果": {}, "因为": {}, "所以": {}, "这样": {}, "那么": {}, "一些": {},
	"已经": {}, "还是": {}, "以及": {}, "可能": {}, "需要": {},
}

// IsStopword 判断一个词是否为停用词（中英文均支持）。
func IsStopword(word string) bool {
	if _, ok := stopwordsEN[word]; ok {
		return true
	}
	_, ok := stopwordsZH[word]
	return ok
}
