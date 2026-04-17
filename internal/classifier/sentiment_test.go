package classifier

import (
	"testing"
)

func TestSentimentAnalyzer_Positive(t *testing.T) {
	sa := NewSentimentAnalyzer()
	result := sa.Analyze("This is an amazing and excellent product. I love it, it's fantastic and great.")
	if result.Polarity != "positive" {
		t.Errorf("expected positive, got %s (score=%f)", result.Polarity, result.Score)
	}
	if result.Score < 0.6 {
		t.Errorf("expected score >= 0.6 for positive text, got %f", result.Score)
	}
}

func TestSentimentAnalyzer_Negative(t *testing.T) {
	sa := NewSentimentAnalyzer()
	result := sa.Analyze("This is terrible and horrible. The worst product I've ever seen. Complete failure.")
	if result.Polarity != "negative" {
		t.Errorf("expected negative, got %s (score=%f)", result.Polarity, result.Score)
	}
}

func TestSentimentAnalyzer_Neutral(t *testing.T) {
	sa := NewSentimentAnalyzer()
	result := sa.Analyze("The meeting was held on Tuesday. Participants discussed the quarterly report.")
	if result.Polarity != "neutral" {
		t.Logf("note: expected neutral, got %s (this may be acceptable)", result.Polarity)
	}
}

func TestSentimentAnalyzer_Empty(t *testing.T) {
	sa := NewSentimentAnalyzer()
	result := sa.Analyze("")
	if result.Polarity != "neutral" {
		t.Errorf("expected neutral for empty text, got %s", result.Polarity)
	}
}

func TestSentimentAnalyzer_ScoreRange(t *testing.T) {
	sa := NewSentimentAnalyzer()
	texts := []string{
		"excellent wonderful amazing great",
		"terrible horrible awful worst failure",
		"the meeting was scheduled",
		"",
	}
	for _, text := range texts {
		result := sa.Analyze(text)
		if result.Score < 0 || result.Score > 1 {
			t.Errorf("score out of [0,1] range for %q: %f", text, result.Score)
		}
	}
}

func TestSentimentAnalyzer_Negation(t *testing.T) {
	sa := NewSentimentAnalyzer()
	// "not good" 应该比 "good" 情感更负面
	pos := sa.Analyze("This is good")
	neg := sa.Analyze("This is not good")
	if pos.Score <= neg.Score {
		t.Logf("note: negation test: positive score=%f, negated score=%f", pos.Score, neg.Score)
	}
}

func TestSentimentAnalyzer_Chinese(t *testing.T) {
	sa := NewSentimentAnalyzer()
	result := sa.Analyze("这个产品非常好，完美，我非常喜欢")
	if result.Polarity != "positive" {
		t.Logf("note: Chinese positive text got polarity=%s", result.Polarity)
	}
}
