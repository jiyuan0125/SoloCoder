package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"points-mall/db"
	"points-mall/models"
	"points-mall/repository"
)

type StartExchangeRequest struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
}

func StartExchange(w http.ResponseWriter, r *http.Request) {
	var req StartExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	product, err := repository.GetProductByID(req.ProductID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if product == nil {
		RespondError(w, http.StatusNotFound, "商品不存在")
		return
	}
	if !product.IsOnline {
		RespondError(w, http.StatusBadRequest, "商品已下架")
		return
	}

	user, err := repository.GetUserByID(req.UserID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		RespondError(w, http.StatusNotFound, "用户不存在")
		return
	}

	if user.PointsBalance < product.Points {
		RespondError(w, http.StatusBadRequest,
			fmt.Sprintf("积分余额不足，当前余额:%d，所需积分:%d", user.PointsBalance, product.Points))
		return
	}

	hasRedeemed, err := repository.CheckDailyRedemption(req.UserID, req.ProductID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if hasRedeemed {
		RespondError(w, http.StatusTooManyRequests, "今日已兑换")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"UPDATE products SET stock = stock - 1, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND stock >= 1",
		req.ProductID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		RespondError(w, http.StatusBadRequest, "库存不足")
		return
	}

	result, err = tx.Exec(
		"UPDATE users SET points_balance = points_balance - ? WHERE id = ? AND points_balance >= ?",
		product.Points, req.UserID, product.Points,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err = result.RowsAffected()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		RespondError(w, http.StatusBadRequest,
			fmt.Sprintf("积分余额不足，当前余额:%d，所需积分:%d", user.PointsBalance, product.Points))
		return
	}

	lockExpiresAt := time.Now().Add(15 * time.Minute)
	result, err = tx.Exec(
		"INSERT INTO exchanges (user_id, product_id, status, points, lock_expires_at) VALUES (?, ?, ?, ?, ?)",
		req.UserID, req.ProductID, models.StatusPendingPay, product.Points,
		lockExpiresAt.Format("2006-01-02 15:04:05"),
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	exchangeID, err := result.LastInsertId()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = tx.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		int(exchangeID), "start_pay", models.StatusNew, models.StatusPendingPay, "开始兑换，锁定库存和积分",
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	exchange, _ := repository.GetExchangeByID(int(exchangeID))
	RespondJSON(w, http.StatusCreated, exchange)
}

func ConfirmExchange(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/exchanges/")
	parts := strings.Split(idStr, "/")
	if len(parts) < 2 || parts[1] != "confirm" {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	exchangeID, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的兑换ID")
		return
	}

	exchange, err := repository.GetExchangeByID(exchangeID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exchange == nil {
		RespondError(w, http.StatusNotFound, "兑换记录不存在")
		return
	}

	if exchange.Status != models.StatusPendingPay {
		RespondError(w, http.StatusBadRequest, "当前状态不允许确认兑换")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"UPDATE exchanges SET status = ?, lock_expires_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?",
		models.StatusRedeemed, exchangeID, models.StatusPendingPay,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		RespondError(w, http.StatusBadRequest, "当前状态不允许确认兑换")
		return
	}

	_, err = tx.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		exchangeID, "confirm", models.StatusPendingPay, models.StatusRedeemed, "确认兑换完成",
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = repository.AddDailyRedemption(exchange.UserID, exchange.ProductID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	code, err := generateRedemptionCodeTx(tx, exchangeID, exchange.ProductID, exchange.UserID, expiresAt)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"exchange_id": exchangeID,
		"code":        code,
		"expires_at":  expiresAt,
	})
}

func generateRedemptionCodeTx(tx *sql.Tx, exchangeID int, productID int, userID int, expiresAt time.Time) (string, error) {
	for i := 0; i < 10; i++ {
		code := generateCode()
		var count int
		err := tx.QueryRow("SELECT COUNT(*) FROM redemption_codes WHERE code = ?", code).Scan(&count)
		if err != nil {
			return "", err
		}
		if count == 0 {
			_, err := tx.Exec(
				"INSERT INTO redemption_codes (code, exchange_id, product_id, user_id, expires_at) VALUES (?, ?, ?, ?, ?)",
				code, exchangeID, productID, userID, expiresAt.Format("2006-01-02 15:04:05"),
			)
			if err != nil {
				return "", err
			}
			return code, nil
		}
	}
	return "", fmt.Errorf("生成兑换码失败")
}

func generateCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < 16; i++ {
		sb.WriteByte(charset[time.Now().UnixNano()%int64(len(charset))])
		time.Sleep(1 * time.Nanosecond)
	}
	return sb.String()
}

func CancelExchange(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/exchanges/")
	parts := strings.Split(idStr, "/")
	if len(parts) < 2 || parts[1] != "cancel" {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	exchangeID, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的兑换ID")
		return
	}

	exchange, err := repository.GetExchangeByID(exchangeID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exchange == nil {
		RespondError(w, http.StatusNotFound, "兑换记录不存在")
		return
	}

	if exchange.Status != models.StatusPendingPay {
		RespondError(w, http.StatusBadRequest, "当前状态不允许取消")
		return
	}

	product, err := repository.GetProductByID(exchange.ProductID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"UPDATE exchanges SET status = ?, lock_expires_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?",
		"cancelled", exchangeID, models.StatusPendingPay,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		RespondError(w, http.StatusBadRequest, "当前状态不允许取消")
		return
	}

	_, err = tx.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		exchangeID, "cancel", models.StatusPendingPay, "cancelled", "用户取消兑换",
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if product != nil && product.IsOnline {
		_, err = tx.Exec(
			"UPDATE products SET stock = stock + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			exchange.ProductID,
		)
		if err != nil {
			RespondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	_, err = tx.Exec(
		"UPDATE users SET points_balance = points_balance + ? WHERE id = ?",
		exchange.Points, exchange.UserID,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "取消成功"})
}

func GetExchange(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/exchanges/")
	parts := strings.Split(path, "/")
	if len(parts) != 1 {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的兑换ID")
		return
	}

	exchange, err := repository.GetExchangeByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exchange == nil {
		RespondError(w, http.StatusNotFound, "兑换记录不存在")
		return
	}

	RespondJSON(w, http.StatusOK, exchange)
}

func GetExchangeDetails(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/exchanges/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "details" {
		RespondError(w, http.StatusNotFound, "路由不存在")
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的兑换ID")
		return
	}

	exchange, err := repository.GetExchangeByID(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exchange == nil {
		RespondError(w, http.StatusNotFound, "兑换记录不存在")
		return
	}

	details, err := repository.GetExchangeDetails(id)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, details)
}

func GetUserExchangeHistory(w http.ResponseWriter, r *http.Request) {
	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userIDStr = strings.TrimSuffix(userIDStr, "/exchanges")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		RespondError(w, http.StatusNotFound, "用户不存在")
		return
	}

	exchanges, err := repository.GetUserExchanges(userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, exchanges)
}

func GetUserCodes(w http.ResponseWriter, r *http.Request) {
	userIDStr := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userIDStr = strings.TrimSuffix(userIDStr, "/codes")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "无效的用户ID")
		return
	}

	user, err := repository.GetUserByID(userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		RespondError(w, http.StatusNotFound, "用户不存在")
		return
	}

	codes, err := repository.GetUserRedemptionCodes(userID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, codes)
}

func RedeemCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "无效的请求体")
		return
	}

	if req.Code == "" {
		RespondError(w, http.StatusBadRequest, "兑换码不能为空")
		return
	}

	code, err := repository.GetRedemptionCodeByCode(req.Code)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if code == nil {
		RespondError(w, http.StatusNotFound, "兑换码不存在")
		return
	}

	if time.Now().After(code.ExpiresAt) {
		RespondError(w, http.StatusBadRequest, "兑换码已过期")
		return
	}

	if code.IsUsed {
		RespondError(w, http.StatusBadRequest, "兑换码已使用")
		return
	}

	exchange, err := repository.GetExchangeByID(code.ExchangeID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exchange == nil {
		RespondError(w, http.StatusNotFound, "兑换记录不存在")
		return
	}

	if exchange.Status != models.StatusRedeemed {
		RespondError(w, http.StatusBadRequest, "当前状态不允许核销")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"UPDATE redemption_codes SET is_used = 1, redeemed_at = CURRENT_TIMESTAMP WHERE code = ? AND is_used = 0",
		req.Code,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	rows, err := result.RowsAffected()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if rows == 0 {
		RespondError(w, http.StatusBadRequest, "兑换码已使用")
		return
	}

	_, err = tx.Exec(
		"UPDATE exchanges SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ?",
		models.StatusUsed, code.ExchangeID, models.StatusRedeemed,
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = tx.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		code.ExchangeID, "redeem", models.StatusRedeemed, models.StatusUsed,
		fmt.Sprintf("兑换码核销: %s", req.Code),
	)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = tx.Commit(); err != nil {
		RespondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"message": "核销成功"})
}
