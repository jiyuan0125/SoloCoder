package common

type MessageType int

const (
	TextMessage MessageType = iota
	BinaryMessage
)

type WSRequest struct {
	Type    MessageType
	Content []byte
}

type WSResponse struct {
	Type    MessageType
	Content []byte
	Error   string
}
