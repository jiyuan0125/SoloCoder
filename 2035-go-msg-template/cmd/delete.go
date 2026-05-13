package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"msg-template/internal/template"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [名称]",
	Short: "删除模板",
	Long:  "删除指定的消息模板。",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		manager := template.NewTemplateManager(templatesDir)

		err := manager.Delete(name)
		if err != nil {
			return err
		}

		fmt.Printf("模板 '%s' 删除成功\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
