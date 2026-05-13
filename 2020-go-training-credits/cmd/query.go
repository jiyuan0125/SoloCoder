package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"credits/pkg/models"
	"credits/pkg/service"
)

var progressCmd = &cobra.Command{
	Use:   "progress",
	Short: "查询学分进度",
	Long:  "查询员工的学分完成进度",
	Run: func(cmd *cobra.Command, args []string) {
		if creditEmployeeID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定员工 ID")
			return
		}

		year := creditYear
		if year == 0 {
			year = time.Now().Year()
		}

		svc := service.NewCreditService(store)
		progress, err := svc.CalculateProgress(creditEmployeeID, year)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "查询进度失败: %v\n", err)
			return
		}

		printJSON(progress)
	},
}

var rankCmd = &cobra.Command{
	Use:   "rank",
	Short: "查询部门排名",
	Long:  "查询部门内员工的学分排名",
	Run: func(cmd *cobra.Command, args []string) {
		if employeeDepartment == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定部门")
			return
		}

		year := creditYear
		if year == 0 {
			year = time.Now().Year()
		}

		svc := service.NewCreditService(store)
		ranking, err := svc.GetDepartmentRanking(employeeDepartment, year)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "查询排名失败: %v\n", err)
			return
		}

		if ranking == nil {
			ranking = []*models.RankItem{}
		}

		printJSON(ranking)
	},
}

var yearlyCmd = &cobra.Command{
	Use:   "yearly",
	Short: "查询年度统计",
	Long:  "查询员工的年度学分统计",
	Run: func(cmd *cobra.Command, args []string) {
		if creditEmployeeID == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "错误: 请指定员工 ID")
			return
		}

		year := creditYear
		if year == 0 {
			year = time.Now().Year()
		}

		svc := service.NewCreditService(store)
		summary, err := svc.GetYearlySummary(creditEmployeeID, year)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "查询年度统计失败: %v\n", err)
			return
		}

		printJSON(summary)
	},
}

func init() {
	progressCmd.Flags().StringVar(&creditEmployeeID, "employee", "", "员工 ID")
	progressCmd.Flags().IntVar(&creditYear, "year", 0, "年份（默认为当前年份）")
	progressCmd.MarkFlagRequired("employee")

	rankCmd.Flags().StringVar(&employeeDepartment, "department", "", "部门名称")
	rankCmd.Flags().IntVar(&creditYear, "year", 0, "年份（默认为当前年份）")
	rankCmd.MarkFlagRequired("department")

	yearlyCmd.Flags().StringVar(&creditEmployeeID, "employee", "", "员工 ID")
	yearlyCmd.Flags().IntVar(&creditYear, "year", 0, "年份（默认为当前年份）")
	yearlyCmd.MarkFlagRequired("employee")
}
