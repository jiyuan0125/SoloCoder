package cmd

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"credits/pkg/models"
	"credits/pkg/service"
)

var creditCmd = &cobra.Command{
	Use:   "credit",
	Short: "学分管理",
	Long:  "学分管理命令 - 记录员工完成培训的学分",
}

var (
	creditEmployeeID string
	creditCourseID   string
	creditYear       int
)

var recordCreditCmd = &cobra.Command{
	Use:   "record",
	Short: "记录学分",
	Long:  "记录员工完成培训课程的学分",
	Run: func(cmd *cobra.Command, args []string) {
		if creditEmployeeID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定员工 ID")
			return
		}
		if creditCourseID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定课程 ID")
			return
		}

		course, err := store.GetCourse(creditCourseID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取课程失败: %v\n", err)
			return
		}

		employee, err := store.GetEmployee(creditEmployeeID)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "获取员工失败: %v\n", err)
			return
		}

		now := time.Now()
		if course.ValidUntil.Before(now) {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 课程已过期，无法记录学分")
			return
		}

		record := &models.CreditRecord{
			ID:          uuid.New().String()[:12],
			EmployeeID:  creditEmployeeID,
			CourseID:    creditCourseID,
			Credits:     course.Credits,
			CourseType:  course.Type,
			CompletedAt: now,
		}

		if err := store.AddCreditRecord(record); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "记录学分失败: %v\n", err)
			return
		}

		svc := service.NewCreditService(store)
		if err := svc.ProcessYearEndCarryOver(creditEmployeeID, now.Year()-1); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "警告: 处理上年结转失败: %v\n", err)
		}

		fmt.Printf("员工 %s (%s) 完成课程 %s，获得 %d 学分\n",
			employee.Name, employee.ID, course.Name, course.Credits)
		printJSON(record)
	},
}

func init() {
	recordCreditCmd.Flags().StringVar(&creditEmployeeID, "employee", "", "员工 ID")
	recordCreditCmd.Flags().StringVar(&creditCourseID, "course", "", "课程 ID")

	recordCreditCmd.MarkFlagRequired("employee")
	recordCreditCmd.MarkFlagRequired("course")

	creditCmd.AddCommand(recordCreditCmd)
}
