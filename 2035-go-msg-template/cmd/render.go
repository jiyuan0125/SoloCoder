package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"msg-template/internal/template"
)

var renderDataFile string
var renderOutputFile string

var renderCmd = &cobra.Command{
	Use:   "render [模板名称]",
	Short: "渲染模板",
	Long: `渲染消息模板，使用 JSON 数据替换变量。

数据来源优先级：
  1. --data 参数指定的 JSON 文件
  2. 标准输入（如果有管道输入）

输出：
  - 默认输出到标准输出
  - 使用 -o 参数输出到文件`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateName := args[0]
		manager := template.NewTemplateManager(templatesDir)

		var data map[string]interface{}
		var err error

		if renderDataFile != "" {
			data, err = template.ReadJSONData(renderDataFile)
			if err != nil {
				return fmt.Errorf("读取数据文件失败: %w", err)
			}
		} else {
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				data, err = template.ReadJSONData("-")
				if err != nil {
					return fmt.Errorf("读取标准输入数据失败: %w", err)
				}
			} else {
				data = make(map[string]interface{})
			}
		}

		var output *os.File
		if renderOutputFile != "" {
			output, err = os.Create(renderOutputFile)
			if err != nil {
				return fmt.Errorf("创建输出文件失败: %w", err)
			}
			defer output.Close()
		} else {
			output = os.Stdout
		}

		if err := manager.Render(templateName, data, output); err != nil {
			return err
		}

		if renderOutputFile != "" {
			fmt.Fprintf(os.Stderr, "渲染结果已写入: %s\n", renderOutputFile)
		}

		return nil
	},
}

func init() {
	renderCmd.Flags().StringVar(&renderDataFile, "data", "", "JSON 数据文件路径（- 表示标准输入）")
	renderCmd.Flags().StringVarP(&renderOutputFile, "output", "o", "", "输出文件路径（默认为标准输出）")
	rootCmd.AddCommand(renderCmd)

	_ = ioutil.Discard
}
