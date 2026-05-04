package common

const (
	ErrCodeSuccess               = 0
	ErrCodeInvalidName           = 1001
	ErrCodeDescriptionTooLong    = 1002
	ErrCodeInvalidStartPrice     = 1003
	ErrCodeInvalidBidIncrement   = 1004
	ErrCodeInvalidDeadline       = 1005
	ErrCodeAuctionNotFound       = 2001
	ErrCodeAuctionAlreadyEnded   = 2002
	ErrCodeBidTooLow             = 2003
	ErrCodeBidAlreadyExceeded    = 2004
	ErrCodeInternalError         = 3001
	ErrCodeInvalidRequest        = 4001
)

var errorMessages = map[int]string{
	ErrCodeSuccess:               "操作成功",
	ErrCodeInvalidName:           "拍品名称不能为空",
	ErrCodeDescriptionTooLong:    "描述不能超过500字",
	ErrCodeInvalidStartPrice:     "起拍价必须大于零",
	ErrCodeInvalidBidIncrement:   "加价幅度必须大于零",
	ErrCodeInvalidDeadline:       "截止时间必须在未来",
	ErrCodeAuctionNotFound:       "拍品不存在",
	ErrCodeAuctionAlreadyEnded:   "该拍品竞拍已结束",
	ErrCodeBidTooLow:             "出价低于当前最高价加价幅度",
	ErrCodeBidAlreadyExceeded:    "出价已被超越",
	ErrCodeInternalError:         "内部错误",
	ErrCodeInvalidRequest:        "请求参数错误",
}

func GetErrorMessage(code int) string {
	if msg, ok := errorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
