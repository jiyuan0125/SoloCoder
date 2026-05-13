package cmd

import (
	"fmt"
	"groupbuy/models"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var activityCmd = &cobra.Command{
	Use:   "activity",
	Short: "活动管理命令",
}

var createActivityCmd = &cobra.Command{
	Use:   "create",
	Short: "创建团购活动",
	RunE:  runCreateActivity,
}

var listActivityCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有活动",
	RunE:  runListActivities,
}

var getActivityCmd = &cobra.Command{
	Use:   "get",
	Short: "查看活动详情",
	RunE:  runGetActivity,
}

func init() {
	rootCmd.AddCommand(activityCmd)
	activityCmd.AddCommand(createActivityCmd, listActivityCmd, getActivityCmd)

	createActivityCmd.Flags().StringP("name", "n", "", "活动名称 (必填)")
	createActivityCmd.Flags().StringP("price", "p", "", "基础价格（分，必填）")
	createActivityCmd.Flags().StringP("deadline", "d", "", "截止时间，格式：2006-01-02 15:04:05 (必填)")
	createActivityCmd.MarkFlagRequired("name")
	createActivityCmd.MarkFlagRequired("price")
	createActivityCmd.MarkFlagRequired("deadline")

	getActivityCmd.Flags().StringP("id", "i", "", "活动ID (必填)")
	getActivityCmd.MarkFlagRequired("id")
}

func runCreateActivity(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	priceStr, _ := cmd.Flags().GetString("price")
	deadlineStr, _ := cmd.Flags().GetString("deadline")

	price, err := strconv.ParseInt(priceStr, 10, 64)
	if err != nil || price <= 0 {
		return fmt.Errorf("价格必须是正整数（单位：分）")
	}

	deadline, err := time.ParseInLocation("2006-01-02 15:04:05", deadlineStr, time.Local)
	if err != nil {
		return fmt.Errorf("时间格式错误，请使用：2006-01-02 15:04:05")
	}

	if deadline.Before(time.Now()) {
		return fmt.Errorf("截止时间必须晚于当前时间")
	}

	activity, err := svc.CreateActivity(name, price, deadline, nil)
	if err != nil {
		return err
	}

	fmt.Println("活动创建成功！")
	fmt.Println("========================================")
	fmt.Printf("活动ID:   %s\n", activity.ActivityID)
	fmt.Printf("活动名称: %s\n", activity.Name)
	fmt.Printf("基础价格: %d 分 (%.2f 元)\n", activity.BasePrice, float64(activity.BasePrice)/100)
	fmt.Printf("截止时间: %s\n", activity.Deadline.Format("2006-01-02 15:04:05"))
	fmt.Println("阶梯规则:")
	for _, t := range activity.Tiers {
		fmt.Printf("  - 满 %d 人 %0.0f 折\n", t.MinPeople, t.Discount*10)
	}
	fmt.Println("========================================")
	return nil
}

func runListActivities(cmd *cobra.Command, args []string) error {
	activities, err := svc.ListActivities()
	if err != nil {
		return err
	}

	if len(activities) == 0 {
		fmt.Println("暂无活动")
		return nil
	}

	fmt.Println("活动列表：")
	fmt.Println("========================================")
	for i, a := range activities {
		fmt.Printf("[%d]\n", i+1)
		fmt.Printf("  ID:     %s\n", a.ActivityID)
		fmt.Printf("  名称:   %s\n", a.Name)
		fmt.Printf("  状态:   %s\n", getActivityStatusText(a.Status))
		fmt.Printf("  价格:   %d 分 (%.2f 元)\n", a.BasePrice, float64(a.BasePrice)/100)
		fmt.Printf("  参与人数: %d 人\n", a.ParticipantCount)
		fmt.Printf("  当前阶梯: %s\n", getCurrentTierText(a.CurrentTier, a.Tiers))
		fmt.Printf("  截止时间: %s\n", a.Deadline.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("========================================")
	return nil
}

func runGetActivity(cmd *cobra.Command, args []string) error {
	id, _ := cmd.Flags().GetString("id")

	activity, err := svc.GetActivity(id)
	if err != nil {
		return err
	}

	fmt.Println("活动详情：")
	fmt.Println("========================================")
	fmt.Printf("活动ID:     %s\n", activity.ActivityID)
	fmt.Printf("活动名称:   %s\n", activity.Name)
	fmt.Printf("状态:       %s\n", getActivityStatusText(activity.Status))
	fmt.Printf("基础价格:   %d 分 (%.2f 元)\n", activity.BasePrice, float64(activity.BasePrice)/100)
	fmt.Printf("参与人数:   %d 人\n", activity.ParticipantCount)
	fmt.Printf("当前阶梯:   %s\n", getCurrentTierText(activity.CurrentTier, activity.Tiers))
	fmt.Printf("创建时间:   %s\n", activity.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("截止时间:   %s\n", activity.Deadline.Format("2006-01-02 15:04:05"))
	if activity.SettledAt != nil {
		fmt.Printf("结算时间:   %s\n", activity.SettledAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("阶梯规则:")
	for _, t := range activity.Tiers {
		fmt.Printf("  - 满 %d 人 %0.0f 折\n", t.MinPeople, t.Discount*10)
	}
	fmt.Println("========================================")
	return nil
}

func getActivityStatusText(status string) string {
	switch status {
	case "active":
		return "进行中"
	case "completed":
		return "已完成"
	case "cancelled":
		return "已取消"
	default:
		return status
	}
}

func getCurrentTierText(tierIndex int, tiers []models.Tier) string {
	if tierIndex < 0 || tierIndex >= len(tiers) {
		return "未达最低阶梯（10人）"
	}
	t := tiers[tierIndex]
	return fmt.Sprintf("满 %d 人 %0.0f 折", t.MinPeople, t.Discount*10)
}
