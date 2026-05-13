package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"msg-template/internal/template"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有模板",
	Long:  "列出 templates/ 目录下的所有模板文件",
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := template.NewTemplateManager(templatesDir)
		templates, err := manager.List()
		if err != nil {
			return err
		}

		if len(templates) == 0 {
			fmt.Println("暂无模板")
			return nil
		}

		fmt.Println("可用模板：")
		for _, t := range templates {
			fmt.Printf("  - %s\n", t)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
