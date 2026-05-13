package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"usercomments/internal/database"
	"usercomments/internal/model"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Paging  *Paging     `json:"paging,omitempty"`
}

type Paging struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int   `json:"total"`
	TotalPage int   `json:"total_page"`
}

type CreateCommentRequest struct {
	ArticleID   int64  `json:"article_id"`
	UserID      int64  `json:"user_id"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	ParentID    *int64 `json:"parent_id,omitempty"`
}

type DeleteCommentRequest struct {
	UserID    int64 `json:"user_id"`
	IsAdmin   bool  `json:"is_admin"`
}

func writeJSON(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func parsePaging(r *http.Request) (page, pageSize int) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page = 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize = 10
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	return
}

func GetArticlesHandler(w http.ResponseWriter, r *http.Request) {
	articles, err := database.GetArticles()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "获取文章列表失败",
		})
		return
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    articles,
	})
}

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "请求参数格式错误",
		})
		return
	}

	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "评论内容不能为空",
		})
		return
	}

	runeCount := len([]rune(req.Content))
	if runeCount > model.MaxContentLength {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   fmt.Sprintf("评论内容长度超过限制，最多 %d 字，当前 %d 字", model.MaxContentLength, runeCount),
		})
		return
	}

	if req.ArticleID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "文章 ID 无效",
		})
		return
	}

	exists, err := database.ArticleExists(req.ArticleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "检查文章是否存在失败",
		})
		return
	}

	if !exists {
		writeJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "文章不存在",
		})
		return
	}

	canComment, err := database.CanUserComment(req.UserID, req.ArticleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "检查评论限制失败",
		})
		return
	}

	if !canComment {
		writeJSON(w, http.StatusTooManyRequests, Response{
			Success: false,
			Error:   fmt.Sprintf("同一用户同一篇文章 %d 秒内不能重复提交评论", model.RateLimitSeconds),
		})
		return
	}

	replyLevel := 1
	var rootParentID *int64
	var actualParentID *int64

	if req.ParentID != nil {
		parentComment, err := database.GetCommentByID(*req.ParentID)
		if err != nil {
			if errors.Is(err, errors.New("comment not found")) {
				writeJSON(w, http.StatusNotFound, Response{
					Success: false,
					Error:   "父评论不存在",
				})
				return
			}
			writeJSON(w, http.StatusInternalServerError, Response{
				Success: false,
				Error:   "获取父评论失败",
			})
			return
		}

		parentLevel := parentComment.ReplyLevel

		if parentLevel >= model.MaxReplyLevel {
			replyLevel = model.MaxReplyLevel
			actualParentID = parentComment.ParentID
			rootParentID = &parentComment.ID
			if parentComment.RootParentID != nil {
				rootParentID = parentComment.RootParentID
			}
		} else {
			replyLevel = parentLevel + 1
			actualParentID = req.ParentID
			if parentLevel == 2 {
				rootParentID = &parentComment.ID
			}
		}
	}

	contentType := req.ContentType
	if contentType != model.ContentTypeRich {
		contentType = model.ContentTypePlain
	}

	comment := &model.Comment{
		ArticleID:    req.ArticleID,
		UserID:       req.UserID,
		Nickname:     req.Nickname,
		Avatar:       req.Avatar,
		Content:      req.Content,
		ContentType:  contentType,
		ParentID:     actualParentID,
		ReplyLevel:   replyLevel,
		RootParentID: rootParentID,
	}

	if err := database.CreateComment(comment); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "创建评论失败",
		})
		return
	}

	notifyCacheUpdate()

	writeJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    comment,
	})
}

func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	articleIDStr := r.URL.Query().Get("article_id")
	sortBy := r.URL.Query().Get("sort_by")

	if articleIDStr == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "缺少 article_id 参数",
		})
		return
	}

	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil || articleID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "article_id 参数无效",
		})
		return
	}

	exists, err := database.ArticleExists(articleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "检查文章是否存在失败",
		})
		return
	}

	if !exists {
		writeJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   "文章不存在",
		})
		return
	}

	page, pageSize := parsePaging(r)

	comments, total, err := database.GetCommentsByArticle(articleID, page, pageSize, sortBy)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "获取评论列表失败",
		})
		return
	}

	treeComments, err := database.BuildCommentTree(comments)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "构建评论树失败",
		})
		return
	}

	totalPage := 0
	if total > 0 {
		totalPage = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    treeComments,
		Paging: &Paging{
			Page:      page,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPage,
		},
	})
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 2 {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "评论 ID 无效",
		})
		return
	}

	commentIDStr := parts[len(parts)-1]
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "评论 ID 无效",
		})
		return
	}

	var req DeleteCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "请求参数格式错误",
		})
		return
	}

	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		if errors.Is(err, errors.New("comment not found")) {
			writeJSON(w, http.StatusNotFound, Response{
				Success: false,
				Error:   "评论不存在",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "获取评论失败",
		})
		return
	}

	if comment.IsDeleted {
		writeJSON(w, http.StatusOK, Response{
			Success: true,
			Data:    map[string]string{"message": "评论已删除"},
		})
		return
	}

	if !req.IsAdmin && comment.UserID != req.UserID {
		writeJSON(w, http.StatusForbidden, Response{
			Success: false,
			Error:   "无权删除他人评论",
		})
		return
	}

	if err := database.SoftDeleteComment(commentID); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "删除评论失败",
		})
		return
	}

	notifyCacheUpdate()

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    map[string]string{"message": "删除成功"},
	})
}

func SearchCommentsHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "缺少 keyword 参数",
		})
		return
	}

	page, pageSize := parsePaging(r)

	comments, total, err := database.SearchComments(keyword, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "搜索评论失败",
		})
		return
	}

	for i := range comments {
		if comments[i].IsDeleted {
			comments[i].Content = "该评论已删除"
		}
	}

	totalPage := 0
	if total > 0 {
		totalPage = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    comments,
		Paging: &Paging{
			Page:      page,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPage,
		},
	})
}

func SearchCommentsByNicknameHandler(w http.ResponseWriter, r *http.Request) {
	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "缺少 nickname 参数",
		})
		return
	}

	page, pageSize := parsePaging(r)

	comments, total, err := database.SearchCommentsByNickname(nickname, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "搜索用户评论失败",
		})
		return
	}

	for i := range comments {
		if comments[i].IsDeleted {
			comments[i].Content = "该评论已删除"
		}
	}

	totalPage := 0
	if total > 0 {
		totalPage = int(math.Ceil(float64(total) / float64(pageSize)))
	}

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    comments,
		Paging: &Paging{
			Page:      page,
			PageSize:  pageSize,
			Total:     total,
			TotalPage: totalPage,
		},
	})
}

func LikeCommentHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "评论 ID 无效",
		})
		return
	}

	commentIDStr := parts[len(parts)-2]
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil || commentID <= 0 {
		writeJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "评论 ID 无效",
		})
		return
	}

	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		if errors.Is(err, errors.New("comment not found")) {
			writeJSON(w, http.StatusNotFound, Response{
				Success: false,
				Error:   "评论不存在",
			})
			return
		}
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "获取评论失败",
		})
		return
	}

	if err := database.LikeComment(commentID); err != nil {
		writeJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   "点赞失败",
		})
		return
	}

	comment.Likes++

	notifyCacheUpdate()

	writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    map[string]int{"likes": comment.Likes},
	})
}

func notifyCacheUpdate() {
	fmt.Println("[Cache] 通知关联模块更新缓存...")
}
