package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"pos-system/pkg/common"
	"pos-system/pkg/core"
)

var store = core.NewStore()

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Server port")
	flag.Parse()

	if envPort := os.Getenv("POS_SERVER_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	http.HandleFunc("/api/products", handleProducts)
	http.HandleFunc("/api/products/barcode", handleProductByBarcode)
	http.HandleFunc("/api/member-levels", handleMemberLevels)
	http.HandleFunc("/api/members", handleMembers)
	http.HandleFunc("/api/orders", handleOrders)
	http.HandleFunc("/api/orders/", handleOrderByID)
	http.HandleFunc("/api/daily-closing", handleDailyClosing)
	http.HandleFunc("/api/stocktaking", handleStocktaking)
	http.HandleFunc("/api/stocktaking/", handleStocktakingByID)
	http.HandleFunc("/api/stocktaking/submit", handleStocktakingSubmit)
	http.HandleFunc("/api/stocktaking/complete", handleStocktakingComplete)
	http.HandleFunc("/api/stock-in", handleStockIn)

	fmt.Printf("POS Server starting on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, common.Response{
		Code:    code,
		Message: message,
	})
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, common.Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func handleProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req common.CreateProductReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		id, err := store.CreateProduct(req.Name, req.Barcode, req.Price, req.StockQty)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeSuccess(w, common.CreateProductResp{ID: id})
		return
	}

	if r.Method == http.MethodGet {
		products := store.GetAllProducts()
		resp := make([]common.GetProductResp, 0, len(products))
		for _, p := range products {
			resp = append(resp, common.GetProductResp{
				ID:       p.ID,
				Name:     p.Name,
				Barcode:  p.Barcode,
				Price:    p.Price,
				StockQty: p.StockQty,
			})
		}
		writeSuccess(w, resp)
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func handleProductByBarcode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	barcode := r.URL.Query().Get("barcode")
	if barcode == "" {
		writeError(w, http.StatusBadRequest, "Barcode is required")
		return
	}

	product, err := store.GetProductByBarcode(barcode)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeSuccess(w, common.GetProductResp{
		ID:       product.ID,
		Name:     product.Name,
		Barcode:  product.Barcode,
		Price:    product.Price,
		StockQty: product.StockQty,
	})
}

func handleMemberLevels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.CreateMemberLevelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	id, err := store.CreateMemberLevel(req.Name, req.DiscountRate)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, common.CreateMemberLevelResp{ID: id})
}

func handleMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.CreateMemberReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	id, err := store.CreateMember(req.Name, req.Phone, req.MemberLevelID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, common.CreateMemberResp{ID: id})
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req common.CreateOrderReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		items := make([]core.OrderInputItem, 0, len(req.Items))
		for _, item := range req.Items {
			items = append(items, core.OrderInputItem{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}

		order, err := store.CreateOrder(req.CashierID, req.MemberID, items, req.PaymentMethod)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeSuccess(w, common.CreateOrderResp{
			OrderID:        order.ID,
			TotalAmount:    order.TotalAmount,
			DiscountAmount: order.DiscountAmount,
			PayableAmount:  order.PayableAmount,
			StockStatus:    string(order.StockStatus),
		})
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func handleOrderByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Path[len("/api/orders/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	order, err := store.GetOrder(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	items := make([]common.OrderItemResp, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, common.OrderItemResp{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Price:       item.Price,
			Quantity:    item.Quantity,
			SubTotal:    item.SubTotal,
		})
	}

	writeSuccess(w, common.GetOrderResp{
		OrderID:        order.ID,
		CashierID:      order.CashierID,
		MemberID:       order.MemberID,
		TotalAmount:    order.TotalAmount,
		DiscountAmount: order.DiscountAmount,
		PayableAmount:  order.PayableAmount,
		PaymentMethod:  order.PaymentMethod,
		PaymentTime:    order.PaymentTime.Format("2006-01-02 15:04:05"),
		StockStatus:    string(order.StockStatus),
		Items:          items,
	})
}

func handleDailyClosing(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req common.DailyClosingReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		closing, err := store.CreateDailyClosing(req.Date)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		stats := make([]common.PaymentStatResp, 0, len(closing.PaymentStats))
		for _, stat := range closing.PaymentStats {
			stats = append(stats, common.PaymentStatResp{
				PaymentMethod: stat.PaymentMethod,
				TotalAmount:   stat.TotalAmount,
				OrderCount:    stat.OrderCount,
			})
		}

		writeSuccess(w, common.DailyClosingResp{
			ClosingID:    closing.ID,
			Date:         closing.Date,
			TotalAmount:  closing.TotalAmount,
			PaymentStats: stats,
			OrderCount:   closing.OrderCount,
			ClosedAt:     closing.ClosedAt.Format("2006-01-02 15:04:05"),
		})
		return
	}

	if r.Method == http.MethodGet {
		date := r.URL.Query().Get("date")
		if date == "" {
			writeError(w, http.StatusBadRequest, "Date is required")
			return
		}

		closing, err := store.GetDailyClosing(date)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}

		stats := make([]common.PaymentStatResp, 0, len(closing.PaymentStats))
		for _, stat := range closing.PaymentStats {
			stats = append(stats, common.PaymentStatResp{
				PaymentMethod: stat.PaymentMethod,
				TotalAmount:   stat.TotalAmount,
				OrderCount:    stat.OrderCount,
			})
		}

		writeSuccess(w, common.DailyClosingResp{
			ClosingID:    closing.ID,
			Date:         closing.Date,
			TotalAmount:  closing.TotalAmount,
			PaymentStats: stats,
			OrderCount:   closing.OrderCount,
			ClosedAt:     closing.ClosedAt.Format("2006-01-02 15:04:05"),
		})
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func handleStocktaking(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req common.CreateStocktakingReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid request")
			return
		}

		stocktaking, err := store.CreateStocktaking(req.OperatorID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeSuccess(w, common.CreateStocktakingResp{
			StocktakingID: stocktaking.ID,
			Status:        string(stocktaking.Status),
		})
		return
	}

	if r.Method == http.MethodGet {
		stocktakings := store.GetAllStocktakings()
		resp := make([]common.StocktakingSummaryResp, 0, len(stocktakings))
		for _, st := range stocktakings {
			summary := common.StocktakingSummaryResp{
				ID:         st.ID,
				Status:     string(st.Status),
				OperatorID: st.OperatorID,
				CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04:05"),
			}
			if !st.CompletedAt.IsZero() {
				summary.CompletedAt = st.CompletedAt.Format("2006-01-02 15:04:05")
			}
			resp = append(resp, summary)
		}
		writeSuccess(w, common.GetStocktakingListResp{Stocktakings: resp})
		return
	}

	writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func handleStocktakingByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Path[len("/api/stocktaking/"):]
	if id == "" {
		writeError(w, http.StatusBadRequest, "Stocktaking ID is required")
		return
	}

	st, err := store.GetStocktaking(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	items := make([]common.StocktakingItemResp, 0, len(st.Items))
	for _, item := range st.Items {
		items = append(items, common.StocktakingItemResp{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Barcode:     item.Barcode,
			Price:       item.Price,
			SnapshotQty: item.SnapshotQty,
			ActualQty:   item.ActualQty,
			DiffQty:     item.DiffQty,
			DiffAmount:  item.DiffAmount,
		})
	}

	pendingOps := store.GetPendingOperations(id)
	pendingOpsResp := make([]common.StocktakingPendingOpResp, 0, len(pendingOps))
	for _, op := range pendingOps {
		pendingOpsResp = append(pendingOpsResp, common.StocktakingPendingOpResp{
			OperationType: string(op.OperationType),
			ProductID:     op.ProductID,
			Quantity:      op.Quantity,
			OperationTime: op.OperationTime.Format("2006-01-02 15:04:05"),
		})
	}

	resp := common.GetStocktakingDetailResp{
		ID:         st.ID,
		Status:     string(st.Status),
		OperatorID: st.OperatorID,
		CreatedAt:  st.CreatedAt.Format("2006-01-02 15:04:05"),
		Items:      items,
		PendingOps: pendingOpsResp,
	}
	if !st.CompletedAt.IsZero() {
		resp.CompletedAt = st.CompletedAt.Format("2006-01-02 15:04:05")
	}

	writeSuccess(w, resp)
}

func handleStocktakingSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.SubmitStocktakingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	actualQtys := make(map[string]int)
	for _, item := range req.Items {
		actualQtys[item.ProductID] = item.ActualQty
	}

	if err := store.SubmitStocktaking(req.StocktakingID, actualQtys); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, nil)
}

func handleStocktakingComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.CompleteStocktakingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := store.CompleteStocktaking(req.StocktakingID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, common.CompleteStocktakingResp{
		StocktakingID: req.StocktakingID,
		Status:        string(core.StocktakingStatusCompleted),
	})
}

func handleStockIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.StockInReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	id, err := store.StockIn(req.ProductID, req.Quantity, req.OperatorID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, common.StockInResp{ID: id})
}
