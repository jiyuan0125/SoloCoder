package rangehandler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ResourceInfo struct {
	Content     []byte
	ContentType string
	ETag        string
	LastModified time.Time
}

func HandleRangeRequest(req *http.Request, info ResourceInfo) *Response {
	size := int64(len(info.Content))

	ifRange := req.Header.Get("If-Range")
	if ifRange != "" {
		if !matchIfRange(ifRange, info) {
			return newFullResponse(info)
		}
	}

	rangeHeader := req.Header.Get("Range")
	if rangeHeader == "" {
		return newFullResponse(info)
	}

	ranges, err := ParseRangeHeader(rangeHeader, size)
	if err != nil {
		if err == ErrRangeUnsatisfiable {
			return &Response{
				StatusCode: http.StatusRequestedRangeNotSatisfiable,
				Headers: http.Header{
					"Content-Range": []string{fmt.Sprintf("bytes */%d", size)},
				},
				Body: nil,
			}
		}
		return newFullResponse(info)
	}

	if len(ranges) == 0 {
		return newFullResponse(info)
	}

	if len(ranges) == 1 {
		return handleSingleRange(ranges[0], info)
	}

	return handleMultipleRanges(ranges, info)
}

func matchIfRange(ifRange string, info ResourceInfo) bool {
	ifRange = strings.TrimSpace(ifRange)

	if strings.HasPrefix(ifRange, "\"") && strings.HasSuffix(ifRange, "\"") {
		ifRange = strings.Trim(ifRange, "\"")
		return ifRange == info.ETag
	}

	parsedTime, err := http.ParseTime(ifRange)
	if err != nil {
		return false
	}

	return !info.LastModified.After(parsedTime)
}

func newFullResponse(info ResourceInfo) *Response {
	headers := http.Header{}
	if info.ContentType != "" {
		headers.Set("Content-Type", info.ContentType)
	}
	if info.ETag != "" {
		headers.Set("ETag", "\""+info.ETag+"\"")
	}
	if !info.LastModified.IsZero() {
		headers.Set("Last-Modified", info.LastModified.Format(http.TimeFormat))
	}
	headers.Set("Accept-Ranges", "bytes")

	return &Response{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       info.Content,
	}
}

func handleSingleRange(r Range, info ResourceInfo) *Response {
	body := info.Content[r.Start : r.End+1]
	size := int64(len(info.Content))

	headers := http.Header{}
	if info.ContentType != "" {
		headers.Set("Content-Type", info.ContentType)
	}
	if info.ETag != "" {
		headers.Set("ETag", "\""+info.ETag+"\"")
	}
	if !info.LastModified.IsZero() {
		headers.Set("Last-Modified", info.LastModified.Format(http.TimeFormat))
	}
	headers.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.Start, r.End, size))
	headers.Set("Accept-Ranges", "bytes")

	return &Response{
		StatusCode: http.StatusPartialContent,
		Headers:    headers,
		Body:       body,
	}
}

func handleMultipleRanges(ranges []Range, info ResourceInfo) *Response {
	boundary := generateBoundary()
	size := int64(len(info.Content))
	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var body []byte
	header := fmt.Sprintf("\r\n--%s\r\n", boundary)

	for _, r := range ranges {
		body = append(body, []byte(header)...)
		body = append(body, []byte("Content-Type: "+contentType+"\r\n")...)
		body = append(body, []byte(fmt.Sprintf("Content-Range: bytes %d-%d/%d\r\n\r\n", r.Start, r.End, size))...)
		body = append(body, info.Content[r.Start:r.End+1]...)
	}

	body = append(body, []byte("\r\n--"+boundary+"--\r\n")...)

	headers := http.Header{}
	headers.Set("Content-Type", "multipart/byteranges; boundary="+boundary)
	if info.ETag != "" {
		headers.Set("ETag", "\""+info.ETag+"\"")
	}
	if !info.LastModified.IsZero() {
		headers.Set("Last-Modified", info.LastModified.Format(http.TimeFormat))
	}
	headers.Set("Accept-Ranges", "bytes")

	return &Response{
		StatusCode: http.StatusPartialContent,
		Headers:    headers,
		Body:       body,
	}
}

func generateBoundary() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(bytes)
}
