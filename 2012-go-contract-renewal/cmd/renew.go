package cmd

import (
	"fmt"

	"contract-manager/internal/storage"

	"github.com/spf13/cobra"
)

var (
	renewID        string
	renewExpiryDate string
)

var renewCmd = &cobra.Command{
	Use:   "renew",
	Short: "手动续约合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		if renewID == "" || renewExpiryDate == "" {
			return fmt.Errorf("必须提供: -id, -expiry-date")
		}

		store, err := storage.Load()
		if err != nil {
			return err
		}

		newContract, err := store.RenewContract(renewID, renewExpiryDate, false)
		if err != nil {
			return err
		}

		fmt.Printf("手动续约成功！新合同号: %s，新到期日: %s\n",
			newContract.ID, storage.FormatDate(newContract.ExpiryDate))
		return nil
	},
}

func init() {
	renewCmd.Flags().StringVar(&renewID, "id", "", "要续约的合同编号（必填）")
	renewCmd.Flags().StringVar(&renewExpiryDate, "expiry-date", "", "新到期日期，格式 YYYY-MM-DD（必填）")
}
