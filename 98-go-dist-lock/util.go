package distlock

import (
	"bytes"
	"runtime"
	"strconv"
	"strings"
)

func getGoroutineID() uint64 {
	buf := make([]byte, 64)
	runtime.Stack(buf, false)
	buf = buf[:bytes.IndexByte(buf, '[')]
	buf = buf[bytes.IndexByte(buf, ' ')+1:]
	idStr := strings.TrimSpace(string(buf))
	id, _ := strconv.ParseUint(idStr, 10, 64)
	return id
}
