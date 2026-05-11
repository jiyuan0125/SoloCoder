package bayes

import (
	"os"
	"testing"
)

func TestClassifier(t *testing.T) {
	classifier := NewClassifier(FeatureTypeBagOfWords)

	samples := []TrainingSample{
		{Text: "免费 优惠 活动 点击 领取", Label: "spam"},
		{Text: "恭喜 获得 免费 礼品", Label: "spam"},
		{Text: "工作 会议 安排 明天", Label: "ham"},
		{Text: "项目 进度 汇报 附件", Label: "ham"},
	}

	err := classifier.Train(samples)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	predictions, bestLabel, err := classifier.Predict("免费 领取 优惠")
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if bestLabel != "spam" {
		t.Errorf("Expected best label 'spam', got '%s'", bestLabel)
	}

	if len(predictions) != 2 {
		t.Errorf("Expected 2 predictions, got %d", len(predictions))
	}
}

func TestTFIDFClassifier(t *testing.T) {
	classifier := NewClassifier(FeatureTypeTFIDF)

	samples := []TrainingSample{
		{Text: "免费 优惠 活动 点击 领取", Label: "spam"},
		{Text: "恭喜 获得 免费 礼品", Label: "spam"},
		{Text: "促销 折扣 限时 免费", Label: "spam"},
		{Text: "工作 会议 安排 明天", Label: "ham"},
		{Text: "项目 进度 汇报 附件", Label: "ham"},
		{Text: "团队 建设 活动 安排", Label: "ham"},
	}

	err := classifier.Train(samples)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	predictions, bestLabel, err := classifier.Predict("工作 会议 安排 项目 进度")
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if bestLabel != "ham" {
		t.Errorf("Expected best label 'ham', got '%s'", bestLabel)
	}

	if len(predictions) != 2 {
		t.Errorf("Expected 2 predictions, got %d", len(predictions))
	}
}

func TestEmptyTrainingData(t *testing.T) {
	classifier := NewClassifier(FeatureTypeBagOfWords)

	err := classifier.Train([]TrainingSample{})
	if err == nil {
		t.Error("Expected error for empty training data")
	}

	_, _, err = classifier.Predict("test text")
	if err == nil {
		t.Error("Expected error when predicting without training")
	}
}

func TestModelPersistence(t *testing.T) {
	classifier1 := NewClassifier(FeatureTypeBagOfWords)

	samples := []TrainingSample{
		{Text: "免费 优惠 活动", Label: "spam"},
		{Text: "工作 会议 安排", Label: "ham"},
	}

	err := classifier1.Train(samples)
	if err != nil {
		t.Fatalf("Training failed: %v", err)
	}

	testFile := "test_model.json"
	defer os.Remove(testFile)

	err = classifier1.Save(testFile)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	classifier2 := NewClassifier(FeatureTypeBagOfWords)
	err = classifier2.Load(testFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	predictions, bestLabel, err := classifier2.Predict("免费 优惠")
	if err != nil {
		t.Fatalf("Prediction after load failed: %v", err)
	}

	if bestLabel != "spam" {
		t.Errorf("Expected best label 'spam', got '%s'", bestLabel)
	}

	if len(predictions) != 2 {
		t.Errorf("Expected 2 predictions, got %d", len(predictions))
	}
}

func TestTokenizer(t *testing.T) {
	tokenizer := NewTokenizer()

	text := "Hello World! This is a test email: test@example.com and URL: https://example.com"
	tokens := tokenizer.Tokenize(text)

	hasEmail := false
	hasURL := false
	hasHello := false

	for _, token := range tokens {
		if token == "test@example.com" {
			hasEmail = true
		}
		if token == "https://example.com" {
			hasURL = true
		}
		if token == "hello" {
			hasHello = true
		}
	}

	if !hasEmail {
		t.Error("Expected email to be preserved")
	}
	if !hasURL {
		t.Error("Expected URL to be preserved")
	}
	if !hasHello {
		t.Error("Expected 'hello' in lowercase")
	}
}
