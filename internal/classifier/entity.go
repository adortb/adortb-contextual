package classifier

import (
	"strings"
	"unicode"
)

// EntityType 实体类型。
type EntityType string

const (
	EntityPerson   EntityType = "person"
	EntityBrand    EntityType = "brand"
	EntityLocation EntityType = "location"
	EntityOrg      EntityType = "organization"
)

// Entity 命名实体。
type Entity struct {
	Type EntityType `json:"type"`
	Text string     `json:"text"`
}

// trieNode Trie 树节点。
type trieNode struct {
	children map[rune]*trieNode
	end      bool
	entity   *entityDef
}

type entityDef struct {
	text       string
	entityType EntityType
}

// EntityRecognizer 基于 Trie 的命名实体识别器。
type EntityRecognizer struct {
	root *trieNode
}

// NewEntityRecognizer 创建内置词典的实体识别器。
func NewEntityRecognizer() *EntityRecognizer {
	er := &EntityRecognizer{root: &trieNode{children: make(map[rune]*trieNode)}}
	er.loadBuiltinDict()
	return er
}

// loadBuiltinDict 加载内置命名实体词典。
func (er *EntityRecognizer) loadBuiltinDict() {
	persons := []string{
		"Elon Musk", "Tim Cook", "Jeff Bezos", "Mark Zuckerberg", "Bill Gates",
		"Steve Jobs", "Sundar Pichai", "Satya Nadella", "Jack Ma", "Ren Zhengfei",
		"Warren Buffett", "Barack Obama", "Donald Trump", "Joe Biden", "Xi Jinping",
		"Vladimir Putin", "Emmanuel Macron", "Angela Merkel", "Justin Trudeau",
		"马云", "任正非", "雷军", "马化腾", "张一鸣", "王兴",
	}
	brands := []string{
		"Tesla", "Apple", "Google", "Microsoft", "Amazon", "Meta", "Netflix",
		"Twitter", "Uber", "Airbnb", "Spotify", "Samsung", "Huawei", "Alibaba",
		"Tencent", "ByteDance", "TikTok", "WeChat", "Baidu", "JD.com",
		"Nike", "Adidas", "Gucci", "Louis Vuitton", "Chanel", "BMW", "Mercedes",
		"Toyota", "Honda", "Ford", "Volkswagen", "OpenAI", "Anthropic",
		"阿里巴巴", "腾讯", "百度", "字节跳动", "京东", "小米", "华为", "滴滴",
	}
	locations := []string{
		"United States", "China", "United Kingdom", "Germany", "France", "Japan",
		"South Korea", "India", "Brazil", "Canada", "Australia", "Russia",
		"New York", "Los Angeles", "San Francisco", "Silicon Valley", "London",
		"Paris", "Berlin", "Tokyo", "Shanghai", "Beijing", "Hong Kong",
		"北京", "上海", "广州", "深圳", "杭州", "成都", "武汉", "西安",
		"中国", "美国", "英国", "日本", "韩国", "欧洲", "亚洲",
	}
	orgs := []string{
		"United Nations", "WHO", "IMF", "World Bank", "NATO", "EU", "WTO",
		"Federal Reserve", "European Central Bank", "SEC", "FBI", "CIA",
		"联合国", "世界卫生组织", "国际货币基金组织",
	}

	for _, p := range persons {
		er.insert(p, EntityPerson)
	}
	for _, b := range brands {
		er.insert(b, EntityBrand)
	}
	for _, l := range locations {
		er.insert(l, EntityLocation)
	}
	for _, o := range orgs {
		er.insert(o, EntityOrg)
	}
}

// insert 向 Trie 插入实体。
func (er *EntityRecognizer) insert(text string, etype EntityType) {
	node := er.root
	lower := strings.ToLower(text)
	for _, r := range lower {
		if node.children == nil {
			node.children = make(map[rune]*trieNode)
		}
		child, ok := node.children[r]
		if !ok {
			child = &trieNode{children: make(map[rune]*trieNode)}
			node.children[r] = child
		}
		node = child
	}
	node.end = true
	node.entity = &entityDef{text: text, entityType: etype}
}

// Recognize 在文本中识别所有命名实体，返回去重结果。
func (er *EntityRecognizer) Recognize(text string) []Entity {
	if text == "" {
		return nil
	}

	lower := strings.ToLower(text)
	runes := []rune(lower)
	n := len(runes)

	seen := make(map[string]struct{})
	var entities []Entity

	for i := 0; i < n; i++ {
		// 仅从单词边界开始匹配
		if i > 0 && isWordChar(runes[i-1]) && isWordChar(runes[i]) {
			continue
		}
		node := er.root
		j := i
		lastMatch := -1
		var lastEntity *entityDef

		for j < n {
			child, ok := node.children[runes[j]]
			if !ok {
				break
			}
			node = child
			j++
			if node.end {
				// 确认单词边界
				if j == n || !isWordChar(runes[j]) {
					lastMatch = j
					lastEntity = node.entity
				}
			}
		}

		if lastMatch > 0 && lastEntity != nil {
			key := lastEntity.text
			if _, dup := seen[key]; !dup {
				seen[key] = struct{}{}
				entities = append(entities, Entity{
					Type: lastEntity.entityType,
					Text: lastEntity.text,
				})
			}
			i = lastMatch - 1
		}
	}
	return entities
}

func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
