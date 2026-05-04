package ordergen

import (
	"strconv"
	"strings"
	"time"
)

func ParseOrderNo(orderNo string) (*OrderInfo, error) {
	if len(orderNo) < TimePartLength+SerialPartLength+RandomPartLength {
		return nil, ErrInvalidOrderNo
	}

	info := &OrderInfo{
		OrderNo: orderNo,
	}

	totalLength := TimePartLength + SerialPartLength + RandomPartLength
	prefixLength := len(orderNo) - totalLength

	if prefixLength > 0 {
		info.Prefix = orderNo[:prefixLength]
	}

	timePart := orderNo[prefixLength : prefixLength+TimePartLength]
	serialPart := orderNo[prefixLength+TimePartLength : prefixLength+TimePartLength+SerialPartLength]
	randomPart := orderNo[prefixLength+TimePartLength+SerialPartLength:]

	timestamp, err := time.ParseInLocation(TimeFormat, timePart, time.Local)
	if err != nil {
		return nil, ErrParseTimeFailed
	}
	info.Timestamp = timestamp

	serialNum, err := strconv.Atoi(serialPart)
	if err != nil {
		return nil, ErrInvalidOrderNo
	}
	info.SerialNum = serialNum

	randomNum, err := strconv.Atoi(randomPart)
	if err != nil {
		return nil, ErrInvalidOrderNo
	}
	info.RandomNum = randomNum

	if !isValidOrderNoFormat(orderNo, prefixLength) {
		return nil, ErrInvalidOrderNo
	}

	return info, nil
}

func isValidOrderNoFormat(orderNo string, prefixLength int) bool {
	for i := prefixLength; i < len(orderNo); i++ {
		c := orderNo[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func HasPrefix(orderNo string, prefix string) bool {
	return strings.HasPrefix(orderNo, prefix)
}

func ExtractTime(orderNo string) (time.Time, error) {
	info, err := ParseOrderNo(orderNo)
	if err != nil {
		return time.Time{}, err
	}
	return info.Timestamp, nil
}

func ExtractPrefix(orderNo string) string {
	info, err := ParseOrderNo(orderNo)
	if err != nil {
		return ""
	}
	return info.Prefix
}
