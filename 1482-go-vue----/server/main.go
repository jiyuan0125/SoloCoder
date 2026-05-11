package main

import (
	"autorepair/common"
	"autorepair/core"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	service *core.Service
}

func NewServer(service *core.Service) *Server {
	return &Server{service: service}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, resp common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeSuccess(w http.ResponseWriter, data interface{}) {
	s.writeJSON(w, http.StatusOK, common.Response{Success: true, Data: data})
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, common.Response{Success: false, Message: message})
}

func (s *Server) parseBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (s *Server) getIDFromPath(path string) (int64, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid path")
	}
	idStr := parts[len(parts)-1]
	return strconv.ParseInt(idStr, 10, 64)
}

func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		orders := s.service.GetOrders()
		s.writeSuccess(w, orders)
	case http.MethodPost:
		var req common.CreateOrderRequest
		if err := s.parseBody(r, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, "请求参数错误")
			return
		}
		order, err := s.service.CreateOrder(req)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.writeSuccess(w, order)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func (s *Server) handleOrder(w http.ResponseWriter, r *http.Request) {
	orderID, err := s.getIDFromPath(r.URL.Path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的工单ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		order, err := s.service.GetOrder(orderID)
		if err != nil {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeSuccess(w, order)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func (s *Server) handleAssignOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.AssignOrderRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	order, err := s.service.AssignOrder(req.OrderID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, order)
}

func (s *Server) handleRepairItems(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.CreateRepairItemRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	item, err := s.service.CreateRepairItem(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, item)
}

func (s *Server) handleCompleteItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.CompleteRepairItemRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	item, err := s.service.CompleteRepairItem(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, item)
}

func (s *Server) handleAddPartToItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.AddPartToItemRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	usedPart, err := s.service.AddPartToItem(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, usedPart)
}

func (s *Server) handleReturnPart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.ReturnPartRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	err := s.service.ReturnPart(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, nil)
}

func (s *Server) handleCompleteOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.CompleteOrderRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	order, err := s.service.CompleteOrder(req.OrderID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, order)
}

func (s *Server) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.CancelOrderRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	order, err := s.service.CancelOrder(req.OrderID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, order)
}

func (s *Server) handleParts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		parts := s.service.GetParts()
		s.writeSuccess(w, parts)
	case http.MethodPost:
		var req common.CreatePartRequest
		if err := s.parseBody(r, &req); err != nil {
			s.writeError(w, http.StatusBadRequest, "请求参数错误")
			return
		}
		part, err := s.service.CreatePart(req)
		if err != nil {
			s.writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.writeSuccess(w, part)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func (s *Server) handlePart(w http.ResponseWriter, r *http.Request) {
	partID, err := s.getIDFromPath(r.URL.Path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "无效的配件ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		part, err := s.service.GetPart(partID)
		if err != nil {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeSuccess(w, part)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func (s *Server) handleUpdatePartPrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.UpdatePartPriceRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	part, err := s.service.UpdatePartPrice(req.PartID, req.UnitPrice)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, part)
}

func (s *Server) handleUpdatePartStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.UpdatePartStockRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	part, err := s.service.UpdatePartStock(req.PartID, req.StockQty)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, part)
}

func (s *Server) handleReplenishments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		todos := s.service.GetReplenishments()
		s.writeSuccess(w, todos)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func (s *Server) handleResolveReplenishment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}

	var req common.ResolveReplenishmentRequest
	if err := s.parseBody(r, &req); err != nil {
		s.writeError(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	todo, err := s.service.ResolveReplenishment(req.TodoID)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeSuccess(w, todo)
}

func main() {
	var portFlag int
	flag.IntVar(&portFlag, "port", 0, "服务端监听端口")
	flag.Parse()

	port := portFlag
	if port == 0 {
		if envPort := os.Getenv("PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil {
				port = p
			}
		}
	}
	if port == 0 {
		port = 9002
	}

	service := core.NewService()
	server := NewServer(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/orders", server.handleOrders)
	mux.HandleFunc("/api/orders/", server.handleOrder)
	mux.HandleFunc("/api/orders/assign", server.handleAssignOrder)
	mux.HandleFunc("/api/orders/complete", server.handleCompleteOrder)
	mux.HandleFunc("/api/orders/cancel", server.handleCancelOrder)

	mux.HandleFunc("/api/repair-items", server.handleRepairItems)
	mux.HandleFunc("/api/repair-items/complete", server.handleCompleteItem)

	mux.HandleFunc("/api/used-parts", server.handleAddPartToItem)
	mux.HandleFunc("/api/used-parts/return", server.handleReturnPart)

	mux.HandleFunc("/api/parts", server.handleParts)
	mux.HandleFunc("/api/parts/", server.handlePart)
	mux.HandleFunc("/api/parts/price", server.handleUpdatePartPrice)
	mux.HandleFunc("/api/parts/stock", server.handleUpdatePartStock)

	mux.HandleFunc("/api/replenishments", server.handleReplenishments)
	mux.HandleFunc("/api/replenishments/resolve", server.handleResolveReplenishment)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("汽车维修管理系统服务端启动中，监听端口: %d\n", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "服务启动失败: %v\n", err)
		os.Exit(1)
	}
}
