package internal

type Granularity string

const (
	GranularityMinute Granularity = "minute"
	GranularityHour   Granularity = "hour"
	GranularityDay    Granularity = "day"
)

type AggregationMethod string

const (
	AggAvg   AggregationMethod = "avg"
	AggMax   AggregationMethod = "max"
	AggMin   AggregationMethod = "min"
	AggSum   AggregationMethod = "sum"
	AggCount AggregationMethod = "count"
)

type ColumnConfig struct {
	Index   int
	Name    string
	Method  AggregationMethod
}

type Config struct {
	InputFile        string
	OutputFile       string
	Granularity      Granularity
	TimeColumnIndex  int
	Columns          []ColumnConfig
	FillMissing      bool
	FillMethod       string
}

type TimeBucket struct {
	Start    int64
	End      int64
	Original string
}

type AggregationState struct {
	Min    float64
	Max    float64
	Sum    float64
	Count  int
}

func NewAggregationState() *AggregationState {
	return &AggregationState{
		Min: 1e308,
		Max: -1e308,
	}
}
