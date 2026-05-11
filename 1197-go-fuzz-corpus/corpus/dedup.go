package corpus

import (
	"os"
	"sort"

	"go-fuzz-corpus/api"
)

func DedupCorpus(targetName, rootPath string) (*api.DedupResponse, error) {
	corpusDir, err := FindCorpusDirForTarget(rootPath, targetName)
	if err != nil {
		return nil, err
	}

	stats, err := ScanCorpusDir(corpusDir)
	if err != nil {
		return nil, err
	}

	hashMap := make(map[string][]api.CorpusFile)
	for _, file := range stats.Files {
		hashMap[file.Hash] = append(hashMap[file.Hash], file)
	}

	response := &api.DedupResponse{
		TargetName:   targetName,
		RemovedCount: 0,
		RemovedFiles: []string{},
	}

	for _, files := range hashMap {
		if len(files) <= 1 {
			continue
		}

		sort.Slice(files, func(i, j int) bool {
			return files[i].AddedTime.Before(files[j].AddedTime)
		})

		for i := 1; i < len(files); i++ {
			err := os.Remove(files[i].Path)
			if err != nil {
				continue
			}
			response.RemovedCount++
			response.RemovedFiles = append(response.RemovedFiles, files[i].Path)
		}
	}

	response.RetainedCount = stats.FileCount - response.RemovedCount

	return response, nil
}
