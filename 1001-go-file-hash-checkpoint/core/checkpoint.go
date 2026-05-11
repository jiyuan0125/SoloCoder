package core

import (
	"math"
	"sync"
	"time"

	"hash-checkpoint/common"
)

type CheckpointManager struct {
	mu        sync.RWMutex
	records   map[string]*common.Progress
}

func NewCheckpointManager() *CheckpointManager {
	return &CheckpointManager{
		records: make(map[string]*common.Progress),
	}
}

func CalculateTotalChunks(fileSize int64, chunkSize int64) int {
	if fileSize <= 0 {
		return 0
	}
	if chunkSize <= 0 {
		chunkSize = common.DefaultChunkSize
	}
	return int(math.Ceil(float64(fileSize) / float64(chunkSize)))
}

func CalculateChunkLen(index int, fileSize int64, chunkSize int64) int64 {
	if fileSize <= 0 {
		return 0
	}
	if chunkSize <= 0 {
		chunkSize = common.DefaultChunkSize
	}
	start := int64(index) * chunkSize
	if start >= fileSize {
		return 0
	}
	remaining := fileSize - start
	if remaining > chunkSize {
		return chunkSize
	}
	return remaining
}

func GenerateFileKey(filePath string, modTime time.Time) string {
	return filePath + "|" + modTime.UTC().Format(time.RFC3339Nano)
}

func (cm *CheckpointManager) StartOrResume(
	filePath string,
	modTime time.Time,
	fileSize int64,
	algorithm common.HashAlgorithm,
	chunkSize int64,
) (*common.Progress, int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := GenerateFileKey(filePath, modTime)
	totalChunks := CalculateTotalChunks(fileSize, chunkSize)

	if existing, ok := cm.records[key]; ok {
		existing.Status = common.StatusInProgress
		return existing, existing.CompletedChunks
	}

	progress := &common.Progress{
		FilePath:        filePath,
		ModTime:         modTime,
		Algorithm:       algorithm,
		ChunkSize:       chunkSize,
		FileSize:        fileSize,
		TotalChunks:     totalChunks,
		CompletedChunks: 0,
		Percentage:      0.0,
		Status:          common.StatusInProgress,
		ChunkHashes:     make([]common.ChunkHash, 0, totalChunks),
		FinalHash:       "",
		StartTime:       time.Now(),
	}
	cm.records[key] = progress
	return progress, 0
}

func (cm *CheckpointManager) SubmitChunk(
	filePath string,
	modTime time.Time,
	chunkIndex int,
	chunkHash string,
	chunkLen int64,
) (*common.Progress, int, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := GenerateFileKey(filePath, modTime)
	progress, ok := cm.records[key]
	if !ok {
		return nil, 0, nil
	}

	if chunkIndex < len(progress.ChunkHashes) {
		if progress.ChunkHashes[chunkIndex].Hash == chunkHash {
			return progress, chunkIndex + 1, nil
		}
		progress.ChunkHashes = progress.ChunkHashes[:chunkIndex]
		progress.CompletedChunks = chunkIndex
		progress.Status = common.StatusChunkMismatch
	}

	progress.ChunkHashes = append(progress.ChunkHashes, common.ChunkHash{
		Index:   chunkIndex,
		Hash:    chunkHash,
		HashLen: chunkLen,
	})
	progress.CompletedChunks = len(progress.ChunkHashes)
	if progress.TotalChunks > 0 {
		progress.Percentage = float64(progress.CompletedChunks) / float64(progress.TotalChunks) * 100.0
	}
	if progress.CompletedChunks == progress.TotalChunks {
		progress.Status = common.StatusCompleted
	} else {
		progress.Status = common.StatusInProgress
	}

	return progress, progress.CompletedChunks, nil
}

func (cm *CheckpointManager) GetProgress(filePath string, modTime time.Time) *common.Progress {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := GenerateFileKey(filePath, modTime)
	progress, ok := cm.records[key]
	if !ok {
		return nil
	}

	clone := *progress
	if len(progress.ChunkHashes) > 0 {
		clone.ChunkHashes = make([]common.ChunkHash, len(progress.ChunkHashes))
		copy(clone.ChunkHashes, progress.ChunkHashes)
	}
	return &clone
}

func (cm *CheckpointManager) CompleteCheck(filePath string, modTime time.Time) (*common.Progress, string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	key := GenerateFileKey(filePath, modTime)
	progress, ok := cm.records[key]
	if !ok {
		return nil, "", nil
	}

	if progress.FileSize <= 0 {
		if progress.FinalHash == "" {
			progress.FinalHash = common.GetEmptyFileHash(progress.Algorithm)
			progress.CompleteTime = time.Now()
			progress.Status = common.StatusCompleted
			progress.Percentage = 100.0
		}
		return progress, progress.FinalHash, nil
	}

	if progress.CompletedChunks != progress.TotalChunks {
		return progress, "", nil
	}

	if progress.FinalHash == "" {
		chunkHashStrings := make([]string, 0, len(progress.ChunkHashes))
		for _, ch := range progress.ChunkHashes {
			chunkHashStrings = append(chunkHashStrings, ch.Hash)
		}
		finalHash, err := CalculateFinalHash(chunkHashStrings, progress.Algorithm)
		if err != nil {
			return progress, "", err
		}
		progress.FinalHash = finalHash
		progress.CompleteTime = time.Now()
		progress.Status = common.StatusCompleted
	}

	return progress, progress.FinalHash, nil
}
