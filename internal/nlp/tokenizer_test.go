package nlp

import (
	"testing"
)

func TestTokenize_English(t *testing.T) {
	tokens := Tokenize("Electric vehicles are transforming the automotive industry.")
	found := false
	for _, tok := range tokens {
		if tok.Text == "electric" && tok.Lang == "en" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected token 'electric' not found in %v", tokens)
	}
}

func TestTokenize_Chinese(t *testing.T) {
	tokens := Tokenize("人工智能技术")
	if len(tokens) == 0 {
		t.Fatal("expected tokens for Chinese text, got none")
	}
	// 验证至少有 unigram
	found := false
	for _, tok := range tokens {
		if tok.Lang == "zh" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Chinese tokens")
	}
}

func TestTokenize_Mixed(t *testing.T) {
	tokens := Tokenize("Tesla's 电动车 technology")
	langs := map[string]bool{}
	for _, tok := range tokens {
		langs[tok.Lang] = true
	}
	if !langs["en"] {
		t.Error("expected English tokens in mixed text")
	}
	if !langs["zh"] {
		t.Error("expected Chinese tokens in mixed text")
	}
}

func TestFilterTokens_StopwordRemoval(t *testing.T) {
	tokens := []Token{
		{Text: "the", Lang: "en"},
		{Text: "electric", Lang: "en"},
		{Text: "car", Lang: "en"},
		{Text: "is", Lang: "en"},
		{Text: "fast", Lang: "en"},
	}
	words := FilterTokens(tokens)
	for _, w := range words {
		if w == "the" || w == "is" {
			t.Errorf("stopword %q should be filtered", w)
		}
	}
}

func TestFilterTokens_MinLength(t *testing.T) {
	tokens := []Token{
		{Text: "a", Lang: "en"},
		{Text: "ok", Lang: "en"},
	}
	words := FilterTokens(tokens)
	for _, w := range words {
		if len([]rune(w)) < 2 {
			t.Errorf("word %q is too short, should be filtered", w)
		}
	}
}
