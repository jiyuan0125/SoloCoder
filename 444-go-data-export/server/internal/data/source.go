package data

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"
)

type UserBehavior struct {
	UserID     string    `json:"user_id"`
	UserName   string    `json:"user_name"`
	Phone      string    `json:"phone"`
	IDCard     string    `json:"idcard"`
	Action     string    `json:"action"`
	Page       string    `json:"page"`
	Timestamp  time.Time `json:"timestamp"`
	DurationMs int       `json:"duration_ms"`
}

type Transaction struct {
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	UserName      string    `json:"user_name"`
	Phone         string    `json:"phone"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
}

type FeatureUsage struct {
	FeatureID   string    `json:"feature_id"`
	FeatureName string    `json:"feature_name"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Phone       string    `json:"phone"`
	UsageCount  int       `json:"usage_count"`
	LastUsedAt  time.Time `json:"last_used_at"`
}

type DataSource struct {
	userBehaviors []UserBehavior
	transactions  []Transaction
	featureUsages []FeatureUsage
	mu            sync.RWMutex
}

func NewDataSource() *DataSource {
	ds := &DataSource{}
	ds.generateMockData()
	return ds
}

func randomPhone() string {
	phones := []string{
		"13800138001", "13900139002", "13600136003",
		"13700137004", "15000150005", "15100151006",
		"18600186007", "18700187008", "18800188009",
		"18900189010",
	}
	return phones[rand.Intn(len(phones))]
}

func (ds *DataSource) generateMockData() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	now := time.Now()
	users := []struct {
		ID     string
		Name   string
		Phone  string
		IDCard string
	}{
		{"u001", "张三", "13800138001", "110101199001011234"},
		{"u002", "李四", "13900139002", "110101199002022345"},
		{"u003", "王五", "13600136003", "110101199003033456"},
		{"u004", "赵六", "13700137004", "110101199004044567"},
		{"u005", "钱七", "15000150005", "110101199005055678"},
	}

	actions := []string{"login", "view", "click", "purchase", "logout"}
	pages := []string{"home", "product", "cart", "checkout", "profile"}
	statuses := []string{"success", "pending", "failed", "refunded"}
	paymentMethods := []string{"alipay", "wechat", "credit_card", "bank_transfer"}
	features := []struct {
		ID   string
		Name string
	}{
		{"f001", "搜索功能"},
		{"f002", "购物车"},
		{"f003", "收藏"},
		{"f004", "分享"},
		{"f005", "优惠券"},
	}

	for i := 0; i < 1000; i++ {
		user := users[rand.Intn(len(users))]
		ds.userBehaviors = append(ds.userBehaviors, UserBehavior{
			UserID:     user.ID,
			UserName:   user.Name,
			Phone:      user.Phone,
			IDCard:     user.IDCard,
			Action:     actions[rand.Intn(len(actions))],
			Page:       pages[rand.Intn(len(pages))],
			Timestamp:  now.Add(-time.Duration(rand.Intn(30*24)) * time.Hour),
			DurationMs: rand.Intn(10000),
		})
	}

	for i := 0; i < 500; i++ {
		user := users[rand.Intn(len(users))]
		ds.transactions = append(ds.transactions, Transaction{
			TransactionID: "txn_" + string(rune(10000+i)),
			UserID:        user.ID,
			UserName:      user.Name,
			Phone:         user.Phone,
			Amount:        float64(rand.Intn(10000)) + 0.99,
			Status:        statuses[rand.Intn(len(statuses))],
			PaymentMethod: paymentMethods[rand.Intn(len(paymentMethods))],
			CreatedAt:     now.Add(-time.Duration(rand.Intn(30*24)) * time.Hour),
		})
	}

	for _, f := range features {
		for _, user := range users {
			ds.featureUsages = append(ds.featureUsages, FeatureUsage{
				FeatureID:   f.ID,
				FeatureName: f.Name,
				UserID:      user.ID,
				UserName:    user.Name,
				Phone:       user.Phone,
				UsageCount:  rand.Intn(100),
				LastUsedAt:  now.Add(-time.Duration(rand.Intn(7*24)) * time.Hour),
			})
		}
	}
}

func getFieldValue(obj interface{}, fieldName string) (interface{}, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Struct {
		return nil, false
	}

	fieldName = strings.ToLower(fieldName)

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == fieldName || strings.EqualFold(fieldType.Name, fieldName) {
			return v.Field(i).Interface(), true
		}
		if idx := strings.Index(jsonTag, ","); idx > 0 {
			if jsonTag[:idx] == fieldName {
				return v.Field(i).Interface(), true
			}
		}
	}
	return nil, false
}

func compareValues(fieldValue interface{}, operator string, filterValue interface{}) bool {
	if fieldValue == nil {
		return false
	}

	operator = strings.ToLower(strings.TrimSpace(operator))

	fv := reflect.ValueOf(fieldValue)
	filterV := reflect.ValueOf(filterValue)

	switch fv.Kind() {
	case reflect.String:
		fStr := fv.String()
		var filterStr string
		if filterV.Kind() == reflect.String {
			filterStr = filterV.String()
		} else {
			filterStr = toStr(filterValue)
		}

		switch operator {
		case "=", "==", "eq", "equals":
			return strings.EqualFold(fStr, filterStr)
		case "!=", "<>", "ne", "notequals":
			return !strings.EqualFold(fStr, filterStr)
		case "contains", "like":
			return strings.Contains(strings.ToLower(fStr), strings.ToLower(filterStr))
		case "startswith":
			return strings.HasPrefix(strings.ToLower(fStr), strings.ToLower(filterStr))
		case "endswith":
			return strings.HasSuffix(strings.ToLower(fStr), strings.ToLower(filterStr))
		default:
			return strings.EqualFold(fStr, filterStr)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fInt := fv.Int()
		filterInt := toInt(filterValue)

		switch operator {
		case "=", "==", "eq", "equals":
			return fInt == filterInt
		case "!=", "<>", "ne", "notequals":
			return fInt != filterInt
		case ">", "gt":
			return fInt > filterInt
		case ">=", "ge":
			return fInt >= filterInt
		case "<", "lt":
			return fInt < filterInt
		case "<=", "le":
			return fInt <= filterInt
		default:
			return fInt == filterInt
		}

	case reflect.Float32, reflect.Float64:
		fFloat := fv.Float()
		filterFloat := toFloat(filterValue)

		switch operator {
		case "=", "==", "eq", "equals":
			return fFloat == filterFloat
		case "!=", "<>", "ne", "notequals":
			return fFloat != filterFloat
		case ">", "gt":
			return fFloat > filterFloat
		case ">=", "ge":
			return fFloat >= filterFloat
		case "<", "lt":
			return fFloat < filterFloat
		case "<=", "le":
			return fFloat <= filterFloat
		default:
			return fFloat == filterFloat
		}

	case reflect.Struct:
		if fv.Type() == reflect.TypeOf(time.Time{}) {
			fTime := fv.Interface().(time.Time)
			filterTime := toTime(filterValue)
			if filterTime.IsZero() {
				return false
			}

			switch operator {
			case "=", "==", "eq", "equals":
				return fTime.Equal(filterTime)
			case "!=", "<>", "ne", "notequals":
				return !fTime.Equal(filterTime)
			case ">", "gt", "after":
				return fTime.After(filterTime)
			case ">=", "ge":
				return fTime.After(filterTime) || fTime.Equal(filterTime)
			case "<", "lt", "before":
				return fTime.Before(filterTime)
			case "<=", "le":
				return fTime.Before(filterTime) || fTime.Equal(filterTime)
			default:
				return fTime.Equal(filterTime)
			}
		}
	}

	return false
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(reflect.ValueOf(v).String()))
}

func toInt(v interface{}) int64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		var result int64
		fmt.Sscanf(val, "%d", &result)
		return result
	default:
		return 0
	}
}

func toFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		var result float64
		fmt.Sscanf(val, "%f", &result)
		return result
	default:
		return 0
	}
}

func toTime(v interface{}) time.Time {
	if v == nil {
		return time.Time{}
	}
	switch val := v.(type) {
	case time.Time:
		return val
	case string:
		layouts := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
			"2006/01/02 15:04:05",
			"2006/01/02",
		}
		for _, layout := range layouts {
			if t, err := time.ParseInLocation(layout, val, time.Local); err == nil {
				return t
			}
		}
		return time.Time{}
	default:
		return time.Time{}
	}
}

func matchFilter(obj interface{}, field string, operator string, value interface{}) bool {
	fieldValue, found := getFieldValue(obj, field)
	if !found {
		return true
	}
	return compareValues(fieldValue, operator, value)
}

func (ds *DataSource) QueryUserBehaviors(filters map[string]interface{}) ([]map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []map[string]interface{}
	for _, b := range ds.userBehaviors {
		if ds.matchAllFilters(b, filters) {
			result = append(result, map[string]interface{}{
				"user_id":     b.UserID,
				"user_name":   b.UserName,
				"phone":       b.Phone,
				"idcard":      b.IDCard,
				"action":      b.Action,
				"page":        b.Page,
				"timestamp":   b.Timestamp,
				"duration_ms": b.DurationMs,
			})
		}
	}
	return result, nil
}

func (ds *DataSource) QueryTransactions(filters map[string]interface{}) ([]map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []map[string]interface{}
	for _, t := range ds.transactions {
		if ds.matchAllFilters(t, filters) {
			result = append(result, map[string]interface{}{
				"transaction_id": t.TransactionID,
				"user_id":        t.UserID,
				"user_name":      t.UserName,
				"phone":          t.Phone,
				"amount":         t.Amount,
				"status":         t.Status,
				"payment_method": t.PaymentMethod,
				"created_at":     t.CreatedAt,
			})
		}
	}
	return result, nil
}

func (ds *DataSource) QueryFeatureUsages(filters map[string]interface{}) ([]map[string]interface{}, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []map[string]interface{}
	for _, u := range ds.featureUsages {
		if ds.matchAllFilters(u, filters) {
			result = append(result, map[string]interface{}{
				"feature_id":   u.FeatureID,
				"feature_name": u.FeatureName,
				"user_id":      u.UserID,
				"user_name":    u.UserName,
				"phone":        u.Phone,
				"usage_count":  u.UsageCount,
				"last_used_at": u.LastUsedAt,
			})
		}
	}
	return result, nil
}

func (ds *DataSource) matchAllFilters(obj interface{}, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}

	for field, filterVal := range filters {
		switch fv := filterVal.(type) {
		case map[string]interface{}:
			for operator, value := range fv {
				if !matchFilter(obj, field, operator, value) {
					return false
				}
			}
		default:
			if !matchFilter(obj, field, "=", filterVal) {
				return false
			}
		}
	}

	return true
}

func (ds *DataSource) Query(dataSource string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	switch dataSource {
	case "user_behavior":
		return ds.QueryUserBehaviors(filters)
	case "transactions":
		return ds.QueryTransactions(filters)
	case "feature_usage":
		return ds.QueryFeatureUsages(filters)
	default:
		return ds.QueryUserBehaviors(filters)
	}
}
