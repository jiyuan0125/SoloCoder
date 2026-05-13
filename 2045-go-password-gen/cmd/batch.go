package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"passwordgen/pkg/generator"
)

var (
	batchCount int
	outputFile string
	batchType  string
	batchLen   int
	batchSpecialChars string
	batchSeparator  string
	batchDictFile   string
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量生成密码",
	Long: `批量生成多个密码并输出到文件，每行一个，确保不重复。最多支持 10000 个。`,
	Run: func(cmd *cobra.Command, args []string) {
		gen := generator.New()

		var pwdType generator.PasswordType
		switch batchType {
		case "numeric":
			pwdType = generator.TypeNumeric
		case "alphanumeric":
			pwdType = generator.TypeAlphanumeric
		case "special":
			pwdType = generator.TypeWithSpecial
		case "readable":
			pwdType = generator.TypeReadable
		default:
			fmt.Println("错误: 未知的密码类型")
			return
		}

		passwords, err := gen.GenerateBatch(batchCount, batchLen, pwdType, batchSpecialChars, batchSeparator, batchDictFile)
		if err != nil {
			fmt.Println("错误:", err)
			return
		}

		if outputFile == "" {
			for _, pwd := range passwords {
				fmt.Println(pwd)
			}
			fmt.Printf("\n已生成 %d 个密码\n", len(passwords))
		} else {
			err = gen.WriteToFile(passwords, outputFile)
			if err != nil {
				fmt.Println("错误:", err)
				return
			}
			fmt.Printf("已成功生成 %d 个密码到文件: %s\n", len(passwords), outputFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.Flags().IntVarP(&batchCount, "number", "n", 10, "要生成的密码数量（最大 10000）")
	batchCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径，不指定则输出到终端")
	batchCmd.Flags().IntVarP(&batchLen, "length", "l", generator.DefaultLength, "密码长度")
	batchCmd.Flags().StringVarP(&batchType, "type", "t", "special", "密码类型: numeric|alphanumeric|special|readable")
	batchCmd.Flags().StringVarP(&batchSpecialChars, "special-chars", "c", generator.DefaultSpecialChars, "自定义特殊字符集")
	batchCmd.Flags().StringVarP(&batchSeparator, "separator", "s", "-", "可读密码的分隔符")
	batchCmd.Flags().StringVarP(&batchDictFile, "dict", "d", "", "可读密码的词典文件路径")
}
