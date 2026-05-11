package columnstore

import "fmt"

type Store struct {
	tables map[string]*Table
}

func NewStore() *Store {
	return &Store{
		tables: make(map[string]*Table),
	}
}

func (s *Store) CreateTable(schema TableSchema) error {
	if _, exists := s.tables[schema.Name]; exists {
		return fmt.Errorf("table %s already exists", schema.Name)
	}
	s.tables[schema.Name] = NewTable(schema)
	return nil
}

func (s *Store) GetTable(name string) (*Table, error) {
	tbl, exists := s.tables[name]
	if !exists {
		return nil, fmt.Errorf("table %s not found", name)
	}
	return tbl, nil
}

func (s *Store) DropTable(name string) error {
	if _, exists := s.tables[name]; !exists {
		return fmt.Errorf("table %s not found", name)
	}
	delete(s.tables, name)
	return nil
}

func (s *Store) ListTables() []string {
	names := make([]string, 0, len(s.tables))
	for name := range s.tables {
		names = append(names, name)
	}
	return names
}

type ColumnStats struct {
	Name             string
	Type             string
	Encoding         string
	CompressionRatio float64
	RowCount         int
}

func (t *Table) ColumnStats() []ColumnStats {
	stats := make([]ColumnStats, len(t.columns))
	for i, col := range t.columns {
		stats[i] = ColumnStats{
			Name:             col.schema.Name,
			Type:             col.schema.Type.String(),
			Encoding:         t.GetColumnEncoding(i).String(),
			CompressionRatio: t.GetColumnCompressionRatio(i),
			RowCount:         t.RowCount(),
		}
	}
	return stats
}
