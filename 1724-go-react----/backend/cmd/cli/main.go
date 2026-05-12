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

var baseURL = "http://localhost:8300"

type CreateTaskRequest struct {
	Semester  string   `json:"semester"`
	StartDate string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	CourseIDs []uint   `json:"course_ids"`
}

type SubmitEvaluationRequest struct {
	StudentID uint   `json:"student_id"`
	CourseID  uint   `json:"course_id"`
	TaskID    uint   `json:"task_id"`
	Answers   []struct {
		QuestionID uint `json:"question_id"`
		Score      int  `json:"score"`
	} `json:"answers"`
	Comment string `json:"comment"`
}

func main() {
	if url := os.Getenv("API_URL"); url != "" {
		baseURL = url
	}

	var rootCmd = &cobra.Command{
		Use:   "te-client",
		Short: "教学评估系统CLI客户端",
	}

	var createTaskCmd = &cobra.Command{
		Use:   "create-task",
		Short: "创建评估任务",
		Run:   createTask,
	}

	var taskSemester, taskStart, taskEnd string
	var courseIDs []uint
	createTaskCmd.Flags().StringVar(&taskSemester, "semester", "", "学期标识 (如 2024-2025-1)")
	createTaskCmd.Flags().StringVar(&taskStart, "start", "", "开始日期 (YYYY-MM-DD)")
	createTaskCmd.Flags().StringVar(&taskEnd, "end", "", "结束日期 (YYYY-MM-DD)")
	createTaskCmd.Flags().UintSliceVar(&courseIDs, "courses", []uint{}, "课程ID列表")
	createTaskCmd.MarkFlagRequired("semester")
	createTaskCmd.MarkFlagRequired("start")
	createTaskCmd.MarkFlagRequired("end")
	createTaskCmd.MarkFlagRequired("courses")

	var submitEvalCmd = &cobra.Command{
		Use:   "submit-evaluation",
		Short: "提交评估",
		Run:   submitEvaluation,
	}

	var studentID, courseID, taskID uint
	var comment string
	submitEvalCmd.Flags().UintVar(&studentID, "student", 0, "学生ID")
	submitEvalCmd.Flags().UintVar(&courseID, "course", 0, "课程ID")
	submitEvalCmd.Flags().UintVar(&taskID, "task", 0, "任务ID")
	submitEvalCmd.Flags().StringVar(&comment, "comment", "", "文字评语")
	submitEvalCmd.MarkFlagRequired("student")
	submitEvalCmd.MarkFlagRequired("course")
	submitEvalCmd.MarkFlagRequired("task")

	var showResultsCmd = &cobra.Command{
		Use:   "show-results",
		Short: "查看结果统计",
		Run:   showResults,
	}

	var resultsTaskID uint
	showResultsCmd.Flags().UintVar(&resultsTaskID, "task", 0, "任务ID")
	showResultsCmd.MarkFlagRequired("task")

	var exportCmd = &cobra.Command{
		Use:   "export-statistics",
		Short: "导出统计数据",
		Run:   exportStatistics,
	}

	var exportTaskID uint
	var outputFile string
	exportCmd.Flags().UintVar(&exportTaskID, "task", 0, "任务ID")
	exportCmd.Flags().StringVar(&outputFile, "output", "", "输出文件路径")
	exportCmd.MarkFlagRequired("task")

	rootCmd.AddCommand(createTaskCmd, submitEvalCmd, showResultsCmd, exportCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createTask(cmd *cobra.Command, args []string) {
	start, _ := time.Parse("2006-01-02", cmd.Flag("start").Value.String())
	end, _ := time.Parse("2006-01-02", cmd.Flag("end").Value.String())

	courseIDs, _ := cmd.Flags().GetUintSlice("courses")

	reqBody := map[string]interface{}{
		"semester":    cmd.Flag("semester").Value.String(),
		"start_date":  start.Format(time.RFC3339),
		"end_date":    end.Format(time.RFC3339),
		"course_ids":  courseIDs,
	}

	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/api/tasks", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusConflict {
		fmt.Printf("错误 (409): %s\n", string(body))
		return
	}
	if resp.StatusCode >= 400 {
		fmt.Printf("错误 (%d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Printf("任务创建成功: %s\n", string(body))
}

func submitEvaluation(cmd *cobra.Command, args []string) {
	studentID, _ := cmd.Flags().GetUint("student")
	courseID, _ := cmd.Flags().GetUint("course")
	taskID, _ := cmd.Flags().GetUint("task")
	comment := cmd.Flag("comment").Value.String()

	answers := make([]map[string]int, 13)
	for i := range answers {
		answers[i] = map[string]int{
			"question_id": i + 1,
			"score":       4,
		}
	}

	reqBody := map[string]interface{}{
		"student_id": studentID,
		"course_id":  courseID,
		"task_id":    taskID,
		"answers":    answers,
		"comment":    comment,
	}

	data, _ := json.Marshal(reqBody)
	resp, err := http.Post(baseURL+"/api/evaluations", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusConflict {
		fmt.Printf("错误 (409): %s\n", string(body))
		return
	}
	if resp.StatusCode == http.StatusBadRequest {
		fmt.Printf("错误 (400): %s\n", string(body))
		return
	}
	if resp.StatusCode >= 400 {
		fmt.Printf("错误 (%d): %s\n", resp.StatusCode, string(body))
		return
	}

	fmt.Printf("评估提交成功: %s\n", string(body))
}

func showResults(cmd *cobra.Command, args []string) {
	taskID, _ := cmd.Flags().GetUint("task")
	resp, err := http.Get(fmt.Sprintf("%s/api/results/task/%d", baseURL, taskID))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func exportStatistics(cmd *cobra.Command, args []string) {
	taskID, _ := cmd.Flags().GetUint("task")
	outputFile := cmd.Flag("output").Value.String()

	resp, err := http.Get(fmt.Sprintf("%s/api/results/export/%d", baseURL, taskID))
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if outputFile == "" {
		outputFile = fmt.Sprintf("statistics_task_%d.json", taskID)
	}

	if err := os.WriteFile(outputFile, body, 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}

	fmt.Printf("统计数据已导出到: %s\n", outputFile)
}
