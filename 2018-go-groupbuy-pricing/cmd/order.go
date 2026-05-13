package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "订单管理命令",
}

var placeOrderCmd = &cobra.Command{
	Use:   "place",
	Short: "参与团购（下单）",
	RunE:  runPlaceOrder,
}

var listOrdersCmd = &cobra.Command{
	Use:   "list",
	Short: "查看用户的订单列表",
	RunE:  runListOrders,
}

func init() {
	rootCmd.AddCommand(orderCmd)
	orderCmd.AddCommand(placeOrderCmd, listOrdersCmd)

	placeOrderCmd.Flags().StringP("activity", "a", "", "活动ID (必填)")
	placeOrderCmd.Flags().StringP("user", "u", "", "用户ID (必填)")
	placeOrderCmd.MarkFlagRequired("activity")
	placeOrderCmd.MarkFlagRequired("user")

	listOrdersCmd.Flags().StringP("user", "u", "", "用户ID (必填)")
	listOrdersCmd.MarkFlagRequired("user")
}

func runPlaceOrder(cmd *cobra.Command, args []string) error {
	activityID, _ := cmd.Flags().GetString("activity")
	userID, _ := cmd.Flags().GetString("user")

	order, err := svc.PlaceOrder(activityID, userID)
	if err != nil {
		return err
	}

	balance, _ := svc.GetBalance(userID)

	fmt.Println("下单成功！")
	fmt.Println("========================================")
	fmt.Printf("订单ID:     %s\n", order.OrderID)
	fmt.Printf("活动ID:     %s\n", order.ActivityID)
	fmt.Printf("用户ID:     %s\n", order.UserID)
	fmt.Printf("支付金额:   %d 分 (%.2f 元)\n", order.PaidPrice, float64(order.PaidPrice)/100)
	fmt.Printf("订单状态:   %s\n", getOrderStatusText(order.Status))
	fmt.Printf("下单时间:   %s\n", order.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("当前余额:   %d 分 (%.2f 元)\n", balance, float64(balance)/100)
	fmt.Println("========================================")
	return nil
}

func runListOrders(cmd *cobra.Command, args []string) error {
	userID, _ := cmd.Flags().GetString("user")

	orders, err := svc.GetOrders(userID)
	if err != nil {
		return err
	}

	if len(orders) == 0 {
		fmt.Println("暂无订单")
		return nil
	}

	fmt.Printf("用户 %s 的订单列表：\n", userID)
	fmt.Println("========================================")
	for i, o := range orders {
		fmt.Printf("[%d]\n", i+1)
		fmt.Printf("  订单ID:   %s\n", o.OrderID)
		fmt.Printf("  活动ID:   %s\n", o.ActivityID)
		fmt.Printf("  支付金额: %d 分 (%.2f 元)\n", o.PaidPrice, float64(o.PaidPrice)/100)
		fmt.Printf("  状态:     %s\n", getOrderStatusText(o.Status))
		fmt.Printf("  时间:     %s\n", o.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("========================================")
	return nil
}

func getOrderStatusText(status string) string {
	switch status {
	case "pending":
		return "待支付"
	case "paid":
		return "已支付"
	case "completed":
		return "已完成"
	case "refunded":
		return "已退款"
	case "cancelled":
		return "已取消"
	default:
		return status
	}
}
