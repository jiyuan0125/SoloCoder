package windowfunc

import "windowfunc/api"

type Engine struct {
	datasets map[string][]api.Record
}

type sortConfig struct {
	orderBy    string
	order      api.SortOrder
	nullsOrder api.NullsOrder
}

type partitionKey struct{}
