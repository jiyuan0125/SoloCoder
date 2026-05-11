package tagparser

import (
	"go-reflect-tags/pkg/common"
)

func ParseJsonTag(tag string) *common.JsonTagInfo {
	if tag == "" {
		return nil
	}

	info := &common.JsonTagInfo{}
	options := splitTagOptions(tag)

	if len(options) == 0 {
		return info
	}

	first := options[0]
	if first == "-" {
		info.Ignored = true
		return info
	}

	info.Name = first

	for i := 1; i < len(options); i++ {
		opt := options[i]
		switch opt {
		case "omitempty":
			info.OmitEmpty = true
		case "string":
			info.String = true
		}
	}

	return info
}
