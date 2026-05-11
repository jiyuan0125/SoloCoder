package bayes

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
)

type FeatureType int

const (
	FeatureTypeBagOfWords FeatureType = iota
	FeatureTypeTFIDF
)

type Classifier struct {
	mu               sync.RWMutex
	classCounts      map[string]int
	wordCounts       map[string]map[string]int
	totalWords       map[string]int
	docWordCounts    map[string]int
	uniqueWords      map[string]struct{}
	totalDocs        int
	featureType      FeatureType
	tokenizer        *Tokenizer
	isTrained        bool
}

type ModelData struct {
	ClassCounts   map[string]int            `json:"class_counts"`
	WordCounts    map[string]map[string]int `json:"word_counts"`
	TotalWords    map[string]int            `json:"total_words"`
	DocWordCounts map[string]int            `json:"doc_word_counts"`
	UniqueWords   []string                  `json:"unique_words"`
	TotalDocs     int                       `json:"total_docs"`
	FeatureType   string                    `json:"feature_type"`
	IsTrained     bool                      `json:"is_trained"`
}

type Prediction struct {
	Label    string
	LogProb  float64
	Prob     float64
}

type TrainingSample struct {
	Text  string
	Label string
}

func NewClassifier(featureType FeatureType) *Classifier {
	return &Classifier{
		classCounts:      make(map[string]int),
		wordCounts:       make(map[string]map[string]int),
		totalWords:       make(map[string]int),
		docWordCounts:    make(map[string]int),
		uniqueWords:      make(map[string]struct{}),
		featureType:      featureType,
		tokenizer:        NewTokenizer(),
		isTrained:        false,
	}
}

func (c *Classifier) Train(samples []TrainingSample) error {
	if len(samples) == 0 {
		return errors.New("training data is empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, sample := range samples {
		if sample.Label == "" {
			continue
		}

		tokens := c.tokenizer.TokenizeChinese(sample.Text)
		
		c.classCounts[sample.Label]++
		c.totalDocs++

		if _, ok := c.wordCounts[sample.Label]; !ok {
			c.wordCounts[sample.Label] = make(map[string]int)
		}

		wordInDoc := make(map[string]struct{})
		for _, token := range tokens {
			c.wordCounts[sample.Label][token]++
			c.totalWords[sample.Label]++
			c.uniqueWords[token] = struct{}{}
			wordInDoc[token] = struct{}{}
		}

		for word := range wordInDoc {
			c.docWordCounts[word]++
		}
	}

	c.isTrained = true
	c.checkEmptyClasses()

	return nil
}

func (c *Classifier) checkEmptyClasses() {
	for label, count := range c.classCounts {
		if count == 0 {
			log.Printf("Warning: class '%s' has no training samples", label)
		}
	}
}

func (c *Classifier) Predict(text string) ([]Prediction, string, error) {
	if !c.isTrained {
		return nil, "", errors.New("classifier not trained yet")
	}

	if c.totalDocs == 0 {
		return nil, "", errors.New("no training data available")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	tokens := c.tokenizer.TokenizeChinese(text)

	vocabSize := len(c.uniqueWords)
	alpha := 1.0

	var predictions []Prediction

	for label, classCount := range c.classCounts {
		if classCount == 0 {
			continue
		}

		priorProb := float64(classCount) / float64(c.totalDocs)
		logProb := math.Log(priorProb)

		for _, word := range tokens {
			var weight float64
			if c.featureType == FeatureTypeTFIDF {
				weight = c.calculateTFIDF(word, tokens, label, vocabSize, alpha)
			} else {
				weight = c.calculateProbability(word, label, vocabSize, alpha)
			}
			
			if weight > 0 {
				logProb += math.Log(weight)
			}
		}

		predictions = append(predictions, Prediction{
			Label:   label,
			LogProb: logProb,
		})
	}

	if len(predictions) == 0 {
		return nil, "", errors.New("no valid classes available for prediction")
	}

	maxLogProb := predictions[0].LogProb
	for _, p := range predictions {
		if p.LogProb > maxLogProb {
			maxLogProb = p.LogProb
		}
	}

	var sumExp float64
	for _, p := range predictions {
		sumExp += math.Exp(p.LogProb - maxLogProb)
	}

	for i := range predictions {
		predictions[i].Prob = math.Exp(predictions[i].LogProb - maxLogProb) / sumExp
	}

	sortPredictions(predictions)

	bestLabel := predictions[0].Label

	return predictions, bestLabel, nil
}

func (c *Classifier) calculateProbability(word, label string, vocabSize int, alpha float64) float64 {
	wordCount := 0
	if counts, ok := c.wordCounts[label]; ok {
		wordCount = counts[word]
	}
	
	totalWords := c.totalWords[label]
	
	return (float64(wordCount) + alpha) / (float64(totalWords) + alpha*float64(vocabSize))
}

func (c *Classifier) calculateTFIDF(word string, tokens []string, label string, vocabSize int, alpha float64) float64 {
	wordCount := 0
	if counts, ok := c.wordCounts[label]; ok {
		wordCount = counts[word]
	}

	totalWords := c.totalWords[label]

	baseProb := (float64(wordCount) + alpha) / (float64(totalWords) + alpha*float64(vocabSize))

	docCount := c.docWordCounts[word]
	idf := math.Log(float64(c.totalDocs+1) / float64(docCount+1))

	tfInClass := 0.0
	if counts, ok := c.wordCounts[label]; ok {
		tfInClass = float64(counts[word])
	}

	tfidfWeight := tfInClass * idf
	if tfidfWeight <= 0 {
		tfidfWeight = alpha
	}

	return baseProb * (1 + math.Log(1+tfidfWeight))
}

func sortPredictions(predictions []Prediction) {
	for i := 0; i < len(predictions)-1; i++ {
		for j := i + 1; j < len(predictions); j++ {
			if predictions[j].LogProb > predictions[i].LogProb {
				predictions[i], predictions[j] = predictions[j], predictions[i]
			}
		}
	}
}

func (c *Classifier) Save(filename string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	uniqueWords := make([]string, 0, len(c.uniqueWords))
	for word := range c.uniqueWords {
		uniqueWords = append(uniqueWords, word)
	}

	featureTypeStr := "bag_of_words"
	if c.featureType == FeatureTypeTFIDF {
		featureTypeStr = "tfidf"
	}

	data := ModelData{
		ClassCounts:   c.classCounts,
		WordCounts:    c.wordCounts,
		TotalWords:    c.totalWords,
		DocWordCounts: c.docWordCounts,
		UniqueWords:   uniqueWords,
		TotalDocs:     c.totalDocs,
		FeatureType:   featureTypeStr,
		IsTrained:     c.isTrained,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to save model to file: %w", err)
	}

	return nil
}

func (c *Classifier) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read model file: %w", err)
	}

	var modelData ModelData
	if err := json.Unmarshal(data, &modelData); err != nil {
		return fmt.Errorf("failed to unmarshal model: %w", err)
	}

	if err := c.validateModelData(&modelData); err != nil {
		return fmt.Errorf("model data validation failed: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.classCounts = modelData.ClassCounts
	c.wordCounts = modelData.WordCounts
	c.totalWords = modelData.TotalWords
	c.docWordCounts = modelData.DocWordCounts
	c.totalDocs = modelData.TotalDocs
	c.isTrained = modelData.IsTrained

	c.uniqueWords = make(map[string]struct{})
	for _, word := range modelData.UniqueWords {
		c.uniqueWords[word] = struct{}{}
	}

	if modelData.FeatureType == "tfidf" {
		c.featureType = FeatureTypeTFIDF
	} else {
		c.featureType = FeatureTypeBagOfWords
	}

	return nil
}

func (c *Classifier) validateModelData(data *ModelData) error {
	if data.TotalDocs < 0 {
		return errors.New("invalid total_docs: must be non-negative")
	}

	if len(data.ClassCounts) == 0 && data.TotalDocs > 0 {
		return errors.New("class_counts is empty but total_docs > 0")
	}

	if data.IsTrained && data.TotalDocs == 0 {
		return errors.New("is_trained is true but no training data")
	}

	for label, count := range data.ClassCounts {
		if count < 0 {
			return fmt.Errorf("invalid count for class '%s': must be non-negative", label)
		}
	}

	return nil
}
