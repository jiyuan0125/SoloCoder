package protocol

import "time"

type Language string

const (
	LanguageChinese Language = "zh"
	LanguageEnglish Language = "en"
)

type FAQContent struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

type FAQ struct {
	ID           string                  `json:"id"`
	CategoryID   string                  `json:"category_id"`
	Content      map[Language]FAQContent `json:"content"`
	SortWeight   int                     `json:"sort_weight"`
	IsEnabled    bool                    `json:"is_enabled"`
	IsPinned     bool                    `json:"is_pinned"`
	ViewCount    int                     `json:"view_count"`
	IsHot        bool                    `json:"is_hot"`
	NeedOptimize bool                    `json:"need_optimize"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

type Category struct {
	ID       string     `json:"id"`
	ParentID string     `json:"parent_id"`
	Name     string     `json:"name"`
	Children []Category `json:"children,omitempty"`
}

type HistoryRecord struct {
	ID          string    `json:"id"`
	FAQID       string    `json:"faq_id"`
	Action      string    `json:"action"`
	OldContent  string    `json:"old_content,omitempty"`
	NewContent  string    `json:"new_content,omitempty"`
	ChangedBy   string    `json:"changed_by"`
	ChangedAt   time.Time `json:"changed_at"`
}

type ClickRecord struct {
	ID        string    `json:"id"`
	FAQID     string    `json:"faq_id"`
	UserID    string    `json:"user_id"`
	IsHelpful bool      `json:"is_helpful"`
	CreatedAt time.Time `json:"created_at"`
}
