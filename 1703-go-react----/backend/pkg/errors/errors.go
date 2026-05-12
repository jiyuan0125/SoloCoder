package errors

import "net/http"

type ErrorCode struct {
	Code    int
	Message string
	Status  int
}

var (
	ErrInvalidIDCardLength = ErrorCode{Code: 10001, Message: "身份证号长度不对，应为18位", Status: http.StatusBadRequest}
	ErrInvalidIDCardFormat = ErrorCode{Code: 10002, Message: "身份证号格式不正确", Status: http.StatusBadRequest}
	ErrInvalidPhone        = ErrorCode{Code: 10003, Message: "手机号格式不正确", Status: http.StatusBadRequest}
	ErrTimeSlotFull        = ErrorCode{Code: 20001, Message: "该时段已满", Status: http.StatusConflict}
	ErrPackageNotFound     = ErrorCode{Code: 30001, Message: "套餐不存在", Status: http.StatusNotFound}
	ErrItemNotFound        = ErrorCode{Code: 30002, Message: "检查项目不存在", Status: http.StatusNotFound}
	ErrAppointmentNotFound = ErrorCode{Code: 40001, Message: "预约不存在", Status: http.StatusNotFound}
	ErrReportNotFound      = ErrorCode{Code: 50001, Message: "报告不存在", Status: http.StatusNotFound}
	ErrModifyLimitExceeded = ErrorCode{Code: 60001, Message: "修改次数超过3次限制", Status: http.StatusBadRequest}
	ErrResultNotFound      = ErrorCode{Code: 60002, Message: "检查结果不存在", Status: http.StatusNotFound}
	ErrInvalidDateRange    = ErrorCode{Code: 70001, Message: "日期范围无效", Status: http.StatusBadRequest}
	ErrReportNotCompleted  = ErrorCode{Code: 50002, Message: "报告未完成所有项目", Status: http.StatusBadRequest}
	ErrInvalidReportStatus = ErrorCode{Code: 50003, Message: "报告状态不允许此操作", Status: http.StatusBadRequest}
	ErrAppointmentCancelled = ErrorCode{Code: 40002, Message: "预约已取消", Status: http.StatusBadRequest}
	ErrInvalidExamDate     = ErrorCode{Code: 40003, Message: "体检日期无效", Status: http.StatusBadRequest}
)

func (e ErrorCode) Error() string {
	return e.Message
}

func (e ErrorCode) GetStatus() int {
	return e.Status
}

func (e ErrorCode) GetCode() int {
	return e.Code
}
