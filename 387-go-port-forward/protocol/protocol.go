package protocol

const DefaultControlPort = 57329
const DefaultTimeout = 300
const DefaultBufferSize = 4096

type MessageType uint8

const (
	MsgTypeAddForward MessageType = iota
	MsgTypeRemoveForward
	MsgTypeListForwards
	MsgTypeGetStats
	MsgTypeShutdown
	MsgTypeResponse
)

type ForwardRule struct {
	LocalPort  int
	RemoteAddr string
}

type ConnectionStats struct {
	ID           uint64
	LocalPort    int
	RemoteAddr   string
	ClientAddr   string
	ConnectedAt  int64
	BytesSent    int64
	BytesReceived int64
	IsActive     bool
}

type ForwardStats struct {
	Rule         ForwardRule
	ActiveConns  int
	TotalConns   int64
	TotalBytes   int64
	Connections  []ConnectionStats
}

type Message struct {
	Type    MessageType
	Payload []byte
}

type AddForwardRequest struct {
	LocalPort  int
	RemoteAddr string
	Verbose    bool
	Timeout    int
	BufferSize int
}

type RemoveForwardRequest struct {
	LocalPort int
}

type Response struct {
	Success bool
	Message string
	Data    []byte
}
