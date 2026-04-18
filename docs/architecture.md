# Architecture — adortb-contextual

## 系统概览

```
广告请求（含页面URL）
    │
    ▼
┌─────────────────────────────────────────────────────┐
│                HTTP API Server                       │
│           POST /v1/analyze                           │
└──────────────────────┬──────────────────────────────┘
                       │
          ┌────────────▼────────────┐
          │     Cache Check         │ 命中 → 直接返回缓存标签
          └────────────┬────────────┘
                   未命中
                       │
          ┌────────────▼────────────┐
          │  Crawler + Extractor    │ 抓取网页，提取正文
          └────────────┬────────────┘
                       │ 正文 text
          ┌────────────▼────────────┐
          │      NLP Pipeline       │
          │  Tokenize → Filter      │
          └────────────┬────────────┘
                       │ tokens / words
     ┌─────────────────┼─────────────────┐
     │                 │                 │
     ▼                 ▼                 ▼
┌─────────┐    ┌───────────────┐   ┌──────────┐
│  TFIDF  │    │IAB Classifier │   │Sentiment │
│Keywords │    │(词典匹配)     │   │Analyzer  │
└────┬────┘    └───────┬───────┘   └────┬─────┘
     │                 │                 │
     └─────────────────┼─────────────────┘
                       │
          ┌────────────▼────────────┐
          │     NER Extractor        │ 命名实体识别
          └────────────┬────────────┘
                       │
          ┌────────────▼────────────┐
          │   Brand Safety Scanner  │ 品牌安全分级
          └────────────┬────────────┘
                       │
          ┌────────────▼────────────┐
          │      Cache Write        │ 写入缓存（TTL=1h）
          └────────────┬────────────┘
                       │
                  完整智能标签结果
```

## 分析流程时序图

```
Client      API Handler    Cache     Crawler    NLP Pipeline
  │              │            │          │            │
  │─ /analyze ──►│            │          │            │
  │              │─ Get(url) ─►│          │            │
  │              │◄── miss ────│          │            │
  │              │─ Fetch(url) ──────────►│            │
  │              │◄── raw HTML ──────────│            │
  │              │─ Extract(html) → text │            │
  │              │─ Tokenize(text) ─────────────────── ►
  │              │◄── tokens/words ────────────────── │
  │              │─ TFIDF.TopKeywords()  │            │
  │              │─ CategoryClassifier.Classify()      │
  │              │─ SentimentAnalyzer.Analyze()        │
  │              │─ EntityExtractor.Extract()           │
  │              │─ SafetyScanner.Scan()               │
  │              │◄── ContextualTags ──────────────── │
  │              │─ Set(url, tags) ────►│              │
  │◄── response ─│            │         │              │
```

## NLP 处理流程

```
原始文本
  "Tesla launches new electric vehicle with breakthrough battery technology"
    │
    ▼ Tokenize(text)
  ["Tesla", "launches", "new", "electric", "vehicle", "with",
   "breakthrough", "battery", "technology"]
    │
    ▼ FilterTokens (去停用词、小写)
  ["tesla", "launches", "electric", "vehicle",
   "breakthrough", "battery", "technology"]
    │
    ├──► TF-IDF TopKeywords(N=5)
    │    [electric:0.082, battery:0.076, breakthrough:0.064, ...]
    │
    ├──► IAB Classify(minConf=0.1)
    │    IAB19(Technology): score=0.18 → conf=1.0
    │    IAB2(Automotive): score=0.06 → conf=0.33
    │    conf≥0.1 → [IAB19, IAB2]
    │
    └──► Sentiment.Analyze()
         positive: {"launches","breakthrough"} → posScore=3.3
         negative: 无
         normalizedScore=(3.3/3.3+1)/2 = 1.0 → "positive"
```

## IAB 分类词典结构

```
iabTaxonomy = [
  {id:"IAB1",  name:"Arts & Entertainment",   keywords:[...28个英中词]},
  {id:"IAB2",  name:"Automotive",             keywords:[...22个英中词]},
  {id:"IAB3",  name:"Business",               keywords:[...20个英中词]},
  {id:"IAB5",  name:"Education",              keywords:[...18个英中词]},
  {id:"IAB7",  name:"Health & Fitness",       keywords:[...20个英中词]},
  {id:"IAB8",  name:"Food & Drink",           keywords:[...18个英中词]},
  {id:"IAB11", name:"Law, Gov & Politics",    keywords:[...20个英中词]},
  {id:"IAB12", name:"News & Weather",         keywords:[...14个英中词]},
  {id:"IAB13", name:"Personal Finance",       keywords:[...18个英中词]},
  {id:"IAB15", name:"Science",                keywords:[...15个英中词]},
  {id:"IAB17", name:"Sports",                 keywords:[...20个英中词]},
  {id:"IAB18", name:"Style & Fashion",        keywords:[...14个英中词]},
  {id:"IAB19", name:"Technology & Computing", keywords:[...18个英中词]},
  {id:"IAB20", name:"Travel",                 keywords:[...14个英中词]},
  {id:"IAB22", name:"Shopping",               keywords:[...14个英中词]},
  {id:"IAB25-3",name:"Gambling",              keywords:[...10个英中词]},
  {id:"IAB24", name:"Uncategorized",          fallback},
]
```

## 数据输入输出

### 输入

| 字段 | 类型 | 说明 |
|------|------|------|
| url | string | 待分析网页URL（互斥于text） |
| text | string | 直接传入文本（互斥于url） |
| min_confidence | float64 | IAB 分类最低置信度（默认0.1） |

### 输出

```json
{
  "categories": [{"id":"IAB19","name":"Technology","confidence":1.0}],
  "keywords":   [{"word":"electric","score":0.082}],
  "entities":   [{"text":"Tesla","type":"BRAND","confidence":0.95}],
  "sentiment":  {"polarity":"positive","score":0.78},
  "brand_safety":{"safe":true,"risk_level":"low","flags":[]}
}
```

## 评估指标

| 指标 | 说明 | 目标 |
|------|------|------|
| IAB 分类准确率 | Top-1 准确率 | > 85% |
| 品牌安全召回率 | 高危内容检出率 | > 95%（宁可误报）|
| P95 分析延迟 | 含抓取（缓存命中） | < 50ms |
| P95 分析延迟 | 含抓取（缓存未命中） | < 2s |
| 缓存命中率 | 热门页面 | > 70% |

## 依赖关系

```
adortb-contextual
└── （无外部存储依赖，缓存在 internal/fetch/cache.go 内存实现）
    （生产环境建议替换为 Redis 缓存）
```
