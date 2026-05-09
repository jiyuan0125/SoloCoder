package orm

import "database/sql"

func NewDB(sqlDB *sql.DB) *DB {
	return &DB{
		sqlDB:          sqlDB,
		converters:     make(map[string]TypeConverter),
		typeConverters: make(map[interface{}]TypeConverter),
	}
}

func (db *DB) RegisterConverter(name string, converter TypeConverter) {
	db.converters[name] = converter
}

func (db *DB) RegisterTypeConverter(typ interface{}, converter TypeConverter) {
	db.typeConverters[typ] = converter
}

func (db *DB) GetConverterByName(name string) TypeConverter {
	return db.converters[name]
}

func (db *DB) GetConverterByType(typ interface{}) TypeConverter {
	return db.typeConverters[typ]
}
