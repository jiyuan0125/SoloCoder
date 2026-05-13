package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const (
	port        = 9803
	archiveDays = 180
)

var db *sql.DB

type Tag struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	UsageCount int       `json:"usage_count"`
	Archived   bool      `json:"archived"`
	LastUsedAt time.Time `json:"last_used_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type ResourceTag struct {
	ID         int64     `json:"id"`
	ResourceType string  `json:"resource_type"`
	ResourceID   string  `json:"resource_id"`
	TagID        int64   `json:"tag_id"`
	TagName      string  `json:"tag_name"`
	TagColor     string  `json:"tag_color"`
	CreatedAt    time.Time `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	var err error
	db, err = initDB()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize database: %v", err))
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/tags", handleTags)
	mux.HandleFunc("/tags/", handleTagByID)
	mux.HandleFunc("/tags/archive/", handleArchiveTag)

	mux.HandleFunc("/resources/tags", handleResourceTags)
	mux.HandleFunc("/resources/tags/", handleResourceTagsByID)

	mux.HandleFunc("/search/resources", handleSearchResources)

	fmt.Printf("Tag System running on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		panic(fmt.Sprintf("failed to start server: %v", err))
	}
}

func initDB() (*sql.DB, error) {
	database, err := sql.Open("sqlite", "tags.db")
	if err != nil {
		return nil, err
	}

	if err = database.Ping(); err != nil {
		return nil, err
	}

	createTablesSQL := `
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			color TEXT NOT NULL DEFAULT '#000000',
			usage_count INTEGER NOT NULL DEFAULT 0,
			archived INTEGER NOT NULL DEFAULT 0,
			last_used_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS resource_tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			tag_id INTEGER NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(resource_type, resource_id, tag_id),
			FOREIGN KEY (tag_id) REFERENCES tags(id)
		);

		CREATE INDEX IF NOT EXISTS idx_resource_tags_resource ON resource_tags(resource_type, resource_id);
		CREATE INDEX IF NOT EXISTS idx_resource_tags_tag ON resource_tags(tag_id);
		CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
	`

	_, err = database.Exec(createTablesSQL)
	if err != nil {
		return nil, err
	}

	return database, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func handleTags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listTags(w, r)
	case http.MethodPost:
		createTag(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleTagByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/tags/")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "tag id is required")
		return
	}

	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tag id")
		return
	}

	getTag(w, id)
}

func handleArchiveTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/tags/archive/")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "tag id is required")
		return
	}

	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		writeError(w, http.StatusBadRequest, "invalid tag id")
		return
	}

	var req struct {
		Confirm bool `json:"confirm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	archiveTag(w, id, req.Confirm)
}

func listTags(w http.ResponseWriter, r *http.Request) {
	includeArchived := r.URL.Query().Get("include_archived") == "true"
	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))

	rows, err := db.Query(`
		SELECT id, name, color, usage_count, archived, last_used_at, created_at
		FROM tags
		WHERE (archived = ? OR ? = true)
			AND (? = '' OR name LIKE ?)
		ORDER BY usage_count DESC, name ASC
	`, 0, includeArchived, search, "%"+search+"%")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tags")
		return
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var tag Tag
		var lastUsedAt sql.NullTime
		if err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Color, &tag.UsageCount, &tag.Archived,
			&lastUsedAt, &tag.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan tags")
			return
		}
		if lastUsedAt.Valid {
			tag.LastUsedAt = lastUsedAt.Time
		}
		tags = append(tags, tag)
	}

	writeJSON(w, http.StatusOK, tags)
}

func createTag(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "tag name is required")
		return
	}

	nameLower := strings.ToLower(req.Name)
	color := req.Color
	if color == "" {
		color = "#000000"
	}

	result, err := db.Exec(`
		INSERT INTO tags (name, color) VALUES (?, ?)
	`, nameLower, color)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			writeError(w, http.StatusConflict, "tag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create tag")
		return
	}

	id, _ := result.LastInsertId()
	getTag(w, id)
}

func getTag(w http.ResponseWriter, id int64) {
	var tag Tag
	var lastUsedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, name, color, usage_count, archived, last_used_at, created_at
		FROM tags WHERE id = ?
	`, id).Scan(
		&tag.ID, &tag.Name, &tag.Color, &tag.UsageCount, &tag.Archived,
		&lastUsedAt, &tag.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get tag")
		return
	}

	if lastUsedAt.Valid {
		tag.LastUsedAt = lastUsedAt.Time
	}

	writeJSON(w, http.StatusOK, tag)
}

func archiveTag(w http.ResponseWriter, id int64, confirm bool) {
	if !confirm {
		writeError(w, http.StatusBadRequest, "archive operation requires confirmation")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	var tag Tag
	var lastUsedAt sql.NullTime
	err = tx.QueryRow(`
		SELECT id, name, color, usage_count, archived, last_used_at, created_at
		FROM tags WHERE id = ?
	`, id).Scan(
		&tag.ID, &tag.Name, &tag.Color, &tag.UsageCount, &tag.Archived,
		&lastUsedAt, &tag.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get tag")
		return
	}

	if tag.Archived {
		writeError(w, http.StatusBadRequest, "tag is already archived")
		return
	}

	if tag.UsageCount > 0 {
		if !lastUsedAt.Valid {
			writeError(w, http.StatusBadRequest, "tag is still in use")
			return
		}
		daysSinceUsed := time.Since(lastUsedAt.Time).Hours() / 24
		if daysSinceUsed < float64(archiveDays) {
			writeError(w, http.StatusBadRequest, fmt.Sprintf(
				"tag has been used within %d days and cannot be archived", archiveDays,
			))
			return
		}
	}

	_, err = tx.Exec(`UPDATE tags SET archived = 1 WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to archive tag")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "tag archived successfully"})
}

func handleResourceTags(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTagToResource(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleResourceTagsByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/resources/tags/")
	parts := strings.SplitN(path, "/", 3)

	if len(parts) < 2 {
		writeError(w, http.StatusBadRequest, "resource type and id are required")
		return
	}

	resourceType := parts[0]
	resourceID := parts[1]

	if resourceType == "" || resourceID == "" {
		writeError(w, http.StatusBadRequest, "resource type and id are required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		getTagsForResource(w, resourceType, resourceID)
	case http.MethodDelete:
		if len(parts) < 3 {
			writeError(w, http.StatusBadRequest, "tag name or id is required")
			return
		}
		removeTagFromResource(w, resourceType, resourceID, parts[2])
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func addTagToResource(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
		TagName      string `json:"tag_name"`
		TagColor     string `json:"tag_color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.ResourceType = strings.TrimSpace(req.ResourceType)
	req.ResourceID = strings.TrimSpace(req.ResourceID)
	req.TagName = strings.ToLower(strings.TrimSpace(req.TagName))

	if req.ResourceType == "" || req.ResourceID == "" || req.TagName == "" {
		writeError(w, http.StatusBadRequest, "resource_type, resource_id, and tag_name are required")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	var tagID int64
	var archived bool
	err = tx.QueryRow(
		`SELECT id, archived FROM tags WHERE name = ?`,
		req.TagName,
	).Scan(&tagID, &archived)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			color := req.TagColor
			if color == "" {
				color = "#000000"
			}
			result, err := tx.Exec(
				`INSERT INTO tags (name, color) VALUES (?, ?)`,
				req.TagName, color,
			)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create tag")
				return
			}
			tagID, _ = result.LastInsertId()
		} else {
			writeError(w, http.StatusInternalServerError, "failed to get tag")
			return
		}
	} else if archived {
		writeError(w, http.StatusBadRequest, "标签已归档")
		return
	}

	var existingID int64
	err = tx.QueryRow(`
		SELECT id FROM resource_tags
		WHERE resource_type = ? AND resource_id = ? AND tag_id = ?
	`, req.ResourceType, req.ResourceID, tagID).Scan(&existingID)

	if err == nil {
		writeError(w, http.StatusConflict, "已存在该标签")
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "failed to check existing tag")
		return
	}

	_, err = tx.Exec(`
		INSERT INTO resource_tags (resource_type, resource_id, tag_id)
		VALUES (?, ?, ?)
	`, req.ResourceType, req.ResourceID, tagID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add tag to resource")
		return
	}

	_, err = tx.Exec(`
		UPDATE tags SET usage_count = usage_count + 1, last_used_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, tagID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tag usage count")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "tag added to resource"})
}

func removeTagFromResource(w http.ResponseWriter, resourceType, resourceID, tagIdentifier string) {
	if !resourceExists(resourceType, resourceID) {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	tagIdentifier = strings.ToLower(strings.TrimSpace(tagIdentifier))
	if tagIdentifier == "" {
		writeError(w, http.StatusBadRequest, "tag identifier is required")
		return
	}

	tx, err := db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start transaction")
		return
	}
	defer tx.Rollback()

	var tagID int64
	err = tx.QueryRow(`
		SELECT id FROM tags WHERE id = ? OR name = ?
	`, tagIdentifier, tagIdentifier).Scan(&tagID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tag not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get tag")
		return
	}

	var relID int64
	err = tx.QueryRow(`
		SELECT id FROM resource_tags
		WHERE resource_type = ? AND resource_id = ? AND tag_id = ?
	`, resourceType, resourceID, tagID).Scan(&relID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "tag not found on resource")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to check tag relation")
		return
	}

	_, err = tx.Exec(`
		DELETE FROM resource_tags WHERE id = ?
	`, relID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to remove tag from resource")
		return
	}

	_, err = tx.Exec(`
		UPDATE tags SET usage_count = MAX(0, usage_count - 1) WHERE id = ?
	`, tagID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update tag usage count")
		return
	}

	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to commit transaction")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "tag removed from resource"})
}

func getTagsForResource(w http.ResponseWriter, resourceType, resourceID string) {
	if !resourceExists(resourceType, resourceID) {
		writeError(w, http.StatusNotFound, "resource not found")
		return
	}

	rows, err := db.Query(`
		SELECT rt.id, rt.resource_type, rt.resource_id, rt.tag_id, 
		       t.name, t.color, rt.created_at
		FROM resource_tags rt
		INNER JOIN tags t ON rt.tag_id = t.id
		WHERE rt.resource_type = ? AND rt.resource_id = ?
		ORDER BY t.name ASC
	`, resourceType, resourceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get resource tags")
		return
	}
	defer rows.Close()

	resourceTags := []ResourceTag{}
	for rows.Next() {
		var rt ResourceTag
		if err := rows.Scan(
			&rt.ID, &rt.ResourceType, &rt.ResourceID, &rt.TagID,
			&rt.TagName, &rt.TagColor, &rt.CreatedAt,
		); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan resource tags")
			return
		}
		resourceTags = append(resourceTags, rt)
	}

	writeJSON(w, http.StatusOK, resourceTags)
}

func handleSearchResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tagsParam := r.URL.Query()["tags"]
	if len(tagsParam) == 0 {
		writeError(w, http.StatusBadRequest, "at least one tag is required")
		return
	}

	var tagNames []string
	for _, t := range tagsParam {
		tagNames = append(tagNames, strings.Split(t, ",")...)
	}

	uniqueTags := make(map[string]bool)
	for _, t := range tagNames {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" {
			uniqueTags[t] = true
		}
	}

	if len(uniqueTags) == 0 {
		writeError(w, http.StatusBadRequest, "at least one valid tag is required")
		return
	}

	tagList := make([]string, 0, len(uniqueTags))
	for t := range uniqueTags {
		tagList = append(tagList, t)
	}

	placeholders := make([]string, len(tagList))
	args := make([]interface{}, len(tagList))
	for i, t := range tagList {
		placeholders[i] = "?"
		args[i] = t
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT rt.resource_type, rt.resource_id
		FROM resource_tags rt
		INNER JOIN tags t ON rt.tag_id = t.id
		WHERE t.name IN (%s)
		GROUP BY rt.resource_type, rt.resource_id
		HAVING COUNT(DISTINCT t.name) = ?
	`, strings.Join(placeholders, ","))

	args = append(args, len(tagList))

	rows, err := db.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to search resources")
		return
	}
	defer rows.Close()

	type Resource struct {
		ResourceType string `json:"resource_type"`
		ResourceID   string `json:"resource_id"`
	}

	resources := []Resource{}
	for rows.Next() {
		var res Resource
		if err := rows.Scan(&res.ResourceType, &res.ResourceID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to scan resources")
			return
		}
		resources = append(resources, res)
	}

	writeJSON(w, http.StatusOK, resources)
}

func resourceExists(resourceType, resourceID string) bool {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM resource_tags 
			WHERE resource_type = ? AND resource_id = ?
		)
	`, resourceType, resourceID).Scan(&exists)

	if err != nil {
		return false
	}

	return exists
}
