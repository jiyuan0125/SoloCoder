package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"godiff/pkg/common"
	"godiff/pkg/diff"
)

const DefaultMoveThreshold = 20

func toCommonDiffType(dt diff.DiffType) common.DiffType {
	switch dt {
	case diff.DiffEqual:
		return common.DiffEqual
	case diff.DiffAdd:
		return common.DiffAdd
	case diff.DiffDelete:
		return common.DiffDelete
	case diff.DiffChange:
		return common.DiffChange
	case diff.DiffMove:
		return common.DiffMove
	default:
		return common.DiffEqual
	}
}

func convertToCommonLineDiffs(lineDiffs []diff.LineDiff) []common.LineDiff {
	result := make([]common.LineDiff, len(lineDiffs))
	for i, ld := range lineDiffs {
		result[i] = common.LineDiff{
			OldLine: ld.OldLine,
			NewLine: ld.NewLine,
			Type:    toCommonDiffType(ld.Type),
			Content: ld.Content,
		}
		if len(ld.WordDiffs) > 0 {
			result[i].WordDiffs = make([]common.WordDiff, len(ld.WordDiffs))
			for j, wd := range ld.WordDiffs {
				result[i].WordDiffs[j] = common.WordDiff{
					Type:    toCommonDiffType(wd.Type),
					Content: wd.Content,
				}
			}
		}
	}
	return result
}

func convertToCommonMovedBlocks(blocks []diff.BlockMove) []common.BlockMove {
	result := make([]common.BlockMove, len(blocks))
	for i, b := range blocks {
		result[i] = common.BlockMove{
			OldStart: b.OldStart,
			OldEnd:   b.OldEnd,
			NewStart: b.NewStart,
			NewEnd:   b.NewEnd,
			Content:  b.Content,
		}
	}
	return result
}

func compareHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req common.CompareRequest
	if err := json.Unmarshal(body, &req); err != nil {
		sendErrorResponse(w, "Invalid JSON format")
		return
	}

	threshold := DefaultMoveThreshold
	if req.MoveThreshold > 0 {
		threshold = req.MoveThreshold
	}

	result := diff.Compare(req.OldText, req.NewText, threshold)

	oldLines := strings.Split(req.OldText, "\n")
	newLines := strings.Split(req.NewText, "\n")
	unifiedDiff := diff.GenerateUnifiedDiff(oldLines, newLines, result)

	response := common.CompareResponse{
		Success:     true,
		LineDiffs:   convertToCommonLineDiffs(result.LineDiffs),
		Stats: common.DiffStats{
			Added:   result.Stats.Added,
			Deleted: result.Stats.Deleted,
			Changed: result.Stats.Changed,
			Moved:   result.Stats.Moved,
		},
		MovedBlocks: convertToCommonMovedBlocks(result.MovedBlocks),
		UnifiedDiff: unifiedDiff,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func sendErrorResponse(w http.ResponseWriter, message string) {
	response := common.CompareResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/api/compare", compareHandler)

	fmt.Println("Diff Server starting on :8080")
	fmt.Println("Endpoint: POST http://localhost:8080/api/compare")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
