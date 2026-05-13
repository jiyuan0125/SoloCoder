package services

import (
	"archive/zip"
	"encoding/json"
	"io"
	"image-batch/models"
	"os"
	"path/filepath"
)

type ZipService struct {
}

func NewZipService() *ZipService {
	return &ZipService{}
}

func (s *ZipService) CreateZip(files []string, result *models.ProcessResult, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	for _, file := range files {
		err = s.addFileToZip(zipWriter, file)
		if err != nil {
			return err
		}
	}

	reportJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	err = s.addDataToZip(zipWriter, "process_report.json", reportJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *ZipService) addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (s *ZipService) addDataToZip(zipWriter *zip.Writer, filename string, data []byte) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}
