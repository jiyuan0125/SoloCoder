package csvutil

import "strings"

type CSVRecord struct {
	Values map[string]string
}

type CSVData struct {
	Headers []string
	Records map[string]CSVRecord
}

type ChangeType string

const (
	ChangeAdd    ChangeType = "ADD"
	ChangeModify ChangeType = "MODIFY"
	ChangeDelete ChangeType = "DELETE"
)

type Change struct {
	Type     ChangeType
	Key      string
	OldRecord *CSVRecord
	NewRecord *CSVRecord
}

func (c *Change) IsModified() bool {
	if c.Type != ChangeModify {
		return false
	}
	if c.OldRecord == nil || c.NewRecord == nil {
		return false
	}
	for key, oldVal := range c.OldRecord.Values {
		newVal, ok := c.NewRecord.Values[key]
		if !ok || strings.TrimSpace(oldVal) != strings.TrimSpace(newVal) {
			return true
		}
	}
	for key := range c.NewRecord.Values {
		if _, ok := c.OldRecord.Values[key]; !ok {
			return true
		}
	}
	return false
}

type SyncResult struct {
	Changes    []Change
	ChangeFile string
	Applied    bool
	Warnings   []string
}

func (r *SyncResult) CountByType(ct ChangeType) int {
	count := 0
	for _, c := range r.Changes {
		if c.Type == ct {
			count++
		}
	}
	return count
}

func (r *SyncResult) TotalChanges() int {
	return len(r.Changes)
}
