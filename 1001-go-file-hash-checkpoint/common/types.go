package common

import "time"

type HashAlgorithm string

const (
	AlgorithmMD5    HashAlgorithm = "MD5"
	AlgorithmSHA256 HashAlgorithm = "SHA256"
)

const (
	DefaultChunkSize = 4 * 1024 * 1024
)

type CheckStatus string

const (
	StatusPending     CheckStatus = "pending"
	StatusInProgress  CheckStatus = "in_progress"
	StatusCompleted   CheckStatus = "completed"
	StatusChunkMismatch CheckStatus = "chunk_mismatch"
)

type FileIdentity struct {
	FilePath    string
	ModTime     time.Time
	FileSize    int64
}

type ChunkHash struct {
	Index    int
	Hash     string
	HashLen  int64
}

type Progress struct {
	FilePath        string
	ModTime         time.Time
	Algorithm       HashAlgorithm
	ChunkSize       int64
	FileSize        int64
	TotalChunks     int
	CompletedChunks int
	Percentage      float64
	Status          CheckStatus
	ChunkHashes     []ChunkHash
	FinalHash       string
	StartTime       time.Time
	CompleteTime    time.Time
}
