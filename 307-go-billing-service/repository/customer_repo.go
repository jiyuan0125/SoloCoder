package repository

import (
	"billing-service/model"
	"database/sql"
	"time"
)

type CustomerRepository struct {
	db *Database
}

func NewCustomerRepository(db *Database) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) Create(customer *model.Customer) error {
	query := `
		INSERT INTO customers (name, current_package, package_start_date, package_end_date, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		customer.Name,
		customer.CurrentPackage,
		customer.PackageStartDate,
		customer.PackageEndDate,
		time.Now(),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	customer.ID = uint(id)

	return nil
}

func (r *CustomerRepository) GetByID(id uint) (*model.Customer, error) {
	query := `
		SELECT id, name, current_package, package_start_date, package_end_date, created_at
		FROM customers
		WHERE id = ?
	`

	var customer model.Customer
	var pkgTypeStr string
	var packageEndDate sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&customer.ID,
		&customer.Name,
		&pkgTypeStr,
		&customer.PackageStartDate,
		&packageEndDate,
		&customer.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	customer.CurrentPackage = model.PackageType(pkgTypeStr)
	if packageEndDate.Valid {
		customer.PackageEndDate = &packageEndDate.Time
	}

	return &customer, nil
}

func (r *CustomerRepository) GetAll() ([]*model.Customer, error) {
	query := `
		SELECT id, name, current_package, package_start_date, package_end_date, created_at
		FROM customers
		ORDER BY id
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []*model.Customer
	for rows.Next() {
		var customer model.Customer
		var pkgTypeStr string
		var packageEndDate sql.NullTime

		err := rows.Scan(
			&customer.ID,
			&customer.Name,
			&pkgTypeStr,
			&customer.PackageStartDate,
			&packageEndDate,
			&customer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		customer.CurrentPackage = model.PackageType(pkgTypeStr)
		if packageEndDate.Valid {
			customer.PackageEndDate = &packageEndDate.Time
		}

		customers = append(customers, &customer)
	}

	return customers, nil
}

func (r *CustomerRepository) UpdatePackage(customerID uint, newPackage model.PackageType, changeDate time.Time) error {
	query := `
		UPDATE customers
		SET current_package = ?, package_start_date = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, newPackage, changeDate, customerID)
	return err
}

func (r *CustomerRepository) AddPackageChange(pkgChange *model.PackageChange) error {
	query := `
		INSERT INTO package_changes (customer_id, old_package, new_package, change_date, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		pkgChange.CustomerID,
		pkgChange.OldPackage,
		pkgChange.NewPackage,
		pkgChange.ChangeDate,
		time.Now(),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	pkgChange.ID = uint(id)

	return nil
}

func (r *CustomerRepository) GetPackageChanges(customerID uint, year int, month int) ([]*model.PackageChange, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	query := `
		SELECT id, customer_id, old_package, new_package, change_date, created_at
		FROM package_changes
		WHERE customer_id = ? AND change_date >= ? AND change_date < ?
		ORDER BY change_date
	`

	rows, err := r.db.Query(query, customerID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*model.PackageChange
	for rows.Next() {
		var change model.PackageChange
		var oldPkgStr, newPkgStr string

		err := rows.Scan(
			&change.ID,
			&change.CustomerID,
			&oldPkgStr,
			&newPkgStr,
			&change.ChangeDate,
			&change.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		change.OldPackage = model.PackageType(oldPkgStr)
		change.NewPackage = model.PackageType(newPkgStr)
		changes = append(changes, &change)
	}

	return changes, nil
}
