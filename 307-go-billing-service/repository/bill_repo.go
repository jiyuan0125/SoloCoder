package repository

import (
	"billing-service/model"
	"database/sql"
	"time"
)

type BillRepository struct {
	db *Database
}

func NewBillRepository(db *Database) *BillRepository {
	return &BillRepository{db: db}
}

func (r *BillRepository) Create(bill *model.Bill) error {
	query := `
		INSERT INTO bills (
			customer_id, bill_year, bill_month, package_fee, sms_fee, storage_fee,
			total_amount, status, due_date, paid_at, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	result, err := r.db.Exec(
		query,
		bill.CustomerID,
		bill.BillYear,
		bill.BillMonth,
		bill.PackageFee,
		bill.SmsFee,
		bill.StorageFee,
		bill.TotalAmount,
		bill.Status,
		bill.DueDate,
		bill.PaidAt,
		now,
		now,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	bill.ID = uint(id)
	bill.CreatedAt = now
	bill.UpdatedAt = now

	return nil
}

func (r *BillRepository) GetByID(id uint) (*model.Bill, error) {
	query := `
		SELECT id, customer_id, bill_year, bill_month, package_fee, sms_fee, storage_fee,
		       total_amount, status, due_date, paid_at, created_at, updated_at
		FROM bills
		WHERE id = ?
	`

	var bill model.Bill
	var statusStr string
	var paidAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&bill.ID,
		&bill.CustomerID,
		&bill.BillYear,
		&bill.BillMonth,
		&bill.PackageFee,
		&bill.SmsFee,
		&bill.StorageFee,
		&bill.TotalAmount,
		&statusStr,
		&bill.DueDate,
		&paidAt,
		&bill.CreatedAt,
		&bill.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	bill.Status = model.BillStatus(statusStr)
	if paidAt.Valid {
		bill.PaidAt = &paidAt.Time
	}

	return &bill, nil
}

func (r *BillRepository) GetByCustomerAndMonth(customerID uint, year int, month int) (*model.Bill, error) {
	query := `
		SELECT id, customer_id, bill_year, bill_month, package_fee, sms_fee, storage_fee,
		       total_amount, status, due_date, paid_at, created_at, updated_at
		FROM bills
		WHERE customer_id = ? AND bill_year = ? AND bill_month = ?
	`

	var bill model.Bill
	var statusStr string
	var paidAt sql.NullTime

	err := r.db.QueryRow(query, customerID, year, month).Scan(
		&bill.ID,
		&bill.CustomerID,
		&bill.BillYear,
		&bill.BillMonth,
		&bill.PackageFee,
		&bill.SmsFee,
		&bill.StorageFee,
		&bill.TotalAmount,
		&statusStr,
		&bill.DueDate,
		&paidAt,
		&bill.CreatedAt,
		&bill.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	bill.Status = model.BillStatus(statusStr)
	if paidAt.Valid {
		bill.PaidAt = &paidAt.Time
	}

	return &bill, nil
}

func (r *BillRepository) GetAll() ([]*model.Bill, error) {
	query := `
		SELECT id, customer_id, bill_year, bill_month, package_fee, sms_fee, storage_fee,
		       total_amount, status, due_date, paid_at, created_at, updated_at
		FROM bills
		ORDER BY bill_year DESC, bill_month DESC, id DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []*model.Bill
	for rows.Next() {
		var bill model.Bill
		var statusStr string
		var paidAt sql.NullTime

		err := rows.Scan(
			&bill.ID,
			&bill.CustomerID,
			&bill.BillYear,
			&bill.BillMonth,
			&bill.PackageFee,
			&bill.SmsFee,
			&bill.StorageFee,
			&bill.TotalAmount,
			&statusStr,
			&bill.DueDate,
			&paidAt,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		bill.Status = model.BillStatus(statusStr)
		if paidAt.Valid {
			bill.PaidAt = &paidAt.Time
		}

		bills = append(bills, &bill)
	}

	return bills, nil
}

func (r *BillRepository) GetByCustomer(customerID uint) ([]*model.Bill, error) {
	query := `
		SELECT id, customer_id, bill_year, bill_month, package_fee, sms_fee, storage_fee,
		       total_amount, status, due_date, paid_at, created_at, updated_at
		FROM bills
		WHERE customer_id = ?
		ORDER BY bill_year DESC, bill_month DESC
	`

	rows, err := r.db.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []*model.Bill
	for rows.Next() {
		var bill model.Bill
		var statusStr string
		var paidAt sql.NullTime

		err := rows.Scan(
			&bill.ID,
			&bill.CustomerID,
			&bill.BillYear,
			&bill.BillMonth,
			&bill.PackageFee,
			&bill.SmsFee,
			&bill.StorageFee,
			&bill.TotalAmount,
			&statusStr,
			&bill.DueDate,
			&paidAt,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		bill.Status = model.BillStatus(statusStr)
		if paidAt.Valid {
			bill.PaidAt = &paidAt.Time
		}

		bills = append(bills, &bill)
	}

	return bills, nil
}

func (r *BillRepository) UpdateStatus(billID uint, status model.BillStatus) error {
	query := `
		UPDATE bills
		SET status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, status, time.Now(), billID)
	return err
}

func (r *BillRepository) MarkAsPaid(billID uint, operator string, remark string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	_, err = tx.Exec(`
		UPDATE bills
		SET status = ?, paid_at = ?, updated_at = ?
		WHERE id = ?
	`, model.BillStatusPaid, now, now, billID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO bill_payments (bill_id, operator, remark, paid_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, billID, operator, remark, now, now)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *BillRepository) GetUnpaidBillsBefore(customerID uint, year int, month int) ([]*model.Bill, error) {
	query := `
		SELECT id, customer_id, bill_year, bill_month, package_fee, sms_fee, storage_fee,
		       total_amount, status, due_date, paid_at, created_at, updated_at
		FROM bills
		WHERE customer_id = ? 
		  AND status != ?
		  AND (bill_year < ? OR (bill_year = ? AND bill_month < ?))
		ORDER BY bill_year, bill_month
	`

	rows, err := r.db.Query(query, customerID, model.BillStatusPaid, year, year, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bills []*model.Bill
	for rows.Next() {
		var bill model.Bill
		var statusStr string
		var paidAt sql.NullTime

		err := rows.Scan(
			&bill.ID,
			&bill.CustomerID,
			&bill.BillYear,
			&bill.BillMonth,
			&bill.PackageFee,
			&bill.SmsFee,
			&bill.StorageFee,
			&bill.TotalAmount,
			&statusStr,
			&bill.DueDate,
			&paidAt,
			&bill.CreatedAt,
			&bill.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		bill.Status = model.BillStatus(statusStr)
		if paidAt.Valid {
			bill.PaidAt = &paidAt.Time
		}

		bills = append(bills, &bill)
	}

	return bills, nil
}

func (r *BillRepository) GetPaymentHistory(billID uint) ([]*model.BillPayment, error) {
	query := `
		SELECT id, bill_id, operator, remark, paid_at, created_at
		FROM bill_payments
		WHERE bill_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*model.BillPayment
	for rows.Next() {
		var payment model.BillPayment
		err := rows.Scan(
			&payment.ID,
			&payment.BillID,
			&payment.Operator,
			&payment.Remark,
			&payment.PaidAt,
			&payment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}
