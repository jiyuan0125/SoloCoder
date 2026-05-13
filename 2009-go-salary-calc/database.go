package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	db.SetMaxOpenConns(1)

	return db, nil
}

func InitTables(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS employees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			base_salary INTEGER NOT NULL,
			allowance INTEGER NOT NULL DEFAULT 0,
			performance REAL NOT NULL DEFAULT 1.0,
			join_date TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS resources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			description TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS resource_relations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER NOT NULL,
			source_type TEXT NOT NULL,
			target_id INTEGER NOT NULL,
			target_type TEXT NOT NULL,
			relation TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS salaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			employee_id INTEGER NOT NULL,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL,
			base_salary INTEGER NOT NULL,
			allowance INTEGER NOT NULL,
			performance REAL NOT NULL,
			monthly_salary INTEGER NOT NULL,
			hourly_rate INTEGER NOT NULL,
			overtime_workday INTEGER NOT NULL DEFAULT 0,
			overtime_weekend INTEGER NOT NULL DEFAULT 0,
			overtime_holiday INTEGER NOT NULL DEFAULT 0,
			total_overtime INTEGER NOT NULL DEFAULT 0,
			social_insurance INTEGER NOT NULL DEFAULT 0,
			taxable_income INTEGER NOT NULL DEFAULT 0,
			income_tax INTEGER NOT NULL DEFAULT 0,
			net_salary INTEGER NOT NULL DEFAULT 0,
			resource_id INTEGER,
			resource_type TEXT,
			status TEXT NOT NULL DEFAULT '新建',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (employee_id) REFERENCES employees(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_salaries_employee_year_month 
		 ON salaries(employee_id, year, month)`,
		`CREATE INDEX IF NOT EXISTS idx_salaries_resource ON salaries(resource_id, resource_type)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_relations_source ON resource_relations(source_id, source_type)`,
		`CREATE INDEX IF NOT EXISTS idx_resource_relations_target ON resource_relations(target_id, target_type)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("执行SQL失败: %w", err)
		}
	}

	return nil
}

func CreateEmployee(db *sql.DB, emp *Employee) error {
	now := time.Now().UTC()
	emp.CreatedAt = now
	emp.UpdatedAt = now

	result, err := db.Exec(
		`INSERT INTO employees (name, base_salary, allowance, performance, join_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		emp.Name, emp.BaseSalary, emp.Allowance, emp.Performance,
		emp.JoinDate.Format(time.RFC3339),
		emp.CreatedAt.Format(time.RFC3339),
		emp.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	emp.ID = id
	return nil
}

func GetEmployee(db *sql.DB, id int64) (*Employee, error) {
	emp := &Employee{}
	var joinDateStr, createdAtStr, updatedAtStr string

	err := db.QueryRow(
		`SELECT id, name, base_salary, allowance, performance, join_date, created_at, updated_at
		 FROM employees WHERE id = ?`,
		id,
	).Scan(&emp.ID, &emp.Name, &emp.BaseSalary, &emp.Allowance, &emp.Performance,
		&joinDateStr, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("员工不存在")
		}
		return nil, err
	}

	emp.JoinDate, _ = time.Parse(time.RFC3339, joinDateStr)
	emp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	emp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return emp, nil
}

func ListEmployees(db *sql.DB) ([]Employee, error) {
	rows, err := db.Query(
		`SELECT id, name, base_salary, allowance, performance, join_date, created_at, updated_at
		 FROM employees ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []Employee
	for rows.Next() {
		emp := Employee{}
		var joinDateStr, createdAtStr, updatedAtStr string
		if err := rows.Scan(&emp.ID, &emp.Name, &emp.BaseSalary, &emp.Allowance, &emp.Performance,
			&joinDateStr, &createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		emp.JoinDate, _ = time.Parse(time.RFC3339, joinDateStr)
		emp.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		emp.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		employees = append(employees, emp)
	}

	return employees, rows.Err()
}

func UpdateEmployee(db *sql.DB, emp *Employee) error {
	now := time.Now().UTC()
	emp.UpdatedAt = now

	result, err := db.Exec(
		`UPDATE employees SET name=?, base_salary=?, allowance=?, performance=?, join_date=?, updated_at=?
		 WHERE id=?`,
		emp.Name, emp.BaseSalary, emp.Allowance, emp.Performance,
		emp.JoinDate.Format(time.RFC3339),
		emp.UpdatedAt.Format(time.RFC3339),
		emp.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("员工不存在")
	}

	return nil
}

func DeleteEmployee(db *sql.DB, id int64) error {
	result, err := db.Exec("DELETE FROM employees WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("员工不存在")
	}

	return nil
}

func CreateResource(db *sql.DB, res *Resource) error {
	now := time.Now().UTC()
	res.CreatedAt = now
	res.UpdatedAt = now

	result, err := db.Exec(
		`INSERT INTO resources (name, type, description, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)`,
		res.Name, res.Type, res.Description,
		res.CreatedAt.Format(time.RFC3339),
		res.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	res.ID = id
	return nil
}

func GetResource(db *sql.DB, id int64) (*Resource, error) {
	res := &Resource{}
	var createdAtStr, updatedAtStr string

	err := db.QueryRow(
		`SELECT id, name, type, description, created_at, updated_at
		 FROM resources WHERE id = ?`,
		id,
	).Scan(&res.ID, &res.Name, &res.Type, &res.Description, &createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("资源不存在")
		}
		return nil, err
	}

	res.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	res.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return res, nil
}

func ListResources(db *sql.DB, resourceType string) ([]Resource, error) {
	var rows *sql.Rows
	var err error

	if resourceType != "" {
		rows, err = db.Query(
			`SELECT id, name, type, description, created_at, updated_at
			 FROM resources WHERE type = ? ORDER BY id DESC`,
			resourceType,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, name, type, description, created_at, updated_at
			 FROM resources ORDER BY id DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		res := Resource{}
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&res.ID, &res.Name, &res.Type, &res.Description,
			&createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		res.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		res.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		resources = append(resources, res)
	}

	return resources, rows.Err()
}

func UpdateResource(db *sql.DB, res *Resource) error {
	now := time.Now().UTC()
	res.UpdatedAt = now

	result, err := db.Exec(
		`UPDATE resources SET name=?, type=?, description=?, updated_at=?
		 WHERE id=?`,
		res.Name, res.Type, res.Description,
		res.UpdatedAt.Format(time.RFC3339),
		res.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("资源不存在")
	}

	return nil
}

func DeleteResource(db *sql.DB, id int64) error {
	result, err := db.Exec("DELETE FROM resources WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("资源不存在")
	}

	return nil
}

func CreateResourceRelation(db *sql.DB, rel *ResourceRelation) error {
	now := time.Now().UTC()
	rel.CreatedAt = now

	result, err := db.Exec(
		`INSERT INTO resource_relations (source_id, source_type, target_id, target_type, relation, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		rel.SourceID, rel.SourceType, rel.TargetID, rel.TargetType, rel.Relation,
		rel.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rel.ID = id
	return nil
}

func ListResourceRelations(db *sql.DB, resourceID int64, resourceType string) ([]ResourceRelation, error) {
	rows, err := db.Query(
		`SELECT id, source_id, source_type, target_id, target_type, relation, created_at
		 FROM resource_relations
		 WHERE (source_id = ? AND source_type = ?) OR (target_id = ? AND target_type = ?)
		 ORDER BY id DESC`,
		resourceID, resourceType, resourceID, resourceType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []ResourceRelation
	for rows.Next() {
		rel := ResourceRelation{}
		var createdAtStr string
		if err := rows.Scan(&rel.ID, &rel.SourceID, &rel.SourceType, &rel.TargetID,
			&rel.TargetType, &rel.Relation, &createdAtStr); err != nil {
			return nil, err
		}
		rel.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		relations = append(relations, rel)
	}

	return relations, rows.Err()
}

func DeleteResourceRelation(db *sql.DB, id int64) error {
	result, err := db.Exec("DELETE FROM resource_relations WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("资源关联不存在")
	}

	return nil
}

func CreateSalaryRecord(db *sql.DB, record *SalaryRecord) error {
	now := time.Now().UTC()
	record.CreatedAt = now
	record.UpdatedAt = now
	record.Status = StatusNew

	result, err := db.Exec(
		`INSERT INTO salaries (
			employee_id, year, month, base_salary, allowance, performance,
			monthly_salary, hourly_rate, overtime_workday, overtime_weekend, overtime_holiday,
			total_overtime, social_insurance, taxable_income, income_tax, net_salary,
			resource_id, resource_type, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.EmployeeID, record.Year, record.Month, record.BaseSalary, record.Allowance, record.Performance,
		record.MonthlySalary, record.HourlyRate, record.OvertimeWorkday, record.OvertimeWeekend, record.OvertimeHoliday,
		record.TotalOvertime, record.SocialInsurance, record.TaxableIncome, record.IncomeTax, record.NetSalary,
		record.ResourceID, record.ResourceType, string(record.Status),
		record.CreatedAt.Format(time.RFC3339), record.UpdatedAt.Format(time.RFC3339),
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			return fmt.Errorf("该员工该月薪资记录已存在")
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	record.ID = id
	return nil
}

func GetSalaryRecord(db *sql.DB, id int64) (*SalaryRecord, error) {
	record := &SalaryRecord{}
	var createdAtStr, updatedAtStr string

	err := db.QueryRow(
		`SELECT id, employee_id, year, month, base_salary, allowance, performance,
		 monthly_salary, hourly_rate, overtime_workday, overtime_weekend, overtime_holiday,
		 total_overtime, social_insurance, taxable_income, income_tax, net_salary,
		 resource_id, resource_type, status, created_at, updated_at
		 FROM salaries WHERE id = ?`,
		id,
	).Scan(&record.ID, &record.EmployeeID, &record.Year, &record.Month,
		&record.BaseSalary, &record.Allowance, &record.Performance,
		&record.MonthlySalary, &record.HourlyRate,
		&record.OvertimeWorkday, &record.OvertimeWeekend, &record.OvertimeHoliday,
		&record.TotalOvertime, &record.SocialInsurance, &record.TaxableIncome,
		&record.IncomeTax, &record.NetSalary,
		&record.ResourceID, &record.ResourceType, &record.Status,
		&createdAtStr, &updatedAtStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("薪资记录不存在")
		}
		return nil, err
	}

	record.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	record.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)

	return record, nil
}

func ListSalaryRecords(db *sql.DB, employeeID int64, year, month int) ([]SalaryRecord, error) {
	query := `SELECT id, employee_id, year, month, base_salary, allowance, performance,
	 monthly_salary, hourly_rate, overtime_workday, overtime_weekend, overtime_holiday,
	 total_overtime, social_insurance, taxable_income, income_tax, net_salary,
	 resource_id, resource_type, status, created_at, updated_at
	 FROM salaries WHERE 1=1`

	args := []interface{}{}

	if employeeID > 0 {
		query += " AND employee_id = ?"
		args = append(args, employeeID)
	}
	if year > 0 {
		query += " AND year = ?"
		args = append(args, year)
	}
	if month > 0 {
		query += " AND month = ?"
		args = append(args, month)
	}

	query += " ORDER BY id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []SalaryRecord
	for rows.Next() {
		record := SalaryRecord{}
		var createdAtStr, updatedAtStr string
		if err := rows.Scan(&record.ID, &record.EmployeeID, &record.Year, &record.Month,
			&record.BaseSalary, &record.Allowance, &record.Performance,
			&record.MonthlySalary, &record.HourlyRate,
			&record.OvertimeWorkday, &record.OvertimeWeekend, &record.OvertimeHoliday,
			&record.TotalOvertime, &record.SocialInsurance, &record.TaxableIncome,
			&record.IncomeTax, &record.NetSalary,
			&record.ResourceID, &record.ResourceType, &record.Status,
			&createdAtStr, &updatedAtStr); err != nil {
			return nil, err
		}
		record.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		record.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAtStr)
		records = append(records, record)
	}

	return records, rows.Err()
}

func UpdateSalaryStatus(db *sql.DB, id int64) (*SalaryRecord, error) {
	record, err := GetSalaryRecord(db, id)
	if err != nil {
		return nil, err
	}

	nextStatus, err := record.Status.Next()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	record.UpdatedAt = now
	record.Status = nextStatus

	_, err = db.Exec(
		`UPDATE salaries SET status = ?, updated_at = ? WHERE id = ?`,
		string(nextStatus), now.Format(time.RFC3339), id,
	)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func GetResourceSummary(db *sql.DB, resourceID int64, resourceType string) (*ResourceSummary, error) {
	resource, err := GetResource(db, resourceID)
	if err != nil {
		return nil, err
	}

	summary := &ResourceSummary{
		ResourceID:   resource.ID,
		ResourceName: resource.Name,
		ResourceType: resource.Type,
	}

	err = db.QueryRow(
		`SELECT COALESCE(SUM(net_salary), 0), COUNT(*)
		 FROM salaries WHERE resource_id = ? AND resource_type = ?`,
		resourceID, resourceType,
	).Scan(&summary.TotalSalary, &summary.RecordCount)
	if err != nil {
		return nil, err
	}

	return summary, nil
}

func ListAllResourceSummaries(db *sql.DB, resourceType string) ([]ResourceSummary, error) {
	query := `SELECT r.id, r.name, r.type, 
	 COALESCE(SUM(s.net_salary), 0), COUNT(s.id)
	 FROM resources r
	 LEFT JOIN salaries s ON s.resource_id = r.id AND s.resource_type = r.type
	 WHERE 1=1`

	args := []interface{}{}

	if resourceType != "" {
		query += " AND r.type = ?"
		args = append(args, resourceType)
	}

	query += " GROUP BY r.id, r.name, r.type ORDER BY r.id DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []ResourceSummary
	for rows.Next() {
		summary := ResourceSummary{}
		if err := rows.Scan(&summary.ResourceID, &summary.ResourceName, &summary.ResourceType,
			&summary.TotalSalary, &summary.RecordCount); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}
