package server

import "errors"

var (
	ErrMemberExists           = errors.New("会员已存在")
	ErrMemberNotFound         = errors.New("会员不存在")
	ErrInvalidPassword        = errors.New("密码错误")
	ErrActiveCardExists       = errors.New("已有生效中的会员卡")
	ErrCardNotFound           = errors.New("会员卡不存在")
	ErrInvalidCardType        = errors.New("无效的卡类型")
	ErrClassNotFound          = errors.New("课程不存在")
	ErrClassInstanceNotFound  = errors.New("课程实例不存在")
	ErrClassNotAvailable      = errors.New("课程不可用")
	ErrClassFull              = errors.New("课程已满")
	ErrAlreadyBooked          = errors.New("已预约该课程")
	ErrBookingNotFound        = errors.New("预约不存在")
	ErrInvalidBookingStatus   = errors.New("预约状态无效")
	ErrCancelTooLate          = errors.New("取消时间已过，上课前2小时内可免费取消")
	ErrNoBookingForClass      = errors.New("未预约该课程")
	ErrAlreadyCheckedIn       = errors.New("已签到")
	ErrCheckInTimeInvalid     = errors.New("签到时间无效，仅在上课前后15分钟内可签到")
	ErrInvalidPhone           = errors.New("手机号格式错误")
	ErrInvalidClassName       = errors.New("课程名称不能为空且不能超过30字")
	ErrInvalidClassWeekday    = errors.New("星期必须在0-6之间")
	ErrInvalidStartTime       = errors.New("开始时间格式错误，应为HH:MM格式")
	ErrNoActiveCard           = errors.New("没有有效的会员卡，请先办卡")
)
