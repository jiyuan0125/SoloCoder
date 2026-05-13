package cmd

import (
	"fmt"
	"os"

	"report-query/internal/storage"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var shortcutCmd = &cobra.Command{
	Use:   "shortcut",
	Short: "管理查询快捷方式",
	Long:  `列出、查看或删除已保存的查询快捷方式`,
}

var shortcutListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有快捷方式",
	Run:   runShortcutList,
}

var shortcutShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "显示快捷方式详情",
	Args:  cobra.ExactArgs(1),
	Run:   runShortcutShow,
}

var shortcutDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "删除快捷方式",
	Args:  cobra.ExactArgs(1),
	Run:   runShortcutDelete,
}

func init() {
	rootCmd.AddCommand(shortcutCmd)
	shortcutCmd.AddCommand(shortcutListCmd)
	shortcutCmd.AddCommand(shortcutShowCmd)
	shortcutCmd.AddCommand(shortcutDeleteCmd)
}

func runShortcutList(cmd *cobra.Command, args []string) {
	store, err := storage.NewStorage()
	if err != nil {
		exitWithError("初始化存储失败: "+err.Error(), 1)
	}

	shortcuts, err := store.ListShortcuts()
	if err != nil {
		exitWithError("获取快捷方式失败: "+err.Error(), 1)
	}

	if len(shortcuts) == 0 {
		fmt.Println("暂无快捷方式")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"名称", "数据库", "表", "字段", "创建时间"})

	for _, sc := range shortcuts {
		table.Append([]string{
			sc.Name,
			sc.DBPath,
			sc.Table,
			sc.Fields,
			sc.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	table.Render()
}

func runShortcutShow(cmd *cobra.Command, args []string) {
	name := args[0]
	store, err := storage.NewStorage()
	if err != nil {
		exitWithError("初始化存储失败: "+err.Error(), 1)
	}

	sc, err := store.GetShortcut(name)
	if err != nil {
		exitWithError(err.Error(), 1)
	}

	fmt.Printf("名称: %s\n", sc.Name)
	fmt.Printf("数据库: %s\n", sc.DBPath)
	fmt.Printf("表: %s\n", sc.Table)
	fmt.Printf("字段: %s\n", sc.Fields)
	if sc.Conditions != "" {
		fmt.Printf("条件: %s\n", sc.Conditions)
	}
	if sc.Where != "" {
		fmt.Printf("WHERE: %s\n", sc.Where)
	}
	if sc.GroupBy != "" {
		fmt.Printf("GROUP BY: %s\n", sc.GroupBy)
	}
	fmt.Printf("创建时间: %s\n", sc.CreatedAt.Format("2006-01-02 15:04:05"))
}

func runShortcutDelete(cmd *cobra.Command, args []string) {
	name := args[0]
	store, err := storage.NewStorage()
	if err != nil {
		exitWithError("初始化存储失败: "+err.Error(), 1)
	}

	if err := store.DeleteShortcut(name); err != nil {
		exitWithError(err.Error(), 1)
	}

	fmt.Printf("已删除快捷方式: %s\n", name)
}
