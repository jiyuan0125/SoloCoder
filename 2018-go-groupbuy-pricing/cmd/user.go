package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "用户余额和提现命令",
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "查看用户余额",
	RunE:  runBalance,
}

var withdrawCmd = &cobra.Command{
	Use:   "withdraw",
	Short: "提现余额（扣1%手续费，最低0.1元）",
	RunE:  runWithdraw,
}

var refundHistoryCmd = &cobra.Command{
	Use:   "refunds",
	Short: "查看退款记录",
	RunE:  runRefundHistory,
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(balanceCmd, withdrawCmd, refundHistoryCmd)

	balanceCmd.Flags().StringP("user", "u", "", "用户ID (必填)")
	balanceCmd.MarkFlagRequired("user")

	withdrawCmd.Flags().StringP("user", "u", "", "用户ID (必填)")
	withdrawCmd.Flags().StringP("amount", "a", "", "提现金额（分，必填）")
	withdrawCmd.MarkFlagRequired("user")
	withdrawCmd.MarkFlagRequired("amount")

	refundHistoryCmd.Flags().StringP("user", "u", "", "用户ID (必填)")
	refundHistoryCmd.MarkFlagRequired("user")
}

func runBalance(cmd *cobra.Command, args []string) error {
	userID, _ := cmd.Flags().GetString("user")

	balance, err := svc.GetBalance(userID)
	if err != nil {
		return err
	}

	fmt.Println("余额查询：")
	fmt.Println("========================================")
	fmt.Printf("用户ID:     %s\n", userID)
	fmt.Printf("当前余额:   %d 分 (%.2f 元)\n", balance, float64(balance)/100)
	fmt.Println("========================================")
	return nil
}

func runWithdraw(cmd *cobra.Command, args []string) error {
	userID, _ := cmd.Flags().GetString("user")
	amountStr, _ := cmd.Flags().GetString("amount")

	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || amount <= 0 {
		return fmt.Errorf("金额必须是正整数（单位：分）")
	}

	netAmount, fee, err := svc.Withdraw(userID, amount)
	if err != nil {
		return err
	}

	newBalance, _ := svc.GetBalance(userID)

	fmt.Println("提现成功！")
	fmt.Println("========================================")
	fmt.Printf("用户ID:       %s\n", userID)
	fmt.Printf("提现金额:     %d 分 (%.2f 元)\n", amount, float64(amount)/100)
	fmt.Printf("手续费:       %d 分 (%.2f 元)\n", fee, float64(fee)/100)
	fmt.Printf("实际到账:     %d 分 (%.2f 元)\n", netAmount, float64(netAmount)/100)
	fmt.Printf("剩余余额:     %d 分 (%.2f 元)\n", newBalance, float64(newBalance)/100)
	fmt.Println("========================================")
	return nil
}

func runRefundHistory(cmd *cobra.Command, args []string) error {
	userID, _ := cmd.Flags().GetString("user")

	records, err := svc.GetRefundRecords(userID)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Println("暂无退款记录")
		return nil
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].CreatedAt.Before(records[j].CreatedAt)
	})

	totalRefund := int64(0)
	fmt.Printf("用户 %s 的退款记录：\n", userID)
	fmt.Println("========================================")
	for i, r := range records {
		totalRefund += r.Amount
		fmt.Printf("[%d]\n", i+1)
		fmt.Printf("  退款ID:   %s\n", r.RefundID)
		fmt.Printf("  活动ID:   %s\n", r.ActivityID)
		fmt.Printf("  订单ID:   %s\n", r.OrderID)
		fmt.Printf("  退款金额: %d 分 (%.2f 元)\n", r.Amount, float64(r.Amount)/100)
		fmt.Printf("  退款原因: %s\n", r.Reason)
		fmt.Printf("  退款时间: %s\n", r.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("累计退款: %d 分 (%.2f 元)\n", totalRefund, float64(totalRefund)/100)
	fmt.Println("========================================")
	return nil
}
