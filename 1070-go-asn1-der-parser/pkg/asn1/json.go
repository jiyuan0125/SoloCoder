package asn1

import (
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/asn1-der-parser/pkg/api"
)

func NodeToJSON(n *Node) *api.NodeJSON {
	if n == nil {
		return nil
	}

	j := &api.NodeJSON{
		Tag: api.TagJSON{
			Class:       tagClassName(n.Tag.Class),
			Constructed: n.Tag.Constructed,
			Number:      n.Tag.Number,
			Name:        tagName(n.Tag),
		},
		Length: n.Length,
	}

	if n.Tag.Constructed {
		if len(n.Value.Children) > 0 {
			j.Children = make([]*api.NodeJSON, len(n.Value.Children))
			for i, c := range n.Value.Children {
				j.Children[i] = NodeToJSON(c)
			}
		}
	} else {
		if len(n.Value.Bytes) > 0 {
			j.ValueHex = hex.EncodeToString(n.Value.Bytes)
		}
		if str, ok := n.AsString(); ok {
			j.ValueStr = str
		}
		if oid, ok := n.AsOID(); ok {
			j.OID = oid
		}
	}

	return j
}

func NodeFromJSON(j *api.NodeJSON) (*Node, error) {
	if j == nil {
		return nil, nil
	}

	tag := Tag{
		Class:       tagClassFromName(j.Tag.Class),
		Constructed: j.Tag.Constructed,
		Number:      j.Tag.Number,
	}

	n := &Node{
		Tag:    tag,
		Length: j.Length,
	}

	if tag.Constructed {
		if len(j.Children) > 0 {
			children := make([]*Node, len(j.Children))
			for i, c := range j.Children {
				child, err := NodeFromJSON(c)
				if err != nil {
					return nil, err
				}
				children[i] = child
			}
			n.Value.Children = children
		}
	} else {
		if j.ValueHex != "" {
			bytes, err := hex.DecodeString(j.ValueHex)
			if err != nil {
				return nil, err
			}
			n.Value.Bytes = bytes
		}
	}

	return n, nil
}

func tagClassName(c TagClass) string {
	switch c {
	case ClassUniversal:
		return "Universal"
	case ClassApplication:
		return "Application"
	case ClassContextSpecific:
		return "Context-specific"
	case ClassPrivate:
		return "Private"
	}
	return strconv.Itoa(int(c))
}

func tagClassFromName(name string) TagClass {
	switch strings.ToLower(name) {
	case "universal":
		return ClassUniversal
	case "application":
		return ClassApplication
	case "context-specific":
		return ClassContextSpecific
	case "private":
		return ClassPrivate
	}
	return ClassUniversal
}

func tagName(t Tag) string {
	if t.Class == ClassUniversal {
		if n, ok := tagNumberNames[t.Number]; ok {
			return n
		}
	}
	return ""
}
