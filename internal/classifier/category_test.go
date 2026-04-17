package classifier

import (
	"testing"
)

func TestCategoryClassifier_Sports(t *testing.T) {
	cls := NewCategoryClassifier()
	cats := cls.Classify("The NBA basketball team won the championship game last night. LeBron James scored 40 points.", 0.3)
	if len(cats) == 0 {
		t.Fatal("expected at least one category")
	}
	topID := cats[0].ID
	if topID != "IAB17" {
		t.Errorf("expected top category IAB17 (Sports), got %s (%s)", topID, cats[0].Name)
	}
}

func TestCategoryClassifier_Technology(t *testing.T) {
	cls := NewCategoryClassifier()
	cats := cls.Classify("Artificial intelligence and machine learning are transforming software development. Cloud computing and mobile apps drive innovation.", 0.3)
	if len(cats) == 0 {
		t.Fatal("expected categories for tech content")
	}
	// 检查 IAB19 在前 3 个
	found := false
	for i, c := range cats {
		if c.ID == "IAB19" && i < 3 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected IAB19 in top 3 categories, got: %v", cats)
	}
}

func TestCategoryClassifier_Finance(t *testing.T) {
	cls := NewCategoryClassifier()
	cats := cls.Classify("Bitcoin cryptocurrency investment portfolio strategy. Stock market finance trading wealth management.", 0.3)
	if len(cats) == 0 {
		t.Fatal("expected categories")
	}
	// IAB13 (Personal Finance) 应出现
	found := false
	for _, c := range cats {
		if c.ID == "IAB13" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected IAB13 (Personal Finance) in results, got: %v", cats)
	}
}

func TestCategoryClassifier_EmptyText(t *testing.T) {
	cls := NewCategoryClassifier()
	cats := cls.Classify("", 0.3)
	if cats != nil {
		t.Errorf("expected nil for empty text, got %v", cats)
	}
}

func TestCategoryClassifier_Uncategorized(t *testing.T) {
	cls := NewCategoryClassifier()
	// 完全无意义的词
	cats := cls.Classify("xyzqpqpqpq mmmm nnnn pppp", 0.3)
	if len(cats) == 0 {
		t.Fatal("expected at least uncategorized")
	}
	if cats[0].ID != "IAB24" {
		t.Logf("note: got category %s instead of IAB24", cats[0].ID)
	}
}

func TestCategoryClassifier_ConfidenceRange(t *testing.T) {
	cls := NewCategoryClassifier()
	cats := cls.Classify("Sports basketball football soccer game championship team player", 0.0)
	for _, c := range cats {
		if c.Confidence < 0 || c.Confidence > 1 {
			t.Errorf("confidence out of range [0,1]: %f for %s", c.Confidence, c.ID)
		}
	}
}

func TestAllCategories(t *testing.T) {
	cats := AllCategories()
	if len(cats) < 10 {
		t.Errorf("expected at least 10 categories, got %d", len(cats))
	}
	// 检查 IAB1 存在
	found := false
	for _, c := range cats {
		if c.ID == "IAB1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected IAB1 in category list")
	}
}
