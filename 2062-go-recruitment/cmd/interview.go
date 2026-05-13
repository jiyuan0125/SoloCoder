package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"recruitment/internal/models"
	"recruitment/internal/services/interview"
)

func NewInterviewCmd(deps *Dependencies) *cobra.Command {
	interviewCmd := &cobra.Command{
		Use:   "interview",
		Short: "面试管理",
		Long:  "面试安排、冲突检查和结果提交",
	}

	interviewCmd.AddCommand(NewInterviewScheduleCmd(deps))
	interviewCmd.AddCommand(NewInterviewListCmd(deps))
	interviewCmd.AddCommand(NewInterviewGetCmd(deps))
	interviewCmd.AddCommand(NewInterviewSubmitCmd(deps))

	return interviewCmd
}

func NewInterviewScheduleCmd(deps *Dependencies) *cobra.Command {
	var (
		candidateID string
		interviewer string
		startStr    string
		endStr      string
		isRetry     bool
	)

	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "安排面试",
		RunE: func(cmd *cobra.Command, args []string) error {
			startTime, err := time.Parse("2006-01-02 15:04", startStr)
			if err != nil {
				return fmt.Errorf("开始时间格式错误，使用: 2006-01-02 15:04")
			}

			endTime, err := time.Parse("2006-01-02 15:04", endStr)
			if err != nil {
				return fmt.Errorf("结束时间格式错误，使用: 2006-01-02 15:04")
			}

			req := interview.ScheduleRequest{
				CandidateID: candidateID,
				Interviewer: interviewer,
				StartTime:   startTime,
				EndTime:     endTime,
				IsRetry:     isRetry,
			}

			i, err := deps.InterviewSvc.Schedule(req)
			if err != nil {
				return err
			}

			fmt.Printf("面试安排成功！\n")
			fmt.Printf("ID: %s\n", i.ID)
			fmt.Printf("面试官: %s\n", i.Interviewer)
			fmt.Printf("时间: %s - %s\n",
				i.StartTime.Format("2006-01-02 15:04"),
				i.EndTime.Format("2006-01-02 15:04"))
			fmt.Printf("轮次: 第%d轮", i.Round)
			if i.IsRetry {
				fmt.Printf(" (加试)")
			}
			fmt.Println()
			return nil
		},
	}

	cmd.Flags().StringVarP(&candidateID, "candidate-id", "c", "", "候选人ID (必填)")
	cmd.Flags().StringVarP(&interviewer, "interviewer", "i", "", "面试官姓名 (必填)")
	cmd.Flags().StringVarP(&startStr, "start", "s", "", "开始时间 (格式: 2006-01-02 15:04) (必填)")
	cmd.Flags().StringVarP(&endStr, "end", "e", "", "结束时间 (格式: 2006-01-02 15:04) (必填)")
	cmd.Flags().BoolVarP(&isRetry, "retry", "r", false, "是否为加试（待定候选人）")

	cmd.MarkFlagRequired("candidate-id")
	cmd.MarkFlagRequired("interviewer")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func NewInterviewListCmd(deps *Dependencies) *cobra.Command {
	var candidateID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出面试",
		RunE: func(cmd *cobra.Command, args []string) error {
			var interviews []models.Interview
			var err error

			if candidateID != "" {
				interviews, err = deps.InterviewSvc.ListByCandidate(candidateID)
			} else {
				interviews, err = deps.InterviewSvc.List()
			}

			if err != nil {
				return err
			}

			if len(interviews) == 0 {
				fmt.Println("暂无面试记录")
				return nil
			}

			fmt.Printf("%-12s %-15s %-15s %-25s %-10s %-8s\n",
				"ID", "候选人", "面试官", "时间", "轮次", "状态")
			for _, i := range interviews {
				roundInfo := fmt.Sprintf("第%d轮", i.Round)
				if i.IsRetry {
					roundInfo += "(加试)"
				}
				fmt.Printf("%-12s %-15s %-15s %-25s %-10s %-8s\n",
					i.ID,
					truncate(i.CandidateID, 13),
					truncate(i.Interviewer, 13),
					i.StartTime.Format("01-02 15:04")+"-"+i.EndTime.Format("15:04"),
					roundInfo,
					i.Status,
				)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&candidateID, "candidate-id", "c", "", "按候选人筛选")

	return cmd
}

func NewInterviewGetCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get [面试ID]",
		Short: "查看面试详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			i, err := deps.InterviewSvc.GetByID(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("ID: %s\n", i.ID)
			fmt.Printf("候选人ID: %s\n", i.CandidateID)
			fmt.Printf("职位ID: %s\n", i.JobID)
			fmt.Printf("面试官: %s\n", i.Interviewer)
			fmt.Printf("时间: %s - %s\n",
				i.StartTime.Format("2006-01-02 15:04"),
				i.EndTime.Format("2006-01-02 15:04"))
			fmt.Printf("轮次: 第%d轮", i.Round)
			if i.IsRetry {
				fmt.Printf(" (加试)")
			}
			fmt.Println()
			fmt.Printf("状态: %s\n", i.Status)
			if i.Result != nil {
				fmt.Printf("结果: %s\n", *i.Result)
			}
			if i.Feedback != "" {
				fmt.Printf("评价: %s\n", i.Feedback)
			}
			if i.SubmittedAt != nil {
				fmt.Printf("提交时间: %s\n", i.SubmittedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func NewInterviewSubmitCmd(deps *Dependencies) *cobra.Command {
	var (
		result   string
		feedback string
	)

	cmd := &cobra.Command{
		Use:   "submit [面试ID]",
		Short: "提交面试结果",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var r models.InterviewResult
			switch result {
			case "pass", "通过":
				r = models.InterviewResultPass
			case "fail", "不通过":
				r = models.InterviewResultFail
			case "pending", "待定":
				r = models.InterviewResultPending
			default:
				return fmt.Errorf("无效的结果值，请使用: pass/fail/pending 或 通过/不通过/待定")
			}

			req := interview.SubmitResultRequest{
				InterviewID: args[0],
				Result:      r,
				Feedback:    feedback,
			}

			err := deps.InterviewSvc.SubmitResult(req)
			if err != nil {
				return err
			}
			fmt.Printf("面试结果已提交: %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVarP(&result, "result", "r", "", "结果: pass/fail/pending (必填)")
	cmd.Flags().StringVarP(&feedback, "feedback", "f", "", "面试评价")

	cmd.MarkFlagRequired("result")

	return cmd
}
