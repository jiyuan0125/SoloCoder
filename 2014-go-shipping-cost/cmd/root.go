package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "shipping-cost",
	Short: "快递运费计算命令行工具",
	Long:  `基于Go和Cobra的快递运费计算命令行工具，支持地址识别、分区定价、批量计算等功能。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.json", "配置文件路径")
}
