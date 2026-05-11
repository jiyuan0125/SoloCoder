package normal

type Database interface {
	Query() string
}

type MySQL struct {
	dsn string
}

func (m *MySQL) Query() string {
	return "mysql data"
}

func NewDatabase(dsn string) Database {
	return &MySQL{dsn: dsn}
}
