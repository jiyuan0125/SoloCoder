package proto

const (
	DefaultPort         = 8080
	DefaultHost         = "localhost"
	DefaultBatchSize    = 100
	DefaultBufferSize   = 4096
)

type ConverterConfig struct {
	BatchSize    int
	SkipBOM      bool
	EscapeQuotes bool
	HandleNull   bool
	NullStrings  []string
	SQLKeywords  map[string]bool
}

func DefaultConverterConfig() *ConverterConfig {
	return &ConverterConfig{
		BatchSize:    DefaultBatchSize,
		SkipBOM:      true,
		EscapeQuotes: true,
		HandleNull:   true,
		NullStrings:  []string{"", "NULL", "null"},
		SQLKeywords:  getSQLKeywords(),
	}
}

func getSQLKeywords() map[string]bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "AND": true, "OR": true,
		"INSERT": true, "UPDATE": true, "DELETE": true, "CREATE": true, "DROP": true,
		"TABLE": true, "DATABASE": true, "INDEX": true, "VIEW": true, "TRIGGER": true,
		"PROCEDURE": true, "FUNCTION": true, "BEGIN": true, "COMMIT": true, "ROLLBACK": true,
		"GRANT": true, "REVOKE": true, "ALTER": true, "ADD": true, "MODIFY": true,
		"COLUMN": true, "ROW": true, "VALUES": true, "SET": true, "JOIN": true,
		"LEFT": true, "RIGHT": true, "INNER": true, "OUTER": true, "ON": true,
		"GROUP": true, "BY": true, "ORDER": true, "HAVING": true, "LIMIT": true,
		"OFFSET": true, "UNION": true, "ALL": true, "DISTINCT": true, "AS": true,
		"IN": true, "NOT": true, "EXISTS": true, "BETWEEN": true, "LIKE": true,
		"NULL": true, "IS": true, "TRUE": true, "FALSE": true, "CASE": true,
		"WHEN": true, "THEN": true, "ELSE": true, "END": true, "CAST": true,
		"CONVERT": true, "COALESCE": true, "IFNULL": true, "NVL": true,
		"COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
		"FIRST": true, "LAST": true, "TOP": true, "PERCENT": true,
	}
	return keywords
}
