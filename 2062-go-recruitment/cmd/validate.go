package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewValidateCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "校验数据一致性",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := deps.Store.ValidateConsistency()
			if err != nil {
				return fmt.Errorf("数据一致性校验失败: %w", err)
			}
			fmt.Println("数据一致性校验通过！")
			return nil
		},
	}
}
