package main

import (
	"log"

	"smart-exam/config"
	"smart-exam/database"
	"smart-exam/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database.Init()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	knowledgeHandler := handlers.NewKnowledgeHandler()
	questionHandler := handlers.NewQuestionHandler()
	examHandler := handlers.NewExamHandler()
	reportHandler := handlers.NewReportHandler()
	wrongAnswerHandler := handlers.NewWrongAnswerHandler()
	billHandler := handlers.NewBillHandler()

	api := r.Group("/api")
	{
		knowledge := api.Group("/knowledge")
		{
			knowledge.GET("", knowledgeHandler.List)
			knowledge.POST("", knowledgeHandler.Create)
			knowledge.GET("/:id", knowledgeHandler.Get)
			knowledge.PUT("/:id", knowledgeHandler.Update)
			knowledge.DELETE("/:id", knowledgeHandler.Delete)
		}

		questions := api.Group("/questions")
		{
			questions.GET("", questionHandler.List)
			questions.POST("", questionHandler.Create)
			questions.GET("/:id", questionHandler.Get)
			questions.PUT("/:id", questionHandler.Update)
			questions.DELETE("/:id", questionHandler.Delete)
			questions.POST("/:id/submit", questionHandler.SubmitForReview)
			questions.POST("/:id/approve", questionHandler.Approve)
			questions.POST("/:id/publish", questionHandler.Publish)
			questions.POST("/:id/offline", questionHandler.Offline)
			questions.POST("/:id/reject", questionHandler.Reject)
		}

		exams := api.Group("/exams")
		{
			exams.POST("/start", examHandler.StartExam)
			exams.GET("/:exam_id/current", examHandler.GetCurrentQuestion)
			exams.POST("/:exam_id/submit", examHandler.SubmitAnswer)
			exams.POST("/:exam_id/incomplete", examHandler.MarkExamIncomplete)
			exams.GET("/:exam_id/report", examHandler.GetExamReport)
		}

		reports := api.Group("/reports")
		{
			reports.GET("/student", reportHandler.GetStudentReport)
			reports.GET("/history", reportHandler.GetExamHistory)
		}

		wrongAnswers := api.Group("/wrong-answers")
		{
			wrongAnswers.GET("", wrongAnswerHandler.List)
			wrongAnswers.POST("", wrongAnswerHandler.Add)
			wrongAnswers.PUT("/:id", wrongAnswerHandler.Update)
			wrongAnswers.DELETE("/:id", wrongAnswerHandler.Delete)
		}

		bills := api.Group("/bills")
		{
			bills.POST("", billHandler.Create)
			bills.GET("", billHandler.List)
			bills.GET("/:id", billHandler.Get)
			bills.POST("/:id/adjust", billHandler.AdjustTotal)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
