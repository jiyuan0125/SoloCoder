package protocol

type EmojiRequest struct {
	Text          string `json:"text"`
	Operation     string `json:"operation"`
	Replacement   string `json:"replacement,omitempty"`
}

type EmojiResponse struct {
	Success       bool        `json:"success"`
	Error         string      `json:"error,omitempty"`
	Result        string      `json:"result,omitempty"`
	EmojiCount    int         `json:"emoji_count,omitempty"`
	Emojis        []EmojiData `json:"emojis,omitempty"`
	IsOnlyEmojis  bool        `json:"is_only_emojis,omitempty"`
}

type EmojiData struct {
	StartByte int    `json:"start_byte"`
	EndByte   int    `json:"end_byte"`
	Emoji     string `json:"emoji"`
}

const (
	OpExtract      = "extract"
	OpCount        = "count"
	OpRemove       = "remove"
	OpReplace      = "replace"
	OpCheckOnly    = "check_only"
	OpContains     = "contains"
)
