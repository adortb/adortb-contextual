package fetch

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// Extracted 从 HTML 提取的结构化内容。
type Extracted struct {
	Title       string
	Description string
	Body        string
	Language    string
}

// ExtractHTML 从原始 HTML 字符串中提取标题、描述和正文。
func ExtractHTML(rawHTML string) Extracted {
	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return Extracted{}
	}

	var (
		title       string
		description string
		bodyTexts   []string
		lang        string
	)

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch strings.ToLower(n.Data) {
			case "html":
				lang = attrVal(n, "lang")
			case "title":
				if n.FirstChild != nil {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "meta":
				name := strings.ToLower(attrVal(n, "name"))
				prop := strings.ToLower(attrVal(n, "property"))
				content := attrVal(n, "content")
				switch {
				case name == "description" || prop == "og:description":
					if description == "" {
						description = content
					}
				case name == "og:title" || prop == "og:title":
					if title == "" {
						title = content
					}
				}
			case "script", "style", "nav", "footer", "header", "aside", "form":
				return // 跳过非内容标签
			case "p", "article", "section", "main", "div", "span", "li", "h1",
				"h2", "h3", "h4", "h5", "h6", "td", "th", "blockquote":
				text := extractText(n)
				text = strings.TrimSpace(text)
				if utf8.RuneCountInString(text) >= 10 {
					bodyTexts = append(bodyTexts, text)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	body := strings.Join(bodyTexts, " ")
	// 限制正文长度，避免内存过大
	body = truncateRunes(body, 5000)

	return Extracted{
		Title:       title,
		Description: description,
		Body:        body,
		Language:    normalizeLanguage(lang),
	}
}

// extractText 递归提取节点的纯文本。
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(extractText(c))
	}
	return sb.String()
}

// attrVal 获取 HTML 节点的属性值。
func attrVal(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if strings.ToLower(a.Key) == key {
			return a.Val
		}
	}
	return ""
}

// normalizeLanguage 标准化语言代码。
func normalizeLanguage(lang string) string {
	if lang == "" {
		return "en"
	}
	lower := strings.ToLower(lang)
	switch {
	case strings.HasPrefix(lower, "zh"):
		return "zh-CN"
	case strings.HasPrefix(lower, "ja"):
		return "ja"
	case strings.HasPrefix(lower, "ko"):
		return "ko"
	default:
		return "en"
	}
}

// truncateRunes 按 rune 数量截断字符串。
func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}
