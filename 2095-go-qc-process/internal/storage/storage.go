package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"qc-process/internal/models"
)

const dataDir = "data"

func ensureDataDir() error {
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return os.MkdirAll(dataDir, 0755)
	}
	return nil
}

func SaveReport(report *models.QCReport) error {
	if err := ensureDataDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("错误：无法序列化质检数据: %v", err)
	}

	filePath := filepath.Join(dataDir, report.ID+".json")
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("错误：无法写入质检数据文件: %v", err)
	}

	return nil
}

func LoadReport(id string) (*models.QCReport, error) {
	if err := ensureDataDir(); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, id+".json")
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("错误：无法读取质检数据文件: %v", err)
	}

	var report models.QCReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("错误：无法解析质检数据: %v", err)
	}

	return &report, nil
}

func ExportReportToJSON(report *models.QCReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("错误：无法序列化质检报告: %v", err)
	}

	if err := ioutil.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("错误：无法写入导出文件: %v", err)
	}

	return nil
}

func ListReports() ([]string, error) {
	if err := ensureDataDir(); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("错误：无法读取数据目录: %v", err)
	}

	var ids []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			id := f.Name()[:len(f.Name())-5]
			ids = append(ids, id)
		}
	}

	return ids, nil
}
