package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"msg-template/internal/template"
)

var updateContent string
var updateFromFile string

var updateCmd = &cobra.Command{
	Use:   "update [名称]",
	Short: "更新现有模板",
	Long:  "更新已存在的消息模板。内容可以通过 -c、-f 参数或标准输入提供。",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		manager := template.NewTemplateManager(templatesDir)

		var content string

		if updateFromFile != "" {
			data, err := ioutil.ReadFile(updateFromFile)
			if err != nil {
				return err
			}
			content = string(data)
		} else if updateContent != "" {
			content = updateContent
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				content = string(data)
			} else {
				return errors.New("请通过 -c、-f 参数或标准输入提供模板内容")
			}
		}

		err := manager.Update(name, content)
		if err != nil {
			return err
		}

		fmt.Printf("模板 '%s' 更新成功\n", name)
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&updateContent, "content", "c", "", "模板内容字符串")
	updateCmd.Flags().StringVarP(&updateFromFile, "file", "f", "", "从文件读取模板内容")
	rootCmd.AddCommand(updateCmd)
}
