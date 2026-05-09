package orm

import (
	"database/sql"
	"reflect"
	"time"
)

type TypeConverter interface {
	ToDB(value interface{}) (interface{}, error)
	FromDB(value interface{}) (interface{}, error)
}

type FieldInfo struct {
	Name           string
	ColumnName     string
	GoType         reflect.Type
	IsPointer      bool
	IsSlice        bool
	IsEmbedded     bool
	EmbeddedFields []*FieldInfo
	ConverterName  string
	Converter      TypeConverter
}

type ModelInfo struct {
	TableName   string
	ModelType   interface{}
	Fields      []*FieldInfo
	ColumnMap   map[string]*FieldInfo
	PrimaryKey  string
}

type DB struct {
	sqlDB         *sql.DB
	converters    map[string]TypeConverter
	typeConverters map[interface{}]TypeConverter
}

type ListOptions struct {
	Where   map[string]interface{}
	OrderBy string
	Limit   int
	Offset  int
}

var (
	supportedTypes = map[interface{}]bool{
		int(0):        true,
		int64(0):      true,
		float64(0):    true,
		"":            true,
		true:          true,
		time.Time{}:   true,
	}
)
