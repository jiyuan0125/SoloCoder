package backup

import (
	"archive/zip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"databackup/pkg/models"
	"databackup/pkg/utils"
)

const (
	ChunkSize           = 64 * 1024 * 1024
	LargeFileThreshold  = 1 * 1024 * 1024 * 1024
	SmallFileChunkSize  = 1024 * 1024
)

type BackupConfig struct {
	SourcePath   string
	TargetPath   string
	IsCompressed bool
	Incremental  bool
	MetadataPath string
	ResumePath   string
	ErrorReport  string
}

type BackupManager struct {
	config       BackupConfig
	ctx          context.Context
	cancel       context.CancelFunc
	record       *models.BackupRecord
	resumeInfo   *models.ResumeInfo
	completedSet map[string]bool
	completedMu  sync.Mutex
	interrupted  bool
}

func NewBackupManager(config BackupConfig) *BackupManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackupManager{
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		completedSet: make(map[string]bool),
	}
}

func (bm *BackupManager) HandleInterrupt() {
	bm.interrupted = true
	bm.cancel()
}

func (bm *BackupManager) checkInterrupted() bool {
	select {
	case <-bm.ctx.Done():
		return true
	default:
		return bm.interrupted
	}
}

func (bm *BackupManager) Backup() error {
	if _, err := os.Stat(bm.config.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("源目录不存在: %s", bm.config.SourcePath)
	}

	bm.record = &models.BackupRecord{
		ID:           utils.GenerateID(),
		StartTime:    time.Now(),
		SourcePath:   bm.config.SourcePath,
		TargetPath:   bm.config.TargetPath,
		IsCompressed: bm.config.IsCompressed,
		Status:       models.StatusInterrupted,
	}

	if err := bm.loadResumeInfo(); err != nil {
		return err
	}

	if bm.config.IsCompressed {
		if err := bm.backupToZip(); err != nil {
			return bm.handleBackupError(err)
		}
	} else {
		if err := bm.backupToDirectory(); err != nil {
			return bm.handleBackupError(err)
		}
	}

	if bm.checkInterrupted() {
		bm.record.Status = models.StatusInterrupted
		if err := bm.saveResumeInfo(); err != nil {
			return err
		}
		return fmt.Errorf("备份被中断，已保存断点信息")
	}

	bm.record.EndTime = time.Now()
	if len(bm.record.FailedFiles) > 0 {
		bm.record.Status = models.StatusPartialFail
	} else {
		bm.record.Status = models.StatusSuccess
	}

	if err := bm.saveMetadata(); err != nil {
		return err
	}

	if err := bm.cleanupResume(); err != nil {
		return err
	}

	if len(bm.record.FailedFiles) > 0 {
		if err := bm.writeErrorReport(); err != nil {
			fmt.Printf("警告: 写入错误报告失败: %v\n", err)
		}
	}

	fmt.Printf("\n备份完成! 状态: %s\n", bm.record.Status)
	fmt.Printf("成功文件数: %d, 失败文件数: %d\n",
		len(bm.record.Files), len(bm.record.FailedFiles))

	return nil
}

func (bm *BackupManager) loadResumeInfo() error {
	if bm.config.ResumePath == "" {
		return nil
	}

	resume, err := utils.ReadResumeInfo(bm.config.ResumePath)
	if err != nil {
		return err
	}

	if resume != nil {
		bm.resumeInfo = resume
		for _, f := range resume.CompletedFiles {
			bm.completedSet[f] = true
		}
		fmt.Printf("检测到中断的备份，将从断点继续 (已完成 %d 个文件)\n", len(resume.CompletedFiles))
		bm.record.ID = resume.BackupID
	} else {
		bm.resumeInfo = &models.ResumeInfo{
			BackupID:       bm.record.ID,
			CompletedFiles: []string{},
		}
	}

	return nil
}

func (bm *BackupManager) saveResumeInfo() error {
	if bm.config.ResumePath == "" {
		return nil
	}

	bm.completedMu.Lock()
	defer bm.completedMu.Unlock()

	files := make([]string, 0, len(bm.completedSet))
	for f := range bm.completedSet {
		files = append(files, f)
	}

	bm.resumeInfo.CompletedFiles = files
	return utils.WriteResumeInfo(bm.config.ResumePath, bm.resumeInfo)
}

func (bm *BackupManager) cleanupResume() error {
	if bm.config.ResumePath == "" {
		return nil
	}
	if _, err := os.Stat(bm.config.ResumePath); err == nil {
		return os.Remove(bm.config.ResumePath)
	}
	return nil
}

func (bm *BackupManager) getLastBackupRecord() *models.BackupRecord {
	meta, err := utils.ReadMetadata(bm.config.MetadataPath)
	if err != nil || meta == nil {
		return nil
	}

	for i := len(meta.BackupRecords) - 1; i >= 0; i-- {
		record := meta.BackupRecords[i]
		if record.SourcePath == bm.config.SourcePath &&
			record.Status == models.StatusSuccess {
			return &record
		}
	}
	return nil
}

func (bm *BackupManager) backupToDirectory() error {
	if err := utils.EnsureDirExists(bm.config.TargetPath); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	lastBackup := bm.getLastBackupRecord()
	var lastFileMap map[string]models.FileInfo
	if bm.config.Incremental && lastBackup != nil {
		lastFileMap = make(map[string]models.FileInfo)
		for _, f := range lastBackup.Files {
			lastFileMap[f.RelativePath] = f
		}
		fmt.Printf("增量备份模式，上一次备份: %s\n", lastBackup.StartTime.Format("2006-01-02 15:04:05"))
	}

	return filepath.Walk(bm.config.SourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if bm.checkInterrupted() {
			return bm.saveResumeInfo()
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(bm.config.SourcePath, path)
		if err != nil {
			return err
		}

		srcInfo := models.FileInfo{
			RelativePath: relPath,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
		}

		if bm.isCompleted(relPath) {
			bm.record.Files = append(bm.record.Files, srcInfo)
			return nil
		}

		if bm.config.Incremental && lastFileMap != nil {
			if lastInfo, ok := lastFileMap[relPath]; ok {
				if !utils.FileChanged(srcInfo, lastInfo) {
					fmt.Printf("跳过未变更文件: %s\n", relPath)
					bm.record.Files = append(bm.record.Files, srcInfo)
					bm.markCompleted(relPath)
					return nil
				}
			}
		}

		dstPath := filepath.Join(bm.config.TargetPath, relPath)
		if err := utils.EnsureDirExists(filepath.Dir(dstPath)); err != nil {
			bm.record.FailedFiles = append(bm.record.FailedFiles, relPath)
			fmt.Printf("创建目标目录失败: %v\n", err)
			return nil
		}

		chunkSize := SmallFileChunkSize
		if info.Size() > LargeFileThreshold {
			chunkSize = ChunkSize
		}

		fmt.Printf("备份文件: %s (%s)\n", relPath, utils.FormatBytes(info.Size()))
		progress := utils.NewProgressBar("  进度", info.Size())

		err = utils.CopyFile(path, dstPath, int64(chunkSize), func(written, total int64) {
			progress.Update(written, total)
		})

		if err != nil {
			bm.record.FailedFiles = append(bm.record.FailedFiles, relPath)
			fmt.Printf("\n备份失败: %s, 错误: %v\n", relPath, err)
			return nil
		}

		if valid := bm.validateFile(path, dstPath); valid {
			bm.record.Files = append(bm.record.Files, srcInfo)
			bm.markCompleted(relPath)
			bm.saveResumeInfo()
		} else {
			bm.record.FailedFiles = append(bm.record.FailedFiles, relPath)
			fmt.Printf("\n校验失败: %s\n", relPath)
		}

		return nil
	})
}

func (bm *BackupManager) backupToZip() error {
	if err := utils.EnsureDirExists(filepath.Dir(bm.config.TargetPath)); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	lastBackup := bm.getLastBackupRecord()
	var lastFileMap map[string]models.FileInfo
	if bm.config.Incremental && lastBackup != nil {
		lastFileMap = make(map[string]models.FileInfo)
		for _, f := range lastBackup.Files {
			lastFileMap[f.RelativePath] = f
		}
		fmt.Printf("增量备份模式，上一次备份: %s\n", lastBackup.StartTime.Format("2006-01-02 15:04:05"))
	}

	tempZip := bm.config.TargetPath + ".tmp"
	zipFile, err := os.Create(tempZip)
	if err != nil {
		return fmt.Errorf("创建压缩包失败: %v", err)
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)

	err = filepath.Walk(bm.config.SourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if bm.checkInterrupted() {
			zw.Close()
			os.Remove(tempZip)
			return bm.saveResumeInfo()
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(bm.config.SourcePath, path)
		if err != nil {
			return err
		}

		srcInfo := models.FileInfo{
			RelativePath: relPath,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
		}

		if bm.isCompleted(relPath) {
			bm.record.Files = append(bm.record.Files, srcInfo)
			return nil
		}

		if bm.config.Incremental && lastFileMap != nil {
			if lastInfo, ok := lastFileMap[relPath]; ok {
				if !utils.FileChanged(srcInfo, lastInfo) {
					fmt.Printf("跳过未变更文件: %s\n", relPath)
					bm.record.Files = append(bm.record.Files, srcInfo)
					bm.markCompleted(relPath)
					return nil
				}
			}
		}

		chunkSize := SmallFileChunkSize
		if info.Size() > LargeFileThreshold {
			chunkSize = ChunkSize
		}

		fmt.Printf("备份文件: %s (%s)\n", relPath, utils.FormatBytes(info.Size()))
		progress := utils.NewProgressBar("  进度", info.Size())

		err = utils.AddFileToZip(zw, path, relPath, int64(chunkSize), func(written, total int64) {
			progress.Update(written, total)
		})

		if err != nil {
			bm.record.FailedFiles = append(bm.record.FailedFiles, relPath)
			fmt.Printf("\n备份失败: %s, 错误: %v\n", relPath, err)
			return nil
		}

		bm.record.Files = append(bm.record.Files, srcInfo)
		bm.markCompleted(relPath)
		bm.saveResumeInfo()

		return nil
	})

	if err != nil {
		zw.Close()
		os.Remove(tempZip)
		return err
	}

	if err := zw.Close(); err != nil {
		os.Remove(tempZip)
		return fmt.Errorf("关闭压缩包失败: %v", err)
	}

	if err := zipFile.Close(); err != nil {
		os.Remove(tempZip)
		return err
	}

	if err := os.Rename(tempZip, bm.config.TargetPath); err != nil {
		os.Remove(tempZip)
		return fmt.Errorf("重命名压缩包失败: %v", err)
	}

	fmt.Println("\n压缩包创建完成，开始校验...")
	if !bm.validateZip() {
		fmt.Println("警告: 压缩包校验失败")
	}

	return nil
}

func (bm *BackupManager) validateFile(srcPath, dstPath string) bool {
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		return false
	}

	return srcInfo.Size() == dstInfo.Size()
}

func (bm *BackupManager) validateZip() bool {
	reader, err := zip.OpenReader(bm.config.TargetPath)
	if err != nil {
		return false
	}
	defer reader.Close()

	zipFileMap := make(map[string]bool)
	for _, f := range reader.File {
		zipFileMap[f.Name] = true
	}

	for _, fileInfo := range bm.record.Files {
		if !zipFileMap[fileInfo.RelativePath] {
			fmt.Printf("校验失败: 文件 %s 不在压缩包中\n", fileInfo.RelativePath)
			return false
		}
	}

	return true
}

func (bm *BackupManager) isCompleted(relPath string) bool {
	bm.completedMu.Lock()
	defer bm.completedMu.Unlock()
	return bm.completedSet[relPath]
}

func (bm *BackupManager) markCompleted(relPath string) {
	bm.completedMu.Lock()
	defer bm.completedMu.Unlock()
	bm.completedSet[relPath] = true
}

func (bm *BackupManager) saveMetadata() error {
	meta, err := utils.ReadMetadata(bm.config.MetadataPath)
	if err != nil {
		return err
	}

	meta.BackupRecords = append(meta.BackupRecords, *bm.record)
	meta.LatestID = bm.record.ID

	return utils.WriteMetadata(bm.config.MetadataPath, meta)
}

func (bm *BackupManager) writeErrorReport() error {
	if bm.config.ErrorReport == "" {
		return nil
	}

	content := fmt.Sprintf("备份错误报告\n")
	content += fmt.Sprintf("备份ID: %s\n", bm.record.ID)
	content += fmt.Sprintf("开始时间: %s\n", bm.record.StartTime.Format("2006-01-02 15:04:05"))
	content += fmt.Sprintf("源目录: %s\n", bm.record.SourcePath)
	content += fmt.Sprintf("目标位置: %s\n", bm.record.TargetPath)
	content += fmt.Sprintf("状态: %s\n", bm.record.Status)
	content += fmt.Sprintf("失败文件列表:\n")
	content += strings.Join(bm.record.FailedFiles, "\n")

	return os.WriteFile(bm.config.ErrorReport, []byte(content), 0644)
}

func (bm *BackupManager) handleBackupError(err error) error {
	bm.record.Status = models.StatusInterrupted
	if saveErr := bm.saveResumeInfo(); saveErr != nil {
		return fmt.Errorf("备份失败: %v, 保存断点也失败: %v", err, saveErr)
	}
	return err
}
