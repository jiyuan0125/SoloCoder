package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"teaching-evaluation-system/pkg/models"
)

var DB *gorm.DB

func Init(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	err = DB.AutoMigrate(
		&models.College{},
		&models.Teacher{},
		&models.Student{},
		&models.Course{},
		&models.Question{},
		&models.EvaluationTask{},
		&models.EvaluationSubmission{},
		&models.Answer{},
		&models.CourseResult{},
		&models.FeedbackItem{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	if err := seedQuestions(); err != nil {
		return err
	}

	return nil
}

func seedQuestions() error {
	var count int64
	DB.Model(&models.Question{}).Count(&count)
	if count > 0 {
		return nil
	}

	questions := []models.Question{
		{Text: "教师认真负责，按时上下课", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryAttitude, CourseType: "", OrderIndex: 1},
		{Text: "教师尊重学生，关心学生成长", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryAttitude, CourseType: "", OrderIndex: 2},

		{Text: "课程内容丰富，符合教学大纲要求", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryContent, CourseType: "", OrderIndex: 3},
		{Text: "教学内容重点突出，难点讲透", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryContent, CourseType: "", OrderIndex: 4},
		{Text: "理论联系实际，内容实用性强", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryContent, CourseType: "", OrderIndex: 5},

		{Text: "教学方法灵活多样，易于理解", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryMethod, CourseType: "", OrderIndex: 6},
		{Text: "教师善于启发引导，互动性好", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryMethod, CourseType: "", OrderIndex: 7},

		{Text: "通过本课程学习，我的知识得到提升", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryEffect, CourseType: "", OrderIndex: 8},
		{Text: "课程考核方式合理，公平公正", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryEffect, CourseType: "", OrderIndex: 9},

		{Text: "我对本课程的总体评价", QuestionType: models.QuestionTypeGeneral, Category: models.CategoryOverall, CourseType: "", OrderIndex: 10},

		{Text: "教师课堂讲解清晰，逻辑严密", QuestionType: models.QuestionTypeSpecial, Category: models.CategoryTheory, CourseType: models.CourseTypeTheory, OrderIndex: 1},
		{Text: "理论深度适中，易于理解", QuestionType: models.QuestionTypeSpecial, Category: models.CategoryTheory, CourseType: models.CourseTypeTheory, OrderIndex: 2},
		{Text: "课堂举例恰当，有助于理解", QuestionType: models.QuestionTypeSpecial, Category: models.CategoryTheory, CourseType: models.CourseTypeTheory, OrderIndex: 3},

		{Text: "实验设计合理，目标明确", QuestionType: models.QuestionTypeSpecial, Category: models.CategoryExperiment, CourseType: models.CourseTypeExperiment, OrderIndex: 1},
		{Text: "实验指导到位，操作规范", QuestionType: models.QuestionTypeSpecial, Category: models.CategoryExperiment, CourseType: models.CourseTypeExperiment, OrderIndex: 2},
		{Text: "实验收获大，能力得到提升", QuestionType: models.QuestionTypeSpecial, Category: models.CategoryExperiment, CourseType: models.CourseTypeExperiment, OrderIndex: 3},

		{Text: "教师运动示范标准，指导专业", QuestionType: models.QuestionTypeSpecial, Category: models.CategorySport, CourseType: models.CourseTypeSport, OrderIndex: 1},
		{Text: "运动强度合理，运动量适中", QuestionType: models.QuestionTypeSpecial, Category: models.CategorySport, CourseType: models.CourseTypeSport, OrderIndex: 2},
		{Text: "安全防护到位，课堂氛围好", QuestionType: models.QuestionTypeSpecial, Category: models.CategorySport, CourseType: models.CourseTypeSport, OrderIndex: 3},
	}

	return DB.Create(&questions).Error
}
