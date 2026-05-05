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

	if isGBK(data) {
		return "GBK"
	}

	return "UTF-8"
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
