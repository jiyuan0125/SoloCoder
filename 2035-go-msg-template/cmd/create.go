package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"msg-template/internal/template"
)

var createContent string
var createFromFile string

var createCmd = &cobra.Command{
	Use:   "create [名称]",
	Short: "创建新模板",
	Long: `创建一个新的消息模板。
模板名称不能重复。
内容可以通过 -c 参数直接指定，或通过 -f 参数从文件读取，或通过标准输入传入。`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		manager := template.NewTemplateManager(templatesDir)

		var content string
		var err error

		if createFromFile != "" {
			data, readErr := ioutil.ReadFile(createFromFile)
			if readErr != nil {
				return readErr
			}
			content = string(data)
		} else if createContent != "" {
			content = createContent
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, readErr := ioutil.ReadAll(os.Stdin)
				if readErr != nil {
					return readErr
				}
				content = string(data)
			} else {
				return errors.New("请通过 -c、-f 参数或标准输入提供模板内容")
			}
		}

		err = manager.Create(name, content)
		if err != nil {
			return err
		}

		fmt.Printf("模板 '%s' 创建成功\n", name)
		return nil
	},
}

func init() {
	createCmd.Flags().StringVarP(&createContent, "content", "c", "", "模板内容字符串")
	createCmd.Flags().StringVarP(&createFromFile, "file", "f", "", "从文件读取模板内容")
	rootCmd.AddCommand(createCmd)
}
