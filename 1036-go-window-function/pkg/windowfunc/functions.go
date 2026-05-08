package windowfunc

import "windowfunc/api"

func applyRowNumber(partition []api.Record, alias string) {
	for i := range partition {
		partition[i][alias] = i + 1
	}
}

func applyRank(partition []api.Record, alias string, sc sortConfig) {
	if len(partition) == 0 {
		return
	}
	orderBy := sc.orderBy
	rank := 1
	count := 1
	for i := range partition {
		if i > 0 {
			if orderBy != "" {
				prev := partition[i-1][orderBy]
				curr := partition[i][orderBy]
				if !valuesEqual(prev, curr) {
					rank = count
				}
			} else {
				rank = count
			}
		}
		partition[i][alias] = rank
		count++
	}
}

func applyDenseRank(partition []api.Record, alias string, sc sortConfig) {
	if len(partition) == 0 {
		return
	}
	orderBy := sc.orderBy
	rank := 1
	for i := range partition {
		if i > 0 {
			if orderBy != "" {
				prev := partition[i-1][orderBy]
				curr := partition[i][orderBy]
				if !valuesEqual(prev, curr) {
					rank++
				}
			} else {
				rank++
			}
		}
		partition[i][alias] = rank
	}
}

func applyLag(partition []api.Record, alias string, field string, offset int) {
	for i := range partition {
		idx := i - offset
		if idx < 0 || idx >= len(partition) {
			partition[i][alias] = nil
		} else {
			partition[i][alias] = partition[idx][field]
		}
	}
}

func applyLead(partition []api.Record, alias string, field string, offset int) {
	for i := range partition {
		idx := i + offset
		if idx < 0 || idx >= len(partition) {
			partition[i][alias] = nil
		} else {
			partition[i][alias] = partition[idx][field]
		}
	}
}
