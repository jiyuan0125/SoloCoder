package encoder

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func Convert(data []byte, fromEncoding string, toEncoding string) ([]byte, error) {
	fromEnc, err := getEncoding(fromEncoding)
	if err != nil {
		return nil, err
	}

	toEnc, err := getEncoding(toEncoding)
	if err != nil {
		return nil, err
	}

	decoded, err := decode(data, fromEnc)
	if err != nil {
		return nil, err
	}

	encoded, err := encode(decoded, toEnc)
	if err != nil {
		return nil, err
	}

	return encoded, nil
}

func getEncoding(name string) (encoding.Encoding, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	switch name {
	case "utf-8", "utf8":
		return unicode.UTF8, nil
	case "utf-16", "utf16":
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM), nil
	case "utf-16le", "utf16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), nil
	case "utf-16be", "utf16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), nil
	case "gbk", "gb2312", "gb18030":
		return simplifiedchinese.GBK, nil
	case "big5":
		return traditionalchinese.Big5, nil
	case "shift_jis", "shiftjis":
		return japanese.ShiftJIS, nil
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", name)
	}
}

func decode(data []byte, enc encoding.Encoding) ([]byte, error) {
	if enc == unicode.UTF8 {
		return data, nil
	}

	reader := transform.NewReader(bytes.NewReader(data), enc.NewDecoder())
	return ioutil.ReadAll(reader)
}

func encode(data []byte, enc encoding.Encoding) ([]byte, error) {
	if enc == unicode.UTF8 {
		return data, nil
	}

	reader := transform.NewReader(bytes.NewReader(data), enc.NewEncoder())
	return ioutil.ReadAll(reader)
}
