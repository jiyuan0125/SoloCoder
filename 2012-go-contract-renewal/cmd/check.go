package cmd

import (
	"fmt"

	"contract-manager/internal/storage"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "检查到期提醒，自动续约已到期且开启自动续约的合同",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.Load()
		if err != nil {
			return err
		}

		alerts, err := store.CheckAndProcess()
		if err != nil {
			return err
		}

		if len(alerts) == 0 {
			fmt.Println("没有需要提醒的合同")
			return nil
		}

		for _, alert := range alerts {
			fmt.Println(alert)
		}

		return nil
	},
}

func init() {}
