# adortb-contextual

九期服务。上下文内容智能（Contextual Intelligence）引擎，通过分析网页内容进行 IAB 分类、NER 命名实体识别、情感分析和品牌安全扫描，为上下文定向广告提供内容理解能力。

## 算法概述

### TF-IDF 关键词提取

基于单文档 TF-IDF 的关键词权重计算：

```
TF(w) = count(w) / total_words

IDF(w) 启发式估算（按词长度分档）：
  len ≤ 2 → df = N × 0.1    （高频短词）
  len ≤ 4 → df = N × 0.01
  len ≤ 6 → df = N × 0.001
  len > 6 → df = N × 0.0001  （低频长词）
  N = 1,000,000（虚拟语料库大小）

TFIDF(w) = TF(w) × (ln(N/df) + 1)
```

`TopKeywords(words, topN)` 返回按 TF-IDF 降序排列的关键词列表。

### IAB 内容分类（IAB Content Taxonomy 3.0）

基于关键词词典的分类器，覆盖 16 个 IAB 标准类别（IAB1~IAB25-3）：

```
score(category) = Σ_{kw ∈ category.keywords ∩ text} tf(kw) × weight(len(kw))

词长权重：
  len ≤ 3 → weight = 1.0
  len ≤ 6 → weight = 1.5
  len > 6 → weight = 2.0   （长词更有判别力）

词组匹配（多词关键词）：
  所有词都出现 → 加分 0.5/total × weight

confidence = score / max_score（归一化到[0,1]）
默认 minConf = 0.1
未匹配任何类别 → IAB24（Uncategorized）
```

### 命名实体识别（NER）

`classifier/entity.go` 基于规则和词典识别：
- 人名（PERSON）
- 组织名（ORGANIZATION）
- 地点（LOCATION）
- 品牌名（BRAND）

### 情感分析

基于情感词典（正/负面词 + 否定词处理）：

```
对每个词 w_i：
  检查前3个词中是否有否定词（negated=true）

  if w_i ∈ positive_words:
    negated ? negScore += score×0.5 : posScore += score
  if w_i ∈ negative_words:
    negated ? posScore += score×0.3 : negScore += score

normalizedScore = (posScore - negScore)/(posScore + negScore) / 2 + 0.5
  → [0,1]，> 0.6 为正面，< 0.4 为负面
```

支持中英双语词典（正面词 ~50个，负面词 ~40个，否定词 ~15个）。

### 品牌安全扫描

`brand_safety/safety_scan.go` 检测高风险内容类别（赌博/暴力/色情/仇恨等），为广告投放提供品牌安全分级。

## 快速开始

```bash
go build -o bin/contextual ./cmd/contextual
./bin/contextual -port 8086

# 内容分析请求
curl -X POST http://localhost:8086/v1/analyze \
  -d '{"url":"https://example.com/article/123"}'

# 或直接传文本
curl -X POST http://localhost:8086/v1/analyze \
  -d '{"text":"Tesla launches new electric vehicle with breakthrough battery technology"}'
```

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/analyze` | 分析URL/文本，返回完整智能标签 |
| GET  | `/v1/categories` | 列出所有 IAB 类别 |
| GET  | `/metrics` | Prometheus 指标 |

## 输出示例

```json
{
  "url": "https://example.com/article/123",
  "categories": [
    {"id": "IAB19", "name": "Technology & Computing", "confidence": 1.0},
    {"id": "IAB2",  "name": "Automotive",             "confidence": 0.31}
  ],
  "keywords": [
    {"word": "electric",    "score": 0.082},
    {"word": "battery",     "score": 0.076},
    {"word": "breakthrough","score": 0.064}
  ],
  "entities": [
    {"text": "Tesla", "type": "BRAND", "confidence": 0.95}
  ],
  "sentiment": {"polarity": "positive", "score": 0.78},
  "brand_safety": {"safe": true, "risk_level": "low"}
}
```

## 技术栈

- **语言**: Go
- **NLP**: 纯 Go 实现（无外部 NLP 库）
- **网页抓取**: `internal/fetch/crawler.go`（HTTP 抓取 + 文本提取）
- **缓存**: `internal/fetch/cache.go`（页面内容缓存，避免重复抓取）

## 目录结构

```
adortb-contextual/
├── client/                   # Go SDK
├── cmd/contextual/
├── internal/
│   ├── api/
│   ├── brand_safety/
│   │   └── safety_scan.go    # 品牌安全分级
│   ├── classifier/
│   │   ├── category.go       # IAB 分类（词典匹配）
│   │   ├── entity.go         # NER 实体识别
│   │   └── sentiment.go      # 情感分析（双语词典）
│   ├── fetch/
│   │   ├── crawler.go        # HTTP 网页抓取
│   │   ├── extractor.go      # 正文提取
│   │   └── cache.go          # 内容缓存
│   └── nlp/
│       ├── tfidf.go          # TF-IDF 计算
│       ├── tokenizer.go      # 分词（中英文）
│       └── stopwords.go      # 停用词过滤
```
