package normal

type Repository struct {
	db Database
}

func (r *Repository) GetData() string {
	return r.db.Query()
}

func NewRepository(db Database) *Repository {
	return &Repository{db: db}
}
