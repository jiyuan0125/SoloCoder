package windowfunc

import (
	"bytes"
	"encoding/json"
	"sort"
	"windowfunc/api"
)

func partitionBy(data []api.Record, partitionBy string) [][]api.Record {
	if partitionBy == "" {
		return [][]api.Record{data}
	}
	groups := make(map[string][]api.Record)
	for _, r := range data {
		key := getPartitionKey(r, partitionBy)
		groups[key] = append(groups[key], r)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([][]api.Record, 0, len(keys))
	for _, k := range keys {
		result = append(result, groups[k])
	}
	return result
}

func getPartitionKey(r api.Record, field string) string {
	v, ok := r[field]
	if !ok || v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}

func orderRows(rows []api.Record, sc sortConfig) {
	if sc.orderBy == "" {
		return
	}
	nullsFirst := sc.nullsOrder == api.NullsFirst
	desc := sc.order == api.SortDesc
	sort.SliceStable(rows, func(i, j int) bool {
		iv := rows[i][sc.orderBy]
		jv := rows[j][sc.orderBy]
		ic := iv == nil
		jc := jv == nil
		if ic && jc {
			return false
		}
		if ic {
			return nullsFirst
		}
		if jc {
			return !nullsFirst
		}
		cmp := compareValues(iv, jv)
		if cmp == 0 {
			return false
		}
		if desc {
			return cmp > 0
		}
		return cmp < 0
	})
}

func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return bytes.Compare(aj, bj)
}

func valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return bytes.Equal(aj, bj)
}
