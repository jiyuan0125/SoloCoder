package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"recruitment/internal/models"
	"recruitment/internal/services/candidate"
)

func NewCandidateCmd(deps *Dependencies) *cobra.Command {
	candidateCmd := &cobra.Command{
		Use:   "candidate",
		Short: "候选人管理",
		Long:  "候选人投递、筛选和淘汰管理",
	}

	candidateCmd.AddCommand(NewCandidateApplyCmd(deps))
	candidateCmd.AddCommand(NewCandidateListCmd(deps))
	candidateCmd.AddCommand(NewCandidateGetCmd(deps))
	candidateCmd.AddCommand(NewCandidateScreenCmd(deps))
	candidateCmd.AddCommand(NewCandidateRejectCmd(deps))

	return candidateCmd
}

func NewCandidateApplyCmd(deps *Dependencies) *cobra.Command {
	var (
		resumeID string
		name     string
		email    string
		phone    string
		jobID    string
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "投递简历",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := candidate.ApplyRequest{
				ResumeID: resumeID,
				Name:     name,
				Email:    email,
				Phone:    phone,
				JobID:    jobID,
			}

			c, err := deps.CandidateSvc.Apply(req)
			if err != nil {
				return err
			}

			fmt.Printf("投递成功！\n")
			fmt.Printf("候选人ID: %s\n", c.ID)
			fmt.Printf("姓名: %s\n", c.Name)
			fmt.Printf("简历ID: %s\n", c.ResumeID)
			fmt.Printf("职位ID: %s\n", c.JobID)
			fmt.Printf("状态: %s\n", c.Status)
			return nil
		},
	}

	cmd.Flags().StringVarP(&resumeID, "resume-id", "i", "", "简历ID (必填)")
	cmd.Flags().StringVarP(&name, "name", "n", "", "姓名 (必填)")
	cmd.Flags().StringVarP(&email, "email", "e", "", "邮箱 (必填)")
	cmd.Flags().StringVarP(&phone, "phone", "p", "", "电话 (必填)")
	cmd.Flags().StringVarP(&jobID, "job-id", "j", "", "职位ID (必填)")

	cmd.MarkFlagRequired("resume-id")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("email")
	cmd.MarkFlagRequired("phone")
	cmd.MarkFlagRequired("job-id")

	return cmd
}

func NewCandidateListCmd(deps *Dependencies) *cobra.Command {
	var jobID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出候选人",
		RunE: func(cmd *cobra.Command, args []string) error {
			var candidates []candidateWithJob
			var err error

			if jobID != "" {
				list, err2 := deps.CandidateSvc.ListByJob(jobID)
				err = err2
				if err == nil {
					for _, c := range list {
						candidates = append(candidates, candidateWithJob{Candidate: c, JobTitle: getJobTitle(deps, c.JobID)})
					}
				}
			} else {
				list, err2 := deps.CandidateSvc.List()
				err = err2
				if err == nil {
					for _, c := range list {
						candidates = append(candidates, candidateWithJob{Candidate: c, JobTitle: getJobTitle(deps, c.JobID)})
					}
				}
			}

			if err != nil {
				return err
			}

			if len(candidates) == 0 {
				fmt.Println("暂无候选人")
				return nil
			}

			fmt.Printf("%-12s %-15s %-15s %-20s %-15s\n", "ID", "姓名", "简历ID", "职位", "状态")
			for _, c := range candidates {
				fmt.Printf("%-12s %-15s %-15s %-20s %-15s\n",
					c.Candidate.ID,
					truncate(c.Candidate.Name, 13),
					truncate(c.Candidate.ResumeID, 13),
					truncate(c.JobTitle, 18),
					c.Candidate.Status,
				)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&jobID, "job-id", "j", "", "按职位筛选")

	return cmd
}

type candidateWithJob struct {
	Candidate models.Candidate
	JobTitle  string
}

func getJobTitle(deps *Dependencies, jobID string) string {
	j, err := deps.JobSvc.GetByID(jobID)
	if err != nil {
		return jobID
	}
	return j.Title
}

func NewCandidateGetCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get [候选人ID]",
		Short: "查看候选人详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := deps.CandidateSvc.GetByID(args[0])
			if err != nil {
				return err
			}

			jobTitle := getJobTitle(deps, c.JobID)

			fmt.Printf("ID: %s\n", c.ID)
			fmt.Printf("姓名: %s\n", c.Name)
			fmt.Printf("邮箱: %s\n", c.Email)
			fmt.Printf("电话: %s\n", c.Phone)
			fmt.Printf("简历ID: %s\n", c.ResumeID)
			fmt.Printf("职位: %s (%s)\n", jobTitle, c.JobID)
			fmt.Printf("状态: %s\n", c.Status)
			fmt.Printf("投递时间: %s\n", c.AppliedAt.Format("2006-01-02 15:04:05"))
			if c.RejectionReason != "" {
				fmt.Printf("淘汰原因: %s\n", c.RejectionReason)
			}
			if c.RejectedAt != nil {
				fmt.Printf("淘汰时间: %s\n", c.RejectedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func NewCandidateScreenCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "screen [候选人ID]",
		Short: "筛选候选人（进入筛选阶段）",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := deps.CandidateSvc.Screen(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("候选人已进入筛选阶段: %s\n", args[0])
			return nil
		},
	}
}

func NewCandidateRejectCmd(deps *Dependencies) *cobra.Command {
	var reason string

	cmd := &cobra.Command{
		Use:   "reject [候选人ID]",
		Short: "淘汰候选人",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if reason == "" {
				return fmt.Errorf("请提供淘汰原因 (--reason)")
			}

			err := deps.CandidateSvc.Reject(args[0], reason)
			if err != nil {
				return err
			}
			fmt.Printf("候选人已淘汰: %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "淘汰原因 (必填)")

	return cmd
}
