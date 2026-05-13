package article

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blog-engine/internal/render"
	"blog-engine/internal/user"
	"blog-engine/pkg/db"
	"blog-engine/pkg/utils"
)

type Article struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Slug            string    `json:"slug"`
	ContentMarkdown string    `json:"content_markdown,omitempty"`
	ContentHTML     string    `json:"content_html,omitempty"`
	Excerpt         string    `json:"excerpt"`
	Status          string    `json:"status"`
	AuthorID        int       `json:"author_id"`
	Author          *AuthorInfo `json:"author,omitempty"`
	CategoryID      *int      `json:"category_id,omitempty"`
	Category        *Category `json:"category,omitempty"`
	Tags            []Tag     `json:"tags,omitempty"`
	Views           int       `json:"views"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
}

type AuthorInfo struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type Category struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ArticleCount int   `json:"article_count,omitempty"`
}

type Tag struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ArticleCount int    `json:"article_count,omitempty"`
}

type CreateArticleRequest struct {
	Title           string   `json:"title"`
	ContentMarkdown string   `json:"content_markdown"`
	Status          string   `json:"status"`
	CategoryID      *int     `json:"category_id"`
	Tags            []string `json:"tags"`
}

type UpdateArticleRequest struct {
	Title           *string   `json:"title"`
	ContentMarkdown *string   `json:"content_markdown"`
	Status          *string   `json:"status"`
	CategoryID      *int      `json:"category_id"`
	Tags            *[]string `json:"tags"`
}

type HotArticle struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	Views      int    `json:"views"`
	RecentViews int   `json:"recent_views"`
}

func CreateArticle(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil {
		utils.JSONError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	var req CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.ContentMarkdown) == "" {
		utils.JSONError(w, http.StatusBadRequest, "Title and content are required")
		return
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}
	if !isValidStatus(status) {
		utils.JSONError(w, http.StatusBadRequest, "Invalid status")
		return
	}

	contentHTML := render.MarkdownToHTML(req.ContentMarkdown)
	excerpt := render.GenerateExcerpt(contentHTML, 200)
	baseSlug := utils.GenerateSlug(req.Title)
	slug := baseSlug
	counter := 1

	for {
		var exists int
		err := db.DB.QueryRow(`SELECT COUNT(*) FROM articles WHERE slug = ?`, slug).Scan(&exists)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Database error")
			return
		}
		if exists == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	tx, err := db.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	var publishedAt *time.Time
	if status == "published" {
		now := time.Now()
		publishedAt = &now
	}

	result, err := tx.Exec(
		`INSERT INTO articles (title, slug, content_markdown, content_html, excerpt, status, author_id, category_id, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.Title, slug, req.ContentMarkdown, contentHTML, excerpt, status, currentUser.ID, req.CategoryID, publishedAt,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create article")
		return
	}

	articleID, _ := result.LastInsertId()

	for _, tagName := range req.Tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}

		tagSlug := utils.GenerateSlug(tagName)

		_, err = tx.Exec(`INSERT OR IGNORE INTO tags (name, slug) VALUES (?, ?)`, tagName, tagSlug)
		if err != nil {
			continue
		}

		var tagID int
		err = tx.QueryRow(`SELECT id FROM tags WHERE slug = ?`, tagSlug).Scan(&tagID)
		if err != nil {
			continue
		}

		_, err = tx.Exec(`INSERT OR IGNORE INTO article_tags (article_id, tag_id) VALUES (?, ?)`, articleID, tagID)
		if err != nil {
			continue
		}
	}

	if err = tx.Commit(); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create article")
		return
	}

	article, err := getArticleByID(int(articleID), currentUser)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to retrieve article")
		return
	}

	utils.JSONSuccess(w, http.StatusCreated, article)
}

func GetArticle(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	idStr := pathParts[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	currentUser := user.UserFromContext(r.Context())
	article, err := getArticleByID(id, currentUser)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "Article not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if article.Status != "published" {
		if currentUser == nil {
			utils.JSONError(w, http.StatusNotFound, "Article not found")
			return
		}
		if currentUser.Role != "admin" && currentUser.ID != article.AuthorID {
			utils.JSONError(w, http.StatusNotFound, "Article not found")
			return
		}
	}

	if article.Status == "published" {
		incrementViewCount(r, id)
	}

	utils.JSONSuccess(w, http.StatusOK, article)
}

func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil {
		utils.JSONError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	idStr := pathParts[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	var existingArticle Article
	err = db.DB.QueryRow(
		`SELECT id, author_id, status FROM articles WHERE id = ?`,
		id,
	).Scan(&existingArticle.ID, &existingArticle.AuthorID, &existingArticle.Status)

	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "Article not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if currentUser.Role != "admin" && currentUser.ID != existingArticle.AuthorID {
		utils.JSONError(w, http.StatusForbidden, "You can only edit your own articles")
		return
	}

	var req UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	updates := []string{}
	args := []interface{}{}

	if req.Title != nil {
		title := *req.Title
		if strings.TrimSpace(title) == "" {
			utils.JSONError(w, http.StatusBadRequest, "Title cannot be empty")
			return
		}
		updates = append(updates, "title = ?")
		args = append(args, title)

		baseSlug := utils.GenerateSlug(title)
		slug := baseSlug
		counter := 1
		for {
			var exists int
			err := tx.QueryRow(`SELECT COUNT(*) FROM articles WHERE slug = ? AND id != ?`, slug, id).Scan(&exists)
			if err != nil {
				utils.JSONError(w, http.StatusInternalServerError, "Database error")
				return
			}
			if exists == 0 {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, counter)
			counter++
		}
		updates = append(updates, "slug = ?")
		args = append(args, slug)
	}

	if req.ContentMarkdown != nil {
		content := *req.ContentMarkdown
		if strings.TrimSpace(content) == "" {
			utils.JSONError(w, http.StatusBadRequest, "Content cannot be empty")
			return
		}
		contentHTML := render.MarkdownToHTML(content)
		excerpt := render.GenerateExcerpt(contentHTML, 200)
		updates = append(updates, "content_markdown = ?, content_html = ?, excerpt = ?")
		args = append(args, content, contentHTML, excerpt)
	}

	if req.CategoryID != nil {
		updates = append(updates, "category_id = ?")
		args = append(args, *req.CategoryID)
	}

	oldStatus := existingArticle.Status
	var newStatus string
	if req.Status != nil {
		newStatus = *req.Status
		if !isValidStatus(newStatus) {
			utils.JSONError(w, http.StatusBadRequest, "Invalid status")
			return
		}
		updates = append(updates, "status = ?")
		args = append(args, newStatus)

		if newStatus == "published" && oldStatus != "published" {
			updates = append(updates, "published_at = CURRENT_TIMESTAMP")
		}
	}

	if len(updates) > 0 {
		updates = append(updates, "updated_at = CURRENT_TIMESTAMP")
		args = append(args, id)

		query := `UPDATE articles SET ` + strings.Join(updates, ", ") + ` WHERE id = ?`
		_, err = tx.Exec(query, args...)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Failed to update article")
			return
		}

		if req.Status != nil {
			_, err = tx.Exec(
				`INSERT INTO article_history (article_id, action, old_status, new_status, performed_by)
				 VALUES (?, 'status_change', ?, ?, ?)`,
				id, oldStatus, newStatus, currentUser.ID,
			)
			if err != nil {
				utils.JSONError(w, http.StatusInternalServerError, "Failed to record history")
				return
			}
		}
	}

	if req.Tags != nil {
		_, err = tx.Exec(`DELETE FROM article_tags WHERE article_id = ?`, id)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Failed to update tags")
			return
		}

		for _, tagName := range *req.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			tagSlug := utils.GenerateSlug(tagName)
			_, err = tx.Exec(`INSERT OR IGNORE INTO tags (name, slug) VALUES (?, ?)`, tagName, tagSlug)
			if err != nil {
				continue
			}

			var tagID int
			err = tx.QueryRow(`SELECT id FROM tags WHERE slug = ?`, tagSlug).Scan(&tagID)
			if err != nil {
				continue
			}

			_, err = tx.Exec(`INSERT OR IGNORE INTO article_tags (article_id, tag_id) VALUES (?, ?)`, id, tagID)
			if err != nil {
				continue
			}
		}
	}

	if err = tx.Commit(); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update article")
		return
	}

	article, err := getArticleByID(id, currentUser)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to retrieve article")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, article)
}

func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil {
		utils.JSONError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	idStr := pathParts[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	var authorID int
	err = db.DB.QueryRow(`SELECT author_id FROM articles WHERE id = ?`, id).Scan(&authorID)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "Article not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if currentUser.Role != "admin" && currentUser.ID != authorID {
		utils.JSONError(w, http.StatusForbidden, "You can only delete your own articles")
		return
	}

	_, err = db.DB.Exec(`DELETE FROM articles WHERE id = ?`, id)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to delete article")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, map[string]string{"message": "Article deleted successfully"})
}

func ListArticles(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	query := r.URL.Query()

	baseQuery := `SELECT a.id, a.title, a.slug, a.excerpt, a.status, a.author_id, a.category_id, a.views, a.created_at, a.updated_at, a.published_at
				  FROM articles a`
	whereClauses := []string{}
	args := []interface{}{}

	if currentUser == nil || (currentUser.Role != "admin") {
		whereClauses = append(whereClauses, "a.status = 'published'")
	}

	if categoryID := query.Get("category_id"); categoryID != "" {
		whereClauses = append(whereClauses, "a.category_id = ?")
		args = append(args, categoryID)
	}

	if categorySlug := query.Get("category"); categorySlug != "" {
		baseQuery += ` JOIN categories c ON a.category_id = c.id`
		whereClauses = append(whereClauses, "c.slug = ?")
		args = append(args, categorySlug)
	}

	if tagSlug := query.Get("tag"); tagSlug != "" {
		baseQuery += ` JOIN article_tags at ON a.id = at.article_id JOIN tags t ON at.tag_id = t.id`
		whereClauses = append(whereClauses, "t.slug = ?")
		args = append(args, tagSlug)
	}

	if authorID := query.Get("author_id"); authorID != "" {
		whereClauses = append(whereClauses, "a.author_id = ?")
		args = append(args, authorID)
	}

	if currentUser != nil && query.Get("mine") == "true" {
		whereClauses = append(whereClauses, "a.author_id = ?")
		args = append(args, currentUser.ID)
	}

	if status := query.Get("status"); status != "" && currentUser != nil {
		if currentUser.Role == "admin" {
			whereClauses = append(whereClauses, "a.status = ?")
			args = append(args, status)
		} else {
			whereClauses = append(whereClauses, "a.status = ? AND a.author_id = ?")
			args = append(args, status, currentUser.ID)
		}
	}

	if len(whereClauses) > 0 {
		baseQuery += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	baseQuery += " ORDER BY a.created_at DESC"

	page := 1
	if p := query.Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}
	pageSize := 10
	if ps := query.Get("page_size"); ps != "" {
		if psi, err := strconv.Atoi(ps); err == nil && psi > 0 && psi <= 100 {
			pageSize = psi
		}
	}

	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	rows, err := db.DB.Query(baseQuery, args...)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	articles := []*Article{}
	for rows.Next() {
		var article Article
		var categoryID sql.NullInt64
		var publishedAt sql.NullTime

		err := rows.Scan(
			&article.ID, &article.Title, &article.Slug, &article.Excerpt, &article.Status,
			&article.AuthorID, &categoryID, &article.Views, &article.CreatedAt, &article.UpdatedAt,
			&publishedAt,
		)
		if err != nil {
			continue
		}

		if categoryID.Valid {
			cid := int(categoryID.Int64)
			article.CategoryID = &cid
		}
		if publishedAt.Valid {
			article.PublishedAt = &publishedAt.Time
		}

		article.Author = getAuthorInfo(article.AuthorID)
		article.Category = getCategoryInfo(article.CategoryID)
		article.Tags = getArticleTags(article.ID)

		articles = append(articles, &article)
	}

	utils.JSONSuccess(w, http.StatusOK, articles)
}

func GetHotArticles(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT a.id, a.title, a.slug, a.views, COALESCE(v.recent_views, 0) as recent_views
		FROM articles a
		LEFT JOIN (
			SELECT article_id, COUNT(*) as recent_views
			FROM article_views
			WHERE view_date >= datetime('now', '-7 days')
			GROUP BY article_id
		) v ON a.id = v.article_id
		WHERE a.status = 'published'
		ORDER BY recent_views DESC, a.views DESC
		LIMIT 10
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	hotArticles := []HotArticle{}
	for rows.Next() {
		var ha HotArticle
		err := rows.Scan(&ha.ID, &ha.Title, &ha.Slug, &ha.Views, &ha.RecentViews)
		if err != nil {
			continue
		}
		hotArticles = append(hotArticles, ha)
	}

	utils.JSONSuccess(w, http.StatusOK, hotArticles)
}

func ListCategories(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT c.id, c.name, c.slug, c.description, COUNT(a.id) as article_count
		FROM categories c
		LEFT JOIN articles a ON c.id = a.category_id AND a.status = 'published'
		GROUP BY c.id
		ORDER BY article_count DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var cat Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.ArticleCount)
		if err != nil {
			continue
		}
		categories = append(categories, cat)
	}

	utils.JSONSuccess(w, http.StatusOK, categories)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil || currentUser.Role != "admin" {
		utils.JSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var cat Category
	if err := json.NewDecoder(r.Body).Decode(&cat); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if strings.TrimSpace(cat.Name) == "" {
		utils.JSONError(w, http.StatusBadRequest, "Name is required")
		return
	}

	slug := cat.Slug
	if slug == "" {
		slug = utils.GenerateSlug(cat.Name)
	}

	result, err := db.DB.Exec(
		`INSERT INTO categories (name, slug, description) VALUES (?, ?, ?)`,
		cat.Name, slug, cat.Description,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}

	id, _ := result.LastInsertId()
	cat.ID = int(id)
	cat.Slug = slug

	utils.JSONSuccess(w, http.StatusCreated, cat)
}

func ListTags(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT t.id, t.name, t.slug, COUNT(at.article_id) as article_count
		FROM tags t
		LEFT JOIN article_tags at ON t.id = at.tag_id
		LEFT JOIN articles a ON at.article_id = a.id AND a.status = 'published'
		GROUP BY t.id
		ORDER BY article_count DESC
	`

	rows, err := db.DB.Query(query)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.ArticleCount)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}

	utils.JSONSuccess(w, http.StatusOK, tags)
}

func getArticleByID(id int, currentUser *utils.UserInfo) (*Article, error) {
	var article Article
	var categoryID sql.NullInt64
	var publishedAt sql.NullTime

	err := db.DB.QueryRow(
		`SELECT id, title, slug, content_markdown, content_html, excerpt, status, author_id, 
		 category_id, views, created_at, updated_at, published_at
		 FROM articles WHERE id = ?`,
		id,
	).Scan(
		&article.ID, &article.Title, &article.Slug, &article.ContentMarkdown, &article.ContentHTML,
		&article.Excerpt, &article.Status, &article.AuthorID, &categoryID, &article.Views,
		&article.CreatedAt, &article.UpdatedAt, &publishedAt,
	)

	if err != nil {
		return nil, err
	}

	if categoryID.Valid {
		cid := int(categoryID.Int64)
		article.CategoryID = &cid
	}
	if publishedAt.Valid {
		article.PublishedAt = &publishedAt.Time
	}

	article.Author = getAuthorInfo(article.AuthorID)
	article.Category = getCategoryInfo(article.CategoryID)
	article.Tags = getArticleTags(article.ID)

	return &article, nil
}

func getAuthorInfo(authorID int) *AuthorInfo {
	var info AuthorInfo
	err := db.DB.QueryRow(
		`SELECT id, username, display_name FROM users WHERE id = ?`,
		authorID,
	).Scan(&info.ID, &info.Username, &info.DisplayName)
	if err != nil {
		return nil
	}
	return &info
}

func getCategoryInfo(categoryID *int) *Category {
	if categoryID == nil {
		return nil
	}
	var cat Category
	err := db.DB.QueryRow(
		`SELECT id, name, slug, description FROM categories WHERE id = ?`,
		*categoryID,
	).Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description)
	if err != nil {
		return nil
	}
	return &cat
}

func getArticleTags(articleID int) []Tag {
	rows, err := db.DB.Query(
		`SELECT t.id, t.name, t.slug FROM tags t
		 JOIN article_tags at ON t.id = at.tag_id
		 WHERE at.article_id = ?`,
		articleID,
	)
	if err != nil {
		return []Tag{}
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var tag Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags
}

func incrementViewCount(r *http.Request, articleID int) {
	visitorIP := utils.GetClientIP(r)
	viewHour := utils.GetCurrentHour()

	result, err := db.DB.Exec(
		`INSERT OR IGNORE INTO article_views (article_id, visitor_ip, view_hour) VALUES (?, ?, ?)`,
		articleID, visitorIP, viewHour,
	)
	if err != nil {
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		db.DB.Exec(`UPDATE articles SET views = views + 1 WHERE id = ?`, articleID)
	}
}

func isValidStatus(status string) bool {
	return status == "draft" || status == "published" || status == "unpublished"
}
