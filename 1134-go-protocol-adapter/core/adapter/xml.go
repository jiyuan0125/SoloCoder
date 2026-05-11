package adapter

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
	"time"
)

type XMLAdapter struct {
	converter *TypeConverter
}

func NewXMLAdapter() *XMLAdapter {
	return &XMLAdapter{
		converter: NewTypeConverter(),
	}
}

func (a *XMLAdapter) Name() string {
	return "xml"
}

func (a *XMLAdapter) CanHandle(format string) bool {
	return strings.EqualFold(format, "xml")
}

func (a *XMLAdapter) Decode(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (a *XMLAdapter) Encode(v interface{}) ([]byte, error) {
	converted := a.convertToXMLSerializable(v)
	return xml.MarshalIndent(converted, "", "  ")
}

func (a *XMLAdapter) convertToXMLSerializable(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return nil
		}
		return a.convertToXMLSerializable(val.Elem().Interface())
	case reflect.Interface:
		if val.IsNil() {
			return nil
		}
		return a.convertToXMLSerializable(val.Elem().Interface())
	case reflect.Map:
		return a.mapToXMLElement(val)
	case reflect.Slice:
		return a.sliceToXMLElements(val)
	default:
		return v
	}
}

type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type xmlMapWrapper struct {
	XMLName xml.Name
	Entries []interface{} `xml:",any"`
}

func (a *XMLAdapter) mapToXMLElement(val reflect.Value) *xmlMapWrapper {
	wrapper := &xmlMapWrapper{
		XMLName: xml.Name{Local: "item"},
		Entries: make([]interface{}, 0, val.Len()),
	}

	for _, key := range val.MapKeys() {
		keyStr := fmt.Sprintf("%v", key.Interface())
		mapValue := val.MapIndex(key)

		if mapValue.Kind() == reflect.Interface && !mapValue.IsNil() {
			mapValue = mapValue.Elem()
		}

		switch mapValue.Kind() {
		case reflect.Map:
			nested := a.mapToXMLElement(mapValue)
			nested.XMLName = xml.Name{Local: keyStr}
			wrapper.Entries = append(wrapper.Entries, nested)
		case reflect.Slice:
			sliceEntries := a.sliceToXMLElementsWithName(mapValue, keyStr)
			wrapper.Entries = append(wrapper.Entries, sliceEntries...)
		default:
			wrapper.Entries = append(wrapper.Entries, &xmlMapEntry{
				XMLName: xml.Name{Local: keyStr},
				Value:   fmt.Sprintf("%v", mapValue.Interface()),
			})
		}
	}

	return wrapper
}

func (a *XMLAdapter) sliceToXMLElements(val reflect.Value) []interface{} {
	return a.sliceToXMLElementsWithName(val, "item")
}

func (a *XMLAdapter) sliceToXMLElementsWithName(val reflect.Value, itemName string) []interface{} {
	entries := make([]interface{}, 0, val.Len())

	for i := 0; i < val.Len(); i++ {
		elem := val.Index(i)
		if elem.Kind() == reflect.Interface && !elem.IsNil() {
			elem = elem.Elem()
		}

		switch elem.Kind() {
		case reflect.Map:
			nested := a.mapToXMLElement(elem)
			nested.XMLName = xml.Name{Local: itemName}
			entries = append(entries, nested)
		default:
			entries = append(entries, &xmlMapEntry{
				XMLName: xml.Name{Local: itemName},
				Value:   fmt.Sprintf("%v", elem.Interface()),
			})
		}
	}

	return entries
}

func (a *XMLAdapter) Priority() int {
	return 100
}

type XMLElement struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Child   []XMLElement `xml:",any"`
}

func (a *XMLAdapter) decodeStruct(data []byte, v interface{}) error {
	var elem XMLElement
	if err := xml.Unmarshal(data, &elem); err != nil {
		return err
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		return ErrInvalidType
	}
	val = val.Elem()

	return a.fillStructFromXML(&elem, val)
}

func (a *XMLAdapter) fillStructFromXML(elem *XMLElement, val reflect.Value) error {
	if val.Kind() != reflect.Struct {
		return nil
	}

	valType := val.Type()

	fieldMap := make(map[string]reflect.Value)
	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		fieldVal := val.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		tag := field.Tag.Get("xml")
		if tag == "" || tag == "-" {
			fieldMap[strings.ToLower(field.Name)] = fieldVal
		} else {
			parts := strings.Split(tag, ",")
			fieldMap[parts[0]] = fieldVal
		}
	}

	for _, attr := range elem.Attrs {
		if fieldVal, ok := fieldMap[strings.ToLower(attr.Name.Local)]; ok {
			if err := a.setFieldFromString(fieldVal, attr.Value); err != nil {
				return err
			}
		}
	}

	if elem.Content != "" {
		for _, fieldVal := range fieldMap {
			if fieldVal.Kind() == reflect.String {
				fieldVal.SetString(elem.Content)
				break
			}
		}
	}

	for _, child := range elem.Child {
		childName := strings.ToLower(child.XMLName.Local)
		if fieldVal, ok := fieldMap[childName]; ok {
			if child.Content != "" {
				if err := a.setFieldFromString(fieldVal, child.Content); err != nil {
					return err
				}
			} else {
				if fieldVal.Kind() == reflect.Ptr {
					if fieldVal.IsNil() {
						fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
					}
					fieldVal = fieldVal.Elem()
				}
				if err := a.fillStructFromXML(&child, fieldVal); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (a *XMLAdapter) setFieldFromString(fieldVal reflect.Value, value string) error {
	if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		}
		if t, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
			fieldVal.Set(reflect.ValueOf(t))
			return nil
		}
	}

	converted, err := a.converter.Convert(value, fieldVal.Type())
	if err != nil {
		return err
	}
	fieldVal.Set(converted)
	return nil
}

var ErrInvalidType = newError("invalid type: expected pointer")

type adapterError struct {
	msg string
}

func newError(msg string) *adapterError {
	return &adapterError{msg: msg}
}

func (e *adapterError) Error() string {
	return e.msg
}
