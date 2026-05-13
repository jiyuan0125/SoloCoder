package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"usercomments/internal/model"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			article_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			nickname TEXT NOT NULL,
			avatar TEXT,
			content TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'plain',
			parent_id INTEGER,
			reply_level INTEGER NOT NULL DEFAULT 1,
			likes INTEGER NOT NULL DEFAULT 0,
			is_deleted INTEGER NOT NULL DEFAULT 0,
			root_parent_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (article_id) REFERENCES articles(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_article_id ON comments(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_parent_id ON comments(parent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_created_at ON comments(created_at)`,
	}

	for _, stmt := range statements {
		_, err := DB.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return seedData()
}

func seedData() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	articles := []struct {
		title string
	}{
		{"Go 语言入门指南"},
		{"SQLite 最佳实践"},
		{"Web 开发进阶技巧"},
	}

	for _, art := range articles {
		_, err := DB.Exec("INSERT INTO articles (title) VALUES (?)", art.title)
		if err != nil {
			return err
		}
	}

	return nil
}

func ArticleExists(articleID int64) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE id = ?)", articleID).Scan(&exists)
	return exists, err
}

func CreateComment(comment *model.Comment) error {
	now := time.Now()
	result, err := DB.Exec(`
		INSERT INTO comments (article_id, user_id, nickname, avatar, content, content_type, 
		                      parent_id, reply_level, likes, is_deleted, root_parent_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, comment.ArticleID, comment.UserID, comment.Nickname, comment.Avatar, comment.Content,
		comment.ContentType, comment.ParentID, comment.ReplyLevel, comment.Likes,
		comment.IsDeleted, comment.RootParentID, now, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	comment.ID = id
	comment.CreatedAt = now
	comment.UpdatedAt = now
	return nil
}

func GetCommentByID(id int64) (*model.Comment, error) {
	var comment model.Comment
	var parentID sql.NullInt64
	var rootParentID sql.NullInt64

	err := DB.QueryRow(`
		SELECT id, article_id, user_id, nickname, avatar, content, content_type,
		       parent_id, reply_level, likes, is_deleted, root_parent_id, created_at, updated_at
		FROM comments WHERE id = ?
	`, id).Scan(
		&comment.ID, &comment.ArticleID, &comment.UserID, &comment.Nickname, &comment.Avatar,
		&comment.Content, &comment.ContentType, &parentID, &comment.ReplyLevel, &comment.Likes,
		&comment.IsDeleted, &rootParentID, &comment.CreatedAt, &comment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("comment not found")
	}
	if err != nil {
		return nil, err
	}

	if parentID.Valid {
		comment.ParentID = &parentID.Int64
	}
	if rootParentID.Valid {
		comment.RootParentID = &rootParentID.Int64
	}

	return &comment, nil
}

func GetCommentsByArticle(articleID int64, page, pageSize int, sortBy string) ([]model.Comment, int, error) {
	offset := (page - 1) * pageSize

	orderClause := "created_at DESC"
	if sortBy == "likes" {
		orderClause = "likes DESC, created_at DESC"
	}

	var total int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM comments WHERE article_id = ? AND parent_id IS NULL
	`, articleID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(`
		SELECT id, article_id, user_id, nickname, avatar, content, content_type,
		       parent_id, reply_level, likes, is_deleted, root_parent_id, created_at, updated_at
		FROM comments 
		WHERE article_id = ? AND parent_id IS NULL
		ORDER BY `+orderClause+`
		LIMIT ? OFFSET ?
	`, articleID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func GetReplies(parentID int64) ([]model.Comment, error) {
	rows, err := DB.Query(`
		SELECT id, article_id, user_id, nickname, avatar, content, content_type,
		       parent_id, reply_level, likes, is_deleted, root_parent_id, created_at, updated_at
		FROM comments 
		WHERE parent_id = ? OR root_parent_id = ?
		ORDER BY created_at ASC
	`, parentID, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanComments(rows)
}

func SoftDeleteComment(commentID int64) error {
	_, err := DB.Exec(`
		UPDATE comments SET is_deleted = 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, commentID)
	return err
}

func SearchComments(keyword string, page, pageSize int) ([]model.Comment, int, error) {
	offset := (page - 1) * pageSize
	searchPattern := "%" + keyword + "%"

	var total int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM comments WHERE content LIKE ?
	`, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(`
		SELECT id, article_id, user_id, nickname, avatar, content, content_type,
		       parent_id, reply_level, likes, is_deleted, root_parent_id, created_at, updated_at
		FROM comments 
		WHERE content LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func SearchCommentsByNickname(nickname string, page, pageSize int) ([]model.Comment, int, error) {
	offset := (page - 1) * pageSize
	searchPattern := "%" + nickname + "%"

	var total int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM comments WHERE nickname LIKE ?
	`, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := DB.Query(`
		SELECT id, article_id, user_id, nickname, avatar, content, content_type,
		       parent_id, reply_level, likes, is_deleted, root_parent_id, created_at, updated_at
		FROM comments 
		WHERE nickname LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, searchPattern, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	comments, err := scanComments(rows)
	if err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func scanComments(rows *sql.Rows) ([]model.Comment, error) {
	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		var parentID sql.NullInt64
		var rootParentID sql.NullInt64

		err := rows.Scan(
			&c.ID, &c.ArticleID, &c.UserID, &c.Nickname, &c.Avatar, &c.Content,
			&c.ContentType, &parentID, &c.ReplyLevel, &c.Likes, &c.IsDeleted,
			&rootParentID, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			c.ParentID = &parentID.Int64
		}
		if rootParentID.Valid {
			c.RootParentID = &rootParentID.Int64
		}

		comments = append(comments, c)
	}
	return comments, nil
}

func CanUserComment(userID int64, articleID int64) (bool, error) {
	var lastCommentTime sql.NullTime
	err := DB.QueryRow(`
		SELECT created_at FROM comments 
		WHERE user_id = ? AND article_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, articleID).Scan(&lastCommentTime)

	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	if !lastCommentTime.Valid {
		return true, nil
	}

	elapsed := time.Since(lastCommentTime.Time)
	canComment := elapsed.Seconds() >= float64(model.RateLimitSeconds)
	return canComment, nil
}

func GetArticles() ([]model.Article, error) {
	rows, err := DB.Query("SELECT id, title FROM articles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var a model.Article
		if err := rows.Scan(&a.ID, &a.Title); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func LikeComment(commentID int64) error {
	_, err := DB.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", commentID)
	return err
}

func BuildCommentTree(comments []model.Comment) ([]model.Comment, error) {
	var tree []model.Comment

	for i := range comments {
		c := comments[i]
		
		if c.IsDeleted {
			c.Content = "该评论已删除"
		}

		if c.ParentID == nil {
			replies, err := GetReplies(c.ID)
			if err != nil {
				return nil, err
			}
			
			for j := range replies {
				if replies[j].IsDeleted {
					replies[j].Content = "该评论已删除"
				}
			}
			
			tree = append(tree, c)
			tree = append(tree, replies...)
		}
	}

	return tree, nil
}

func GetRootParentID(parentID int64) (int64, error) {
	var rootParentID int64
	var level int

	err := DB.QueryRow(`
		WITH RECURSIVE comment_hierarchy(id, parent_id, root_parent_id, reply_level) AS (
			SELECT id, parent_id, COALESCE(root_parent_id, id) as root_parent_id, reply_level
			FROM comments WHERE id = ?
			UNION ALL
			SELECT c.id, c.parent_id, ch.root_parent_id, c.reply_level
			FROM comments c
			INNER JOIN comment_hierarchy ch ON c.id = ch.parent_id
			WHERE c.parent_id IS NOT NULL
		)
		SELECT root_parent_id, MAX(reply_level) as level
		FROM comment_hierarchy
	`, parentID).Scan(&rootParentID, &level)

	if err != nil {
		return 0, fmt.Errorf("failed to get root parent: %w", err)
	}

	return rootParentID, nil
}

func GetParentLevel(parentID int64) (int, error) {
	var level int
	err := DB.QueryRow(`
		SELECT reply_level FROM comments WHERE id = ?
	`, parentID).Scan(&level)
	return level, err
}
