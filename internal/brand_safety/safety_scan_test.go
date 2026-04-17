package brand_safety

import (
	"testing"
)

func TestScanner_SafeContent(t *testing.T) {
	s := NewScanner()
	result := s.Scan("The new electric vehicle model offers excellent range and performance.")
	if result.Score < 0.9 {
		t.Errorf("expected safe content score >= 0.9, got %f (flags: %v)", result.Score, result.Flags)
	}
	if result.BlockLevel != "none" {
		t.Errorf("expected block_level=none for safe content, got %s", result.BlockLevel)
	}
}

func TestScanner_GamblingContent(t *testing.T) {
	s := NewScanner()
	result := s.Scan("Online casino gambling poker slots betting jackpot winners")
	found := false
	for _, f := range result.Flags {
		if f == "gambling" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected gambling flag, got flags: %v", result.Flags)
	}
	if result.Score >= 1.0 {
		t.Errorf("expected score < 1.0 for gambling content, got %f", result.Score)
	}
}

func TestScanner_ViolentContent(t *testing.T) {
	s := NewScanner()
	result := s.Scan("terrorist bomb attack violence killing massacre")
	if result.BlockLevel == "none" {
		t.Errorf("expected warn or block for violent content, got none (score=%f)", result.Score)
	}
}

func TestScanner_EmptyContent(t *testing.T) {
	s := NewScanner()
	result := s.Scan("")
	if result.Score != 1.0 {
		t.Errorf("expected score=1.0 for empty content, got %f", result.Score)
	}
}

func TestScanner_ScoreRange(t *testing.T) {
	s := NewScanner()
	texts := []string{
		"Normal news article about technology",
		"Casino gambling poker betting adult explicit violence",
		"",
		"hate speech racism violent terrorism drugs",
	}
	for _, text := range texts {
		result := s.Scan(text)
		if result.Score < 0 || result.Score > 1 {
			t.Errorf("score out of [0,1] for %q: %f", text, result.Score)
		}
	}
}
