package tagparser

import (
	"go-reflect-tags/pkg/common"
)

func ParseDbTag(tag string) *common.DbTagInfo {
	if tag == "" {
		return nil
	}

	info := &common.DbTagInfo{}
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
		key, _ := parseKeyValue(opt)
		switch common.DbIndexType(key) {
		case common.DbIndexPK:
			info.IndexType = common.DbIndexPK
		case common.DbIndexUnique:
			info.IndexType = common.DbIndexUnique
		case common.DbIndexIndex:
			info.IndexType = common.DbIndexIndex
		}
	}

	return info
}
