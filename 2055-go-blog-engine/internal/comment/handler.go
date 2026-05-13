package comment

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"blog-engine/internal/user"
	"blog-engine/pkg/db"
	"blog-engine/pkg/utils"
)

type Comment struct {
	ID         int       `json:"id"`
	ArticleID  int       `json:"article_id"`
	ParentID   *int      `json:"parent_id,omitempty"`
	ReplyLevel int       `json:"reply_level"`
	UserID     *int      `json:"user_id,omitempty"`
	GuestName  string    `json:"guest_name,omitempty"`
	GuestEmail string    `json:"guest_email,omitempty"`
	Author     *AuthorInfo `json:"author,omitempty"`
	Content    string    `json:"content"`
	Status     string    `json:"status"`
	IsDeleted  bool      `json:"is_deleted"`
	CreatedAt  time.Time `json:"created_at"`
	Replies    []*Comment `json:"replies,omitempty"`
}

type AuthorInfo struct {
	ID          int    `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type CreateCommentRequest struct {
	ArticleID  int    `json:"article_id"`
	ParentID   *int   `json:"parent_id"`
	GuestName  string `json:"guest_name"`
	GuestEmail string `json:"guest_email"`
	Content    string `json:"content"`
}

type ReviewCommentRequest struct {
	Action string `json:"action"`
}

const maxReplyLevel = 5

func CreateComment(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ArticleID == 0 {
		utils.JSONError(w, http.StatusBadRequest, "Article ID is required")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		utils.JSONError(w, http.StatusBadRequest, "Content is required")
		return
	}

	var articleExists int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM articles WHERE id = ? AND status = 'published'`, req.ArticleID).Scan(&articleExists)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	if articleExists == 0 {
		utils.JSONError(w, http.StatusNotFound, "Article not found")
		return
	}

	if currentUser == nil {
		if strings.TrimSpace(req.GuestName) == "" {
			utils.JSONError(w, http.StatusBadRequest, "Guest name is required")
			return
		}
	}

	replyLevel := 1
	parentID := req.ParentID
	if parentID != nil {
		var parentLevel int
		var parentParentID sql.NullInt64
		err := db.DB.QueryRow(
			`SELECT reply_level, parent_id FROM comments WHERE id = ?`,
			*parentID,
		).Scan(&parentLevel, &parentParentID)

		if err == sql.ErrNoRows {
			utils.JSONError(w, http.StatusNotFound, "Parent comment not found")
			return
		} else if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Database error")
			return
		}

		if parentLevel >= maxReplyLevel {
			replyLevel = maxReplyLevel
			if parentParentID.Valid {
				pid := int(parentParentID.Int64)
				parentID = &pid
			}
		} else {
			replyLevel = parentLevel + 1
		}
	}

	status := "approved"
	if containsSensitiveWords(req.Content) {
		status = "pending"
	}

	tx, err := db.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	var userID *int
	if currentUser != nil {
		userID = &currentUser.ID
	}

	result, err := tx.Exec(
		`INSERT INTO comments (article_id, parent_id, reply_level, user_id, guest_name, guest_email, content, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ArticleID, parentID, replyLevel, userID, req.GuestName, req.GuestEmail, req.Content, status,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	commentID, _ := result.LastInsertId()

	if status != "approved" {
		var performedBy *int
		_, err = tx.Exec(
			`INSERT INTO comment_history (comment_id, action, old_status, new_status, performed_by)
			 VALUES (?, 'create', NULL, ?, ?)`,
			commentID, status, performedBy,
		)
		if err != nil {
			utils.JSONError(w, http.StatusInternalServerError, "Failed to record history")
			return
		}
	}

	if err = tx.Commit(); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to create comment")
		return
	}

	comment, err := getCommentByID(int(commentID), currentUser)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to retrieve comment")
		return
	}

	utils.JSONSuccess(w, http.StatusCreated, comment)
}

func ListComments(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	query := r.URL.Query()

	articleIDStr := query.Get("article_id")
	if articleIDStr == "" {
		utils.JSONError(w, http.StatusBadRequest, "Article ID is required")
		return
	}

	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid article ID")
		return
	}

	whereClauses := []string{"article_id = ?"}
	args := []interface{}{articleID}

	if currentUser == nil || currentUser.Role != "admin" {
		whereClauses = append(whereClauses, "status = 'approved'")
	} else if status := query.Get("status"); status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, status)
	}

	whereClauses = append(whereClauses, "parent_id IS NULL")

	rows, err := db.DB.Query(
		`SELECT id, article_id, parent_id, reply_level, user_id, guest_name, guest_email, 
		 content, status, is_deleted, created_at
		 FROM comments 
		 WHERE `+strings.Join(whereClauses, " AND ")+`
		 ORDER BY created_at DESC`,
		args...,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	comments := []*Comment{}
	for rows.Next() {
		var comment Comment
		var parentID, userID sql.NullInt64

		err := rows.Scan(
			&comment.ID, &comment.ArticleID, &parentID, &comment.ReplyLevel, &userID,
			&comment.GuestName, &comment.GuestEmail, &comment.Content, &comment.Status,
			&comment.IsDeleted, &comment.CreatedAt,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			comment.ParentID = &pid
		}
		if userID.Valid {
			uid := int(userID.Int64)
			comment.UserID = &uid
		}

		comment.Author = getCommentAuthor(&comment)
		comment.Replies = getReplies(comment.ID, currentUser)

		if comment.IsDeleted {
			comment.Content = "[原评论已删除]"
		}

		comments = append(comments, &comment)
	}

	utils.JSONSuccess(w, http.StatusOK, comments)
}

func ListPendingComments(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil || currentUser.Role != "admin" {
		utils.JSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	rows, err := db.DB.Query(
		`SELECT id, article_id, parent_id, reply_level, user_id, guest_name, guest_email,
		 content, status, is_deleted, created_at
		 FROM comments
		 WHERE status = 'pending'
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer rows.Close()

	comments := []*Comment{}
	for rows.Next() {
		var comment Comment
		var parentID, userID sql.NullInt64

		err := rows.Scan(
			&comment.ID, &comment.ArticleID, &parentID, &comment.ReplyLevel, &userID,
			&comment.GuestName, &comment.GuestEmail, &comment.Content, &comment.Status,
			&comment.IsDeleted, &comment.CreatedAt,
		)
		if err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			comment.ParentID = &pid
		}
		if userID.Valid {
			uid := int(userID.Int64)
			comment.UserID = &uid
		}

		comment.Author = getCommentAuthor(&comment)
		comments = append(comments, &comment)
	}

	utils.JSONSuccess(w, http.StatusOK, comments)
}

func ReviewComment(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil || currentUser.Role != "admin" {
		utils.JSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	idStr := pathParts[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var req ReviewCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Action != "approve" && req.Action != "reject" {
		utils.JSONError(w, http.StatusBadRequest, "Invalid action")
		return
	}

	var oldStatus string
	err = db.DB.QueryRow(`SELECT status FROM comments WHERE id = ?`, id).Scan(&oldStatus)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "Comment not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	newStatus := "approved"
	if req.Action == "reject" {
		newStatus = "rejected"
	}

	tx, err := db.DB.Begin()
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE comments SET status = ? WHERE id = ?`, newStatus, id)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to update comment")
		return
	}

	_, err = tx.Exec(
		`INSERT INTO comment_history (comment_id, action, old_status, new_status, performed_by)
		 VALUES (?, 'review', ?, ?, ?)`,
		id, oldStatus, newStatus, currentUser.ID,
	)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to record history")
		return
	}

	if err = tx.Commit(); err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to review comment")
		return
	}

	comment, err := getCommentByID(id, currentUser)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to retrieve comment")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, comment)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	currentUser := user.UserFromContext(r.Context())
	if currentUser == nil {
		utils.JSONError(w, http.StatusUnauthorized, "Not logged in")
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		utils.JSONError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	idStr := pathParts[1]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.JSONError(w, http.StatusBadRequest, "Invalid comment ID")
		return
	}

	var commentUserID sql.NullInt64
	var isDeleted bool
	err = db.DB.QueryRow(`SELECT user_id, is_deleted FROM comments WHERE id = ?`, id).Scan(&commentUserID, &isDeleted)
	if err == sql.ErrNoRows {
		utils.JSONError(w, http.StatusNotFound, "Comment not found")
		return
	} else if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Database error")
		return
	}

	if isDeleted {
		utils.JSONError(w, http.StatusBadRequest, "Comment already deleted")
		return
	}

	if currentUser.Role != "admin" {
		if !commentUserID.Valid || int(commentUserID.Int64) != currentUser.ID {
			utils.JSONError(w, http.StatusForbidden, "You can only delete your own comments")
			return
		}
	}

	_, err = db.DB.Exec(`UPDATE comments SET is_deleted = 1 WHERE id = ?`, id)
	if err != nil {
		utils.JSONError(w, http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	utils.JSONSuccess(w, http.StatusOK, map[string]string{"message": "Comment deleted successfully"})
}

func getCommentByID(id int, currentUser *utils.UserInfo) (*Comment, error) {
	var comment Comment
	var parentID, userID sql.NullInt64

	err := db.DB.QueryRow(
		`SELECT id, article_id, parent_id, reply_level, user_id, guest_name, guest_email,
		 content, status, is_deleted, created_at
		 FROM comments WHERE id = ?`,
		id,
	).Scan(
		&comment.ID, &comment.ArticleID, &parentID, &comment.ReplyLevel, &userID,
		&comment.GuestName, &comment.GuestEmail, &comment.Content, &comment.Status,
		&comment.IsDeleted, &comment.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		pid := int(parentID.Int64)
		comment.ParentID = &pid
	}
	if userID.Valid {
		uid := int(userID.Int64)
		comment.UserID = &uid
	}

	comment.Author = getCommentAuthor(&comment)

	if comment.IsDeleted {
		comment.Content = "[原评论已删除]"
	}

	return &comment, nil
}

func getReplies(parentID int, currentUser *utils.UserInfo) []*Comment {
	whereClauses := []string{"parent_id = ?"}
	args := []interface{}{parentID}

	if currentUser == nil || currentUser.Role != "admin" {
		whereClauses = append(whereClauses, "status = 'approved'")
	}

	rows, err := db.DB.Query(
		`SELECT id, article_id, parent_id, reply_level, user_id, guest_name, guest_email,
		 content, status, is_deleted, created_at
		 FROM comments 
		 WHERE `+strings.Join(whereClauses, " AND ")+`
		 ORDER BY created_at ASC`,
		args...,
	)
	if err != nil {
		return []*Comment{}
	}
	defer rows.Close()

	replies := []*Comment{}
	for rows.Next() {
		var reply Comment
		var pID, uID sql.NullInt64

		err := rows.Scan(
			&reply.ID, &reply.ArticleID, &pID, &reply.ReplyLevel, &uID,
			&reply.GuestName, &reply.GuestEmail, &reply.Content, &reply.Status,
			&reply.IsDeleted, &reply.CreatedAt,
		)
		if err != nil {
			continue
		}

		if pID.Valid {
			pid := int(pID.Int64)
			reply.ParentID = &pid
		}
		if uID.Valid {
			uid := int(uID.Int64)
			reply.UserID = &uid
		}

		reply.Author = getCommentAuthor(&reply)
		reply.Replies = getReplies(reply.ID, currentUser)

		if reply.IsDeleted {
			reply.Content = "[原评论已删除]"
		}

		replies = append(replies, &reply)
	}

	return replies
}

func getCommentAuthor(comment *Comment) *AuthorInfo {
	if comment.UserID != nil {
		var info AuthorInfo
		err := db.DB.QueryRow(
			`SELECT id, username, display_name FROM users WHERE id = ?`,
			*comment.UserID,
		).Scan(&info.ID, &info.Username, &info.DisplayName)
		if err == nil {
			return &info
		}
	}
	return nil
}

func containsSensitiveWords(content string) bool {
	var count int
	err := db.DB.QueryRow(
		`SELECT COUNT(*) FROM sensitive_words WHERE INSTR(LOWER(?), LOWER(word)) > 0`,
		content,
	).Scan(&count)
	return err == nil && count > 0
}
