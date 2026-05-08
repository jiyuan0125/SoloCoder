package windowfunc

import (
	"fmt"
	"sort"
	"windowfunc/api"
)

func NewEngine() *Engine {
	return &Engine{
		datasets: make(map[string][]api.Record),
	}
}

func (e *Engine) UploadDataset(name string, data []api.Record) error {
	if name == "" {
		return fmt.Errorf("dataset name cannot be empty")
	}
	copied := make([]api.Record, len(data))
	for i, r := range data {
		copied[i] = make(api.Record, len(r))
		for k, v := range r {
			copied[i][k] = v
		}
	}
	e.datasets[name] = copied
	return nil
}

func (e *Engine) ListDatasets() []api.DatasetInfo {
	infos := make([]api.DatasetInfo, 0, len(e.datasets))
	for name, data := range e.datasets {
		infos = append(infos, api.DatasetInfo{Name: name, Size: len(data)})
	}
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})
	return infos
}

func (e *Engine) Query(name string, funcs []api.WindowFunction) ([]api.Record, error) {
	data, ok := e.datasets[name]
	if !ok {
		return nil, fmt.Errorf("dataset not found: %s", name)
	}
	if len(funcs) == 0 {
		copied := make([]api.Record, len(data))
		for i, r := range data {
			copied[i] = make(api.Record, len(r))
			for k, v := range r {
				copied[i][k] = v
			}
		}
		return copied, nil
	}
	result := make([]api.Record, len(data))
	for i, r := range data {
		result[i] = make(api.Record, len(r)+len(funcs))
		for k, v := range r {
			result[i][k] = v
		}
	}
	for _, fn := range funcs {
		if err := applyWindowFunction(result, fn); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func applyWindowFunction(data []api.Record, fn api.WindowFunction) error {
	alias := fn.Alias
	if alias == "" {
		alias = string(fn.Name)
		if fn.Field != "" {
			alias = alias + "_" + fn.Field
		}
	}
	partitions := partitionBy(data, fn.PartitionBy)
	sc := sortConfig{
		orderBy:    fn.OrderBy,
		order:      fn.Order,
		nullsOrder: fn.NullsOrder,
	}
	for _, p := range partitions {
		orderRows(p, sc)
		switch fn.Name {
		case api.FuncRowNumber:
			applyRowNumber(p, alias)
		case api.FuncRank:
			applyRank(p, alias, sc)
		case api.FuncDenseRank:
			applyDenseRank(p, alias, sc)
		case api.FuncLag:
			offset := fn.Offset
			if offset <= 0 {
				offset = 1
			}
			applyLag(p, alias, fn.Field, offset)
		case api.FuncLead:
			offset := fn.Offset
			if offset <= 0 {
				offset = 1
			}
			applyLead(p, alias, fn.Field, offset)
		default:
			return fmt.Errorf("unknown window function: %s", fn.Name)
		}
	}
	return nil
}
