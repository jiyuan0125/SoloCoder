package errors

import "errors"

var (
	ErrDonorNoDuplicate     = errors.New("捐献者编号重复")
	ErrOrganScoreTooLow     = errors.New("器官功能评分低于60分，不能用于移植")
	ErrInvalidOrganStatus   = errors.New("器官状态流转不合法")
	ErrBloodTypeMismatch    = errors.New("血型不匹配")
	ErrOrganAlreadyMatched  = errors.New("器官已匹配给其他受体")
	ErrTransplantNotFound   = errors.New("移植手术记录不存在")
	ErrInvalidPostOpStatus  = errors.New("术后状态流转不合法")
	ErrOrganColdIschemiaTimeout = errors.New("器官已超过冷缺血时间上限")
	ErrInvalidPRA           = errors.New("PRA值必须在0到100之间")
	ErrNotFound             = errors.New("资源不存在")
	ErrInvalidRequest       = errors.New("请求参数无效")
)

type AppError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func FromError(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	switch err {
	case ErrDonorNoDuplicate, ErrOrganAlreadyMatched:
		return NewAppError(409, err.Error(), err)
	case ErrOrganScoreTooLow, ErrInvalidOrganStatus, ErrBloodTypeMismatch,
		ErrInvalidPostOpStatus, ErrOrganColdIschemiaTimeout, ErrInvalidPRA, ErrInvalidRequest:
		return NewAppError(400, err.Error(), err)
	case ErrTransplantNotFound, ErrNotFound:
		return NewAppError(404, err.Error(), err)
	default:
		return NewAppError(500, "服务器内部错误", err)
	}
}
