package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"databackup/pkg/backup"
	"databackup/pkg/restore"
	"databackup/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "databackup",
		Short: "数据备份工具",
		Long:  "一个功能强大的数据备份工具，支持增量备份、断点续传、大文件分块、压缩等功能",
	}

	backupCmd = &cobra.Command{
		Use:   "backup",
		Short: "备份指定目录",
		Long:  "备份源目录到目标位置（本地目录或压缩包）",
		RunE:  runBackup,
	}

	restoreCmd = &cobra.Command{
		Use:   "restore",
		Short: "恢复备份",
		Long:  "根据备份记录恢复文件到指定目录",
		RunE:  runRestore,
	}

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "列出备份记录",
		Long:  "显示所有备份记录的详细信息",
		RunE:  runList,
	}

	sourcePath    string
	targetPath    string
	compressed    bool
	incremental   bool
	metadataPath  string
	resumePath    string
	errorReport   string
	backupID      string
	restoreTarget string
	restoreLog    string
)

func init() {
	homeDir, _ := os.UserHomeDir()
	defaultMeta := filepath.Join(homeDir, ".databackup", "metadata.json")
	defaultResume := filepath.Join(homeDir, ".databackup", "resume.json")

	backupCmd.Flags().StringVarP(&sourcePath, "source", "s", "", "源目录路径 (必填)")
	backupCmd.Flags().StringVarP(&targetPath, "target", "t", "", "目标路径 (必填)")
	backupCmd.Flags().BoolVarP(&compressed, "compressed", "c", false, "备份到压缩包 (.zip)")
	backupCmd.Flags().BoolVarP(&incremental, "incremental", "i", false, "增量备份模式")
	backupCmd.Flags().StringVar(&metadataPath, "metadata", defaultMeta, "元数据文件路径")
	backupCmd.Flags().StringVar(&resumePath, "resume", defaultResume, "断点信息文件路径")
	backupCmd.Flags().StringVar(&errorReport, "error-report", "", "错误报告文件路径")

	backupCmd.MarkFlagRequired("source")
	backupCmd.MarkFlagRequired("target")

	restoreCmd.Flags().StringVarP(&backupID, "id", "i", "latest", "要恢复的备份ID (latest 表示最新)")
	restoreCmd.Flags().StringVarP(&restoreTarget, "target", "t", "", "恢复目标目录 (必填)")
	restoreCmd.Flags().StringVar(&metadataPath, "metadata", defaultMeta, "元数据文件路径")
	restoreCmd.Flags().StringVar(&restoreLog, "log", "", "恢复日志文件路径")

	restoreCmd.MarkFlagRequired("target")

	listCmd.Flags().StringVar(&metadataPath, "metadata", defaultMeta, "元数据文件路径")

	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(listCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runBackup(cmd *cobra.Command, args []string) error {
	if err := utils.EnsureDirExists(filepath.Dir(metadataPath)); err != nil {
		return err
	}

	config := backup.BackupConfig{
		SourcePath:   sourcePath,
		TargetPath:   targetPath,
		IsCompressed: compressed,
		Incremental:  incremental,
		MetadataPath: metadataPath,
		ResumePath:   resumePath,
		ErrorReport:  errorReport,
	}

	manager := backup.NewBackupManager(config)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		println("\n\n检测到中断信号，正在保存断点...")
		manager.HandleInterrupt()
	}()

	return manager.Backup()
}

func runRestore(cmd *cobra.Command, args []string) error {
	config := restore.RestoreConfig{
		BackupID:     backupID,
		TargetPath:   restoreTarget,
		MetadataPath: metadataPath,
		LogPath:      restoreLog,
	}

	manager := restore.NewRestoreManager(config)
	return manager.Restore()
}

func runList(cmd *cobra.Command, args []string) error {
	return restore.ListBackups(metadataPath)
}
