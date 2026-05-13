package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"passwordgen/pkg/generator"
)

var strengthCmd = &cobra.Command{
	Use:   "strength [密码]",
	Short: "评估密码强度",
	Long: `评估给定密码的强度，输出弱/中/强三个等级。
分析因素包括：密码长度、字符种类、常见模式（连续数字、重复字符）等。`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		password := args[0]
		gen := generator.New()

		result := gen.EvaluateStrength(password)

		fmt.Printf("密码: %s\n", password)
		fmt.Printf("强度等级: %s\n", result.Level)
		fmt.Printf("总分数: %d/100\n", result.Score)
		fmt.Println("详细得分:")
		fmt.Printf("  - 长度得分: %d\n", result.Details["length"])
		fmt.Printf("  - 字符类型得分: %d\n", result.Details["char_types"])
		fmt.Printf("  - 模式检测得分: %d\n", result.Details["patterns"])
	},
}

func init() {
	rootCmd.AddCommand(strengthCmd)
}
