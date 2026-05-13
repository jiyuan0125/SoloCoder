package cmd

import (
	"fmt"

	"contract-manager/internal/models"
	"contract-manager/internal/storage"

	"github.com/spf13/cobra"
)

var (
	addID         string
	addName       string
	addParty      string
	addSignDate   string
	addExpiryDate string
	addAutoRenew  bool
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "添加新合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addID == "" || addName == "" || addParty == "" || addSignDate == "" || addExpiryDate == "" {
			return fmt.Errorf("必须提供: -id, -name, -party, -sign-date, -expiry-date")
		}

		signDate, err := storage.ParseDate(addSignDate)
		if err != nil {
			return fmt.Errorf("签约日期格式错误: %v", err)
		}

		expiryDate, err := storage.ParseDate(addExpiryDate)
		if err != nil {
			return fmt.Errorf("到期日期格式错误: %v", err)
		}

		store, err := storage.Load()
		if err != nil {
			return err
		}

		contract := &models.Contract{
			ID:         addID,
			Name:       addName,
			Party:      addParty,
			SignDate:   signDate,
			ExpiryDate: expiryDate,
			AutoRenew:  addAutoRenew,
			Terminated: false,
		}

		if err := store.AddContract(contract); err != nil {
			return err
		}

		fmt.Printf("合同添加成功: %s\n", addID)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addID, "id", "", "合同编号（必填）")
	addCmd.Flags().StringVar(&addName, "name", "", "合同名称（必填）")
	addCmd.Flags().StringVar(&addParty, "party", "", "签约方（必填）")
	addCmd.Flags().StringVar(&addSignDate, "sign-date", "", "签约日期，格式 YYYY-MM-DD（必填）")
	addCmd.Flags().StringVar(&addExpiryDate, "expiry-date", "", "到期日期，格式 YYYY-MM-DD（必填）")
	addCmd.Flags().BoolVar(&addAutoRenew, "auto-renew", false, "是否自动续约（默认否）")
}
