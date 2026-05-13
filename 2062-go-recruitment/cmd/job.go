package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"recruitment/internal/services/job"
)

func NewJobCmd(deps *Dependencies) *cobra.Command {
	jobCmd := &cobra.Command{
		Use:   "job",
		Short: "职位管理",
		Long:  "发布、查询和关闭职位",
	}

	jobCmd.AddCommand(NewJobCreateCmd(deps))
	jobCmd.AddCommand(NewJobListCmd(deps))
	jobCmd.AddCommand(NewJobGetCmd(deps))
	jobCmd.AddCommand(NewJobCloseCmd(deps))

	return jobCmd
}

func NewJobCreateCmd(deps *Dependencies) *cobra.Command {
	var (
		title        string
		department   string
		requirements string
		salaryMin    float64
		salaryMax    float64
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "发布新职位",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := job.CreateJobRequest{
				Title:        title,
				Department:   department,
				Requirements: requirements,
				SalaryMin:    salaryMin,
				SalaryMax:    salaryMax,
			}

			j, err := deps.JobSvc.Create(req)
			if err != nil {
				return err
			}

			fmt.Printf("职位创建成功！\n")
			fmt.Printf("ID: %s\n", j.ID)
			fmt.Printf("标题: %s\n", j.Title)
			fmt.Printf("部门: %s\n", j.Department)
			fmt.Printf("状态: %s\n", j.Status)
			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "职位标题 (必填)")
	cmd.Flags().StringVarP(&department, "department", "d", "", "部门 (必填)")
	cmd.Flags().StringVarP(&requirements, "requirements", "r", "", "职位要求")
	cmd.Flags().Float64VarP(&salaryMin, "salary-min", "m", 0, "最低薪资")
	cmd.Flags().Float64VarP(&salaryMax, "salary-max", "x", 0, "最高薪资")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("department")

	return cmd
}

func NewJobListCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有职位",
		RunE: func(cmd *cobra.Command, args []string) error {
			jobs, err := deps.JobSvc.List()
			if err != nil {
				return err
			}

			if len(jobs) == 0 {
				fmt.Println("暂无职位")
				return nil
			}

			fmt.Printf("%-12s %-20s %-15s %-15s %-10s\n", "ID", "标题", "部门", "薪资范围", "状态")
			for _, j := range jobs {
				fmt.Printf("%-12s %-20s %-15s %-15.0f-%-10.0f %s\n",
					j.ID,
					truncate(j.Title, 18),
					truncate(j.Department, 13),
					j.SalaryMin,
					j.SalaryMax,
					j.Status,
				)
			}
			return nil
		},
	}
}

func NewJobGetCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get [职位ID]",
		Short: "查看职位详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			j, err := deps.JobSvc.GetByID(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("ID: %s\n", j.ID)
			fmt.Printf("标题: %s\n", j.Title)
			fmt.Printf("部门: %s\n", j.Department)
			fmt.Printf("薪资范围: %.0f - %.0f\n", j.SalaryMin, j.SalaryMax)
			fmt.Printf("状态: %s\n", j.Status)
			fmt.Printf("发布时间: %s\n", j.CreatedAt.Format("2006-01-02 15:04:05"))
			if j.ClosedAt != nil {
				fmt.Printf("关闭时间: %s\n", j.ClosedAt.Format("2006-01-02 15:04:05"))
			}
			fmt.Printf("要求: %s\n", j.Requirements)
			return nil
		},
	}
}

func NewJobCloseCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "close [职位ID]",
		Short: "关闭职位",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := deps.JobSvc.Close(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("职位已关闭: %s\n", args[0])
			return nil
		},
	}
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + ".."
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
