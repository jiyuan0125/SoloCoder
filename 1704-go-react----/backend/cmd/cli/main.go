package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var baseURL string

type Patient struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Plan struct {
	ID            uint       `json:"id"`
	PatientID     uint       `json:"patientId"`
	Name          string     `json:"name"`
	AssessmentType string    `json:"assessmentType"`
	StartDate     time.Time  `json:"startDate"`
	DurationWeeks int        `json:"durationWeeks"`
	Exercises     []Exercise `json:"exercises"`
}

type Exercise struct {
	Name            string `json:"name"`
	GoalDescription string `json:"goalDescription"`
	FrequencyType   string `json:"frequencyType"`
	FrequencyCount  int    `json:"frequencyCount"`
	DurationMinutes int    `json:"durationMinutes"`
	DifficultyLevel int    `json:"difficultyLevel"`
}

type TrainingRecord struct {
	TaskID            uint   `json:"taskId"`
	ExerciseID        uint   `json:"exerciseId"`
	PlanID            uint   `json:"planId"`
	PatientID         uint   `json:"patientId"`
	ActualDuration    int    `json:"actualDuration"`
	QualityScore      int    `json:"qualityScore"`
	SubjectiveFeeling string `json:"subjectiveFeeling"`
	Notes             string `json:"notes"`
}

type Assessment struct {
	ID             uint               `json:"id"`
	CurrentScore   float64            `json:"currentScore"`
	Notes          string             `json:"notes"`
	Indicators     []AssessmentIndicator `json:"indicators"`
}

type AssessmentIndicator struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func main() {
	baseURL = os.Getenv("API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api"
	}

	var rootCmd = &cobra.Command{
		Use:   "rehab-cli",
		Short: "康复管理系统 CLI",
	}

	rootCmd.AddCommand(createPlanCmd())
	rootCmd.AddCommand(logTrainingCmd())
	rootCmd.AddCommand(recordAssessmentCmd())
	rootCmd.AddCommand(showSummaryCmd())
	rootCmd.AddCommand(listPatientsCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func listPatientsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-patients",
		Short: "列出所有患者",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := http.Get(baseURL + "/patients")
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var patients []Patient
			json.Unmarshal(body, &patients)

			fmt.Println("=== 患者列表 ===")
			for _, p := range patients {
				fmt.Printf("ID: %d, 姓名: %s\n", p.ID, p.Name)
			}
		},
	}
}

func createPlanCmd() *cobra.Command {
	var patientID uint
	var planName string
	var assessmentType string
	var durationWeeks int

	cmd := &cobra.Command{
		Use:   "create-plan",
		Short: "创建康复计划",
		Run: func(cmd *cobra.Command, args []string) {
			exercises := []Exercise{
				{
					Name:            "膝关节屈伸训练",
					GoalDescription: "恢复膝关节活动范围",
					FrequencyType:   "daily",
					FrequencyCount:  3,
					DurationMinutes: 15,
					DifficultyLevel: 2,
				},
			}

			plan := Plan{
				PatientID:      patientID,
				Name:           planName,
				AssessmentType: assessmentType,
				DurationWeeks:  durationWeeks,
				StartDate:      time.Now(),
				Exercises:      exercises,
			}

			jsonData, _ := json.Marshal(plan)
			resp, err := http.Post(baseURL+"/plans", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != 201 {
				fmt.Printf("创建失败 (%d): %s\n", resp.StatusCode, string(body))
				return
			}

			var createdPlan Plan
			json.Unmarshal(body, &createdPlan)
			fmt.Printf("计划创建成功！ID: %d\n", createdPlan.ID)
		},
	}

	cmd.Flags().UintVar(&patientID, "patient-id", 0, "患者ID")
	cmd.Flags().StringVar(&planName, "name", "康复计划", "计划名称")
	cmd.Flags().StringVar(&assessmentType, "assessment", "joint_range", "评估类型")
	cmd.Flags().IntVar(&durationWeeks, "weeks", 4, "持续周数")
	cmd.MarkFlagRequired("patient-id")

	return cmd
}

func logTrainingCmd() *cobra.Command {
	var taskID, exerciseID, planID, patientID, actualDuration, qualityScore int
	var subjectiveFeeling, notes string

	cmd := &cobra.Command{
		Use:   "log-training",
		Short: "录入训练记录",
		Run: func(cmd *cobra.Command, args []string) {
			record := TrainingRecord{
				TaskID:            uint(taskID),
				ExerciseID:        uint(exerciseID),
				PlanID:            uint(planID),
				PatientID:         uint(patientID),
				ActualDuration:    actualDuration,
				QualityScore:      qualityScore,
				SubjectiveFeeling: subjectiveFeeling,
				Notes:             notes,
			}

			jsonData, _ := json.Marshal(record)
			resp, err := http.Post(baseURL+"/training/records", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != 201 {
				fmt.Printf("录入失败 (%d): %s\n", resp.StatusCode, string(body))
				return
			}
			fmt.Println("训练记录录入成功！")
		},
	}

	cmd.Flags().IntVar(&taskID, "task-id", 0, "任务ID")
	cmd.Flags().IntVar(&exerciseID, "exercise-id", 0, "训练项目ID")
	cmd.Flags().IntVar(&planID, "plan-id", 0, "计划ID")
	cmd.Flags().IntVar(&patientID, "patient-id", 0, "患者ID")
	cmd.Flags().IntVar(&actualDuration, "duration", 0, "实际时长（分钟）")
	cmd.Flags().IntVar(&qualityScore, "score", 0, "完成质量评分（1-10）")
	cmd.Flags().StringVar(&subjectiveFeeling, "feeling", "normal", "主观感受（easy/normal/hard/very_hard）")
	cmd.Flags().StringVar(&notes, "notes", "", "备注")
	cmd.MarkFlagRequired("task-id")
	cmd.MarkFlagRequired("exercise-id")
	cmd.MarkFlagRequired("plan-id")
	cmd.MarkFlagRequired("patient-id")

	return cmd
}

func recordAssessmentCmd() *cobra.Command {
	var assessmentID int
	var score float64
	var notes string

	cmd := &cobra.Command{
		Use:   "record-assessment",
		Short: "录入评估结果",
		Run: func(cmd *cobra.Command, args []string) {
			assessment := Assessment{
				ID:           uint(assessmentID),
				CurrentScore: score,
				Notes:        notes,
				Indicators: []AssessmentIndicator{
					{Name: "综合评分", Value: score, Unit: "分"},
				},
			}

			jsonData, _ := json.Marshal(assessment)
			req, _ := http.NewRequest("POST", fmt.Sprintf("%s/assessments/%d/record", baseURL, assessmentID), bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != 200 {
				fmt.Printf("录入失败 (%d): %s\n", resp.StatusCode, string(body))
				return
			}
			fmt.Println("评估记录录入成功！")
		},
	}

	cmd.Flags().IntVar(&assessmentID, "id", 0, "评估ID")
	cmd.Flags().Float64Var(&score, "score", 0, "评分（0-100）")
	cmd.Flags().StringVar(&notes, "notes", "", "备注")
	cmd.MarkFlagRequired("id")
	cmd.MarkFlagRequired("score")

	return cmd
}

func showSummaryCmd() *cobra.Command {
	var patientID uint

	cmd := &cobra.Command{
		Use:   "show-summary",
		Short: "查看患者汇总数据",
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := http.Get(fmt.Sprintf("%s/summary/patients/%d", baseURL, patientID))
			if err != nil {
				fmt.Println("请求失败:", err)
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, body, "", "  ")
			fmt.Println(prettyJSON.String())
		},
	}

	cmd.Flags().UintVar(&patientID, "patient-id", 0, "患者ID")
	cmd.MarkFlagRequired("patient-id")

	return cmd
}
