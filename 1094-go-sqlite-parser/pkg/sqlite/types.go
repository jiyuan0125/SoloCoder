package sqlite

type Header struct {
	MagicString        string
	PageSize           uint16
	FileFormatWrite    uint8
	FileFormatRead     uint8
	ReservedSpace      uint8
	MaxEmbeddedPayload uint8
	MinEmbeddedPayload uint8
	LeafPayloadFraction uint8
	FileChangeCounter  uint32
	PageCount          uint32
	FirstFreelistPage  uint32
	FreelistPageCount  uint32
	SchemaCookie       uint32
	SchemaFormat       uint32
	DefaultPageCache   uint32
	AutoVacuumTop      uint32
	IncrementalVacuum  uint32
	TextEncoding       uint32
	UserVersion        uint32
	ApplicationID      uint32
	VersionValidFor    uint32
	SQLiteVersion      uint32
}

type TableInfo struct {
	Name    string
	SQL     string
	RootPage uint32
}

type Database struct {
	pages    map[uint32][]byte
	pageSize uint32
	header   *Header
	tables   map[string]*TableInfo
}

type Row struct {
	values []interface{}
}

func (r *Row) GetInteger(index int) (int64, bool) {
	if index < 0 || index >= len(r.values) {
		return 0, false
	}
	v, ok := r.values[index].(int64)
	return v, ok
}

func (r *Row) GetText(index int) (string, bool) {
	if index < 0 || index >= len(r.values) {
		return "", false
	}
	v, ok := r.values[index].(string)
	return v, ok
}

func (r *Row) GetBlob(index int) ([]byte, bool) {
	if index < 0 || index >= len(r.values) {
		return nil, false
	}
	v, ok := r.values[index].([]byte)
	return v, ok
}

func (r *Row) IsNull(index int) bool {
	if index < 0 || index >= len(r.values) {
		return true
	}
	return r.values[index] == nil
}

func (r *Row) Values() []interface{} {
	result := make([]interface{}, len(r.values))
	for i, v := range r.values {
		switch val := v.(type) {
		case []byte:
			result[i] = val
		default:
			result[i] = v
		}
	}
	return result
}

func GetIntegers(row map[string]interface{}, column string) (int64, bool) {
	v, ok := row[column]
	if !ok {
		return 0, false
	}
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int8:
		return int64(val), true
	case int16:
		return int64(val), true
	case int32:
		return int64(val), true
	case int64:
		return val, true
	case float64:
		return int64(val), true
	default:
		return 0, false
	}
}

func GetTexts(row map[string]interface{}, column string) (string, bool) {
	v, ok := row[column]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func GetBlobs(row map[string]interface{}, column string) ([]byte, bool) {
	v, ok := row[column]
	if !ok {
		return nil, false
	}
	b, ok := v.([]byte)
	return b, ok
}
