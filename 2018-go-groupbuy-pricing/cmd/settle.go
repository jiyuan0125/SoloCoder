package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var settleCmd = &cobra.Command{
	Use:   "settle",
	Short: "活动结算命令",
}

var settleActivityCmd = &cobra.Command{
	Use:   "activity",
	Short: "结算活动（活动截止后执行）",
	RunE:  runSettleActivity,
}

func init() {
	rootCmd.AddCommand(settleCmd)
	settleCmd.AddCommand(settleActivityCmd)

	settleActivityCmd.Flags().StringP("id", "i", "", "活动ID (必填)")
	settleActivityCmd.MarkFlagRequired("id")
}

func runSettleActivity(cmd *cobra.Command, args []string) error {
	activityID, _ := cmd.Flags().GetString("id")

	activityBefore, err := svc.GetActivity(activityID)
	if err != nil {
		return err
	}

	participantCount := activityBefore.ParticipantCount

	if err := svc.SettleActivity(activityID); err != nil {
		return err
	}

	activityAfter, err := svc.GetActivity(activityID)
	if err != nil {
		return err
	}

	fmt.Println("活动结算完成！")
	fmt.Println("========================================")
	fmt.Printf("活动ID:     %s\n", activityAfter.ActivityID)
	fmt.Printf("活动名称:   %s\n", activityAfter.Name)
	fmt.Printf("参与人数:   %d 人\n", participantCount)
	fmt.Printf("最终状态:   %s\n", getActivityStatusText(activityAfter.Status))
	if activityAfter.Status == "cancelled" {
		fmt.Println("结算结果:   未达到最低10人，活动已取消，所有参与者全额退款")
	} else {
		fmt.Printf("最终阶梯:   %s\n", getCurrentTierText(activityAfter.CurrentTier, activityAfter.Tiers))
		fmt.Println("结算结果:   已按最终阶梯价结算，差价已退还至各用户余额")
	}
	if activityAfter.SettledAt != nil {
		fmt.Printf("结算时间:   %s\n", activityAfter.SettledAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("========================================")
	return nil
}
