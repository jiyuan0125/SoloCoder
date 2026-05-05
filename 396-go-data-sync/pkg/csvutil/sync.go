package csvutil

import (
	"os"
)

func SyncCSV(sourceFile, targetFile, keyColumn string, apply bool) (*SyncResult, error) {
	result := &SyncResult{
		Changes:  []Change{},
		Applied:  false,
		Warnings: []string{},
	}
	
	if keyColumn == "" {
		var err error
		keyColumn, err = GetFirstColumnHeader(sourceFile)
		if err != nil {
			return nil, err
		}
	}
	
	sourceData, sourceWarnings, err := ReadCSV(sourceFile, keyColumn)
	if err != nil {
		return nil, err
	}
	result.Warnings = append(result.Warnings, sourceWarnings...)
	
	var targetData *CSVData
	
	if _, err := os.Stat(targetFile); err == nil {
		var targetWarnings []string
		targetData, targetWarnings, err = ReadCSV(targetFile, keyColumn)
		if err != nil {
			return nil, err
		}
		result.Warnings = append(result.Warnings, targetWarnings...)
	}
	
	changes := CompareCSV(sourceData, targetData)
	result.Changes = changes
	
	if len(changes) > 0 {
		changeFile, err := WriteChangeFile(changes, sourceData.Headers, targetFile)
		if err != nil {
			return nil, err
		}
		result.ChangeFile = changeFile
	}
	
	if apply && len(changes) > 0 {
		err := ApplyChanges(sourceData, changes, targetFile, keyColumn)
		if err != nil {
			return nil, err
		}
		result.Applied = true
	}
	
	return result, nil
}
