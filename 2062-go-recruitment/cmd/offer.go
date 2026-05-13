package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"recruitment/internal/services/offer"
)

func NewOfferCmd(deps *Dependencies) *cobra.Command {
	offerCmd := &cobra.Command{
		Use:   "offer",
		Short: "Offer 管理",
		Long:  "Offer 生成、发放和管理",
	}

	offerCmd.AddCommand(NewOfferCreateCmd(deps))
	offerCmd.AddCommand(NewOfferListCmd(deps))
	offerCmd.AddCommand(NewOfferGetCmd(deps))
	offerCmd.AddCommand(NewOfferAcceptCmd(deps))
	offerCmd.AddCommand(NewOfferRejectCmd(deps))
	offerCmd.AddCommand(NewOfferBudgetCmd(deps))

	return offerCmd
}

func NewOfferCreateCmd(deps *Dependencies) *cobra.Command {
	var (
		candidateID string
		salary      float64
		startStr    string
		validDays   int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "生成 Offer",
		RunE: func(cmd *cobra.Command, args []string) error {
			startDate, err := time.Parse("2006-01-02", startStr)
			if err != nil {
				return fmt.Errorf("入职日期格式错误，使用: 2006-01-02")
			}

			req := offer.CreateOfferRequest{
				CandidateID: candidateID,
				Salary:      salary,
				StartDate:   startDate,
				ValidDays:   validDays,
			}

			o, err := deps.OfferSvc.Create(req)
			if err != nil {
				return err
			}

			fmt.Printf("Offer 生成成功！\n")
			fmt.Printf("ID: %s\n", o.ID)
			fmt.Printf("候选人ID: %s\n", o.CandidateID)
			fmt.Printf("薪资: %.0f\n", o.Salary)
			fmt.Printf("入职日期: %s\n", o.StartDate.Format("2006-01-02"))
			fmt.Printf("有效期至: %s\n", o.ValidUntil.Format("2006-01-02 15:04:05"))
			fmt.Printf("状态: %s\n", o.Status)
			return nil
		},
	}

	cmd.Flags().StringVarP(&candidateID, "candidate-id", "c", "", "候选人ID (必填)")
	cmd.Flags().Float64VarP(&salary, "salary", "s", 0, "薪资 (必填)")
	cmd.Flags().StringVarP(&startStr, "start-date", "d", "", "入职日期 (格式: 2006-01-02) (必填)")
	cmd.Flags().IntVarP(&validDays, "valid-days", "v", 7, "Offer 有效期（天）")

	cmd.MarkFlagRequired("candidate-id")
	cmd.MarkFlagRequired("salary")
	cmd.MarkFlagRequired("start-date")

	return cmd
}

func NewOfferListCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有 Offer",
		RunE: func(cmd *cobra.Command, args []string) error {
			offers, err := deps.OfferSvc.List()
			if err != nil {
				return err
			}

			if len(offers) == 0 {
				fmt.Println("暂无 Offer")
				return nil
			}

			fmt.Printf("%-12s %-15s %-12s %-15s %-15s\n",
				"ID", "候选人", "薪资", "入职日期", "状态")
			for _, o := range offers {
				fmt.Printf("%-12s %-15s %-12.0f %-15s %-15s\n",
					o.ID,
					truncate(o.CandidateID, 13),
					o.Salary,
					o.StartDate.Format("2006-01-02"),
					o.Status,
				)
			}
			return nil
		},
	}
}

func NewOfferGetCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get [OfferID]",
		Short: "查看 Offer 详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o, err := deps.OfferSvc.GetByID(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("ID: %s\n", o.ID)
			fmt.Printf("候选人ID: %s\n", o.CandidateID)
			fmt.Printf("职位ID: %s\n", o.JobID)
			fmt.Printf("薪资: %.0f\n", o.Salary)
			fmt.Printf("入职日期: %s\n", o.StartDate.Format("2006-01-02"))
			fmt.Printf("有效期至: %s\n", o.ValidUntil.Format("2006-01-02 15:04:05"))
			fmt.Printf("状态: %s\n", o.Status)
			fmt.Printf("创建时间: %s\n", o.CreatedAt.Format("2006-01-02 15:04:05"))
			if o.RepliedAt != nil {
				fmt.Printf("回复时间: %s\n", o.RepliedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func NewOfferAcceptCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "accept [OfferID]",
		Short: "接受 Offer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := deps.OfferSvc.Accept(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Offer 已接受: %s\n", args[0])
			return nil
		},
	}
}

func NewOfferRejectCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "reject [OfferID]",
		Short: "拒绝 Offer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := deps.OfferSvc.Reject(args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Offer 已拒绝: %s\n", args[0])
			return nil
		},
	}
}

func NewOfferBudgetCmd(deps *Dependencies) *cobra.Command {
	budgetCmd := &cobra.Command{
		Use:   "budget",
		Short: "预算管理",
	}

	budgetCmd.AddCommand(NewBudgetInitCmd(deps))
	budgetCmd.AddCommand(NewBudgetAdjustCmd(deps))
	budgetCmd.AddCommand(NewBudgetListCmd(deps))
	budgetCmd.AddCommand(NewBudgetGetCmd(deps))

	return budgetCmd
}

func NewBudgetInitCmd(deps *Dependencies) *cobra.Command {
	var (
		jobID       string
		totalAmount float64
		stagesStr   string
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "初始化职位预算",
		RunE: func(cmd *cobra.Command, args []string) error {
			stages := strings.Split(stagesStr, ",")
			for i := range stages {
				stages[i] = strings.TrimSpace(stages[i])
			}

			req := offer.InitBudgetRequest{
				JobID:       jobID,
				TotalAmount: totalAmount,
				Stages:      stages,
			}

			err := deps.OfferSvc.InitBudget(req)
			if err != nil {
				return err
			}

			fmt.Printf("预算初始化成功！职位ID: %s, 总金额: %.0f\n", jobID, totalAmount)
			return nil
		},
	}

	cmd.Flags().StringVarP(&jobID, "job-id", "j", "", "职位ID (必填)")
	cmd.Flags().Float64VarP(&totalAmount, "total", "t", 0, "总金额 (必填)")
	cmd.Flags().StringVarP(&stagesStr, "stages", "s", "", "阶段列表，逗号分隔，如: 筛选,面试,Offer (必填)")

	cmd.MarkFlagRequired("job-id")
	cmd.MarkFlagRequired("total")
	cmd.MarkFlagRequired("stages")

	return cmd
}

func NewBudgetAdjustCmd(deps *Dependencies) *cobra.Command {
	var (
		jobID    string
		newTotal float64
	)

	cmd := &cobra.Command{
		Use:   "adjust",
		Short: "调整总金额，各阶段按比例重新分配",
		RunE: func(cmd *cobra.Command, args []string) error {
			req := offer.AdjustBudgetRequest{
				JobID:    jobID,
				NewTotal: newTotal,
			}

			err := deps.OfferSvc.AdjustBudget(req)
			if err != nil {
				return err
			}

			fmt.Printf("预算调整成功！职位ID: %s, 新总金额: %.0f\n", jobID, newTotal)
			return nil
		},
	}

	cmd.Flags().StringVarP(&jobID, "job-id", "j", "", "职位ID (必填)")
	cmd.Flags().Float64VarP(&newTotal, "new-total", "n", 0, "新总金额 (必填)")

	cmd.MarkFlagRequired("job-id")
	cmd.MarkFlagRequired("new-total")

	return cmd
}

func NewBudgetListCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有预算",
		RunE: func(cmd *cobra.Command, args []string) error {
			budgets, err := deps.OfferSvc.ListBudgets()
			if err != nil {
				return err
			}

			if len(budgets) == 0 {
				fmt.Println("暂无预算")
				return nil
			}

			fmt.Printf("%-12s %-15s %-10s\n", "职位ID", "总金额", "阶段数")
			for _, b := range budgets {
				fmt.Printf("%-12s %-15.0f %-10d\n",
					b.JobID,
					b.TotalAmount,
					len(b.Stages),
				)
			}
			return nil
		},
	}
}

func NewBudgetGetCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "get [职位ID]",
		Short: "查看职位预算详情",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := deps.OfferSvc.GetBudget(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("职位ID: %s\n", b.JobID)
			fmt.Printf("总金额: %.0f\n", b.TotalAmount)
			fmt.Printf("更新时间: %s\n", b.UpdatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println("各阶段计划金额:")
			for i, s := range b.Stages {
				ratio := s.PlannedAmount / b.TotalAmount * 100
				fmt.Printf("  %d. %s: %.0f (%.1f%%)\n", i+1, s.Stage, s.PlannedAmount, ratio)
			}
			return nil
		},
	}
}
