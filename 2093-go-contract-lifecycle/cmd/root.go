package cmd

import (
	"os"
	"path/filepath"

	"contract-lifecycle/service"
	"contract-lifecycle/storage"

	"github.com/spf13/cobra"
)

var (
	dataDir  string
	store    *storage.Storage
	svc      *service.ContractService
	resourceID string
)

var rootCmd = &cobra.Command{
	Use:   "contract",
	Short: "合同全生命周期管理系统",
	Long:  `一个基于命令行的合同全生命周期管理系统，支持合同创建、状态流转、到期提醒、资源管理等功能。`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initStorage)

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	rootCmd.PersistentFlags().StringVar(&dataDir, "data-dir", filepath.Join(home, ".contract-data"), "数据存储目录")
	rootCmd.PersistentFlags().StringVar(&resourceID, "resource", "", "关联的资源ID")

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(remindCmd)
	rootCmd.AddCommand(terminateCmd)
	rootCmd.AddCommand(renewCmd)
	rootCmd.AddCommand(breachCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(resourceCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(summaryCmd)
}

func initStorage() {
	store = storage.NewStorage(dataDir)
	svc = service.NewContractService(store)
}
