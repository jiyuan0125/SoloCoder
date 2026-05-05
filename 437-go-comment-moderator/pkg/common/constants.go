package common

import "time"

const (
	MaxCommentLength         = 500
	MaxBatchOperations       = 50
	ConsecutiveRejectsLimit  = 5
	ManualReviewDays         = 7
	MaxComments24h           = 50
	EditWindowMinutes        = 30
	MaxEditTimes             = 1
	ReportThreshold          = 10
)

const (
	DefaultPort              = 8080
	DefaultHost              = "localhost"
)

var RejectReasonTemplates = map[SensitiveWordLevel]string{
	LevelSevere: "您的评论包含严重违规内容，已被拒绝。",
	LevelMedium: "您的评论包含中等违规内容，已被拒绝。",
	LevelMild:   "您的评论包含轻微违规内容，已被拒绝。",
}

func GenerateID() string {
	return time.Now().Format("20060102150405.000000000")
}

func GetTodayString() string {
	return time.Now().Format("2006-01-02")
}
