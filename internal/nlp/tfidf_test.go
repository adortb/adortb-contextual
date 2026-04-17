package nlp

import (
	"testing"
)

func TestTFIDF_TopKeywords_Stability(t *testing.T) {
	tf := NewTFIDF()
	words := []string{
		"electric", "vehicle", "tesla", "battery", "charging",
		"electric", "vehicle", "tesla", "electric",
		"technology", "innovation", "battery",
	}

	// 多次调用结果应一致（稳定性）
	kw1 := tf.TopKeywords(words, 5)
	kw2 := tf.TopKeywords(words, 5)

	if len(kw1) != len(kw2) {
		t.Fatalf("result length differs: %d vs %d", len(kw1), len(kw2))
	}
	for i := range kw1 {
		if kw1[i].Word != kw2[i].Word {
			t.Errorf("unstable result at pos %d: %s vs %s", i, kw1[i].Word, kw2[i].Word)
		}
	}
}

func TestTFIDF_TopKeywords_TopNRespected(t *testing.T) {
	tf := NewTFIDF()
	words := []string{"a1", "b2", "c3", "d4", "e5", "f6", "g7", "h8", "i9", "j10",
		"a1", "b2", "c3", "d4", "e5"}
	kw := tf.TopKeywords(words, 3)
	if len(kw) > 3 {
		t.Errorf("expected at most 3 keywords, got %d", len(kw))
	}
}

func TestTFIDF_TopKeywords_Empty(t *testing.T) {
	tf := NewTFIDF()
	kw := tf.TopKeywords(nil, 10)
	if kw != nil {
		t.Errorf("expected nil for empty input, got %v", kw)
	}
}

func TestTFIDF_ScoreOrdering(t *testing.T) {
	tf := NewTFIDF()
	// "electric" 出现最多，应排在前面
	words := []string{
		"electric", "electric", "electric", "vehicle", "technology",
	}
	kw := tf.TopKeywords(words, 5)
	if len(kw) == 0 {
		t.Fatal("expected at least one keyword")
	}
	// 检查结果按分数降序
	for i := 1; i < len(kw); i++ {
		if kw[i].Score > kw[i-1].Score {
			t.Errorf("keywords not sorted by score: %v > %v", kw[i].Score, kw[i-1].Score)
		}
	}
}

func TestWordFrequency(t *testing.T) {
	words := []string{"go", "go", "rust", "python", "go"}
	freq := WordFrequency(words)
	if freq["go"] != 3 {
		t.Errorf("expected freq[go]=3, got %d", freq["go"])
	}
	if freq["rust"] != 1 {
		t.Errorf("expected freq[rust]=1, got %d", freq["rust"])
	}
}
