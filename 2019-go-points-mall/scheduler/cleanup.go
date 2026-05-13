package scheduler

import (
	"log"
	"time"

	"points-mall/db"
	"points-mall/models"
	"points-mall/repository"
)

func StartExpiredLockCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			cleanupExpiredLocks()
		}
	}()
	log.Println("过期锁清理任务已启动")
}

func cleanupExpiredLocks() {
	exchanges, err := repository.GetExpiredLocks()
	if err != nil {
		log.Printf("获取过期锁失败: %v", err)
		return
	}

	for _, exchange := range exchanges {
		cleanupSingleExchange(exchange)
	}
}

func cleanupSingleExchange(exchange *models.Exchange) {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("开始事务失败: %v", err)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		"UPDATE exchanges SET status = ?, lock_expires_at = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = ? AND lock_expires_at < datetime('now')",
		"expired", exchange.ID, models.StatusPendingPay,
	)
	if err != nil {
		log.Printf("更新兑换状态失败: %v", err)
		return
	}
	rows, err := result.RowsAffected()
	if err != nil || rows == 0 {
		return
	}

	_, err = tx.Exec(
		"INSERT INTO exchange_details (exchange_id, action_type, old_status, new_status, description) VALUES (?, ?, ?, ?, ?)",
		exchange.ID, "expire", models.StatusPendingPay, "expired", "库存锁定超时，自动取消",
	)
	if err != nil {
		log.Printf("添加明细失败: %v", err)
		return
	}

	var isOnline int
	err = tx.QueryRow("SELECT is_online FROM products WHERE id = ?", exchange.ProductID).Scan(&isOnline)
	if err != nil {
		log.Printf("查询商品状态失败: %v", err)
		return
	}

	if isOnline == 1 {
		_, err = tx.Exec(
			"UPDATE products SET stock = stock + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			exchange.ProductID,
		)
		if err != nil {
			log.Printf("释放库存失败: %v", err)
			return
		}
	} else {
		_, err = tx.Exec(
			"UPDATE products SET stock = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			exchange.ProductID,
		)
		if err != nil {
			log.Printf("设置库存为零失败: %v", err)
			return
		}
	}

	_, err = tx.Exec(
		"UPDATE users SET points_balance = points_balance + ? WHERE id = ?",
		exchange.Points, exchange.UserID,
	)
	if err != nil {
		log.Printf("退还积分失败: %v", err)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("提交事务失败: %v", err)
		return
	}

	log.Printf("已释放兑换记录 %d 的锁定库存和积分", exchange.ID)
}
