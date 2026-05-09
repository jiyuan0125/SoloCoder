package natsort

import (
	"sort"
)

func Sort(strings []string, opts Options) []string {
	result := make([]string, len(strings))
	copy(result, strings)

	items := make([]indexedString, len(strings))
	for i, s := range strings {
		items[i] = indexedString{index: i, str: s}
	}

	sort.SliceStable(items, func(i, j int) bool {
		cmp := Compare(items[i].str, items[j].str, opts)
		if cmp != 0 {
			return cmp < 0
		}
		return items[i].index < items[j].index
	})

	for i, item := range items {
		result[i] = item.str
	}

	return result
}

type indexedString struct {
	index int
	str   string
}
