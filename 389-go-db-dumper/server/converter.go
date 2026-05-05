package server

import (
	"os"
	"path/filepath"

	"go-db-dumper/proto"
)

type Converter struct {
	config    *proto.ConverterConfig
	csvParser *CSVParser
	sqlGen    *SQLGenerator
}

func NewConverter(config *proto.ConverterConfig) *Converter {
	if config == nil {
		config = proto.DefaultConverterConfig()
	}

	return &Converter{
		config:    config,
		csvParser: NewCSVParser(config.SkipBOM),
		sqlGen:    NewSQLGenerator(config),
	}
}

func (c *Converter) Convert(inputFile, outputFile, tableName string) (*ConversionResult, error) {
	if tableName == "" {
		tableName = c.extractTableName(inputFile)
	}

	csvData, err := c.parseCSV(inputFile)
	if err != nil {
		return nil, err
	}

	result, err := c.generateSQL(csvData, outputFile, tableName, inputFile)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Converter) parseCSV(inputFile string) (*CSVData, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return c.csvParser.Parse(file)
}

func (c *Converter) generateSQL(csvData *CSVData, outputFile, tableName, sourceFile string) (*ConversionResult, error) {
	file, err := os.Create(outputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sqlCount, err := c.sqlGen.Generate(csvData, tableName, sourceFile, file)
	if err != nil {
		return nil, err
	}

	return &ConversionResult{
		RecordsRead: len(csvData.Rows),
		SQLCount:    sqlCount,
		TableName:   tableName,
		OutputFile:  outputFile,
	}, nil
}

func (c *Converter) extractTableName(inputFile string) string {
	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	if ext == "" {
		return base
	}
	return base[:len(base)-len(ext)]
}

type ConversionResult struct {
	RecordsRead int
	SQLCount    int
	TableName   string
	OutputFile  string
}
