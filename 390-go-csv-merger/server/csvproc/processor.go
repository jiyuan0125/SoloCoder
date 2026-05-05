package csvproc

import (
	"fmt"
	"os"

	"csv-merger/common"
)

func ProcessMergeRequest(request *common.MergeRequest) (*common.MergeResponse, error) {
	response := &common.MergeResponse{}

	if len(request.InputFiles) == 0 {
		response.Success = false
		response.Error = "no input files specified"
		return response, nil
	}

	var existingFiles []string
	for _, filePath := range request.InputFiles {
		if _, err := os.Stat(filePath); err == nil {
			existingFiles = append(existingFiles, filePath)
		}
	}

	if len(existingFiles) == 0 {
		response.Success = false
		response.Error = "all input files do not exist"
		return response, nil
	}

	var csvFiles []*CSVFile
	for _, filePath := range existingFiles {
		csvFile, err := ReadCSVFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		if len(csvFile.Rows) > 0 {
			csvFiles = append(csvFiles, csvFile)
		}
	}

	if len(csvFiles) == 0 {
		response.Success = false
		response.Error = "all input files are empty (only headers)"
		return response, nil
	}

	if err := ValidateHeaders(csvFiles); err != nil {
		response.Success = false
		response.Error = err.Error()
		return response, nil
	}

	options := MergeOptions{
		DedupColumns: request.DedupColumns,
		SortColumn:   request.SortColumn,
		SortOrder:    request.SortOrder,
		CustomHeader: request.CustomHeader,
	}

	recordsMerged, err := MergeCSVFiles(csvFiles, request.OutputFile, options)
	if err != nil {
		response.Success = false
		response.Error = err.Error()
		return response, nil
	}

	response.Success = true
	response.RecordsMerged = recordsMerged
	response.Message = fmt.Sprintf("successfully merged %d records into %s", recordsMerged, request.OutputFile)

	return response, nil
}
