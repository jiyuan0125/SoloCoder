package cmd

import (
	"groupbuy/service"
	"groupbuy/storage"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var svc *service.Service
var dataDir string

var rootCmd = &cobra.Command{
	Use:   "groupbuy",
	Short: "团购阶梯价命令行工具",
	Long: `一个基于 Go + cobra 的团购阶梯价命令行工具，支持：
- 团购活动管理（创建、列表、查看）
- 阶梯价规则（满10人9折，满50人8折，满100人7折，满500人6折）
- 订单管理（下单、查看订单）
- 差价自动退还到虚拟余额
- 活动结算（未达最低人数自动取消并全额退款）
- 余额提现（1%手续费，最低0.1元）
- 退款记录查询`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initService)
	rootCmd.PersistentFlags().StringVar(&dataDir, "data", getDataDir(), "数据存储目录")
}

func getDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".groupbuy")
}

func initService() {
	dataFile := filepath.Join(dataDir, "data.json")
	store := storage.NewStorage(dataFile)
	svc = service.NewService(store)
}
