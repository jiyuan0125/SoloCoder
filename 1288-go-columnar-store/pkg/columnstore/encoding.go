package columnstore

type ColumnEncoder interface {
	Append(v Value)
	Get(idx int) Value
	Size() int
	CompressionRatio() float64
	EncodingType() EncodingType
	EstimateSize(values []Value) int
	RebuildFromValues(values []Value)
}

type DictionaryEncoder struct {
	dict    []Value
	indices []int
	colType ColumnType
}

func NewDictionaryEncoder(colType ColumnType) *DictionaryEncoder {
	return &DictionaryEncoder{
		colType: colType,
	}
}

func (e *DictionaryEncoder) EncodingType() EncodingType {
	return EncodingDictionary
}

func (e *DictionaryEncoder) Append(v Value) {
	idx := -1
	for i, dv := range e.dict {
		if dv.Equals(v) {
			idx = i
			break
		}
	}
	if idx == -1 {
		idx = len(e.dict)
		e.dict = append(e.dict, v)
	}
	e.indices = append(e.indices, idx)
}

func (e *DictionaryEncoder) Get(idx int) Value {
	if idx < 0 || idx >= len(e.indices) {
		return Value{Null: true}
	}
	dictIdx := e.indices[idx]
	if dictIdx < 0 || dictIdx >= len(e.dict) {
		return Value{Null: true}
	}
	return e.dict[dictIdx]
}

func (e *DictionaryEncoder) Size() int {
	return len(e.indices)
}

func (e *DictionaryEncoder) CompressionRatio() float64 {
	if len(e.indices) == 0 {
		return 1.0
	}
	rawSize := len(e.indices) * valueSize(e.colType)
	compressedSize := len(e.dict)*valueSize(e.colType) + len(e.indices)*8
	if rawSize == 0 {
		return 1.0
	}
	return float64(rawSize) / float64(compressedSize)
}

func (e *DictionaryEncoder) EstimateSize(values []Value) int {
	if len(values) == 0 {
		return 0
	}
	uniqueCount := make(map[int]struct{})
	dict := make([]Value, 0)
	for _, v := range values {
		found := false
		for i, dv := range dict {
			if dv.Equals(v) {
				uniqueCount[i] = struct{}{}
				found = true
				break
			}
		}
		if !found {
			dict = append(dict, v)
			uniqueCount[len(dict)-1] = struct{}{}
		}
	}
	return len(dict)*valueSize(e.colType) + len(values)*8
}

func (e *DictionaryEncoder) RebuildFromValues(values []Value) {
	e.dict = make([]Value, 0)
	e.indices = make([]int, 0)
	for _, v := range values {
		e.Append(v)
	}
}

type RunLengthEncoder struct {
	runs    []Run
	colType ColumnType
}

type Run struct {
	Value  Value
	Length int
}

func NewRunLengthEncoder(colType ColumnType) *RunLengthEncoder {
	return &RunLengthEncoder{
		colType: colType,
	}
}

func (e *RunLengthEncoder) EncodingType() EncodingType {
	return EncodingRunLength
}

func (e *RunLengthEncoder) Append(v Value) {
	if len(e.runs) > 0 {
		lastRun := &e.runs[len(e.runs)-1]
		if lastRun.Value.Equals(v) {
			lastRun.Length++
			return
		}
	}
	e.runs = append(e.runs, Run{Value: v, Length: 1})
}

func (e *RunLengthEncoder) Get(idx int) Value {
	if idx < 0 {
		return Value{Null: true}
	}
	currentIdx := 0
	for _, run := range e.runs {
		if idx < currentIdx+run.Length {
			return run.Value
		}
		currentIdx += run.Length
	}
	return Value{Null: true}
}

func (e *RunLengthEncoder) Size() int {
	total := 0
	for _, run := range e.runs {
		total += run.Length
	}
	return total
}

func (e *RunLengthEncoder) CompressionRatio() float64 {
	totalSize := e.Size()
	if totalSize == 0 {
		return 1.0
	}
	rawSize := totalSize * valueSize(e.colType)
	compressedSize := len(e.runs) * (valueSize(e.colType) + 8)
	if rawSize == 0 {
		return 1.0
	}
	return float64(rawSize) / float64(compressedSize)
}

func (e *RunLengthEncoder) EstimateSize(values []Value) int {
	if len(values) == 0 {
		return 0
	}
	runCount := 1
	prev := values[0]
	for i := 1; i < len(values); i++ {
		if !values[i].Equals(prev) {
			runCount++
			prev = values[i]
		}
	}
	return runCount * (valueSize(e.colType) + 8)
}

func (e *RunLengthEncoder) RebuildFromValues(values []Value) {
	e.runs = make([]Run, 0)
	for _, v := range values {
		e.Append(v)
	}
}

func valueSize(t ColumnType) int {
	switch t {
	case ColumnTypeInt64, ColumnTypeFloat64:
		return 8
	case ColumnTypeString:
		return 32
	default:
		return 16
	}
}

func ChooseOptimalEncoding(colType ColumnType, values []Value) EncodingType {
	if len(values) == 0 {
		return EncodingDictionary
	}
	dictEnc := NewDictionaryEncoder(colType)
	rleEnc := NewRunLengthEncoder(colType)
	dictSize := dictEnc.EstimateSize(values)
	rleSize := rleEnc.EstimateSize(values)
	if dictSize <= rleSize {
		return EncodingDictionary
	}
	return EncodingRunLength
}

func NewEncoder(colType ColumnType, encType EncodingType, values []Value) ColumnEncoder {
	var enc ColumnEncoder
	switch encType {
	case EncodingRunLength:
		enc = NewRunLengthEncoder(colType)
	default:
		enc = NewDictionaryEncoder(colType)
	}
	if values != nil {
		for _, v := range values {
			enc.Append(v)
		}
	}
	return enc
}
