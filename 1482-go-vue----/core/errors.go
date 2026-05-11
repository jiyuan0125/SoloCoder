package core

import "errors"

var (
	ErrOrderNotFound           = errors.New("工单不存在")
	ErrOrderInvalidStatus      = errors.New("工单状态不正确")
	ErrOrderAlreadyCompleted   = errors.New("工单已完成")
	ErrOrderAlreadyCancelled   = errors.New("工单已取消")
	ErrItemNotFound            = errors.New("维修项目不存在")
	ErrPartNotFound            = errors.New("配件不存在")
	ErrPartCodeDuplicate       = errors.New("配件编码已存在")
	ErrInsufficientStock       = errors.New("库存不足")
	ErrUsedPartNotFound        = errors.New("出库配件记录不存在")
	ErrUsedPartAlreadyReturned = errors.New("配件已全部退回")
	ErrInvalidQuantity         = errors.New("数量无效")
	ErrInvalidPrice            = errors.New("价格无效")
	ErrInvalidHours            = errors.New("工时无效")
	ErrOvertimeReasonRequired  = errors.New("超时需要填写原因")
	ErrTodoNotFound            = errors.New("补货待办不存在")
	ErrTodoAlreadyResolved     = errors.New("补货待办已处理")
)
