package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"contract-lifecycle/models"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "创建新合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		partiesStr, _ := cmd.Flags().GetString("parties")
		signDateStr, _ := cmd.Flags().GetString("sign-date")
		effectiveDateStr, _ := cmd.Flags().GetString("effective-date")
		expiryDateStr, _ := cmd.Flags().GetString("expiry-date")
		amountStr, _ := cmd.Flags().GetString("amount")
		content, _ := cmd.Flags().GetString("content")

		parties := strings.Split(partiesStr, ",")
		for i, p := range parties {
			parties[i] = strings.TrimSpace(p)
		}

		signDate, err := parseDate(signDateStr)
		if err != nil {
			return fmt.Errorf("签署日期格式错误: %v", err)
		}

		effectiveDate, err := parseDate(effectiveDateStr)
		if err != nil {
			return fmt.Errorf("生效日期格式错误: %v", err)
		}

		expiryDate, err := parseDate(expiryDateStr)
		if err != nil {
			return fmt.Errorf("到期日期格式错误: %v", err)
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("金额格式错误: %v", err)
		}

		contract := &models.Contract{
			Title:         title,
			Parties:       parties,
			SignDate:      signDate,
			EffectiveDate: effectiveDate,
			ExpiryDate:    expiryDate,
			Amount:        amount,
			Content:       content,
		}

		if err := svc.CreateContract(contract); err != nil {
			return err
		}

		fmt.Printf("合同创建成功！\nID: %s\n标题: %s\n状态: %s\n", contract.ID, contract.Title, contract.Status)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		contracts, err := svc.ListContracts()
		if err != nil {
			return err
		}

		if len(contracts) == 0 {
			fmt.Println("[]")
			return nil
		}

		fmt.Println("合同列表:")
		for _, c := range contracts {
			fmt.Printf("  ID: %s\n    标题: %s\n    状态: %s\n    金额: %.2f\n    到期日期: %s\n\n",
				c.ID, c.Title, c.Status, c.Amount, formatDate(c.ExpiryDate))
		}
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "显示合同详情",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contract, err := svc.GetContract(args[0])
		if err != nil {
			return err
		}

		data, _ := json.MarshalIndent(contract, "", "  ")
		fmt.Println(string(data))
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "更新合同信息",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		updates := &models.Contract{}

		if title, _ := cmd.Flags().GetString("title"); title != "" {
			updates.Title = title
		}
		if partiesStr, _ := cmd.Flags().GetString("parties"); partiesStr != "" {
			parties := strings.Split(partiesStr, ",")
			for i, p := range parties {
				parties[i] = strings.TrimSpace(p)
			}
			updates.Parties = parties
		}
		if signDateStr, _ := cmd.Flags().GetString("sign-date"); signDateStr != "" {
			if date, err := parseDate(signDateStr); err == nil {
				updates.SignDate = date
			} else {
				return err
			}
		}
		if effectiveDateStr, _ := cmd.Flags().GetString("effective-date"); effectiveDateStr != "" {
			if date, err := parseDate(effectiveDateStr); err == nil {
				updates.EffectiveDate = date
			} else {
				return err
			}
		}
		if expiryDateStr, _ := cmd.Flags().GetString("expiry-date"); expiryDateStr != "" {
			if date, err := parseDate(expiryDateStr); err == nil {
				updates.ExpiryDate = date
			} else {
				return err
			}
		}
		if amountStr, _ := cmd.Flags().GetString("amount"); amountStr != "" {
			if amount, err := strconv.ParseFloat(amountStr, 64); err == nil {
				updates.Amount = amount
			} else {
				return err
			}
		}
		if content, _ := cmd.Flags().GetString("content"); content != "" {
			updates.Content = content
		}

		if err := svc.UpdateContract(args[0], updates); err != nil {
			return err
		}

		fmt.Println("合同更新成功！")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status [id] [action]",
	Short: "变更合同状态 (submit/reject/approve/sign/perform/expire)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := svc.TransitionStatus(args[0], args[1], resourceID); err != nil {
			return err
		}

		contract, err := svc.GetContract(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("状态变更成功！\n当前状态: %s\n", contract.Status)
		return nil
	},
}

var remindCmd = &cobra.Command{
	Use:   "remind",
	Short: "检查并发送到期提醒",
	RunE: func(cmd *cobra.Command, args []string) error {
		reminders, err := svc.CheckReminders(resourceID)
		if err != nil {
			return err
		}

		if len(reminders) == 0 {
			fmt.Println("暂无到期提醒")
			return nil
		}

		fmt.Println("到期提醒:")
		for _, r := range reminders {
			fmt.Println("  " + r)
		}
		return nil
	},
}

var terminateCmd = &cobra.Command{
	Use:   "terminate [id]",
	Short: "终止并归档合同",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := svc.TerminateContract(args[0], resourceID); err != nil {
			return err
		}

		fmt.Println("合同已终止并归档")
		return nil
	},
}

var renewCmd = &cobra.Command{
	Use:   "renew [original-id]",
	Short: "续约合同（创建新合同关联原合同）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		partiesStr, _ := cmd.Flags().GetString("parties")
		signDateStr, _ := cmd.Flags().GetString("sign-date")
		effectiveDateStr, _ := cmd.Flags().GetString("effective-date")
		expiryDateStr, _ := cmd.Flags().GetString("expiry-date")
		amountStr, _ := cmd.Flags().GetString("amount")
		content, _ := cmd.Flags().GetString("content")

		parties := strings.Split(partiesStr, ",")
		for i, p := range parties {
			parties[i] = strings.TrimSpace(p)
		}

		signDate, err := parseDate(signDateStr)
		if err != nil {
			return err
		}

		effectiveDate, err := parseDate(effectiveDateStr)
		if err != nil {
			return err
		}

		expiryDate, err := parseDate(expiryDateStr)
		if err != nil {
			return err
		}

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return err
		}

		newContract := &models.Contract{
			Title:         title,
			Parties:       parties,
			SignDate:      signDate,
			EffectiveDate: effectiveDate,
			ExpiryDate:    expiryDate,
			Amount:        amount,
			Content:       content,
		}

		if err := svc.RenewContract(args[0], newContract, resourceID); err != nil {
			return err
		}

		fmt.Printf("续约合同创建成功！\nID: %s\n标题: %s\n原合同ID: %s\n", newContract.ID, newContract.Title, args[0])
		return nil
	},
}

var breachCmd = &cobra.Command{
	Use:   "breach [contract-id]",
	Short: "记录履行中合同的违约行为",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		description, _ := cmd.Flags().GetString("description")
		dateStr, _ := cmd.Flags().GetString("date")

		date, err := parseDate(dateStr)
		if err != nil {
			date = time.Now()
		}

		if err := svc.RecordBreach(args[0], description, date, resourceID); err != nil {
			return err
		}

		fmt.Println("违约记录已添加")
		return nil
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs [contract-id]",
	Short: "查看操作日志（可选指定合同ID）",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contractID := ""
		if len(args) > 0 {
			contractID = args[0]
		}

		logs, err := svc.GetLogs(contractID)
		if err != nil {
			return err
		}

		if len(logs) == 0 {
			fmt.Println("[]")
			return nil
		}

		for _, log := range logs {
			data, _ := json.Marshal(log)
			fmt.Println(string(data))
		}
		return nil
	},
}

func init() {
	createCmd.Flags().String("title", "", "合同标题")
	createCmd.Flags().String("parties", "", "签约方（逗号分隔）")
	createCmd.Flags().String("sign-date", "", "签署日期 (YYYY-MM-DD)")
	createCmd.Flags().String("effective-date", "", "生效日期 (YYYY-MM-DD)")
	createCmd.Flags().String("expiry-date", "", "到期日期 (YYYY-MM-DD)")
	createCmd.Flags().String("amount", "", "合同金额")
	createCmd.Flags().String("content", "", "合同内容")
	createCmd.MarkFlagRequired("title")
	createCmd.MarkFlagRequired("parties")
	createCmd.MarkFlagRequired("sign-date")
	createCmd.MarkFlagRequired("effective-date")
	createCmd.MarkFlagRequired("expiry-date")
	createCmd.MarkFlagRequired("amount")

	updateCmd.Flags().String("title", "", "合同标题")
	updateCmd.Flags().String("parties", "", "签约方（逗号分隔）")
	updateCmd.Flags().String("sign-date", "", "签署日期 (YYYY-MM-DD)")
	updateCmd.Flags().String("effective-date", "", "生效日期 (YYYY-MM-DD)")
	updateCmd.Flags().String("expiry-date", "", "到期日期 (YYYY-MM-DD)")
	updateCmd.Flags().String("amount", "", "合同金额")
	updateCmd.Flags().String("content", "", "合同内容")

	breachCmd.Flags().String("description", "", "违约描述")
	breachCmd.Flags().String("date", "", "违约日期 (YYYY-MM-DD)")
	breachCmd.MarkFlagRequired("description")

	renewCmd.Flags().String("title", "", "新合同标题")
	renewCmd.Flags().String("parties", "", "签约方（逗号分隔）")
	renewCmd.Flags().String("sign-date", "", "签署日期 (YYYY-MM-DD)")
	renewCmd.Flags().String("effective-date", "", "生效日期 (YYYY-MM-DD)")
	renewCmd.Flags().String("expiry-date", "", "到期日期 (YYYY-MM-DD)")
	renewCmd.Flags().String("amount", "", "合同金额")
	renewCmd.Flags().String("content", "", "合同内容")
	renewCmd.MarkFlagRequired("title")
	renewCmd.MarkFlagRequired("parties")
	renewCmd.MarkFlagRequired("sign-date")
	renewCmd.MarkFlagRequired("effective-date")
	renewCmd.MarkFlagRequired("expiry-date")
	renewCmd.MarkFlagRequired("amount")
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
