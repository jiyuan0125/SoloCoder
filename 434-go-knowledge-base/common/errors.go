package common

const (
	ErrCodeSuccess        = 0
	ErrCodeBadRequest     = 400
	ErrCodeUnauthorized   = 401
	ErrCodeForbidden      = 403
	ErrCodeNotFound       = 404
	ErrCodeConflict       = 409
	ErrCodeInternal       = 500

	ErrCodeArticleNotFound    = 1001
	ErrCodeVersionNotFound    = 1002
	ErrCodeAlreadyArchived    = 1003
	ErrCodeNotAuthor          = 1004
	ErrCodeDraftOnly          = 1005
	ErrCodeAccessDenied       = 1006
	ErrCodeInvalidStatus      = 1007
)

var errMessages = map[int]string{
	ErrCodeSuccess:         "成功",
	ErrCodeBadRequest:      "请求参数错误",
	ErrCodeUnauthorized:    "未授权",
	ErrCodeForbidden:       "禁止访问",
	ErrCodeNotFound:        "资源不存在",
	ErrCodeConflict:        "资源冲突",
	ErrCodeInternal:        "服务器内部错误",
	ErrCodeArticleNotFound: "文章不存在",
	ErrCodeVersionNotFound: "版本不存在",
	ErrCodeAlreadyArchived: "文章已归档，无法操作",
	ErrCodeNotAuthor:       "无权限操作此文章",
	ErrCodeDraftOnly:       "只有草稿状态的文章可编辑",
	ErrCodeAccessDenied:    "无访问权限",
	ErrCodeInvalidStatus:   "无效的文章状态",
}

func GetErrMessage(code int) string {
	if msg, ok := errMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
