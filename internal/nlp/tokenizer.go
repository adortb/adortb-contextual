package nlp

import (
	"strings"
	"unicode"
)

// Token 表示一个分词结果。
type Token struct {
	Text string
	Lang string // "zh" | "en"
}

// Tokenize 对混合文本进行分词。
// 英文按空格/标点切分；中文按字符 bigram + unigram 切分。
func Tokenize(text string) []Token {
	text = strings.ToLower(text)
	tokens := make([]Token, 0, len(text)/3)

	runes := []rune(text)
	i := 0
	for i < len(runes) {
		r := runes[i]
		if isCJK(r) {
			// 中文：unigram + bigram
			tokens = append(tokens, Token{Text: string(r), Lang: "zh"})
			if i+1 < len(runes) && isCJK(runes[i+1]) {
				tokens = append(tokens, Token{Text: string(runes[i : i+2]), Lang: "zh"})
			}
			i++
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// 英文 / 数字：收集连续字符
			j := i
			for j < len(runes) && (unicode.IsLetter(runes[j]) || unicode.IsDigit(runes[j])) {
				j++
			}
			word := string(runes[i:j])
			tokens = append(tokens, Token{Text: word, Lang: "en"})
			i = j
		} else {
			i++
		}
	}
	return tokens
}

// FilterTokens 过滤停用词并返回有效词列表（最小长度 2）。
func FilterTokens(tokens []Token) []string {
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if len([]rune(t.Text)) < 2 {
			continue
		}
		if IsStopword(t.Text) {
			continue
		}
		out = append(out, t.Text)
	}
	return out
}

// isCJK 判断是否为 CJK 字符（中日韩）。
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) ||
		(r >= 0x3400 && r <= 0x4DBF) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0x20000 && r <= 0x2A6DF)
}
