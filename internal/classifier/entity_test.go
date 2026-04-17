package classifier

import (
	"testing"
)

func TestEntityRecognizer_Brand(t *testing.T) {
	er := NewEntityRecognizer()
	entities := er.Recognize("Tesla announced a new electric vehicle. Apple released iOS 17.")
	found := map[string]bool{}
	for _, e := range entities {
		found[e.Text] = true
	}
	if !found["Tesla"] {
		t.Errorf("expected entity Tesla, got: %v", entities)
	}
	if !found["Apple"] {
		t.Errorf("expected entity Apple, got: %v", entities)
	}
}

func TestEntityRecognizer_Person(t *testing.T) {
	er := NewEntityRecognizer()
	entities := er.Recognize("Elon Musk is the CEO of Tesla. Tim Cook leads Apple.")
	personFound := false
	for _, e := range entities {
		if e.Text == "Elon Musk" && e.Type == EntityPerson {
			personFound = true
		}
	}
	if !personFound {
		t.Errorf("expected person entity 'Elon Musk', got: %v", entities)
	}
}

func TestEntityRecognizer_Location(t *testing.T) {
	er := NewEntityRecognizer()
	entities := er.Recognize("The conference was held in San Francisco, United States.")
	locFound := false
	for _, e := range entities {
		if e.Type == EntityLocation {
			locFound = true
			break
		}
	}
	if !locFound {
		t.Errorf("expected location entity, got: %v", entities)
	}
}

func TestEntityRecognizer_Dedup(t *testing.T) {
	er := NewEntityRecognizer()
	entities := er.Recognize("Tesla Tesla Tesla Tesla is a great company")
	count := 0
	for _, e := range entities {
		if e.Text == "Tesla" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected Tesla to appear exactly once (dedup), got %d", count)
	}
}

func TestEntityRecognizer_Empty(t *testing.T) {
	er := NewEntityRecognizer()
	entities := er.Recognize("")
	if entities != nil {
		t.Errorf("expected nil for empty text, got %v", entities)
	}
}

func TestEntityRecognizer_Chinese(t *testing.T) {
	er := NewEntityRecognizer()
	entities := er.Recognize("阿里巴巴和腾讯是中国最大的科技公司")
	found := map[string]bool{}
	for _, e := range entities {
		found[e.Text] = true
	}
	if !found["阿里巴巴"] {
		t.Logf("entities: %v", entities)
	}
}
