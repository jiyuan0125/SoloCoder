package ordergen

import (
	"errors"
	"time"
)

const (
	TimeFormat      = "20060102150405"
	SerialNumMax    = 999999
	RandomNumMax    = 9999
	TimePartLength  = 14
	SerialPartLength = 6
	RandomPartLength = 4
)

var (
	ErrSerialNumExhausted = errors.New("serial number exhausted, please wait for next millisecond")
	ErrInvalidOrderNo     = errors.New("invalid order number format")
	ErrParseTimeFailed    = errors.New("failed to parse time from order number")
)

type OrderGeneratorConfig struct {
	Prefix      string
	StoragePath string
}

type PersistedState struct {
	LastSecond    string `json:"last_second"`
	LastSerialNum int    `json:"last_serial_num"`
	Prefix        string `json:"prefix"`
}

type OrderInfo struct {
	Prefix    string
	Timestamp time.Time
	SerialNum int
	RandomNum int
	OrderNo   string
}
