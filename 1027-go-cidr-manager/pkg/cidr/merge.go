package cidr

func Merge(list CIDRList) CIDRList {
	if len(list) <= 1 {
		result := make(CIDRList, len(list))
		copy(result, list)
		return result
	}

	sorted := make(CIDRList, len(list))
	copy(sorted, list)
	sorted.Sort()

	changed := true
	for changed {
		changed = false
		var newList CIDRList
		i := 0
		for i < len(sorted) {
			if i+1 < len(sorted) {
				merged, _ := sorted[i].mergeWith(sorted[i+1])
				if merged != nil {
					newList = append(newList, merged)
					i += 2
					changed = true
					continue
				}
			}
			newList = append(newList, sorted[i])
			i++
		}
		sorted = newList
	}

	return sorted
}

func TryMergeTwo(a, b *CIDR) (*CIDR, error) {
	return a.mergeWith(b)
}
