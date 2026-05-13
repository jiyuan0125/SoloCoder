package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"shipping-cost/pkg/address"
	"shipping-cost/pkg/calculator"
	"shipping-cost/pkg/config"
	syncpkg "shipping-cost/pkg/sync"
)

var (
	senderAddr   string
	receiverAddr string
	weight       float64
	length       float64
	width        float64
	height       float64
)

var calculateCmd = &cobra.Command{
	Use:   "calculate",
	Short: "计算单条快递运费",
	Long:  `输入寄件地址、收件地址、重量和尺寸，计算并输出运费。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Load(cfgFile); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		in := &calculator.Input{
			Sender:   senderAddr,
			Receiver: receiverAddr,
			Weight:   weight,
			Length:   length,
			Width:    width,
			Height:   height,
		}

		result, err := calculator.Calculate(in)
		if err != nil {
			if err == address.ErrAddressNotFound {
				fmt.Println("无法识别该地址")
			} else {
				fmt.Println("计算出错:", err)
			}
			os.Exit(1)
		}

		fmt.Println("========== 运费计算结果 ==========")
		fmt.Printf("寄件地址: %s\n", senderAddr)
		fmt.Printf("收件地址: %s\n", receiverAddr)
		fmt.Printf("分区: %s\n", result.Zone)
		fmt.Printf("实际重量: %.2f kg\n", result.Weight)
		fmt.Printf("体积重量: %.2f kg\n", result.VolumeWeight)
		fmt.Printf("计费重量: %.2f kg\n", result.ChargeWeight)
		fmt.Printf("计费单价: %d 分/kg\n", result.PricePerKG)
		fmt.Printf("首重: %d kg\n", result.FirstWeight)
		fmt.Printf("续重单位数: %d (每0.5kg)\n", result.ContinueWeight)
		fmt.Printf("运费总计: %d 分 (%.2f 元)\n", result.TotalFen, result.TotalYuan)
		fmt.Println("==================================")

		syncpkg.OnComplete()
	},
}

func init() {
	rootCmd.AddCommand(calculateCmd)

	calculateCmd.Flags().StringVarP(&senderAddr, "sender", "s", "", "寄件地址")
	calculateCmd.Flags().StringVarP(&receiverAddr, "receiver", "r", "", "收件地址")
	calculateCmd.Flags().Float64VarP(&weight, "weight", "w", 0.0, "实际重量 (kg)")
	calculateCmd.Flags().Float64VarP(&length, "length", "l", 0.0, "长度 (cm)")
	calculateCmd.Flags().Float64VarP(&width, "width", "W", 0.0, "宽度 (cm)")
	calculateCmd.Flags().Float64VarP(&height, "height", "H", 0.0, "高度 (cm)")

	calculateCmd.MarkFlagRequired("sender")
	calculateCmd.MarkFlagRequired("receiver")
	calculateCmd.MarkFlagRequired("weight")
	calculateCmd.MarkFlagRequired("length")
	calculateCmd.MarkFlagRequired("width")
	calculateCmd.MarkFlagRequired("height")
}
