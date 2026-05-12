package services

import (
	"errors"
	"math/rand"
	"time"

	"smart-exam/database"
	"smart-exam/models"

	"gorm.io/gorm"
)

const (
	MinDifficulty     = 1
	MaxDifficulty     = 5
	InitialDifficulty = 3
	TotalQuestions    = 20
	MinPoolSize       = 3
)

type AdaptiveService struct {
	db *gorm.DB
}

func NewAdaptiveService() *AdaptiveService {
	return &AdaptiveService{db: database.DB}
}

func (s *AdaptiveService) GetNextDifficulty(
	currentDiff int,
	consecutiveCorrect int,
	consecutiveWrong int,
	lastAnswerCorrect bool,
) int {
	newDiff := currentDiff
	newConsecutiveCorrect := consecutiveCorrect
	newConsecutiveWrong := consecutiveWrong

	if lastAnswerCorrect {
		newConsecutiveCorrect++
		newConsecutiveWrong = 0
		if newConsecutiveCorrect >= 2 {
			if currentDiff < MaxDifficulty {
				newDiff = currentDiff + 1
			}
			newConsecutiveCorrect = 0
		}
	} else {
		newConsecutiveWrong++
		newConsecutiveCorrect = 0
		if newConsecutiveWrong >= 2 {
			if currentDiff > MinDifficulty {
				newDiff = currentDiff - 1
			}
			newConsecutiveWrong = 0
		}
	}

	if newDiff < MinDifficulty {
		newDiff = MinDifficulty
	}
	if newDiff > MaxDifficulty {
		newDiff = MaxDifficulty
	}

	return newDiff
}

func (s *AdaptiveService) GetQuestionPool(difficulty int, usedQuestionIDs []uint) ([]models.Question, error) {
	var pool []models.Question
	query := s.db.Where("status = ? AND difficulty = ?", models.QuestionStatusPublished, difficulty)

	if len(usedQuestionIDs) > 0 {
		query = query.Where("id NOT IN ?", usedQuestionIDs)
	}

	err := query.Find(&pool).Error
	if err != nil {
		return nil, err
	}

	if len(pool) >= MinPoolSize {
		return pool, nil
	}

	return s.expandPool(difficulty, usedQuestionIDs)
}

func (s *AdaptiveService) expandPool(targetDiff int, usedQuestionIDs []uint) ([]models.Question, error) {
	var pool []models.Question
	distances := []int{0, 1, 2, 3, 4}

	for _, dist := range distances {
		difficulties := []int{}
		if dist == 0 {
			difficulties = append(difficulties, targetDiff)
		} else {
			if targetDiff-dist >= MinDifficulty {
				difficulties = append(difficulties, targetDiff-dist)
			}
			if targetDiff+dist <= MaxDifficulty {
				difficulties = append(difficulties, targetDiff+dist)
			}
		}

		for _, diff := range difficulties {
			var questions []models.Question
			query := s.db.Where("status = ? AND difficulty = ?", models.QuestionStatusPublished, diff)

			if len(usedQuestionIDs) > 0 {
				query = query.Where("id NOT IN ?", usedQuestionIDs)
			}

			err := query.Find(&questions).Error
			if err != nil {
				return nil, err
			}

			pool = append(pool, questions...)

			if len(pool) >= MinPoolSize {
				return pool, nil
			}
		}
	}

	if len(pool) == 0 {
		return nil, errors.New("no available questions in question pool")
	}

	return pool, nil
}

func (s *AdaptiveService) PickRandomQuestion(pool []models.Question) (*models.Question, error) {
	if len(pool) == 0 {
		return nil, errors.New("empty question pool")
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(pool))
	return &pool[index], nil
}

func (s *AdaptiveService) IsObjectiveQuestion(qType models.QuestionType) bool {
	return qType == models.QuestionTypeSingleChoice ||
		qType == models.QuestionTypeMultipleChoice ||
		qType == models.QuestionTypeTrueFalse ||
		qType == models.QuestionTypeFillBlank
}

func (s *AdaptiveService) CheckAnswer(question *models.Question, studentAnswer string) (bool, error) {
	if !s.IsObjectiveQuestion(question.Type) {
		return false, nil
	}

	return studentAnswer == question.CorrectAnswer, nil
}

func (s *AdaptiveService) CalculateMasteryLevel(correct, total int) models.MasteryLevel {
	if total == 0 {
		return models.MasteryLevelWeak
	}

	accuracy := float64(correct) / float64(total) * 100

	switch {
	case accuracy >= 80:
		return models.MasteryLevelMastered
	case accuracy >= 50:
		return models.MasteryLevelReinforce
	default:
		return models.MasteryLevelWeak
	}
}

func (s *AdaptiveService) GetKnowledgePointAncestors(kpID uint) ([]uint, error) {
	var ancestors []uint
	var currentID *uint = &kpID

	for currentID != nil {
		var kp models.KnowledgePoint
		if err := s.db.First(&kp, *currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, err
		}
		ancestors = append([]uint{kp.ID}, ancestors...)
		currentID = kp.ParentID
	}

	return ancestors, nil
}
