package converter

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
	"strings"
)

func JsonToMap(jsonContent []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(jsonContent, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func MapToXml(data map[string]interface{}) ([]byte, error) {
	var builder strings.Builder
	encoder := xml.NewEncoder(&builder)
	encoder.Indent("", "  ")
	
	for rootName, rootValue := range data {
		if err := writeElement(encoder, rootName, rootValue); err != nil {
			return nil, err
		}
	}
	
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	
	return []byte(builder.String()), nil
}

func writeElement(encoder *xml.Encoder, name string, value interface{}) error {
	start := xml.StartElement{Name: xml.Name{Local: name}}
	
	switch v := value.(type) {
	case map[string]interface{}:
		attrs, content, children, cdata := extractMapContent(v)
		
		for _, attr := range attrs {
			start.Attr = append(start.Attr, attr)
		}
		
		if err := encoder.EncodeToken(start); err != nil {
			return err
		}
		
		if cdata != "" {
			if err := encoder.EncodeToken(xml.Directive("[CDATA[" + cdata + "]]")); err != nil {
				return err
			}
		} else if content != "" {
			if err := encoder.EncodeToken(xml.CharData([]byte(content))); err != nil {
				return err
			}
		}
		
		for childName, childValue := range children {
			switch cv := childValue.(type) {
			case []interface{}:
				for _, item := range cv {
					if err := writeElement(encoder, childName, item); err != nil {
						return err
					}
				}
			default:
				if err := writeElement(encoder, childName, cv); err != nil {
					return err
				}
			}
		}
		
		return encoder.EncodeToken(start.End())
		
	case []interface{}:
		for _, item := range v {
			if err := writeElement(encoder, name, item); err != nil {
				return err
			}
		}
		return nil
		
	default:
		if err := encoder.EncodeToken(start); err != nil {
			return err
		}
		if err := encoder.EncodeToken(xml.CharData([]byte(interfaceToString(v)))); err != nil {
			return err
		}
		return encoder.EncodeToken(start.End())
	}
}

func extractMapContent(m map[string]interface{}) ([]xml.Attr, string, map[string]interface{}, string) {
	var attrs []xml.Attr
	var content string
	var cdata string
	children := make(map[string]interface{})
	
	for k, v := range m {
		if strings.HasPrefix(k, "@") {
			attrs = append(attrs, xml.Attr{
				Name:  xml.Name{Local: strings.TrimPrefix(k, "@")},
				Value: interfaceToString(v),
			})
		} else if k == "#text" {
			content = interfaceToString(v)
		} else if k == "#cdata" {
			cdata = interfaceToString(v)
		} else {
			children[k] = v
		}
	}
	
	return attrs, content, children, cdata
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}
