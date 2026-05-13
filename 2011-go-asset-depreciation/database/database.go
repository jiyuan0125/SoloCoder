package database

import (
	"asset-depreciation/models"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	return createTables()
}

func createTables() error {
	sqlStatements := `
	CREATE TABLE IF NOT EXISTS assets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		original_value INTEGER NOT NULL,
		salvage_rate REAL NOT NULL DEFAULT 0.05,
		useful_life_months INTEGER NOT NULL,
		purchase_date TEXT NOT NULL,
		purchase_day INTEGER NOT NULL,
		department TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'draft',
		method TEXT NOT NULL DEFAULT 'straight_line',
		salvage_value INTEGER NOT NULL,
		accumulated_depr INTEGER NOT NULL DEFAULT 0,
		current_book_value INTEGER NOT NULL,
		last_depr_month TEXT NOT NULL DEFAULT '',
		scrapped_date TEXT,
		scrapped_value INTEGER,
		disposal_gain_loss INTEGER,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS operation_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		asset_id INTEGER NOT NULL,
		from_status TEXT NOT NULL,
		to_status TEXT NOT NULL,
		action TEXT NOT NULL,
		remark TEXT NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (asset_id) REFERENCES assets(id)
	);

	CREATE TABLE IF NOT EXISTS depreciation_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		asset_id INTEGER NOT NULL,
		depr_month TEXT NOT NULL,
		depr_amount INTEGER NOT NULL,
		accumulated_depr INTEGER NOT NULL,
		book_value INTEGER NOT NULL,
		is_partial_month INTEGER NOT NULL DEFAULT 0,
		days_in_month INTEGER NOT NULL DEFAULT 30,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (asset_id) REFERENCES assets(id),
		UNIQUE(asset_id, depr_month)
	);

	CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
	CREATE INDEX IF NOT EXISTS idx_assets_department ON assets(department);
	CREATE INDEX IF NOT EXISTS idx_operation_asset ON operation_history(asset_id);
	CREATE INDEX IF NOT EXISTS idx_depr_asset ON depreciation_records(asset_id);
	`

	_, err := DB.Exec(sqlStatements)
	return err
}

func CreateAsset(req *models.CreateAssetRequest) (*models.Asset, error) {
	salvageValue := int64(float64(req.OriginalValue) * req.SalvageRate)
	bookValue := req.OriginalValue

	t, err := time.Parse("2006-01-02", req.PurchaseDate)
	if err != nil {
		return nil, err
	}
	purchaseDay := t.Day()

	query := `
		INSERT INTO assets (
			name, original_value, salvage_rate, useful_life_months,
			purchase_date, purchase_day, department, status, method,
			salvage_value, current_book_value
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := DB.Exec(
		query,
		req.Name,
		req.OriginalValue,
		req.SalvageRate,
		req.UsefulLifeYears*12,
		req.PurchaseDate,
		purchaseDay,
		req.Department,
		models.StatusDraft,
		req.Method,
		salvageValue,
		bookValue,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetAssetByID(id)
}

func GetAssetByID(id int64) (*models.Asset, error) {
	query := `
		SELECT id, name, original_value, salvage_rate, useful_life_months,
		       purchase_date, purchase_day, department, status, method,
		       salvage_value, accumulated_depr, current_book_value, last_depr_month,
		       scrapped_date, scrapped_value, disposal_gain_loss, created_at, updated_at
		FROM assets WHERE id = ?
	`

	var asset models.Asset
	err := DB.QueryRow(query, id).Scan(
		&asset.ID,
		&asset.Name,
		&asset.OriginalValue,
		&asset.SalvageRate,
		&asset.UsefulLifeMonths,
		&asset.PurchaseDate,
		&asset.PurchaseDay,
		&asset.Department,
		&asset.Status,
		&asset.Method,
		&asset.SalvageValue,
		&asset.AccumulatedDepr,
		&asset.CurrentBookValue,
		&asset.LastDeprMonth,
		&asset.ScrappedDate,
		&asset.ScrappedValue,
		&asset.DisposalGainLoss,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func ListAssets() ([]*models.Asset, error) {
	query := `
		SELECT id, name, original_value, salvage_rate, useful_life_months,
		       purchase_date, purchase_day, department, status, method,
		       salvage_value, accumulated_depr, current_book_value, last_depr_month,
		       scrapped_date, scrapped_value, disposal_gain_loss, created_at, updated_at
		FROM assets ORDER BY id DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := []*models.Asset{}
	for rows.Next() {
		var asset models.Asset
		err := rows.Scan(
			&asset.ID,
			&asset.Name,
			&asset.OriginalValue,
			&asset.SalvageRate,
			&asset.UsefulLifeMonths,
			&asset.PurchaseDate,
			&asset.PurchaseDay,
			&asset.Department,
			&asset.Status,
			&asset.Method,
			&asset.SalvageValue,
			&asset.AccumulatedDepr,
			&asset.CurrentBookValue,
			&asset.LastDeprMonth,
			&asset.ScrappedDate,
			&asset.ScrappedValue,
			&asset.DisposalGainLoss,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}
	return assets, nil
}

func UpdateAssetStatus(assetID int64, fromStatus, toStatus models.AssetStatus, action, remark string) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE assets 
		SET status = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND status = ?
	`
	result, err := tx.Exec(query, toStatus, assetID, fromStatus)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	_, err = tx.Exec(`
		INSERT INTO operation_history (asset_id, from_status, to_status, action, remark)
		VALUES (?, ?, ?, ?, ?)
	`, assetID, fromStatus, toStatus, action, remark)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func AddRemarkToHistory(historyID int64, remark string) error {
	query := `
		UPDATE operation_history 
		SET remark = remark || ? 
		WHERE id = ?
	`
	separator := "\n"
	if remark != "" {
		separator = "\n" + remark
	}
	_, err := DB.Exec(query, separator, historyID)
	return err
}

func GetOperationHistory(assetID int64) ([]*models.OperationHistory, error) {
	query := `
		SELECT id, asset_id, from_status, to_status, action, remark, created_at
		FROM operation_history WHERE asset_id = ? ORDER BY id DESC
	`
	rows, err := DB.Query(query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	histories := []*models.OperationHistory{}
	for rows.Next() {
		var h models.OperationHistory
		err := rows.Scan(&h.ID, &h.AssetID, &h.FromStatus, &h.ToStatus, &h.Action, &h.Remark, &h.CreatedAt)
		if err != nil {
			return nil, err
		}
		histories = append(histories, &h)
	}
	return histories, nil
}

func GetOperationHistoryByID(historyID int64) (*models.OperationHistory, error) {
	query := `
		SELECT id, asset_id, from_status, to_status, action, remark, created_at
		FROM operation_history WHERE id = ?
	`
	var h models.OperationHistory
	err := DB.QueryRow(query, historyID).Scan(
		&h.ID, &h.AssetID, &h.FromStatus, &h.ToStatus, &h.Action, &h.Remark, &h.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func GetDepreciationRecords(assetID int64) ([]*models.DepreciationRecord, error) {
	query := `
		SELECT id, asset_id, depr_month, depr_amount, accumulated_depr,
		       book_value, is_partial_month, days_in_month, created_at
		FROM depreciation_records WHERE asset_id = ? ORDER BY depr_month ASC
	`
	rows, err := DB.Query(query, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []*models.DepreciationRecord{}
	for rows.Next() {
		var r models.DepreciationRecord
		var isPartialInt int
		err := rows.Scan(
			&r.ID, &r.AssetID, &r.DeprMonth, &r.DeprAmount, &r.AccumulatedDepr,
			&r.BookValue, &isPartialInt, &r.DaysInMonth, &r.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		r.IsPartialMonth = isPartialInt != 0
		records = append(records, &r)
	}
	return records, nil
}

func GetDepartmentSummary() ([]*models.DepartmentSummary, error) {
	query := `
		SELECT 
			department,
			COUNT(*) as asset_count,
			SUM(accumulated_depr) as total_depr
		FROM assets 
		WHERE status != 'scrapped'
		GROUP BY department
		ORDER BY department
	`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := []*models.DepartmentSummary{}
	for rows.Next() {
		var s models.DepartmentSummary
		err := rows.Scan(&s.Department, &s.AssetCount, &s.CurrentDeprTotal)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, &s)
	}
	return summaries, nil
}

func InsertDepreciationRecord(tx *sql.Tx, record *models.DepreciationRecord) error {
	query := `
		INSERT INTO depreciation_records (
			asset_id, depr_month, depr_amount, accumulated_depr,
			book_value, is_partial_month, days_in_month
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	isPartialInt := 0
	if record.IsPartialMonth {
		isPartialInt = 1
	}
	_, err := tx.Exec(
		query,
		record.AssetID,
		record.DeprMonth,
		record.DeprAmount,
		record.AccumulatedDepr,
		record.BookValue,
		isPartialInt,
		record.DaysInMonth,
	)
	return err
}

func UpdateAssetDepreciation(tx *sql.Tx, assetID int64, accumulatedDepr, bookValue int64, lastDeprMonth string) error {
	query := `
		UPDATE assets 
		SET accumulated_depr = ?, current_book_value = ?, last_depr_month = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := tx.Exec(query, accumulatedDepr, bookValue, lastDeprMonth, assetID)
	return err
}

func ScrapAsset(assetID int64, scrappedValue int64, remark string) (*models.Asset, error) {
	tx, err := DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	asset, err := GetAssetByID(assetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, nil
	}

	gainLoss := scrappedValue - asset.CurrentBookValue
	now := time.Now().Format("2006-01-02")

	_, err = tx.Exec(`
		UPDATE assets 
		SET status = ?, scrapped_date = ?, scrapped_value = ?, disposal_gain_loss = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, models.StatusScrapped, now, scrappedValue, gainLoss, assetID)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		INSERT INTO operation_history (asset_id, from_status, to_status, action, remark)
		VALUES (?, ?, ?, 'scrap', ?)
	`, assetID, asset.Status, models.StatusScrapped, remark)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetAssetByID(assetID)
}
