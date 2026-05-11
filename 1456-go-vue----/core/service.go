package core

import (
	"archivesystem/common"
	"sync"
	"time"
)

type ArchiveService struct {
	mu            sync.RWMutex
	archives      map[string]*common.Archive
	borrowRecords []*common.BorrowRecord
	destroyRecords []*common.DestroyRecord
	idCounter     map[string]int
	borrowID      int64
	destroyID     int64
}

func NewArchiveService() *ArchiveService {
	return &ArchiveService{
		archives:       make(map[string]*common.Archive),
		borrowRecords:  make([]*common.BorrowRecord, 0),
		destroyRecords: make([]*common.DestroyRecord, 0),
		idCounter:      make(map[string]int),
	}
}

func today() string {
	return time.Now().Format("2006-01-02")
}

func currentYear() string {
	return time.Now().Format("2006")
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
