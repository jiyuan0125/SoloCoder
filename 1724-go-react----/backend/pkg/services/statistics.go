package services

import (
	"sort"
	"time"

	"teaching-evaluation-system/pkg/database"
	"teaching-evaluation-system/pkg/models"
	"teaching-evaluation-system/pkg/utils"
)

type QuestionScores struct {
	Attitude  []int `json:"attitude"`
	Content   []int `json:"content"`
	Method    []int `json:"method"`
	Effect    []int `json:"effect"`
	Special   []int `json:"special"`
	General   []int `json:"general"`
}

func CalculateStatistics(taskID uint) ([]models.CourseResult, error) {
	var task models.EvaluationTask
	if err := database.DB.Preload("Courses").First(&task, taskID).Error; err != nil {
		return nil, err
	}

	var results []models.CourseResult

	for _, course := range task.Courses {
		result, err := calculateCourseResult(taskID, course)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}

	results = assignRanks(results)

	tx := database.DB.Begin()
	for _, result := range results {
		if err := tx.Save(&result).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if result.IsLowScore {
			if err := createFeedbackItem(&result); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	task.Status = models.TaskStatusCompleted
	database.DB.Save(&task)

	return results, nil
}

func calculateCourseResult(taskID uint, course models.Course) (*models.CourseResult, error) {
	var submissions []models.EvaluationSubmission
	err := database.DB.Where("task_id = ? AND course_id = ?", taskID, course.ID).Find(&submissions).Error
	if err != nil {
		return nil, err
	}

	totalStudents := 50
	submittedCount := len(submissions)
	submissionRate := float64(submittedCount) / float64(totalStudents) * 100

	var dataQuality models.DataQuality
	switch {
	case submissionRate < 50:
		dataQuality = models.DataQualityInsufficient
	case submissionRate >= 50 && submissionRate < 70:
		dataQuality = models.DataQualityReference
	default:
		dataQuality = models.DataQualityValid
	}

	scores, err := collectQuestionScores(submissions, course.CourseType)
	if err != nil {
		return nil, err
	}

	attitudeAvg := calculateAverage(scores.Attitude)
	contentAvg := calculateAverage(scores.Content)
	methodAvg := calculateAverage(scores.Method)
	effectAvg := calculateAverage(scores.Effect)
	specialAvg := calculateAverage(scores.Special)
	generalAvg := calculateAverage(scores.General)

	overallScore := utils.RoundToTwoDecimals(generalAvg*0.6 + specialAvg*0.4)

	result := &models.CourseResult{
		TaskID:         taskID,
		CourseID:       course.ID,
		CollegeID:      course.CollegeID,
		TeacherID:      course.TeacherID,
		CourseType:     course.CourseType,
		TotalStudents:  totalStudents,
		SubmittedCount: submittedCount,
		SubmissionRate: utils.RoundToTwoDecimals(submissionRate),
		DataQuality:    dataQuality,
		AttitudeScore:  attitudeAvg,
		ContentScore:   contentAvg,
		MethodScore:    methodAvg,
		EffectScore:    effectAvg,
		SpecialScore:   specialAvg,
		GeneralAverage: generalAvg,
		OverallScore:   overallScore,
		IsLowScore:     overallScore < 3.5 && dataQuality != models.DataQualityInsufficient,
	}

	return result, nil
}

func collectQuestionScores(submissions []models.EvaluationSubmission, courseType models.CourseType) (*QuestionScores, error) {
	scores := &QuestionScores{}

	if len(submissions) == 0 {
		return scores, nil
	}

	var submissionIDs []uint
	for _, s := range submissions {
		submissionIDs = append(submissionIDs, s.ID)
	}

	var answers []models.Answer
	if err := database.DB.Preload("Question").Where("submission_id IN ?", submissionIDs).Find(&answers).Error; err != nil {
		return nil, err
	}

	for _, answer := range answers {
		switch answer.Question.Category {
		case models.CategoryAttitude:
			scores.Attitude = append(scores.Attitude, answer.Score)
			scores.General = append(scores.General, answer.Score)
		case models.CategoryContent:
			scores.Content = append(scores.Content, answer.Score)
			scores.General = append(scores.General, answer.Score)
		case models.CategoryMethod:
			scores.Method = append(scores.Method, answer.Score)
			scores.General = append(scores.General, answer.Score)
		case models.CategoryEffect:
			scores.Effect = append(scores.Effect, answer.Score)
			scores.General = append(scores.General, answer.Score)
		case models.CategoryOverall:
			scores.General = append(scores.General, answer.Score)
		case models.CategoryTheory, models.CategoryExperiment, models.CategorySport:
			scores.Special = append(scores.Special, answer.Score)
		}
	}

	return scores, nil
}

func calculateAverage(scores []int) float64 {
	if len(scores) == 0 {
		return 0
	}
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return utils.RoundToTwoDecimals(float64(sum) / float64(len(scores)))
}

func assignRanks(results []models.CourseResult) []models.CourseResult {
	validResults := make([]models.CourseResult, 0)
	for _, r := range results {
		if r.DataQuality != models.DataQualityInsufficient {
			validResults = append(validResults, r)
		}
	}

	sort.Slice(validResults, func(i, j int) bool {
		return validResults[i].OverallScore > validResults[j].OverallScore
	})

	collegeGroups := make(map[uint][]int)
	teacherGroups := make(map[uint][]int)
	typeGroups := make(map[models.CourseType][]int)

	for idx, r := range validResults {
		collegeGroups[r.CollegeID] = append(collegeGroups[r.CollegeID], idx)
		teacherGroups[r.TeacherID] = append(teacherGroups[r.TeacherID], idx)
		typeGroups[r.CourseType] = append(typeGroups[r.CourseType], idx)
	}

	for i := range validResults {
		validResults[i].CollegeRank = getRankInGroup(collegeGroups[validResults[i].CollegeID], i)
		validResults[i].TeacherRank = getRankInGroup(teacherGroups[validResults[i].TeacherID], i)
		validResults[i].TypeRank = getRankInGroup(typeGroups[validResults[i].CourseType], i)
	}

	return validResults
}

func getRankInGroup(groupIndices []int, currentIdx int) int {
	for rank, idx := range groupIndices {
		if idx == currentIdx {
			return rank + 1
		}
	}
	return -1
}

func createFeedbackItem(result *models.CourseResult) error {
	feedback := models.FeedbackItem{
		CourseResultID: result.ID,
		TeacherID:      result.TeacherID,
		Status:         models.FeedbackStatusPending,
		PlanDeadline:   time.Now().AddDate(0, 0, 30),
		NextReviewAt:   time.Now().AddDate(1, 0, 0),
	}
	return database.DB.Create(&feedback).Error
}

func GetCourseResults(taskID uint) ([]models.CourseResult, error) {
	var results []models.CourseResult
	err := database.DB.Where("task_id = ?", taskID).Order("overall_score DESC").Find(&results).Error
	return results, err
}

func GetComments(courseResultID uint) ([]string, error) {
	var result models.CourseResult
	if err := database.DB.First(&result, courseResultID).Error; err != nil {
		return nil, err
	}

	var submissions []models.EvaluationSubmission
	err := database.DB.Where("task_id = ? AND course_id = ?", result.TaskID, result.CourseID).
		Where("filtered_comment != ''").
		Select("filtered_comment").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}

	var comments []string
	for _, s := range submissions {
		comments = append(comments, s.FilteredComment)
	}
	return comments, nil
}
