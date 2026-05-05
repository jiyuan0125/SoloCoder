package comparator

import (
    "sort"

    "github.com/config-diff/internal/parser"
    "github.com/config-diff/internal/protocol"
)

type Comparator struct{}

func NewComparator() *Comparator {
    return &Comparator{}
}

func (c *Comparator) Compare(config1, config2 parser.Config, ignoreKeys []string) ([]protocol.DiffResult, bool) {
    ignoreMap := make(map[string]bool)
    for _, key := range ignoreKeys {
        ignoreMap[key] = true
    }

    diffs := []protocol.DiffResult{}

    allKeys := make(map[string]bool)
    for key := range config1 {
        allKeys[key] = true
    }
    for key := range config2 {
        allKeys[key] = true
    }

    sortedKeys := make([]string, 0, len(allKeys))
    for key := range allKeys {
        sortedKeys = append(sortedKeys, key)
    }
    sort.Strings(sortedKeys)

    for _, key := range sortedKeys {
        if ignoreMap[key] {
            continue
        }

        val1, exists1 := config1[key]
        val2, exists2 := config2[key]

        if exists1 && !exists2 {
            diffs = append(diffs, protocol.DiffResult{
                Type:     protocol.DiffRemoved,
                Key:      key,
                OldValue: val1,
            })
        } else if !exists1 && exists2 {
            diffs = append(diffs, protocol.DiffResult{
                Type:     protocol.DiffAdded,
                Key:      key,
                NewValue: val2,
            })
        } else {
            if val1 != val2 {
                diffs = append(diffs, protocol.DiffResult{
                    Type:     protocol.DiffModified,
                    Key:      key,
                    OldValue: val1,
                    NewValue: val2,
                })
            }
        }
    }

    return diffs, len(diffs) == 0
}

func (c *Comparator) FormatDiff(diff protocol.DiffResult) string {
    switch diff.Type {
    case protocol.DiffAdded:
        return "+ " + diff.Key + ": " + diff.NewValue
    case protocol.DiffRemoved:
        return "- " + diff.Key + ": " + diff.OldValue
    case protocol.DiffModified:
        return "~ " + diff.Key + ": " + diff.OldValue + " -> " + diff.NewValue
    default:
        return ""
    }
}

func (c *Comparator) FormatDiffs(diffs []protocol.DiffResult) []string {
    result := make([]string, 0, len(diffs))
    for _, diff := range diffs {
        result = append(result, c.FormatDiff(diff))
    }
    return result
}
