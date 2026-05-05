package common

import "time"

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type CreateArticleRequest struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	AuthorID      string   `json:"author_id"`
	AccessLevel   AccessLevel `json:"access_level"`
	DepartmentIDs []string `json:"department_ids,omitempty"`
}

type UpdateArticleRequest struct {
	Title         string   `json:"title"`
	Content       string   `json:"content"`
	Category      string   `json:"category"`
	Tags          []string `json:"tags"`
	AuthorID      string   `json:"author_id"`
	AccessLevel   AccessLevel `json:"access_level"`
	DepartmentIDs []string `json:"department_ids,omitempty"`
	RelatedArticleIDs []string `json:"related_article_ids,omitempty"`
}

type PublishArticleRequest struct {
	AuthorID string `json:"author_id"`
}

type ArchiveArticleRequest struct {
	OperatorID string `json:"operator_id"`
}

type RollbackRequest struct {
	VersionNum int    `json:"version_num"`
	AuthorID   string `json:"author_id"`
}

type SearchRequest struct {
	Keyword    string `json:"keyword"`
	UserID     string `json:"user_id"`
	Department string `json:"department"`
	IsLoggedIn bool   `json:"is_logged_in"`
}

type SearchResult struct {
	Article       *Article `json:"article"`
	IsArchived    bool     `json:"is_archived"`
	Relevance     float64  `json:"relevance"`
}

type HotArticle struct {
	ArticleID   string `json:"article_id"`
	Title       string `json:"title"`
	ViewCount   int    `json:"view_count"`
	ReferenceCount int `json:"reference_count"`
}

type Stats struct {
	TotalArticles int64 `json:"total_articles"`
	TotalSize     int64 `json:"total_size_bytes"`
	DraftCount    int   `json:"draft_count"`
	PublishedCount int  `json:"published_count"`
	ArchivedCount int  `json:"archived_count"`
}

type VersionCompareResult struct {
	ArticleID    string `json:"article_id"`
	Version1     int    `json:"version_1"`
	Version2     int    `json:"version_2"`
	DiffContent  string `json:"diff_content"`
	CreatedAt1   time.Time `json:"created_at_1"`
	CreatedAt2   time.Time `json:"created_at_2"`
}
