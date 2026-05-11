package columnstore

import "fmt"

type Column struct {
	schema  ColumnSchema
	encoder ColumnEncoder
	values  []Value
}

type Table struct {
	schema  TableSchema
	columns []*Column
}

func NewTable(schema TableSchema) *Table {
	t := &Table{
		schema:  schema,
		columns: make([]*Column, len(schema.Columns)),
	}
	for i, colSchema := range schema.Columns {
		t.columns[i] = &Column{
			schema: colSchema,
		}
		encType := colSchema.Encoding
		if encType == EncodingUnknown {
			encType = EncodingDictionary
		}
		t.columns[i].encoder = NewEncoder(colSchema.Type, encType, nil)
	}
	return t
}

func (t *Table) Schema() TableSchema {
	return t.schema
}

func (t *Table) AddColumn(colSchema ColumnSchema) error {
	if t.schema.ColumnIndex(colSchema.Name) != -1 {
		return fmt.Errorf("column %s already exists", colSchema.Name)
	}
	rowCount := t.RowCount()
	newColumn := &Column{
		schema: colSchema,
		values: make([]Value, rowCount),
	}
	for i := range newColumn.values {
		newColumn.values[i] = Value{Null: true}
	}
	encType := colSchema.Encoding
	if encType == EncodingUnknown {
		encType = EncodingDictionary
	}
	newColumn.encoder = NewEncoder(colSchema.Type, encType, newColumn.values)
	t.schema.Columns = append(t.schema.Columns, colSchema)
	t.columns = append(t.columns, newColumn)
	return nil
}

func (t *Table) InsertRow(row Row) error {
	if len(row.Values) != len(t.columns) {
		return fmt.Errorf("row has %d values, expected %d", len(row.Values), len(t.columns))
	}
	for i, val := range row.Values {
		t.columns[i].values = append(t.columns[i].values, val)
		t.columns[i].encoder.Append(val)
	}
	return nil
}

func (t *Table) InsertRows(rows []Row) error {
	for _, row := range rows {
		if err := t.InsertRow(row); err != nil {
			return err
		}
	}
	return nil
}

func (t *Table) RowCount() int {
	if len(t.columns) == 0 {
		return 0
	}
	return len(t.columns[0].values)
}

func (t *Table) GetValue(colIdx, rowIdx int) Value {
	if colIdx < 0 || colIdx >= len(t.columns) {
		return Value{Null: true}
	}
	col := t.columns[colIdx]
	if rowIdx < 0 || rowIdx >= len(col.values) {
		return Value{Null: true}
	}
	return col.encoder.Get(rowIdx)
}

func (t *Table) GetColumnEncoding(colIdx int) EncodingType {
	if colIdx < 0 || colIdx >= len(t.columns) {
		return EncodingUnknown
	}
	return t.columns[colIdx].encoder.EncodingType()
}

func (t *Table) GetColumnCompressionRatio(colIdx int) float64 {
	if colIdx < 0 || colIdx >= len(t.columns) {
		return 1.0
	}
	return t.columns[colIdx].encoder.CompressionRatio()
}

func (t *Table) SetColumnEncoding(colIdx int, encType EncodingType) error {
	if colIdx < 0 || colIdx >= len(t.columns) {
		return fmt.Errorf("column index %d out of range", colIdx)
	}
	if encType == EncodingUnknown {
		return fmt.Errorf("unknown encoding type")
	}
	col := t.columns[colIdx]
	values := make([]Value, len(col.values))
	for i := range col.values {
		values[i] = col.encoder.Get(i)
	}
	col.schema.Encoding = encType
	col.encoder = NewEncoder(col.schema.Type, encType, values)
	return nil
}

func (t *Table) ColumnIndex(name string) int {
	return t.schema.ColumnIndex(name)
}

func (t *Table) Columns() []*Column {
	return t.columns
}

func (c *Column) Schema() ColumnSchema {
	return c.schema
}

func (c *Column) Values() []Value {
	return c.values
}
