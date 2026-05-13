package converter

import (
	"encoding/json"
	"encoding/xml"
	"regexp"
	"strings"
)

var cdataRegex = regexp.MustCompile(`<!\[CDATA\[(.*?)\]\]>`)
var cdataPlaceholder = "____CDATA____"

type Node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Text    string     `xml:",chardata"`
	Nodes   []Node     `xml:",any"`
}

type CDataMarker struct {
	Content string
}

func XmlToJson(xmlContent []byte) (map[string]interface{}, error) {
	xmlStr := string(xmlContent)
	
	cdataContents := []string{}
	xmlStr = cdataRegex.ReplaceAllStringFunc(xmlStr, func(match string) string {
		content := match[9 : len(match)-3]
		cdataContents = append(cdataContents, content)
		return cdataPlaceholder
	})
	
	var root Node
	if err := xml.Unmarshal([]byte(xmlStr), &root); err != nil {
		return nil, err
	}
	
	cdataIndex := 0
	result := make(map[string]interface{})
	result[root.XMLName.Local] = nodeToMap(&root, &cdataIndex, cdataContents)
	return result, nil
}

func nodeToMap(node *Node, cdataIndex *int, cdataContents []string) map[string]interface{} {
	result := make(map[string]interface{})
	
	for _, attr := range node.Attrs {
		result["@"+attr.Name.Local] = attr.Value
	}
	
	text := strings.TrimSpace(node.Text)
	if text == cdataPlaceholder && *cdataIndex < len(cdataContents) {
		result["#cdata"] = cdataContents[*cdataIndex]
		*cdataIndex++
	} else if text != "" {
		result["#text"] = text
	}
	
	childMap := make(map[string]interface{})
	for i := range node.Nodes {
		child := &node.Nodes[i]
		childName := child.XMLName.Local
		childValue := nodeToMap(child, cdataIndex, cdataContents)
		
		if existing, ok := childMap[childName]; ok {
			switch v := existing.(type) {
			case []interface{}:
				childMap[childName] = append(v, childValue)
			default:
				childMap[childName] = []interface{}{existing, childValue}
			}
		} else {
			childMap[childName] = childValue
		}
	}
	
	for k, v := range childMap {
		result[k] = v
	}
	
	return result
}

func MapToJson(data map[string]interface{}) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}
