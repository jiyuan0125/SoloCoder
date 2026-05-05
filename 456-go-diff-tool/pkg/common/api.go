package common

type DiffType string

const (
	DiffEqual   DiffType = "equal"
	DiffAdd     DiffType = "add"
	DiffDelete  DiffType = "delete"
	DiffChange  DiffType = "change"
	DiffMove    DiffType = "move"
)

type LineDiff struct {
	OldLine   int        `json:"old_line"`
	NewLine   int        `json:"new_line"`
	Type      DiffType   `json:"type"`
	Content   string     `json:"content"`
	WordDiffs []WordDiff `json:"word_diffs,omitempty"`
}

type WordDiff struct {
	Type    DiffType `json:"type"`
	Content string   `json:"content"`
}

type BlockMove struct {
	OldStart int    `json:"old_start"`
	OldEnd   int    `json:"old_end"`
	NewStart int    `json:"new_start"`
	NewEnd   int    `json:"new_end"`
	Content  string `json:"content"`
}

type DiffStats struct {
	Added   int `json:"added"`
	Deleted int `json:"deleted"`
	Changed int `json:"changed"`
	Moved   int `json:"moved"`
}

type CompareRequest struct {
	OldText       string `json:"old_text"`
	NewText       string `json:"new_text"`
	MoveThreshold int    `json:"move_threshold,omitempty"`
}

type CompareResponse struct {
	Success    bool       `json:"success"`
	LineDiffs  []LineDiff `json:"line_diffs"`
	Stats      DiffStats  `json:"stats"`
	MovedBlocks []BlockMove `json:"moved_blocks,omitempty"`
	UnifiedDiff string    `json:"unified_diff,omitempty"`
	Error      string     `json:"error,omitempty"`
}
