package csvutil

import (
	"encoding/csv"
	"os"
)

func ApplyChanges(source *CSVData, changes []Change, targetFile string, keyColumn string) error {
	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	if len(source.Headers) > 0 {
		if err := writer.Write(source.Headers); err != nil {
			return err
		}
	}
	
	for _, rec := range source.Records {
		row := make([]string, len(source.Headers))
		for i, h := range source.Headers {
			row[i] = rec.Values[h]
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	
	return nil
}
