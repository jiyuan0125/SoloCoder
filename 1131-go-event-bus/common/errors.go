package common

const (
	CodeSuccess = 0
	CodeError   = 1000

	CodeTopicExist    = 1001
	CodeTopicNotExist = 1002
	CodeSubNotExist   = 1003
	CodeEventNotExist = 1004
	CodeInvalidParam  = 1005
)

var codeMessages = map[int]string{
	CodeSuccess:      "success",
	CodeError:        "internal error",
	CodeTopicExist:   "topic already exists",
	CodeTopicNotExist: "topic does not exist",
	CodeSubNotExist:  "subscriber does not exist",
	CodeEventNotExist: "event does not exist",
	CodeInvalidParam: "invalid parameter",
}

func CodeMessage(code int) string {
	if msg, ok := codeMessages[code]; ok {
		return msg
	}
	return "unknown error"
}
