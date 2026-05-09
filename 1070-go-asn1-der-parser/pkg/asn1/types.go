package asn1

import (
	"fmt"
	"strings"
)

type TagClass int

const (
	ClassUniversal       TagClass = 0
	ClassApplication     TagClass = 1
	ClassContextSpecific TagClass = 2
	ClassPrivate         TagClass = 3
)

type Tag struct {
	Class       TagClass
	Constructed bool
	Number      int
}

type Value struct {
	Bytes    []byte
	Children []*Node
}

type Node struct {
	Tag    Tag
	Length int
	Value  Value
}

const (
	TagNumberINTEGER          = 0x02
	TagNumberBITSTRING        = 0x03
	TagNumberOCTETSTRING      = 0x04
	TagNumberNULL             = 0x05
	TagNumberOID              = 0x06
	TagNumberUTF8String       = 0x0C
	TagNumberSEQUENCE         = 0x10
	TagNumberSET              = 0x11
	TagNumberPrintableString  = 0x13
	TagNumberUTCTime          = 0x17
	TagNumberGeneralizedTime  = 0x18
)

var tagNumberNames = map[int]string{
	TagNumberINTEGER:         "INTEGER",
	TagNumberBITSTRING:       "BIT STRING",
	TagNumberOCTETSTRING:     "OCTET STRING",
	TagNumberNULL:            "NULL",
	TagNumberOID:             "OID",
	TagNumberUTF8String:      "UTF8String",
	TagNumberSEQUENCE:        "SEQUENCE",
	TagNumberSET:             "SET",
	TagNumberPrintableString: "PrintableString",
	TagNumberUTCTime:         "UTCTime",
	TagNumberGeneralizedTime: "GeneralizedTime",
}

func (t Tag) String() string {
	class := ""
	switch t.Class {
	case ClassUniversal:
		class = "Universal"
	case ClassApplication:
		class = "Application"
	case ClassContextSpecific:
		class = "Context-specific"
	case ClassPrivate:
		class = "Private"
	}

	constructed := "Primitive"
	if t.Constructed {
		constructed = "Constructed"
	}

	name := fmt.Sprintf("%d", t.Number)
	if t.Class == ClassUniversal {
		if n, ok := tagNumberNames[t.Number]; ok {
			name = n
		}
	}

	return fmt.Sprintf("[%s] %s (0x%x) %s", class, name, t.Number, constructed)
}

func (n *Node) Dump(indent int) string {
	prefix := strings.Repeat("  ", indent)
	lines := []string{
		fmt.Sprintf("%sTag: %s", prefix, n.Tag),
		fmt.Sprintf("%sLength: %d", prefix, n.Length),
	}

	if n.Tag.Constructed && len(n.Value.Children) > 0 {
		lines = append(lines, fmt.Sprintf("%sValue (children: %d):", prefix, len(n.Value.Children)))
		for _, child := range n.Value.Children {
			lines = append(lines, child.Dump(indent+1))
		}
	} else if len(n.Value.Bytes) > 0 {
		lines = append(lines, fmt.Sprintf("%sValue (bytes: %x)", prefix, n.Value.Bytes))
		if str, ok := n.AsString(); ok {
			lines = append(lines, fmt.Sprintf("%sValue (string): %s", prefix, str))
		}
		if oid, ok := n.AsOID(); ok {
			lines = append(lines, fmt.Sprintf("%sValue (OID): %s", prefix, oid))
		}
	}

	return strings.Join(lines, "\n")
}

func (n *Node) AsString() (string, bool) {
	switch {
	case n.Tag.Class == ClassUniversal && !n.Tag.Constructed:
		switch n.Tag.Number {
		case TagNumberUTF8String, TagNumberPrintableString, TagNumberUTCTime, TagNumberGeneralizedTime:
			return string(n.Value.Bytes), true
		}
	}
	return "", false
}

func (n *Node) AsOID() (string, bool) {
	if n.Tag.Class == ClassUniversal && !n.Tag.Constructed && n.Tag.Number == TagNumberOID {
		if oid, err := decodeOID(n.Value.Bytes); err == nil {
			return oid, true
		}
	}
	return "", false
}
