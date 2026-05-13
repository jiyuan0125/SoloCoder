package cmd

import (
	"fmt"
	"os"

	"report-query/internal/storage"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "管理查询历史",
	Long:  `查看或清除查询历史记录`,
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出查询历史",
	Run:   runHistoryList,
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "清除查询历史",
	Run:   runHistoryClear,
}

var historyLimit int

func init() {
	rootCmd.AddCommand(historyCmd)
	historyCmd.AddCommand(historyListCmd)
	historyCmd.AddCommand(historyClearCmd)

	historyListCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "显示最近的 N 条记录")
}

func runHistoryList(cmd *cobra.Command, args []string) {
	store, err := storage.NewStorage()
	if err != nil {
		exitWithError("初始化存储失败: "+err.Error(), 1)
	}

	history, err := store.ListHistory()
	if err != nil {
		exitWithError("获取历史记录失败: "+err.Error(), 1)
	}

	if len(history) == 0 {
		fmt.Println("暂无历史记录")
		return
	}

	start := 0
	if len(history) > historyLimit {
		start = len(history) - historyLimit
	}
	history = history[start:]

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "时间", "SQL"})
	table.SetAutoWrapText(true)
	table.SetColWidth(80)

	for _, h := range history {
		table.Append([]string{
			fmt.Sprintf("%d", h.ID),
			h.Timestamp.Format("2006-01-02 15:04:05"),
			h.Query,
		})
	}

	table.Render()
}

func runHistoryClear(cmd *cobra.Command, args []string) {
	store, err := storage.NewStorage()
	if err != nil {
		exitWithError("初始化存储失败: "+err.Error(), 1)
	}

	if err := store.ClearHistory(); err != nil {
		exitWithError("清除历史记录失败: "+err.Error(), 1)
	}

	fmt.Println("已清除所有历史记录")
}
