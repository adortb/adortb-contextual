# CLAUDE.md — adortb-contextual

## 项目角色

九期服务：上下文内容智能。实时分析网页内容，为上下文定向提供 IAB 分类和品牌安全信号。对抓取延迟和分类准确率都有要求。

## 关键函数与复杂度

| 函数 | 文件 | 复杂度 | 说明 |
|------|------|--------|------|
| `TFIDF.TopKeywords` | `nlp/tfidf.go` | O(W·log W) | W=词数，sort 主导 |
| `TFIDF.estimateIDF` | `nlp/tfidf.go` | O(len(word)) | 按字符长度分档 |
| `CategoryClassifier.Classify` | `classifier/category.go` | O(C·K·W) | C=类别数(16), K=关键词数 |
| `CategoryClassifier.scoreEntry` | `classifier/category.go` | O(K·W) | 词组匹配需遍历词组所有词 |
| `SentimentAnalyzer.Analyze` | `classifier/sentiment.go` | O(W·3) | W词，回溯3个否定词 |
| `nlp.Tokenize` | `nlp/tokenizer.go` | O(len(text)) | 中英文分词 |

## IAB 分类器规范

- **词典（iabTaxonomy）** 在 `category.go` 编译期初始化，修改后需重新编译
- 新增 IAB 类别：在 `iabTaxonomy` slice 中添加 `iabEntry`，关键词需覆盖中英文
- `minConf` 默认 0.1，过低会导致分类噪声多，过高会漏分
- 分数归一化：以最高分作为基准，其余按比例
- 多类别返回：`confidence ≥ minConf` 的所有类别均返回（不只返回第一个）
- 未分类 fallback：`{ID: "IAB24", Name: "Uncategorized", Confidence: 1.0}`

## 情感分析注意事项

- 否定词回溯窗口：前3个词（`max(0, i-3)` 到 `i`）
- 否定词反转效果（不是完全取反）：
  - 正面词被否定 → 计入负面（×0.5权重）
  - 负面词被否定 → 计入正面（×0.3权重）
- `total < 0.01` 时返回 neutral（无情感词覆盖）
- `normalizedScore = (diff/total + 1) / 2` 映射到[0,1]

## TF-IDF IDF 估算规范

当前 IDF 用词长度启发式估算（无实际语料库），精度有限：
- 短词（≤2字符）IDF 低，如 "is", "be"（需配合 stopwords 过滤）
- 长词（>6字符）IDF 高，如 "breakthrough", "electric"
- **不要改动 df 分档比例**，这是经过调参的经验值

## 网页抓取注意事项

- 抓取超时：建议 5-10 秒，超时返回错误（不阻塞分析流程）
- 缓存 TTL：页面内容缓存 1 小时（新闻类可更短）
- `extractor.go` 提取正文时需过滤导航栏、页脚等干扰元素
- User-Agent 需设置为真实浏览器UA，避免被反爬

## 新增关键词词典规范

1. 关键词统一小写
2. 中英文混合（支持双语）
3. 词组（多词短语）用空格分隔，`scoreEntry` 会做部分匹配
4. 词典变更后运行回归测试验证分类精度

## 测试

```bash
go test -race ./...
go test -v ./internal/nlp/ -run TestTFIDF
go test -v ./internal/classifier/ -run TestCategory
go test -v ./internal/classifier/ -run TestSentiment
go test -v ./internal/brand_safety/ -run TestSafety
```

关键测试：
- `nlp/tfidf_test.go` — 关键词排序正确性、边界（空输入、全停用词）
- `nlp/tokenizer_test.go` — 中英文分词
- `classifier/category_test.go` — IAB 分类准确性
- `classifier/sentiment_test.go` — 否定词处理、中英文情感词
- `classifier/entity_test.go` — 实体识别
- `brand_safety/safety_scan_test.go` — 高风险内容检测
- `api/handler_test.go` — 端到端接口测试
