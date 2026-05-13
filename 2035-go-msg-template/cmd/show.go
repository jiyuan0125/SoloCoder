package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"msg-template/internal/template"
)

var showCmd = &cobra.Command{
	Use:   "show [名称]",
	Short: "显示模板内容",
	Long:  "显示指定模板的原始内容。",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		manager := template.NewTemplateManager(templatesDir)

		content, err := manager.Read(name)
		if err != nil {
			return err
		}

		fmt.Print(content)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
