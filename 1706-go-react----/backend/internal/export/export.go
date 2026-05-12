package main

import (
	"encoding/csv"
	"fmt"
	"hospital-pharmacy/pkg/database"
	"hospital-pharmacy/pkg/models"
	"hospital-pharmacy/pkg/utils"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "export",
	Short: "药房管理系统数据导出工具",
}

var stockCmd = &cobra.Command{
	Use:   "stock [filename]",
	Short: "导出当前库存明细",
	Run: func(cmd *cobra.Command, args []string) {
		filename := "stock.csv"
		if len(args) > 0 {
			filename = args[0]
		}
		exportStock(filename)
	},
}

var transactionsCmd = &cobra.Command{
	Use:   "transactions [filename]",
	Short: "导出指定月份的出入库流水",
	Run: func(cmd *cobra.Command, args []string) {
		filename := "transactions.csv"
		if len(args) > 0 {
			filename = args[0]
		}
		month, _ := cmd.Flags().GetString("month")
		exportTransactions(filename, month)
	},
}

var expiryCmd = &cobra.Command{
	Use:   "expiry [filename]",
	Short: "导出近效期和过期药品清单",
	Run: func(cmd *cobra.Command, args []string) {
		filename := "expiry.csv"
		if len(args) > 0 {
			filename = args[0]
		}
		exportExpiry(filename)
	},
}

func init() {
	transactionsCmd.Flags().String("month", "", "月份 (格式: YYYY-MM)")
	rootCmd.AddCommand(stockCmd)
	rootCmd.AddCommand(transactionsCmd)
	rootCmd.AddCommand(expiryCmd)
}

func main() {
	if err := database.Init(); err != nil {
		fmt.Println("数据库初始化失败:", err)
		os.Exit(1)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func exportStock(filename string) {
	var drugs []models.Drug
	if err := database.DB.Preload("Category").Find(&drugs).Error; err != nil {
		fmt.Println("查询库存失败:", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"药品编码", "通用名", "商品名", "类别", "规格", "剂型", "单位", "库存数量", "零售价", "库存金额"})

	for _, drug := range drugs {
		categoryName := ""
		if drug.Category.Name != "" {
			categoryName = drug.Category.Name
		}
		stockValue := float64(drug.CurrentStock) * drug.RetailPrice
		writer.Write([]string{
			drug.DrugCode,
			drug.GenericName,
			drug.BrandName,
			categoryName,
			drug.Specification,
			string(drug.DosageForm),
			drug.Unit,
			strconv.Itoa(drug.CurrentStock),
			fmt.Sprintf("%.2f", drug.RetailPrice),
			fmt.Sprintf("%.2f", stockValue),
		})
	}

	fmt.Printf("库存明细已导出到: %s\n", filename)
}

func exportTransactions(filename, month string) {
	var transactions []models.StockTransaction
	query := database.DB.Preload("Drug")

	if month != "" {
		t, err := time.Parse("2006-01", month)
		if err != nil {
			fmt.Println("月份格式错误，应为 YYYY-MM")
			return
		}
		startOfMonth := t
		endOfMonth := t.AddDate(0, 1, 0)
		query = query.Where("created_at >= ? AND created_at < ?", startOfMonth, endOfMonth)
	}

	if err := query.Order("created_at DESC").Find(&transactions).Error; err != nil {
		fmt.Println("查询流水失败:", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"时间", "类型", "药品编码", "药品名称", "数量", "批号", "供应商", "操作人1", "操作人2"})

	for _, t := range transactions {
		transType := "入库"
		if t.TransactionType == "out" {
			transType = "出库"
		}
		drugCode := ""
		drugName := ""
		if t.Drug.ID > 0 {
			drugCode = t.Drug.DrugCode
			drugName = t.Drug.GenericName
		}
		writer.Write([]string{
			t.CreatedAt.Format("2006-01-02 15:04:05"),
			transType,
			drugCode,
			drugName,
			strconv.Itoa(t.Quantity),
			t.BatchNumber,
			t.Supplier,
			t.Operator1,
			t.Operator2,
		})
	}

	fmt.Printf("出入库流水已导出到: %s\n", filename)
}

func exportExpiry(filename string) {
	var items []models.StockItem
	now := time.Now()
	nearDate := now.AddDate(0, 0, 180)
	if err := database.DB.Preload("Drug").
		Where("expiry_date <= ? AND quantity > 0", nearDate).
		Order("expiry_date ASC").
		Find(&items).Error; err != nil {
		fmt.Println("查询效期药品失败:", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("创建文件失败:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"药品编码", "药品名称", "批号", "库存数量", "生产日期", "有效期", "效期状态"})

	for _, item := range items {
		status := utils.GetExpiryStatus(item.ExpiryDate)
		statusText := "正常"
		switch status {
		case "near":
			statusText = "近效期"
		case "critical":
			statusText = "临期"
		case "expired":
			statusText = "过期"
		}
		writer.Write([]string{
			item.Drug.DrugCode,
			item.Drug.GenericName,
			item.BatchNumber,
			strconv.Itoa(item.Quantity),
			item.ProductionDate.Format("2006-01-02"),
			item.ExpiryDate.Format("2006-01-02"),
			statusText,
		})
	}

	fmt.Printf("近效期药品清单已导出到: %s\n", filename)
}
