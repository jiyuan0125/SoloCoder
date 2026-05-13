package repository

import (
	"database/sql"
	"math/rand"
	"strings"
	"time"

	"points-mall/db"
	"points-mall/models"
)

func GenerateRedemptionCode(exchangeID int, productID int, userID int, expiresAt time.Time) (*models.RedemptionCode, error) {
	for i := 0; i < 10; i++ {
		code := generateCode()
		exists, err := codeExists(code)
		if err != nil {
			return nil, err
		}
		if !exists {
			return createRedemptionCode(code, exchangeID, productID, userID, expiresAt)
		}
	}
	return nil, sql.ErrNoRows
}

func generateCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < 16; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}
	return sb.String()
}

func codeExists(code string) (bool, error) {
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM redemption_codes WHERE code = ?", code).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func createRedemptionCode(code string, exchangeID int, productID int, userID int, expiresAt time.Time) (*models.RedemptionCode, error) {
	_, err := db.DB.Exec(
		"INSERT INTO redemption_codes (code, exchange_id, product_id, user_id, expires_at) VALUES (?, ?, ?, ?, ?)",
		code, exchangeID, productID, userID, expiresAt.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		return nil, err
	}
	return GetRedemptionCodeByCode(code)
}

func GetRedemptionCodeByCode(code string) (*models.RedemptionCode, error) {
	rc := &models.RedemptionCode{}
	var redeemedAt sql.NullString
	err := db.DB.QueryRow(
		"SELECT id, code, exchange_id, product_id, user_id, is_used, expires_at, redeemed_at, created_at FROM redemption_codes WHERE code = ?",
		code,
	).Scan(&rc.ID, &rc.Code, &rc.ExchangeID, &rc.ProductID, &rc.UserID, &rc.IsUsed, &rc.ExpiresAt, &redeemedAt, &rc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if redeemedAt.Valid {
		t, _ := time.Parse("2006-01-02 15:04:05", redeemedAt.String)
		rc.RedeemedAt = &t
	}
	return rc, nil
}

func MarkCodeAsUsed(code string) error {
	_, err := db.DB.Exec(
		"UPDATE redemption_codes SET is_used = 1, redeemed_at = CURRENT_TIMESTAMP WHERE code = ?",
		code,
	)
	return err
}

func GetUserRedemptionCodes(userID int) ([]*models.RedemptionCode, error) {
	rows, err := db.DB.Query(
		"SELECT id, code, exchange_id, product_id, user_id, is_used, expires_at, redeemed_at, created_at FROM redemption_codes WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []*models.RedemptionCode
	for rows.Next() {
		rc := &models.RedemptionCode{}
		var redeemedAt sql.NullString
		err := rows.Scan(&rc.ID, &rc.Code, &rc.ExchangeID, &rc.ProductID, &rc.UserID, &rc.IsUsed, &rc.ExpiresAt, &redeemedAt, &rc.CreatedAt)
		if err != nil {
			return nil, err
		}
		if redeemedAt.Valid {
			t, _ := time.Parse("2006-01-02 15:04:05", redeemedAt.String)
			rc.RedeemedAt = &t
		}
		codes = append(codes, rc)
	}
	return codes, nil
}
