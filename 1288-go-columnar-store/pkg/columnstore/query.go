package columnstore

import "fmt"

type QueryResult struct {
	Columns []ColumnSchema
	Rows    []Row
}

func (t *Table) Query(columnNames []string, predicates []Predicate) (*QueryResult, error) {
	if t.RowCount() == 0 {
		result := &QueryResult{
			Columns: make([]ColumnSchema, 0),
			Rows:    make([]Row, 0),
		}
		for _, name := range columnNames {
			idx := t.ColumnIndex(name)
			if idx == -1 {
				return nil, fmt.Errorf("column %s not found", name)
			}
			result.Columns = append(result.Columns, t.schema.Columns[idx])
		}
		return result, nil
	}
	var selectedColIndices []int
	if len(columnNames) == 0 {
		for i := range t.columns {
			selectedColIndices = append(selectedColIndices, i)
		}
	} else {
		for _, name := range columnNames {
			idx := t.ColumnIndex(name)
			if idx == -1 {
				return nil, fmt.Errorf("column %s not found", name)
			}
			selectedColIndices = append(selectedColIndices, idx)
		}
	}
	matchedRowIndices := t.evaluatePredicates(predicates)
	result := &QueryResult{
		Columns: make([]ColumnSchema, len(selectedColIndices)),
		Rows:    make([]Row, len(matchedRowIndices)),
	}
	for i, colIdx := range selectedColIndices {
		result.Columns[i] = t.schema.Columns[colIdx]
	}
	for i, rowIdx := range matchedRowIndices {
		row := Row{
			Values: make([]Value, len(selectedColIndices)),
		}
		for j, colIdx := range selectedColIndices {
			row.Values[j] = t.GetValue(colIdx, rowIdx)
		}
		result.Rows[i] = row
	}
	return result, nil
}

func (t *Table) evaluatePredicates(predicates []Predicate) []int {
	rowCount := t.RowCount()
	if len(predicates) == 0 {
		indices := make([]int, rowCount)
		for i := 0; i < rowCount; i++ {
			indices[i] = i
		}
		return indices
	}
	result := make([]int, 0)
	for rowIdx := 0; rowIdx < rowCount; rowIdx++ {
		if t.matchPredicates(rowIdx, predicates) {
			result = append(result, rowIdx)
		}
	}
	return result
}

func (t *Table) matchPredicates(rowIdx int, predicates []Predicate) bool {
	for _, pred := range predicates {
		colIdx := t.ColumnIndex(pred.Column)
		if colIdx == -1 {
			return false
		}
		val := t.GetValue(colIdx, rowIdx)
		if !matchPredicate(val, pred) {
			return false
		}
	}
	return true
}

func matchPredicate(v Value, pred Predicate) bool {
	switch pred.Op {
	case OpEq:
		return v.Equals(pred.Value)
	case OpNeq:
		return !v.Equals(pred.Value)
	case OpLt:
		return v.Compare(pred.Value) < 0
	case OpLte:
		return v.Compare(pred.Value) <= 0
	case OpGt:
		return v.Compare(pred.Value) > 0
	case OpGte:
		return v.Compare(pred.Value) >= 0
	case OpIsNull:
		return v.Null
	case OpIsNotNull:
		return !v.Null
	default:
		return false
	}
}

func (t *Table) ExportAll() (*QueryResult, error) {
	return t.Query(nil, nil)
}

func (r *QueryResult) ToRows() []Row {
	return r.Rows
}
