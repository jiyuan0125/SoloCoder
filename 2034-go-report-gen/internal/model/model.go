package model

import (
	"time"
)

type FieldType string

const (
	FieldTypeString  FieldType = "string"
	FieldTypeNumber  FieldType = "number"
	FieldTypeDate    FieldType = "date"
	FieldTypeBoolean FieldType = "boolean"
)

type AggregationType string

const (
	AggSum   AggregationType = "sum"
	AggAvg   AggregationType = "avg"
	AggCount AggregationType = "count"
	AggMin   AggregationType = "min"
	AggMax   AggregationType = "max"
)

type DateGroupType string

const (
	DateGroupYear  DateGroupType = "year"
	DateGroupMonth DateGroupType = "month"
	DateGroupWeek  DateGroupType = "week"
	DateGroupDay   DateGroupType = "day"
)

type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatHTML OutputFormat = "html"
)

type DataSourceType string

const (
	DataSourceCSV      DataSourceType = "csv"
	DataSourceMySQL    DataSourceType = "mysql"
	DataSourcePostgres DataSourceType = "postgres"
)

type Column struct {
	Name string    `yaml:"name"`
	Type FieldType `yaml:"type"`
}

type DataSource struct {
	Type      DataSourceType `yaml:"type"`
	Path      string         `yaml:"path"`
	Delimiter string         `yaml:"delimiter"`
	Host      string         `yaml:"host"`
	Port      int            `yaml:"port"`
	Database  string         `yaml:"database"`
	Table     string         `yaml:"table"`
	Query     string         `yaml:"query"`
	Username  string         `yaml:"username"`
	Password  string         `yaml:"password"`
	Columns   []Column       `yaml:"columns"`
}

type FilterOperator string

const (
	FilterEq    FilterOperator = "eq"
	FilterNe    FilterOperator = "ne"
	FilterGt    FilterOperator = "gt"
	FilterGte   FilterOperator = "gte"
	FilterLt    FilterOperator = "lt"
	FilterLte   FilterOperator = "lte"
	FilterLike  FilterOperator = "like"
	FilterIn    FilterOperator = "in"
	FilterBetween FilterOperator = "between"
)

type Filter struct {
	Field    string         `yaml:"field"`
	Operator FilterOperator `yaml:"operator"`
	Value    interface{}    `yaml:"value"`
}

type GroupBy struct {
	Field     string        `yaml:"field"`
	DateGroup DateGroupType `yaml:"date_group"`
}

type Aggregation struct {
	Field string          `yaml:"field"`
	Type  AggregationType `yaml:"type"`
	Alias string          `yaml:"alias"`
}

type AnomalyRule struct {
	Field      string      `yaml:"field"`
	Operator   string      `yaml:"operator"`
	Threshold  interface{} `yaml:"threshold"`
}

type Template struct {
	Name          string         `yaml:"name"`
	Title         string         `yaml:"title"`
	Columns       []string       `yaml:"columns"`
	Filters       []Filter       `yaml:"filters"`
	GroupBy       []GroupBy      `yaml:"group_by"`
	Aggregations  []Aggregation  `yaml:"aggregations"`
	AnomalyRules  []AnomalyRule  `yaml:"anomaly_rules"`
	SortBy        []string       `yaml:"sort_by"`
	Limit         int            `yaml:"limit"`
}

type ReportConfig struct {
	Name         string       `yaml:"name"`
	Description  string       `yaml:"description"`
	DataSource   DataSource   `yaml:"data_source"`
	Template     Template     `yaml:"template"`
	OutputFormat OutputFormat `yaml:"output_format"`
	OutputPath   string       `yaml:"output_path"`
}

type DataRow struct {
	Values map[string]interface{}
}

type ProcessedValue struct {
	Value     interface{}
	IsAnomaly bool
}

type ProcessedRow struct {
	Values map[string]ProcessedValue
}

type GroupKey struct {
	Fields map[string]interface{}
}

type GroupData struct {
	Key    GroupKey
	Values map[string]interface{}
	Count  int
}

type ReportResult struct {
	Name    string
	Title   string
	Headers []string
	Rows    []ProcessedRow
	Groups  []GroupData
	HasAnomaly bool
}

func (c FieldType) String() string {
	return string(c)
}

func (a AggregationType) String() string {
	return string(a)
}

func (d DateGroupType) String() string {
	return string(d)
}

func (o OutputFormat) String() string {
	return string(o)
}

func (d DataSourceType) String() string {
	return string(d)
}

func IsValidDateGroupType(s string) bool {
	switch DateGroupType(s) {
	case DateGroupYear, DateGroupMonth, DateGroupWeek, DateGroupDay:
		return true
	default:
		return false
	}
}

func ParseDateGroupType(s string) (DateGroupType, bool) {
	d := DateGroupType(s)
	return d, IsValidDateGroupType(s)
}

func IsValidAggregationType(s string) bool {
	switch AggregationType(s) {
	case AggSum, AggAvg, AggCount, AggMin, AggMax:
		return true
	default:
		return false
	}
}

func ParseAggregationType(s string) (AggregationType, bool) {
	a := AggregationType(s)
	return a, IsValidAggregationType(s)
}

func GetWeekStart(t time.Time) time.Time {
	offset := int(t.Weekday())
	if offset == 0 {
		offset = 7
	}
	return t.AddDate(0, 0, -(offset - 1))
}

func GetISOWeek(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}
