package common

import "time"

type StartCheckRequest struct {
	FilePath    string
	ModTime     time.Time
	FileSize    int64
	Algorithm   HashAlgorithm
	ChunkSize   int64
}

type StartCheckResponse struct {
	Success         bool
	Message         string
	NextChunkIndex  int
	Progress        *Progress
}

type SubmitChunkRequest struct {
	FilePath    string
	ModTime     time.Time
	ChunkIndex  int
	ChunkHash   string
	ChunkLen    int64
}

type SubmitChunkResponse struct {
	Success         bool
	Message         string
	NextChunkIndex  int
	Progress        *Progress
}

type GetProgressRequest struct {
	FilePath    string
	ModTime     time.Time
}

type GetProgressResponse struct {
	Success  bool
	Message  string
	Progress *Progress
}

type CompleteCheckRequest struct {
	FilePath    string
	ModTime     time.Time
}

type CompleteCheckResponse struct {
	Success   bool
	Message   string
	FinalHash string
}

type ErrorResponse struct {
	Success bool
	Error   string
}
