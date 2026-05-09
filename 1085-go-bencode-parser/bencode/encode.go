package bencode

import (
	"fmt"
	"sort"
	"strconv"
)

func Encode(v interface{}) ([]byte, error) {
	e := &encoder{}
	err := e.encodeValue(v)
	return e.buf, err
}

type encoder struct {
	buf []byte
}

func (e *encoder) encodeValue(v interface{}) error {
	switch val := v.(type) {
	case int:
		return e.encodeInteger(int64(val))
	case int8:
		return e.encodeInteger(int64(val))
	case int16:
		return e.encodeInteger(int64(val))
	case int32:
		return e.encodeInteger(int64(val))
	case int64:
		return e.encodeInteger(val)
	case uint:
		return e.encodeInteger(int64(val))
	case uint8:
		return e.encodeInteger(int64(val))
	case uint16:
		return e.encodeInteger(int64(val))
	case uint32:
		return e.encodeInteger(int64(val))
	case uint64:
		return e.encodeInteger(int64(val))
	case string:
		return e.encodeByteString([]byte(val))
	case []byte:
		return e.encodeByteString(val)
	case []interface{}:
		return e.encodeList(val)
	case map[string]interface{}:
		return e.encodeDict(val)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (e *encoder) encodeInteger(v int64) error {
	e.buf = append(e.buf, 'i')
	e.buf = strconv.AppendInt(e.buf, v, 10)
	e.buf = append(e.buf, 'e')
	return nil
}

func (e *encoder) encodeByteString(v []byte) error {
	e.buf = strconv.AppendInt(e.buf, int64(len(v)), 10)
	e.buf = append(e.buf, ':')
	e.buf = append(e.buf, v...)
	return nil
}

func (e *encoder) encodeList(v []interface{}) error {
	e.buf = append(e.buf, 'l')
	for _, elem := range v {
		if err := e.encodeValue(elem); err != nil {
			return err
		}
	}
	e.buf = append(e.buf, 'e')
	return nil
}

func (e *encoder) encodeDict(v map[string]interface{}) error {
	e.buf = append(e.buf, 'd')

	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if err := e.encodeByteString([]byte(k)); err != nil {
			return err
		}
		if err := e.encodeValue(v[k]); err != nil {
			return err
		}
	}

	e.buf = append(e.buf, 'e')
	return nil
}
