package csvproc

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func DetectEncoding(data []byte) string {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return "UTF-8-BOM"
	}

	if isValidUTF8(data) {
		return "UTF-8"
	}

	if isGBK(data) {
		return "GBK"
	}

	return "UTF-8"
}

func isValidUTF8(data []byte) bool {
	length := len(data)
	var i int = 0

	for i < length {
		b := data[i]

		if b <= 0x7F {
			i++
			continue
		}

		if b >= 0xC0 && b <= 0xDF {
			if i+1 >= length {
				return false
			}
			if !isUTF8ContinuationByte(data[i+1]) {
				return false
			}
			i += 2
			continue
		}

		if b >= 0xE0 && b <= 0xEF {
			if i+2 >= length {
				return false
			}
			if !isUTF8ContinuationByte(data[i+1]) || !isUTF8ContinuationByte(data[i+2]) {
				return false
			}
			if b == 0xE0 {
				if data[i+1] < 0xA0 {
					return false
				}
			}
			if b == 0xED {
				if data[i+1] > 0x9F {
					return false
				}
			}
			i += 3
			continue
		}

		if b >= 0xF0 && b <= 0xF7 {
			if i+3 >= length {
				return false
			}
			if !isUTF8ContinuationByte(data[i+1]) || !isUTF8ContinuationByte(data[i+2]) || !isUTF8ContinuationByte(data[i+3]) {
				return false
			}
			if b == 0xF0 {
				if data[i+1] < 0x90 {
					return false
				}
			}
			if b == 0xF4 {
				if data[i+1] > 0x8F {
					return false
				}
			}
			i += 4
			continue
		}

		return false
	}

	return true
}

func isUTF8ContinuationByte(b byte) bool {
	return b >= 0x80 && b <= 0xBF
}

func isGBK(data []byte) bool {
	length := len(data)
	var i int = 0
	for i < length {
		if data[i] <= 0x7f {
			i++
			continue
		} else {
			if (data[i] >= 0x81 && data[i] <= 0xFE) && (i+1 < length) {
				if data[i+1] >= 0x40 && data[i+1] <= 0xFE && data[i+1] != 0x7F {
					i += 2
					continue
				} else {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

func ConvertToUTF8(data []byte, encoding string) ([]byte, error) {
	if encoding == "UTF-8" {
		return data, nil
	}

	if encoding == "UTF-8-BOM" {
		return data[3:], nil
	}

	if encoding == "GBK" {
		reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
		utf8Data, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return utf8Data, nil
	}

	return data, nil
}
