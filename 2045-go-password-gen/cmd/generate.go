package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"passwordgen/pkg/generator"
)

var (
	length      int
	passwordType string
	specialChars string
	separator   string
	dictFile    string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成单个密码",
	Long: `生成指定类型的密码，支持纯数字、字母数字混合、包含特殊字符、可读密码等类型。

密码类型:
  numeric     - 纯数字密码
  alphanumeric - 字母数字混合密码
  special     - 包含特殊字符的密码
  readable    - 可读密码（用单词拼接）`,
	Run: func(cmd *cobra.Command, args []string) {
		gen := generator.New()

		var password string
		var err error

		switch passwordType {
		case "numeric":
			password, err = gen.Generate(length, generator.TypeNumeric, "")
		case "alphanumeric":
			password, err = gen.Generate(length, generator.TypeAlphanumeric, "")
		case "special":
			password, err = gen.Generate(length, generator.TypeWithSpecial, specialChars)
		case "readable":
			password, err = gen.GenerateReadable(length, separator, dictFile)
		default:
			err = fmt.Errorf("未知的密码类型: %s，支持的类型: numeric, alphanumeric, special, readable", passwordType)
		}

		if err != nil {
			fmt.Println("错误:", err)
			return
		}

		fmt.Println(password)

		strength := gen.EvaluateStrength(password)
		fmt.Printf("密码强度: %s (分数: %d)\n", strength.Level, strength.Score)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().IntVarP(&length, "length", "l", generator.DefaultLength, "密码长度（最小 4 位）")
	generateCmd.Flags().StringVarP(&passwordType, "type", "t", "special", "密码类型: numeric|alphanumeric|special|readable")
	generateCmd.Flags().StringVarP(&specialChars, "special-chars", "c", generator.DefaultSpecialChars, "自定义特殊字符集")
	generateCmd.Flags().StringVarP(&separator, "separator", "s", "-", "可读密码的分隔符")
	generateCmd.Flags().StringVarP(&dictFile, "dict", "d", "", "可读密码的词典文件路径")
}
