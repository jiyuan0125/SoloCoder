package main

import (
	"encoding/json"
	"kb/common"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) respond(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resp := common.Response{
		Code:    code,
		Message: common.GetErrMessage(code),
		Data:    data,
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var req common.CreateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	if req.Title == "" || req.Content == "" || req.AuthorID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	article, _, err := h.store.CreateArticle(&req)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, article)
}

func (h *Handler) GetArticle(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	if articleID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	userID := r.URL.Query().Get("user_id")
	department := r.URL.Query().Get("department")
	isLoggedIn := r.URL.Query().Get("is_logged_in") == "true"

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.Status == common.StatusDraft && article.AuthorID != userID {
		h.respond(w, common.ErrCodeAccessDenied, nil)
		return
	}

	if !h.canAccess(article, userID, department, isLoggedIn) {
		h.respond(w, common.ErrCodeAccessDenied, nil)
		return
	}

	h.store.IncrementViewCount(articleID)

	result := map[string]interface{}{
		"article":        article,
		"related_articles": h.store.GetRelatedArticles(articleID),
	}

	h.respond(w, common.ErrCodeSuccess, result)
}

func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	articleID = strings.TrimSuffix(articleID, "/update")

	if articleID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	var req common.UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.Status == common.StatusArchived {
		h.respond(w, common.ErrCodeAlreadyArchived, nil)
		return
	}

	if article.AuthorID != req.AuthorID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	if article.Status != common.StatusDraft {
		h.respond(w, common.ErrCodeDraftOnly, nil)
		return
	}

	updatedArticle, _, err := h.store.UpdateArticle(articleID, &req)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	if updatedArticle == nil {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, updatedArticle)
}

func (h *Handler) PublishArticle(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	articleID = strings.TrimSuffix(articleID, "/publish")

	if articleID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	var req common.PublishArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.Status == common.StatusArchived {
		h.respond(w, common.ErrCodeAlreadyArchived, nil)
		return
	}

	if article.AuthorID != req.AuthorID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	publishedArticle, err := h.store.PublishArticle(articleID, req.AuthorID)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	if publishedArticle == nil {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, publishedArticle)
}

func (h *Handler) ArchiveArticle(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	articleID = strings.TrimSuffix(articleID, "/archive")

	if articleID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	var req common.ArchiveArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.Status == common.StatusArchived {
		h.respond(w, common.ErrCodeAlreadyArchived, nil)
		return
	}

	if article.AuthorID != req.OperatorID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	archivedArticle, err := h.store.ArchiveArticle(articleID, req.OperatorID)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	if archivedArticle == nil {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, archivedArticle)
}

func (h *Handler) GetVersions(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	articleID = strings.TrimSuffix(articleID, "/versions")

	if articleID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	userID := r.URL.Query().Get("user_id")

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.AuthorID != userID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	versions, _ := h.store.GetVersions(articleID)
	h.respond(w, common.ErrCodeSuccess, versions)
}

func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/articles/"), "/")
	if len(parts) < 3 || parts[1] != "versions" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	articleID := parts[0]
	versionNum, err := strconv.Atoi(parts[2])
	if err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	userID := r.URL.Query().Get("user_id")

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.AuthorID != userID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	version, exists := h.store.GetVersion(articleID, versionNum)
	if !exists {
		h.respond(w, common.ErrCodeVersionNotFound, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, version)
}

func (h *Handler) RollbackArticle(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/articles/")
	articleID = strings.TrimSuffix(articleID, "/rollback")

	if articleID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	var req common.RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.Status == common.StatusArchived {
		h.respond(w, common.ErrCodeAlreadyArchived, nil)
		return
	}

	if article.AuthorID != req.AuthorID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	rolledBackArticle, newVersion, err := h.store.RollbackToVersion(articleID, req.VersionNum, req.AuthorID)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	if rolledBackArticle == nil {
		h.respond(w, common.ErrCodeVersionNotFound, nil)
		return
	}

	result := map[string]interface{}{
		"article":      rolledBackArticle,
		"new_version":  newVersion,
	}

	h.respond(w, common.ErrCodeSuccess, result)
}

func (h *Handler) CompareVersions(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/articles/"), "/")
	if len(parts) < 5 || parts[1] != "versions" || parts[3] != "compare" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	articleID := parts[0]
	v1, err1 := strconv.Atoi(parts[2])
	v2, err2 := strconv.Atoi(parts[4])
	if err1 != nil || err2 != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	userID := r.URL.Query().Get("user_id")

	article, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	if article.AuthorID != userID {
		h.respond(w, common.ErrCodeNotAuthor, nil)
		return
	}

	version1, exists1 := h.store.GetVersion(articleID, v1)
	version2, exists2 := h.store.GetVersion(articleID, v2)
	if !exists1 || !exists2 {
		h.respond(w, common.ErrCodeVersionNotFound, nil)
		return
	}

	result := CompareVersions(version1, version2)
	h.respond(w, common.ErrCodeSuccess, result)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	var req common.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	results := h.store.Search(req.Keyword, req.UserID, req.Department, req.IsLoggedIn)

	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})

	h.respond(w, common.ErrCodeSuccess, results)
}

func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/favorites/")
	userID := r.URL.Query().Get("user_id")

	if articleID == "" || userID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	_, exists := h.store.GetArticle(articleID)
	if !exists {
		h.respond(w, common.ErrCodeArticleNotFound, nil)
		return
	}

	err := h.store.AddFavorite(userID, articleID)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, nil)
}

func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	articleID := strings.TrimPrefix(r.URL.Path, "/favorites/")
	userID := r.URL.Query().Get("user_id")

	if articleID == "" || userID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	err := h.store.RemoveFavorite(userID, articleID)
	if err != nil {
		h.respond(w, common.ErrCodeInternal, nil)
		return
	}

	h.respond(w, common.ErrCodeSuccess, nil)
}

func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimPrefix(r.URL.Path, "/favorites/")
	if userID == "" {
		h.respond(w, common.ErrCodeBadRequest, nil)
		return
	}

	favorites := h.store.GetFavorites(userID)

	articles := make([]*common.Article, 0)
	for _, fav := range favorites {
		if article, exists := h.store.GetArticle(fav.ArticleID); exists {
			articles = append(articles, article)
		}
	}

	h.respond(w, common.ErrCodeSuccess, articles)
}

func (h *Handler) GetHotArticles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	hot := h.store.GetHotArticles(limit)
	h.respond(w, common.ErrCodeSuccess, hot)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.store.GetStats()
	h.respond(w, common.ErrCodeSuccess, stats)
}

func (h *Handler) canAccess(article *common.Article, userID string, department string, isLoggedIn bool) bool {
	switch article.AccessLevel {
	case common.AccessPublic:
		return true
	case common.AccessLoggedIn:
		return isLoggedIn
	case common.AccessDepartment:
		if !isLoggedIn {
			return false
		}
		for _, deptID := range article.DepartmentIDs {
			if deptID == department {
				return true
			}
		}
		return false
	default:
		return true
	}
}
