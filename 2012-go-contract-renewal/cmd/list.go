package cmd

import (
	"fmt"

	"contract-manager/internal/storage"

	"github.com/spf13/cobra"
)

var (
	listParty   string
	listSortByExpiry bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.Load()
		if err != nil {
			return err
		}

		contracts := store.GetContracts(listParty, listSortByExpiry)

		if len(contracts) == 0 {
			fmt.Println("没有找到合同")
			return nil
		}

		fmt.Printf("%-15s %-30s %-20s %-12s %-12s %-10s %-10s\n",
			"编号", "名称", "签约方", "签约日期", "到期日期", "自动续约", "状态")
		fmt.Println("---------------------------------------------------------------------------------------------------")

		for _, c := range contracts {
			status := "有效"
			if c.Terminated {
				status = "已终止"
			}
			autoRenew := "否"
			if c.AutoRenew {
				autoRenew = "是"
			}
			fmt.Printf("%-15s %-30s %-20s %-12s %-12s %-10s %-10s\n",
				c.ID,
				truncate(c.Name, 28),
				truncate(c.Party, 18),
				storage.FormatDate(c.SignDate),
				storage.FormatDate(c.ExpiryDate),
				autoRenew,
				status)
		}

		return nil
	},
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func init() {
	listCmd.Flags().StringVar(&listParty, "party", "", "按签约方筛选")
	listCmd.Flags().BoolVar(&listSortByExpiry, "sort-by-expiry", false, "按到期日期排序（升序）")
}
