package database

import (
	"database/sql"
	"restaurant-preorder/models"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone TEXT NOT NULL,
			dining_date TEXT NOT NULL,
			time_slot TEXT NOT NULL,
			guest_count INTEGER NOT NULL,
			status TEXT NOT NULL,
			resource_type TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_phone_date ON orders(phone, dining_date)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_date_slot ON orders(dining_date, time_slot)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_resource ON orders(resource_type, resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)`,
		`CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id INTEGER NOT NULL,
			dish_name TEXT NOT NULL,
			quantity INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id)`,
		`CREATE TABLE IF NOT EXISTS blacklist (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone TEXT NOT NULL UNIQUE,
			end_time DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreateOrder(order *models.Order, items []models.OrderItem) (int64, error) {
	tx, err := DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.Exec(`
		INSERT INTO orders (phone, dining_date, time_slot, guest_count, status, resource_type, resource_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, order.Phone, order.DiningDate, order.TimeSlot, order.GuestCount, order.Status, order.ResourceType, order.ResourceID, now, now)
	if err != nil {
		return 0, err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, dish_name, quantity, created_at)
			VALUES (?, ?, ?, ?)
		`, orderID, item.DishName, item.Quantity, now)
		if err != nil {
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return orderID, nil
}

func GetOrderByID(id int64) (*models.Order, error) {
	order := &models.Order{}
	err := DB.QueryRow(`
		SELECT id, phone, dining_date, time_slot, guest_count, status, resource_type, resource_id, created_at, updated_at
		FROM orders WHERE id = ?
	`, id).Scan(&order.ID, &order.Phone, &order.DiningDate, &order.TimeSlot, &order.GuestCount, &order.Status, &order.ResourceType, &order.ResourceID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func GetOrderWithItems(id int64) (*models.OrderWithItems, error) {
	order, err := GetOrderByID(id)
	if err != nil {
		return nil, err
	}

	items, err := GetOrderItems(id)
	if err != nil {
		return nil, err
	}

	return &models.OrderWithItems{
		Order: *order,
		Items: items,
	}, nil
}

func GetOrderItems(orderID int64) ([]models.OrderItem, error) {
	rows, err := DB.Query(`
		SELECT id, order_id, dish_name, quantity, created_at
		FROM order_items WHERE order_id = ?
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		item := models.OrderItem{}
		err := rows.Scan(&item.ID, &item.OrderID, &item.DishName, &item.Quantity, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func UpdateOrderStatus(id int64, status models.OrderStatus) error {
	_, err := DB.Exec(`
		UPDATE orders SET status = ?, updated_at = ? WHERE id = ?
	`, status, time.Now(), id)
	return err
}

func CountOrdersByDateAndSlot(date string, slot models.TimeSlot, excludeStatuses ...models.OrderStatus) (int, error) {
	query := `SELECT COUNT(*) FROM orders WHERE dining_date = ? AND time_slot = ?`
	args := []interface{}{date, slot}

	if len(excludeStatuses) > 0 {
		query += ` AND status NOT IN (`
		for i, s := range excludeStatuses {
			if i > 0 {
				query += `,`
			}
			query += `?`
			args = append(args, s)
		}
		query += `)`
	}

	var count int
	err := DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

func HasOrderOnDate(phone string, date string) (bool, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM orders 
		WHERE phone = ? AND dining_date = ? AND status NOT IN (?, ?)
	`, phone, date, models.StatusCancelled, models.StatusNoShow).Scan(&count)
	return count > 0, err
}

func CountNoShows(phone string) (int, error) {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM orders 
		WHERE phone = ? AND status = ? AND updated_at >= ?
	`, phone, models.StatusNoShow, thirtyDaysAgo).Scan(&count)
	return count, err
}

func IsBlacklisted(phone string) (bool, error) {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM blacklist 
		WHERE phone = ? AND end_time > ?
	`, phone, time.Now()).Scan(&count)
	return count > 0, err
}

func AddToBlacklist(phone string, endTime time.Time) error {
	_, err := DB.Exec(`
		INSERT INTO blacklist (phone, end_time)
		VALUES (?, ?)
		ON CONFLICT(phone) DO UPDATE SET end_time = ?
	`, phone, endTime, endTime)
	return err
}

func GetOrdersForNoShowCheck(slotEndTime time.Time, gracePeriod time.Duration) ([]models.Order, error) {
	cutoff := slotEndTime.Add(-gracePeriod)
	rows, err := DB.Query(`
		SELECT id, phone, dining_date, time_slot, guest_count, status, resource_type, resource_id, created_at, updated_at
		FROM orders WHERE status = ?
	`, models.StatusConfirmed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		order := models.Order{}
		err := rows.Scan(&order.ID, &order.Phone, &order.DiningDate, &order.TimeSlot, &order.GuestCount, &order.Status, &order.ResourceType, &order.ResourceID, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}

		orderSlotEnd, err := GetSlotEndTime(order.DiningDate, order.TimeSlot)
		if err != nil {
			continue
		}

		if orderSlotEnd.Before(cutoff) {
			orders = append(orders, order)
		}
	}
	return orders, nil
}

func GetSlotEndTime(date string, slot models.TimeSlot) (time.Time, error) {
	layout := "2006-01-02 15:04"
	var timeStr string
	if slot == models.Lunch {
		timeStr = date + " 14:00"
	} else {
		timeStr = date + " 21:00"
	}
	return time.Parse(layout, timeStr)
}

func GetKitchenSummary(date string) (*models.KitchenSummary, *models.KitchenSummary, error) {
	lunchSummary := &models.KitchenSummary{
		Date:      date,
		TimeSlot:  models.Lunch,
		Dishes:    make(map[string]int),
		GeneratedAt: time.Now(),
	}
	dinnerSummary := &models.KitchenSummary{
		Date:      date,
		TimeSlot:  models.Dinner,
		Dishes:    make(map[string]int),
		GeneratedAt: time.Now(),
	}

	rows, err := DB.Query(`
		SELECT o.time_slot, oi.dish_name, SUM(oi.quantity)
		FROM orders o
		JOIN order_items oi ON o.id = oi.order_id
		WHERE o.dining_date = ? AND o.status NOT IN (?, ?)
		GROUP BY o.time_slot, oi.dish_name
	`, date, models.StatusCancelled, models.StatusNoShow)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var slot models.TimeSlot
		var dishName string
		var quantity int
		err := rows.Scan(&slot, &dishName, &quantity)
		if err != nil {
			return nil, nil, err
		}

		if slot == models.Lunch {
			lunchSummary.Dishes[dishName] += quantity
		} else {
			dinnerSummary.Dishes[dishName] += quantity
		}
	}

	return lunchSummary, dinnerSummary, nil
}

func GetResourceSummary(resourceType models.ResourceType, resourceID string) (*models.ResourceSummary, error) {
	summary := &models.ResourceSummary{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Orders:       []models.Order{},
	}

	rows, err := DB.Query(`
		SELECT id, phone, dining_date, time_slot, guest_count, status, resource_type, resource_id, created_at, updated_at
		FROM orders 
		WHERE resource_type = ? AND resource_id = ?
		ORDER BY created_at DESC
	`, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		order := models.Order{}
		err := rows.Scan(&order.ID, &order.Phone, &order.DiningDate, &order.TimeSlot, &order.GuestCount, &order.Status, &order.ResourceType, &order.ResourceID, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		summary.Orders = append(summary.Orders, order)
	}

	summary.TotalOrders = len(summary.Orders)
	return summary, nil
}

func GetAllOrders() ([]models.Order, error) {
	rows, err := DB.Query(`
		SELECT id, phone, dining_date, time_slot, guest_count, status, resource_type, resource_id, created_at, updated_at
		FROM orders ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		order := models.Order{}
		err := rows.Scan(&order.ID, &order.Phone, &order.DiningDate, &order.TimeSlot, &order.GuestCount, &order.Status, &order.ResourceType, &order.ResourceID, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
