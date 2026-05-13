package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"shipping-cost/pkg/calculator"
	"shipping-cost/pkg/config"
	syncpkg "shipping-cost/pkg/sync"
)

var (
	inputFile  string
	outputFile string
)

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量计算快递运费",
	Long:  `从CSV文件批量导入寄件清单，计算运费后导出到新的CSV文件。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Load(cfgFile); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		inFile, err := os.Open(inputFile)
		if err != nil {
			fmt.Printf("无法打开输入文件: %v\n", err)
			os.Exit(1)
		}
		defer inFile.Close()

		reader := csv.NewReader(inFile)

		header, err := reader.Read()
		if err != nil {
			fmt.Printf("无法读取CSV表头: %v\n", err)
			os.Exit(1)
		}

		colIdx := map[string]int{}
		for i, col := range header {
			colIdx[col] = i
		}

		requiredCols := []string{"寄件地址", "收件地址", "重量(kg)", "长度(cm)", "宽度(cm)", "高度(cm)"}
		for _, col := range requiredCols {
			if _, exists := colIdx[col]; !exists {
				fmt.Printf("CSV缺少必要列: %s\n", col)
				os.Exit(1)
			}
		}

		outFile, err := os.Create(outputFile)
		if err != nil {
			fmt.Printf("无法创建输出文件: %v\n", err)
			os.Exit(1)
		}
		defer outFile.Close()

		writer := csv.NewWriter(outFile)
		defer writer.Flush()

		outputHeader := append(header, "分区", "实际重量(kg)", "体积重量(kg)", "计费重量(kg)", "单价(分/kg)", "首重(kg)", "续重单位数", "运费(分)", "运费(元)")
		if err := writer.Write(outputHeader); err != nil {
			fmt.Printf("写入输出文件失败: %v\n", err)
			os.Exit(1)
		}

		successCount := 0
		failCount := 0

		lineNum := 1

		for {
			lineNum++
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("第 %d 行读取错误，跳过: %v\n", lineNum, err)
				failCount++
				continue
			}

			if len(record) < len(requiredCols) {
				fmt.Printf("第 %d 行格式错误，跳过\n", lineNum)
				failCount++
				continue
			}

			weight, err := strconv.ParseFloat(record[colIdx["重量(kg)"]], 64)
			if err != nil {
				fmt.Printf("第 %d 行重量格式错误，跳过\n", lineNum)
				failCount++
				continue
			}

			length, err := strconv.ParseFloat(record[colIdx["长度(cm)"]], 64)
			if err != nil {
				fmt.Printf("第 %d 行长度格式错误，跳过\n", lineNum)
				failCount++
				continue
			}

			width, err := strconv.ParseFloat(record[colIdx["宽度(cm)"]], 64)
			if err != nil {
				fmt.Printf("第 %d 行宽度格式错误，跳过\n", lineNum)
				failCount++
				continue
			}

			height, err := strconv.ParseFloat(record[colIdx["高度(cm)"]], 64)
			if err != nil {
				fmt.Printf("第 %d 行高度格式错误，跳过\n", lineNum)
				failCount++
				continue
			}

			in := &calculator.Input{
				Sender:   record[colIdx["寄件地址"]],
				Receiver: record[colIdx["收件地址"]],
				Weight:   weight,
				Length:   length,
				Width:    width,
				Height:   height,
			}

			result, err := calculator.Calculate(in)
			if err != nil {
				fmt.Printf("第 %d 行计算失败，跳过: %v\n", lineNum, err)
				failCount++
				continue
			}

			outputRow := make([]string, len(outputHeader))
			copy(outputRow, record)

			offset := len(header)
			outputRow[offset] = result.Zone
			outputRow[offset+1] = fmt.Sprintf("%.2f", result.Weight)
			outputRow[offset+2] = fmt.Sprintf("%.2f", result.VolumeWeight)
			outputRow[offset+3] = fmt.Sprintf("%.2f", result.ChargeWeight)
			outputRow[offset+4] = fmt.Sprintf("%d", result.PricePerKG)
			outputRow[offset+5] = fmt.Sprintf("%d", result.FirstWeight)
			outputRow[offset+6] = fmt.Sprintf("%d", result.ContinueWeight)
			outputRow[offset+7] = fmt.Sprintf("%d", result.TotalFen)
			outputRow[offset+8] = fmt.Sprintf("%.2f", result.TotalYuan)

			if err := writer.Write(outputRow); err != nil {
				fmt.Printf("第 %d 行写入输出失败，跳过: %v\n", lineNum, err)
				failCount++
				continue
			}

			successCount++
		}

		fmt.Println("========== 批量处理完成 ==========")
		fmt.Printf("成功处理: %d 行\n", successCount)
		fmt.Printf("失败跳过: %d 行\n", failCount)
		fmt.Printf("输出文件: %s\n", outputFile)
		fmt.Println("==================================")

		syncpkg.OnComplete()
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入CSV文件路径")
	batchCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出CSV文件路径")

	batchCmd.MarkFlagRequired("input")
	batchCmd.MarkFlagRequired("output")
}
