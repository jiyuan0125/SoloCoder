package querybuilder

import (
	"encoding/json"
)

type ConditionType string

const (
	ConditionTypeEqual    ConditionType = "equal"
	ConditionTypeRange    ConditionType = "range"
	ConditionTypeIn       ConditionType = "in"
	ConditionTypeLike     ConditionType = "like"
)

type LikeType string

const (
	LikeTypeBoth  LikeType = "both"
	LikeTypePrefix LikeType = "prefix"
	LikeTypeSuffix LikeType = "suffix"
)

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

type Condition struct {
	Field string
	Type  ConditionType
	Value interface{}
}

type EqualCondition struct {
	Field string
	Value interface{}
}

type RangeCondition struct {
	Field string
	Start interface{}
	End   interface{}
}

type InCondition struct {
	Field  string
	Values []interface{}
}

type LikeCondition struct {
	Field    string
	Value    string
	LikeType LikeType
}

type SortCondition struct {
	Field string
	Order SortOrder
}

type QueryParams struct {
	Page     int
	PageSize int
}

type Query struct {
	Conditions  []Condition
	SortOrders  []SortCondition
	Pagination  QueryParams
}

type QueryBuilder struct {
	conditions  map[string]Condition
	sortOrders  map[string]SortOrder
	pagination  QueryParams
}

func New() *QueryBuilder {
	return &QueryBuilder{
		conditions: make(map[string]Condition),
		sortOrders: make(map[string]SortOrder),
		pagination: QueryParams{
			Page:     1,
			PageSize: 10,
		},
	}
}

func (qb *QueryBuilder) clone() *QueryBuilder {
	newQb := New()
	
	for k, v := range qb.conditions {
		newQb.conditions[k] = v
	}
	
	for k, v := range qb.sortOrders {
		newQb.sortOrders[k] = v
	}
	
	newQb.pagination = qb.pagination
	
	return newQb
}

func (qb *QueryBuilder) Equal(field string, value interface{}) *QueryBuilder {
	if value == nil || value == "" {
		return qb
	}
	
	newQb := qb.clone()
	newQb.conditions[field] = Condition{
		Field: field,
		Type:  ConditionTypeEqual,
		Value: value,
	}
	
	return newQb
}

func (qb *QueryBuilder) Range(field string, start interface{}, end interface{}) *QueryBuilder {
	if start == nil && end == nil {
		return qb
	}
	
	if start == "" && end == "" {
		return qb
	}
	
	newQb := qb.clone()
	newQb.conditions[field] = Condition{
		Field: field,
		Type:  ConditionTypeRange,
		Value: RangeCondition{
			Field: field,
			Start: start,
			End:   end,
		},
	}
	
	return newQb
}

func (qb *QueryBuilder) In(field string, values ...interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	
	nonNilValues := make([]interface{}, 0)
	for _, v := range values {
		if v != nil && v != "" {
			nonNilValues = append(nonNilValues, v)
		}
	}
	
	if len(nonNilValues) == 0 {
		return qb
	}
	
	newQb := qb.clone()
	newQb.conditions[field] = Condition{
		Field: field,
		Type:  ConditionTypeIn,
		Value: InCondition{
			Field:  field,
			Values: nonNilValues,
		},
	}
	
	return newQb
}

func (qb *QueryBuilder) Like(field string, value string, likeType ...LikeType) *QueryBuilder {
	if value == "" {
		return qb
	}
	
	lt := LikeTypeBoth
	if len(likeType) > 0 {
		lt = likeType[0]
	}
	
	newQb := qb.clone()
	newQb.conditions[field] = Condition{
		Field: field,
		Type:  ConditionTypeLike,
		Value: LikeCondition{
			Field:    field,
			Value:    value,
			LikeType: lt,
		},
	}
	
	return newQb
}

func (qb *QueryBuilder) Sort(field string, order SortOrder) *QueryBuilder {
	if field == "" {
		return qb
	}
	
	newQb := qb.clone()
	newQb.sortOrders[field] = order
	
	return newQb
}

func (qb *QueryBuilder) Asc(field string) *QueryBuilder {
	return qb.Sort(field, SortOrderAsc)
}

func (qb *QueryBuilder) Desc(field string) *QueryBuilder {
	return qb.Sort(field, SortOrderDesc)
}

func (qb *QueryBuilder) Pagination(page, pageSize int) *QueryBuilder {
	newQb := qb.clone()
	
	if page < 1 {
		newQb.pagination.Page = 1
	} else {
		newQb.pagination.Page = page
	}
	
	if pageSize < 1 {
		newQb.pagination.PageSize = 10
	} else if pageSize > 100 {
		newQb.pagination.PageSize = 100
	} else {
		newQb.pagination.PageSize = pageSize
	}
	
	return newQb
}

func (qb *QueryBuilder) Reset() *QueryBuilder {
	return New()
}

func (qb *QueryBuilder) Build() Query {
	conditions := make([]Condition, 0, len(qb.conditions))
	for _, cond := range qb.conditions {
		processed := qb.processCondition(cond)
		if processed != nil {
			conditions = append(conditions, *processed)
		}
	}
	
	sortOrders := make([]SortCondition, 0, len(qb.sortOrders))
	for field, order := range qb.sortOrders {
		sortOrders = append(sortOrders, SortCondition{
			Field: field,
			Order: order,
		})
	}
	
	return Query{
		Conditions: conditions,
		SortOrders: sortOrders,
		Pagination: qb.pagination,
	}
}

func (qb *QueryBuilder) processCondition(cond Condition) *Condition {
	switch cond.Type {
	case ConditionTypeRange:
		rangeCond, ok := cond.Value.(RangeCondition)
		if !ok {
			return &cond
		}
		
		start, startOk := toFloat64(rangeCond.Start)
		end, endOk := toFloat64(rangeCond.End)
		
		if startOk && endOk && start > end {
			rangeCond.Start, rangeCond.End = rangeCond.End, rangeCond.Start
			cond.Value = rangeCond
		} else if !startOk || !endOk {
			startStr, startOkStr := toString(rangeCond.Start)
			endStr, endOkStr := toString(rangeCond.End)
			
			if startOkStr && endOkStr && startStr > endStr {
				rangeCond.Start, rangeCond.End = rangeCond.End, rangeCond.Start
				cond.Value = rangeCond
			}
		}
		
		return &cond
		
	case ConditionTypeLike:
		likeCond, ok := cond.Value.(LikeCondition)
		if !ok {
			return &cond
		}
		
		switch likeCond.LikeType {
		case LikeTypeBoth:
			likeCond.Value = "%" + likeCond.Value + "%"
		case LikeTypePrefix:
			likeCond.Value = "%" + likeCond.Value
		case LikeTypeSuffix:
			likeCond.Value = likeCond.Value + "%"
		}
		
		cond.Value = likeCond
		return &cond
		
	case ConditionTypeIn:
		inCond, ok := cond.Value.(InCondition)
		if !ok {
			return &cond
		}
		
		if len(inCond.Values) == 0 {
			return nil
		}
		
		return &cond
		
	default:
		return &cond
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}

func toString(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	switch val := v.(type) {
	case string:
		return val, true
	default:
		return "", false
	}
}

func (q Query) ToJSON() (string, error) {
	data, err := json.Marshal(q)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
