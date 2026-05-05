package csvutil

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"time"
)

func WriteChangeFile(changes []Change, sourceHeaders []string, targetFile string) (string, error) {
	if len(changes) == 0 {
		return "", nil
	}
	
	timestamp := time.Now().Format("20060102_150405")
	dir := filepath.Dir(targetFile)
	base := filepath.Base(targetFile)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	
	changeFileName := filepath.Join(dir, name+"_changes_"+timestamp+".csv")
	
	file, err := os.Create(changeFileName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	headers := []string{"Operation", "Key"}
	
	usedHeaders := make(map[string]bool)
	for _, h := range sourceHeaders {
		usedHeaders[h] = true
		headers = append(headers, "Old_"+h)
	}
	for _, h := range sourceHeaders {
		headers = append(headers, "New_"+h)
	}
	
	if err := writer.Write(headers); err != nil {
		return "", err
	}
	
	for _, change := range changes {
		row := []string{
			string(change.Type),
			change.Key,
		}
		
		for _, h := range sourceHeaders {
			oldVal := ""
			if change.OldRecord != nil {
				oldVal = change.OldRecord.Values[h]
			}
			row = append(row, oldVal)
		}
		
		for _, h := range sourceHeaders {
			newVal := ""
			if change.NewRecord != nil {
				newVal = change.NewRecord.Values[h]
			}
			row = append(row, newVal)
		}
		
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	
	return changeFileName, nil
}
