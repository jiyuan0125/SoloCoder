package service

import (
	"asset-depreciation/database"
	"asset-depreciation/models"
	"database/sql"
	"fmt"
	"time"
)

func CalculateMonthlyDepreciation(asset *models.Asset, month string) (int64, bool, error) {
	if asset.CurrentBookValue <= asset.SalvageValue {
		return 0, false, nil
	}

	depreciableValue := asset.OriginalValue - asset.SalvageValue
	if asset.AccumulatedDepr >= depreciableValue {
		return 0, false, nil
	}

	t, err := time.Parse("2006-01", month)
	if err != nil {
		return 0, false, err
	}

	isFirstMonth := false
	purchaseDate, err := time.Parse("2006-01-02", asset.PurchaseDate)
	if err != nil {
		return 0, false, err
	}

	purchaseYearMonth := purchaseDate.Format("2006-01")
	currentYearMonth := t.Format("2006-01")
	isFirstMonth = purchaseYearMonth == currentYearMonth

	var monthlyRate float64
	if asset.Method == models.MethodDoubleDeclining {
		straightLineRate := 1.0 / float64(asset.UsefulLifeMonths)
		monthlyRate = 2.0 * straightLineRate
	}

	var fullMonthAmount int64
	var isPartial bool

	if asset.Method == models.MethodStraightLine {
		remainingMonths := asset.UsefulLifeMonths - getMonthsDepreciated(asset, t)
		remainingDeprValue := asset.CurrentBookValue - asset.SalvageValue

		if remainingMonths <= 1 || remainingDeprValue <= 0 {
			fullMonthAmount = remainingDeprValue
		} else {
			fullMonthAmount = depreciableValue / int64(asset.UsefulLifeMonths)
		}
	} else {
		remainingDeprValue := asset.CurrentBookValue - asset.SalvageValue
		if remainingDeprValue <= 0 {
			return 0, false, nil
		}

		calcAmount := int64(float64(asset.CurrentBookValue) * monthlyRate)

		if calcAmount <= 0 || asset.CurrentBookValue-calcAmount <= asset.SalvageValue {
			fullMonthAmount = remainingDeprValue
		} else {
			fullMonthAmount = calcAmount
		}
	}

	if isFirstMonth {
		isPartial = true
		usedDays := 30 - asset.PurchaseDay + 1
		if usedDays < 1 {
			usedDays = 1
		}
		if usedDays > 30 {
			usedDays = 30
		}

		prorated := int64(float64(fullMonthAmount) * float64(usedDays) / 30.0)

		if asset.CurrentBookValue-prorated < asset.SalvageValue {
			prorated = asset.CurrentBookValue - asset.SalvageValue
		}
		return prorated, isPartial, nil
	}

	isLastMonth := isLastDepreciationMonth(asset, t)
	if isLastMonth {
		remaining := asset.CurrentBookValue - asset.SalvageValue
		if remaining > 0 {
			return remaining, false, nil
		}
		return 0, false, nil
	}

	if asset.Method == models.MethodDoubleDeclining {
		nextBookValue := asset.CurrentBookValue - fullMonthAmount
		nextMonths := asset.UsefulLifeMonths - getMonthsDepreciated(asset, t) - 1

		if nextBookValue < asset.SalvageValue {
			return asset.CurrentBookValue - asset.SalvageValue, isPartial, nil
		}

		if nextMonths <= 1 && nextBookValue > asset.SalvageValue {
			return fullMonthAmount, isPartial, nil
		}
	}

	return fullMonthAmount, isPartial, nil
}

func getMonthsDepreciated(asset *models.Asset, currentTime time.Time) int {
	if asset.LastDeprMonth == "" {
		return 0
	}

	lastTime, err := time.Parse("2006-01", asset.LastDeprMonth)
	if err != nil {
		return 0
	}

	purchaseTime, err := time.Parse("2006-01-02", asset.PurchaseDate)
	if err != nil {
		return 0
	}

	lastYearMonth := lastTime.Year()*12 + int(lastTime.Month())
	purchaseYearMonth := purchaseTime.Year()*12 + int(purchaseTime.Month())

	return lastYearMonth - purchaseYearMonth + 1
}

func isLastDepreciationMonth(asset *models.Asset, currentTime time.Time) bool {
	purchaseTime, err := time.Parse("2006-01-02", asset.PurchaseDate)
	if err != nil {
		return false
	}

	endTime := purchaseTime.AddDate(0, asset.UsefulLifeMonths-1, 0)

	currentYearMonth := currentTime.Year()*12 + int(currentTime.Month())
	endYearMonth := endTime.Year()*12 + int(endTime.Month())

	return currentYearMonth >= endYearMonth
}

func ProcessMonthlyDepreciation() error {
	assets, err := database.ListAssets()
	if err != nil {
		return err
	}

	now := time.Now()
	currentMonth := now.Format("2006-01")

	for _, asset := range assets {
		if asset.Status != models.StatusRunning && asset.Status != models.StatusApproved {
			continue
		}
		if asset.LastDeprMonth == currentMonth {
			continue
		}

		if err := ProcessAssetDepreciation(asset, currentMonth); err != nil {
			fmt.Printf("Error processing asset %d: %v\n", asset.ID, err)
			continue
		}
	}

	return nil
}

func ProcessAssetDepreciation(asset *models.Asset, month string) error {
	amount, isPartial, err := CalculateMonthlyDepreciation(asset, month)
	if err != nil {
		return err
	}

	if amount <= 0 {
		tx, err := database.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		if asset.Status == models.StatusRunning {
			if err := updateStatusInTx(tx, asset.ID, models.StatusRunning, models.StatusCompleted, "complete", "折旧完成"); err != nil {
				return err
			}
		}

		if err := database.UpdateAssetDepreciation(tx, asset.ID, asset.AccumulatedDepr, asset.CurrentBookValue, month); err != nil {
			return err
		}

		return tx.Commit()
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	newAccumulated := asset.AccumulatedDepr + amount
	newBookValue := asset.CurrentBookValue - amount

	record := &models.DepreciationRecord{
		AssetID:         asset.ID,
		DeprMonth:       month,
		DeprAmount:      amount,
		AccumulatedDepr: newAccumulated,
		BookValue:       newBookValue,
		IsPartialMonth:  isPartial,
		DaysInMonth:     30,
	}

	if err := database.InsertDepreciationRecord(tx, record); err != nil {
		return err
	}

	if err := database.UpdateAssetDepreciation(tx, asset.ID, newAccumulated, newBookValue, month); err != nil {
		return err
	}

	if newBookValue <= asset.SalvageValue && asset.Status == models.StatusRunning {
		if err := updateStatusInTx(tx, asset.ID, models.StatusRunning, models.StatusCompleted, "complete", "折旧完成"); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func updateStatusInTx(tx *sql.Tx, assetID int64, fromStatus, toStatus models.AssetStatus, action, remark string) error {
	query := `
		UPDATE assets 
		SET status = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND status = ?
	`
	_, err := tx.Exec(query, toStatus, assetID, fromStatus)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO operation_history (asset_id, from_status, to_status, action, remark)
		VALUES (?, ?, ?, ?, ?)
	`, assetID, fromStatus, toStatus, action, remark)
	return err
}
