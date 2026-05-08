package xpathengine

import (
	"encoding/xml"
	"io"
	"strings"
)

type Parser struct {
	nsStack []map[string]string
}

func NewParser() *Parser {
	return &Parser{
		nsStack: []map[string]string{{}},
	}
}

func (p *Parser) pushNS() {
	newNs := make(map[string]string)
	for k, v := range p.nsStack[len(p.nsStack)-1] {
		newNs[k] = v
	}
	p.nsStack = append(p.nsStack, newNs)
}

func (p *Parser) popNS() {
	if len(p.nsStack) > 1 {
		p.nsStack = p.nsStack[:len(p.nsStack)-1]
	}
}

func (p *Parser) currentNS() map[string]string {
	return p.nsStack[len(p.nsStack)-1]
}

func (p *Parser) parseQName(name xml.Name) QName {
	prefix := ""
	local := name.Local
	if len(name.Space) > 0 {
		for pfx, uri := range p.currentNS() {
			if uri == name.Space && pfx != "" {
				prefix = pfx
				break
			}
		}
	}
	return QName{Prefix: prefix, Local: local}
}

func (p *Parser) Parse(xmlData string) (*Node, error) {
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	
	doc := &Node{
		Type:     DocumentNode,
		Name:     QName{Local: "#document"},
		Children: []*Node{},
	}
	
	var currentNode *Node = doc
	p.nsStack = []map[string]string{{}}
	
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		
		switch tok := token.(type) {
		case xml.StartElement:
			p.pushNS()
			
			for _, attr := range tok.Attr {
				if attr.Name.Local == "xmlns" && attr.Name.Space == "" {
					p.currentNS()[""] = attr.Value
				} else if attr.Name.Space == "xmlns" {
					p.currentNS()[attr.Name.Local] = attr.Value
				}
			}
			
			namespace := ""
			if ns, exists := p.currentNS()[""]; exists {
				namespace = ns
			}
			if len(tok.Name.Space) > 0 {
				namespace = tok.Name.Space
			}
			
			elem := &Node{
				Type:      ElementNode,
				Name:      p.parseQName(tok.Name),
				Namespace: namespace,
				Parent:    currentNode,
				Children:  []*Node{},
			}
			
			for _, attr := range tok.Attr {
				if attr.Name.Space == "xmlns" || (attr.Name.Local == "xmlns" && attr.Name.Space == "") {
					continue
				}
				
				attrNS := ""
				if len(attr.Name.Space) > 0 {
					attrNS = attr.Name.Space
				}
				
				attrNode := &Node{
					Type:      AttributeNode,
					Name:      p.parseQName(attr.Name),
					Namespace: attrNS,
					Text:      attr.Value,
					Parent:    elem,
				}
				elem.Attributes = append(elem.Attributes, attrNode)
			}
			
			currentNode.Children = append(currentNode.Children, elem)
			currentNode = elem
			
		case xml.EndElement:
			if currentNode.Parent != nil {
				currentNode = currentNode.Parent
			}
			p.popNS()
			
		case xml.CharData:
			text := strings.TrimSpace(string(tok))
			if len(text) > 0 {
				textNode := &Node{
					Type:   TextNode,
					Name:   QName{Local: "#text"},
					Text:   string(tok),
					Parent: currentNode,
				}
				currentNode.Children = append(currentNode.Children, textNode)
			}
			
		case xml.Comment:
		case xml.ProcInst:
		case xml.Directive:
		}
	}
	
	return doc, nil
}
