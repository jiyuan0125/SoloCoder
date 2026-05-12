package services

import (
	"errors"
	"time"

	"teaching-evaluation-system/pkg/database"
	"teaching-evaluation-system/pkg/models"
	"teaching-evaluation-system/pkg/utils"

	"gorm.io/gorm"
)

type EvaluationRequest struct {
	StudentID uint            `json:"student_id" binding:"required"`
	CourseID  uint            `json:"course_id" binding:"required"`
	TaskID    uint            `json:"task_id" binding:"required"`
	Answers   []AnswerRequest `json:"answers" binding:"required"`
	Comment   string          `json:"comment"`
}

type AnswerRequest struct {
	QuestionID uint `json:"question_id" binding:"required"`
	Score      int  `json:"score" binding:"required"`
}

type CreateTaskRequest struct {
	Semester    string    `json:"semester" binding:"required"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	CourseIDs   []uint    `json:"course_ids" binding:"required"`
}

func GetQuestionsForCourse(courseType models.CourseType) ([]models.Question, error) {
	var questions []models.Question
	err := database.DB.Where("question_type = ? OR course_type = ?", models.QuestionTypeGeneral, courseType).
		Order("course_type ASC, order_index ASC").
		Find(&questions).Error
	return questions, err
}

func GetEvaluationTask(taskID uint) (*models.EvaluationTask, error) {
	var task models.EvaluationTask
	err := database.DB.Preload("Courses").First(&task, taskID).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func GetAllTasks() ([]models.EvaluationTask, error) {
	var tasks []models.EvaluationTask
	err := database.DB.Preload("Courses").Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func CreateTask(req CreateTaskRequest) (*models.EvaluationTask, error) {
	var existingTask models.EvaluationTask
	err := database.DB.Where("semester = ?", req.Semester).First(&existingTask).Error
	if err == nil {
		return nil, errors.New("该学期评估任务已存在")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("结束日期不能早于开始日期")
	}

	var courses []models.Course
	if err := database.DB.Find(&courses, req.CourseIDs).Error; err != nil {
		return nil, err
	}

	if len(courses) != len(req.CourseIDs) {
		return nil, errors.New("存在无效的课程ID")
	}

	task := models.EvaluationTask{
		Semester:  req.Semester,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Status:    models.TaskStatusDraft,
		Courses:   courses,
	}

	if err := database.DB.Create(&task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

func TaskAction(taskID uint, action string, userRole string) error {
	var task models.EvaluationTask
	if err := database.DB.First(&task, taskID).Error; err != nil {
		return err
	}

	switch action {
	case "approve":
		switch task.Status {
		case models.TaskStatusDraft:
			if userRole != "admin" && userRole != "reviewer" {
				return errors.New("无权限")
			}
			task.Status = models.TaskStatusReview
		case models.TaskStatusReview:
			if userRole != "admin" && userRole != "final_reviewer" {
				return errors.New("无权限")
			}
			task.Status = models.TaskStatusFinal
		case models.TaskStatusFinal:
			if userRole != "admin" && userRole != "signer" {
				return errors.New("无权限")
			}
			task.Status = models.TaskStatusActive
		}
	case "reject":
		if task.Status == models.TaskStatusFinal {
			task.Status = models.TaskStatusReview
		} else {
			return errors.New("当前状态不可拒绝")
		}
	case "cancel":
		if task.Status == models.TaskStatusCompleted || task.Status == models.TaskStatusActive {
			return errors.New("当前状态不可取消")
		}
		task.Status = models.TaskStatusCancelled
	default:
		return errors.New("无效的操作")
	}

	return database.DB.Save(&task).Error
}

func SubmitEvaluation(req EvaluationRequest) (*models.EvaluationSubmission, error) {
	var task models.EvaluationTask
	if err := database.DB.First(&task, req.TaskID).Error; err != nil {
		return nil, errors.New("评估任务不存在")
	}

	now := time.Now()
	if now.Before(task.StartDate) {
		return nil, errors.New("评估尚未开始")
	}
	if now.After(task.EndDate) {
		return nil, errors.New("评估已结束")
	}
	if task.Status != models.TaskStatusActive {
		return nil, errors.New("评估任务未激活")
	}

	var existing models.EvaluationSubmission
	err := database.DB.Where("student_id = ? AND course_id = ? AND task_id = ?", req.StudentID, req.CourseID, req.TaskID).First(&existing).Error
	if err == nil {
		return nil, errors.New("您已对该课程完成评估")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var course models.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		return nil, errors.New("课程不存在")
	}

	questions, err := GetQuestionsForCourse(course.CourseType)
	if err != nil {
		return nil, err
	}

	if len(req.Answers) != len(questions) {
		return nil, errors.New("答案数量不匹配")
	}

	for _, answer := range req.Answers {
		if answer.Score < 1 || answer.Score > 5 {
			return nil, errors.New("评分必须在1到5之间")
		}
	}

	filteredComment := utils.FilterSensitiveWords(utils.TruncateComment(req.Comment, 500))

	submission := models.EvaluationSubmission{
		StudentID:       req.StudentID,
		CourseID:        req.CourseID,
		TaskID:          req.TaskID,
		SubmittedAt:     now,
		Comment:         req.Comment,
		FilteredComment: filteredComment,
	}

	tx := database.DB.Begin()
	if err := tx.Create(&submission).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, answer := range req.Answers {
		dbAnswer := models.Answer{
			SubmissionID: submission.ID,
			QuestionID:   answer.QuestionID,
			Score:        answer.Score,
		}
		if err := tx.Create(&dbAnswer).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &submission, nil
}
