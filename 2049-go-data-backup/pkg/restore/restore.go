package restore

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"databackup/pkg/models"
	"databackup/pkg/utils"
)

type RestoreConfig struct {
	BackupID     string
	TargetPath   string
	MetadataPath string
	LogPath      string
}

type RestoreManager struct {
	config      RestoreConfig
	record      *models.BackupRecord
	skipped     []string
	restored    []string
	failed      []string
}

func NewRestoreManager(config RestoreConfig) *RestoreManager {
	return &RestoreManager{
		config:  config,
		skipped: []string{},
		restored: []string{},
		failed:   []string{},
	}
}

func (rm *RestoreManager) findBackupRecord() error {
	meta, err := utils.ReadMetadata(rm.config.MetadataPath)
	if err != nil {
		return fmt.Errorf("读取元数据失败: %v", err)
	}

	var targetRecord *models.BackupRecord
	if rm.config.BackupID == "" || rm.config.BackupID == "latest" {
		if len(meta.BackupRecords) == 0 {
			return fmt.Errorf("没有找到备份记录")
		}
		for i := len(meta.BackupRecords) - 1; i >= 0; i-- {
			rec := meta.BackupRecords[i]
			if rec.Status == models.StatusSuccess || rec.Status == models.StatusPartialFail {
				targetRecord = &rec
				break
			}
		}
		if targetRecord == nil {
			return fmt.Errorf("没有找到可用的备份记录")
		}
	} else {
		for i := range meta.BackupRecords {
			if meta.BackupRecords[i].ID == rm.config.BackupID {
				targetRecord = &meta.BackupRecords[i]
				break
			}
		}
		if targetRecord == nil {
			return fmt.Errorf("未找到备份 ID: %s", rm.config.BackupID)
		}
	}

	rm.record = targetRecord
	fmt.Printf("找到备份记录: %s\n", targetRecord.ID)
	fmt.Printf("备份时间: %s\n", targetRecord.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("源目录: %s\n", targetRecord.SourcePath)
	fmt.Printf("备份位置: %s\n", targetRecord.TargetPath)
	fmt.Printf("文件数: %d\n", len(targetRecord.Files))

	return nil
}

func (rm *RestoreManager) Restore() error {
	if err := rm.findBackupRecord(); err != nil {
		return err
	}

	if err := utils.EnsureDirExists(rm.config.TargetPath); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	if rm.record.IsCompressed {
		return rm.restoreFromZip()
	}
	return rm.restoreFromDirectory()
}

func (rm *RestoreManager) restoreFromDirectory() error {
	backupDir := rm.record.TargetPath

	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("备份目录不存在: %s", backupDir)
	}

	for _, fileInfo := range rm.record.Files {
		targetPath := filepath.Join(rm.config.TargetPath, fileInfo.RelativePath)

		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("跳过已存在文件: %s\n", fileInfo.RelativePath)
			rm.skipped = append(rm.skipped, fileInfo.RelativePath)
			continue
		}

		srcPath := filepath.Join(backupDir, fileInfo.RelativePath)
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			fmt.Printf("跳过不存在的备份文件: %s\n", fileInfo.RelativePath)
			rm.failed = append(rm.failed, fileInfo.RelativePath)
			continue
		}

		fmt.Printf("恢复文件: %s\n", fileInfo.RelativePath)

		if err := utils.EnsureDirExists(filepath.Dir(targetPath)); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
			rm.failed = append(rm.failed, fileInfo.RelativePath)
			continue
		}

		if err := rm.copyFile(srcPath, targetPath, fileInfo.Size); err != nil {
			fmt.Printf("恢复失败: %v\n", err)
			rm.failed = append(rm.failed, fileInfo.RelativePath)
			continue
		}

		rm.restored = append(rm.restored, fileInfo.RelativePath)
	}

	rm.writeLog()
	rm.printSummary()
	return nil
}

func (rm *RestoreManager) restoreFromZip() error {
	zipPath := rm.record.TargetPath

	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return fmt.Errorf("备份压缩包不存在: %s", zipPath)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开压缩包失败: %v", err)
	}
	defer reader.Close()

	zipFileMap := make(map[string]*zip.File)
	for _, f := range reader.File {
		zipFileMap[f.Name] = f
	}

	for _, fileInfo := range rm.record.Files {
		targetPath := filepath.Join(rm.config.TargetPath, fileInfo.RelativePath)

		if _, err := os.Stat(targetPath); err == nil {
			fmt.Printf("跳过已存在文件: %s\n", fileInfo.RelativePath)
			rm.skipped = append(rm.skipped, fileInfo.RelativePath)
			continue
		}

		zipFile, ok := zipFileMap[fileInfo.RelativePath]
		if !ok {
			fmt.Printf("跳过不存在的备份文件: %s\n", fileInfo.RelativePath)
			rm.failed = append(rm.failed, fileInfo.RelativePath)
			continue
		}

		fmt.Printf("恢复文件: %s\n", fileInfo.RelativePath)

		if err := utils.EnsureDirExists(filepath.Dir(targetPath)); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
			rm.failed = append(rm.failed, fileInfo.RelativePath)
			continue
		}

		if err := rm.extractFile(zipFile, targetPath); err != nil {
			fmt.Printf("恢复失败: %v\n", err)
			rm.failed = append(rm.failed, fileInfo.RelativePath)
			continue
		}

		rm.restored = append(rm.restored, fileInfo.RelativePath)
	}

	rm.writeLog()
	rm.printSummary()
	return nil
}

func (rm *RestoreManager) copyFile(src, dst string, size int64) error {
	progress := utils.NewProgressBar("  进度", size)
	return utils.CopyFile(src, dst, 1024*1024, func(written, total int64) {
		progress.Update(written, total)
	})
}

func (rm *RestoreManager) extractFile(zipFile *zip.File, dstPath string) error {
	rc, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	progress := utils.NewProgressBar("  进度", zipFile.FileInfo().Size())
	buffer := make([]byte, 1024*1024)
	var totalWritten int64

	for {
		n, readErr := rc.Read(buffer)
		if n > 0 {
			_, writeErr := dstFile.Write(buffer[:n])
			if writeErr != nil {
				return writeErr
			}
			totalWritten += int64(n)
			progress.Update(totalWritten, zipFile.FileInfo().Size())
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}

func (rm *RestoreManager) writeLog() {
	if rm.config.LogPath == "" {
		return
	}

	content := fmt.Sprintf("恢复日志\n")
	content += fmt.Sprintf("备份ID: %s\n", rm.record.ID)
	content += fmt.Sprintf("恢复时间: %s\n", rm.config.TargetPath)
	content += fmt.Sprintf("恢复文件数: %d\n", len(rm.restored))
	content += fmt.Sprintf("跳过文件数: %d\n", len(rm.skipped))
	content += fmt.Sprintf("失败文件数: %d\n", len(rm.failed))

	if len(rm.skipped) > 0 {
		content += fmt.Sprintf("\n跳过的文件:\n")
		content += strings.Join(rm.skipped, "\n")
	}

	if len(rm.failed) > 0 {
		content += fmt.Sprintf("\n失败的文件:\n")
		content += strings.Join(rm.failed, "\n")
	}

	if err := os.WriteFile(rm.config.LogPath, []byte(content), 0644); err != nil {
		fmt.Printf("警告: 写入恢复日志失败: %v\n", err)
	}
}

func (rm *RestoreManager) printSummary() {
	fmt.Printf("\n恢复完成!\n")
	fmt.Printf("成功恢复: %d 个文件\n", len(rm.restored))
	fmt.Printf("跳过 (已存在): %d 个文件\n", len(rm.skipped))
	fmt.Printf("失败: %d 个文件\n", len(rm.failed))
}

func ListBackups(metadataPath string) error {
	meta, err := utils.ReadMetadata(metadataPath)
	if err != nil {
		return fmt.Errorf("读取元数据失败: %v", err)
	}

	if len(meta.BackupRecords) == 0 {
		fmt.Println("没有备份记录")
		return nil
	}

	fmt.Println("备份记录列表:")
	fmt.Println("========================================")
	for i := len(meta.BackupRecords) - 1; i >= 0; i-- {
		rec := meta.BackupRecords[i]
		isLatest := rec.ID == meta.LatestID
		fmt.Printf("\n%s\n", strings.Repeat("-", 40))
		fmt.Printf("备份ID: %s%s\n", rec.ID, map[bool]string{true: " (最新)", false: ""}[isLatest])
		fmt.Printf("开始时间: %s\n", rec.StartTime.Format("2006-01-02 15:04:05"))
		if !rec.EndTime.IsZero() {
			fmt.Printf("结束时间: %s\n", rec.EndTime.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("源目录: %s\n", rec.SourcePath)
		fmt.Printf("目标位置: %s\n", rec.TargetPath)
		fmt.Printf("压缩包: %v\n", rec.IsCompressed)
		fmt.Printf("状态: %s\n", rec.Status)
		fmt.Printf("文件数: %d\n", len(rec.Files))
		if len(rec.FailedFiles) > 0 {
			fmt.Printf("失败文件数: %d\n", len(rec.FailedFiles))
		}
	}

	return nil
}
