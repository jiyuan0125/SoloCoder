package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"search-service/database"
	"search-service/models"
)

const (
	titleWeight      = 2.0
	bodyWeight       = 1.0
	customFieldWeight = 1.0
	highlightPrefix  = "<mark>"
	highlightSuffix  = "</mark>"
)

func Search(keyword string, field string) ([]*models.SearchResult, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, fmt.Errorf("keyword is required")
	}

	docs, err := ListDocuments()
	if err != nil {
		return nil, err
	}

	searchInAllFields := true
	if field != "" {
		fieldExists := checkFieldExists(field, docs)
		if fieldExists {
			searchInAllFields = false
		}
	}

	results := []*models.SearchResult{}
	keywordLower := strings.ToLower(keyword)

	for _, doc := range docs {
		score := 0.0
		highlights := []string{}

		if searchInAllFields || strings.ToLower(field) == "title" {
			if count, matched := matchAndHighlight(doc.Title, keywordLower); count > 0 {
				score += float64(count) * titleWeight
				highlights = append(highlights, matched...)
			}
		}

		if searchInAllFields || strings.ToLower(field) == "body" {
			if count, matched := matchAndHighlight(doc.Body, keywordLower); count > 0 {
				score += float64(count) * bodyWeight
				highlights = append(highlights, matched...)
			}
		}

		if doc.CustomFields != nil {
			for name, value := range doc.CustomFields {
				if !searchInAllFields && strings.ToLower(field) != strings.ToLower(name) {
					continue
				}
				valStr := convertToString(value)
				if count, matched := matchAndHighlight(valStr, keywordLower); count > 0 {
					score += float64(count) * customFieldWeight
					highlights = append(highlights, matched...)
				}
			}
		}

		if score > 0 {
			results = append(results, &models.SearchResult{
				Document:  *doc,
				Score:     score,
				Highlights: highlights,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if err := RecordSearchHistory(keyword); err != nil {
		fmt.Printf("Failed to record search history: %v\n", err)
	}

	return results, nil
}

func checkFieldExists(field string, docs []*models.Document) bool {
	fieldLower := strings.ToLower(field)
	if fieldLower == "title" || fieldLower == "body" {
		return true
	}

	for _, doc := range docs {
		if doc.CustomFields != nil {
			for name := range doc.CustomFields {
				if strings.ToLower(name) == fieldLower {
					return true
				}
			}
		}
	}

	return false
}

func matchAndHighlight(text string, keywordLower string) (int, []string) {
	if text == "" {
		return 0, nil
	}

	textLower := strings.ToLower(text)
	keywordLen := len(keywordLower)
	count := 0
	highlights := []string{}

	idx := strings.Index(textLower, keywordLower)
	for idx != -1 {
		count++
		start := idx
		end := idx + keywordLen

		contextStart := max(0, start-20)
		contextEnd := min(len(text), end+20)

		contextPrefix := ""
		contextSuffix := ""
		if contextStart > 0 {
			contextPrefix = "..."
		}
		if contextEnd < len(text) {
			contextSuffix = "..."
		}

		originalMatch := text[start:end]
		before := text[contextStart:start]
		after := text[end:contextEnd]
		highlighted := contextPrefix + before + highlightPrefix + originalMatch + highlightSuffix + after + contextSuffix

		highlights = append(highlights, highlighted)

		idx = strings.Index(textLower[end:], keywordLower)
		if idx != -1 {
			idx += end
		}
	}

	return count, highlights
}

func convertToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	}
}

func RecordSearchHistory(keyword string) error {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil
	}

	_, err := database.DB.Exec(`INSERT INTO search_history (keyword) VALUES (?)`, keyword)
	return err
}

func GetHotKeywords() ([]*models.HotKeyword, error) {
	query := `
		SELECT keyword, COUNT(*) as count
		FROM search_history
		WHERE searched_at >= date('now', '-7 days')
		GROUP BY keyword
		ORDER BY count DESC, keyword ASC
		LIMIT 20
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.HotKeyword{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	keywords := []*models.HotKeyword{}
	for rows.Next() {
		hk := &models.HotKeyword{}
		if err := rows.Scan(&hk.Keyword, &hk.Count); err != nil {
			return nil, err
		}
		keywords = append(keywords, hk)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keywords, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func HighlightText(text string, keyword string) string {
	if text == "" || keyword == "" {
		return text
	}

	keywordLower := strings.ToLower(keyword)
	textLower := strings.ToLower(text)
	keywordLen := len(keywordLower)

	var result strings.Builder
	lastIdx := 0

	idx := strings.Index(textLower[lastIdx:], keywordLower)
	for idx != -1 {
		absIdx := lastIdx + idx
		result.WriteString(text[lastIdx:absIdx])
		result.WriteString(highlightPrefix)
		result.WriteString(text[absIdx : absIdx+keywordLen])
		result.WriteString(highlightSuffix)
		lastIdx = absIdx + keywordLen
		idx = strings.Index(textLower[lastIdx:], keywordLower)
	}

	result.WriteString(text[lastIdx:])
	return result.String()
}
